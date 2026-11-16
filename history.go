package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/egoist/mygo"
)

const quizHistoryDirectory = "测试记录"

// One record represents one finished attempt, including words skipped because
// their authored choices are missing. Each correct word is always worth 2 points.
type QuizRecord struct {
	ID           string `json:"id"`
	Book         string `json:"book"`
	CompletedAt  string `json:"completedAt"`
	Date         string `json:"date"`
	Correct      int    `json:"correct"`
	Wrong        int    `json:"wrong"`
	Skipped      int    `json:"skipped"`
	Total        int    `json:"total"`
	Score        int    `json:"score"`
	Attempt      int    `json:"attempt"`
	DailyAttempt int    `json:"dailyAttempt"`
}

type QuizHistoryState struct {
	Folder  string       `json:"folder"`
	Records []QuizRecord `json:"records"`
}

type quizHistoryFile struct {
	Version int          `json:"version"`
	Records []QuizRecord `json:"records"`
}

var HistoryChangedEvent = mygo.NewEvent[string]("history-changed")

func (l *Library) QuizHistory() (QuizHistoryState, error) {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()
	return readQuizHistory(l.core.cfg.Folder)
}

func readQuizHistory(wordFolder string) (QuizHistoryState, error) {
	s := QuizHistoryState{Folder: filepath.Join(wordFolder, quizHistoryDirectory), Records: []QuizRecord{}}
	path := filepath.Join(s.Folder, "records.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, fmt.Errorf("无法读取测试记录 %s：%w", path, err)
	}
	var file quizHistoryFile
	if err := json.Unmarshal(data, &file); err != nil {
		return s, fmt.Errorf("测试记录文件损坏，请先备份并检查 %s：%w", path, err)
	}
	if file.Version != 1 || file.Records == nil {
		return s, fmt.Errorf("测试记录格式不受支持，请先备份并检查 %s", path)
	}
	seen := map[string]bool{}
	for _, r := range file.Records {
		completed, err := time.Parse(time.RFC3339Nano, r.CompletedAt)
		if err != nil || r.ID == "" || seen[r.ID] || strings.TrimSpace(r.Book) == "" || r.Date != completed.Format("2006-01-02") || r.Correct < 0 || r.Wrong < 0 || r.Skipped < 0 || r.Total <= 0 || r.Correct+r.Wrong != r.Total || r.Score != r.Correct*2 || r.Attempt <= 0 || r.DailyAttempt <= 0 {
			return s, fmt.Errorf("测试记录内容不完整，请先备份并检查 %s", path)
		}
		seen[r.ID] = true
	}
	s.Records = file.Records
	return s, nil
}

// Called under the core lock. Read before writing so an unreadable or corrupt
// existing file can never be silently replaced with an empty history.
func saveQuizRecord(q *quizSession, completed time.Time) (bool, error) {
	if q.historySaved {
		return false, nil
	}
	history, err := readQuizHistory(q.folder)
	if err != nil {
		return false, err
	}
	for _, r := range history.Records {
		if r.ID == q.recordID {
			q.historySaved = true
			return false, nil
		}
	}
	date := completed.Format("2006-01-02")
	attempt, dailyAttempt := 1, 1
	for _, r := range history.Records {
		if r.Book == q.book {
			attempt = max(attempt, r.Attempt+1)
			if r.Date == date {
				dailyAttempt = max(dailyAttempt, r.DailyAttempt+1)
			}
		}
	}
	r := QuizRecord{
		ID: q.recordID, Book: q.book, CompletedAt: completed.Format(time.RFC3339Nano), Date: date,
		Correct: q.score, Wrong: len(q.questions) - q.score, Skipped: q.skipped,
		Total: len(q.questions), Score: q.score * 2, Attempt: attempt, DailyAttempt: dailyAttempt,
	}
	history.Records = append(history.Records, r)
	data, err := json.MarshalIndent(quizHistoryFile{Version: 1, Records: history.Records}, "", "  ")
	if err != nil {
		return false, err
	}
	if err := os.MkdirAll(history.Folder, 0755); err != nil {
		return false, fmt.Errorf("无法创建测试记录文件夹 %s：%w", history.Folder, err)
	}
	if err := atomicWriteQuizHistory(history.Folder, data); err != nil {
		return false, fmt.Errorf("测试成绩尚未保存，请检查文件夹权限后重新选择本题答案：%w", err)
	}
	q.historySaved = true
	return true, nil
}

func atomicWriteQuizHistory(folder string, data []byte) error {
	f, err := os.CreateTemp(folder, "records-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(name, filepath.Join(folder, "records.json"))
}

func newQuizRecordID() string { return rand.Text() }
