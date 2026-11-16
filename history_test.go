package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"fuckenglish/internal/wordbook"
)

func finishQuizForHistory(t *testing.T, l *Library, correctCount int) LibraryState {
	t.Helper()
	s, err := l.StartQuiz(englishToChinese)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < s.Quiz.Total; i++ {
		l.core.mu.Lock()
		choice := l.core.quiz.questions[i].correct
		l.core.mu.Unlock()
		if i >= correctCount {
			choice = (choice + 1) % 4
		}
		s, err = l.AnswerQuiz(s.Quiz.Token, choice)
		if err != nil {
			t.Fatal(err)
		}
		if i+1 < s.Quiz.Total {
			s, err = l.NextQuiz(s.Quiz.Token)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	return s
}

func TestQuizHistoryFixedTwoPointsSkippedAndAutomaticSave(t *testing.T) {
	c := quizCore(t)
	valid := c.words[c.cfg.Selected][0]
	words := make([]wordbook.Word, 50)
	for i := 0; i < 48; i++ {
		words[i] = valid
	}
	words[48] = wordbook.Word{English: "missing1", Meaning: "缺选项"}
	words[49] = wordbook.Word{English: "missing2", Meaning: "缺选项"}
	c.words[c.cfg.Selected] = words
	l := &Library{c}
	history, err := l.QuizHistory()
	if err != nil || history.Records == nil || len(history.Records) != 0 {
		t.Fatalf("empty history: %+v %v", history, err)
	}
	if _, err := os.Stat(history.Folder); !os.IsNotExist(err) {
		t.Fatalf("reading history created a folder: %v", err)
	}
	s := finishQuizForHistory(t, l, 40)
	if s.Quiz.Completed || !s.Quiz.Answered {
		t.Fatalf("last answer feedback should remain visible: %+v", s.Quiz)
	}
	history, err = l.QuizHistory()
	if err != nil || len(history.Records) != 1 {
		t.Fatalf("last answer did not save: %+v %v", history, err)
	}
	r := history.Records[0]
	if r.Correct != 40 || r.Wrong != 8 || r.Skipped != 2 || r.Total != 48 || r.Score != 80 || r.Attempt != 1 || r.DailyAttempt != 1 || r.Book != "list1.xlsx" || r.ID == "" {
		t.Fatalf("fixed scoring: %+v", r)
	}
	if _, err := time.Parse(time.RFC3339Nano, r.CompletedAt); err != nil {
		t.Fatalf("timestamp: %v", err)
	}
	_, _ = l.AnswerQuiz(s.Quiz.Token, 0)
	_, _ = l.NextQuiz(s.Quiz.Token)
	_, _ = l.NextQuiz(s.Quiz.Token)
	again, _ := l.QuizHistory()
	if !reflect.DeepEqual(again, history) {
		t.Fatalf("duplicate answer or finish duplicated result: %+v", again)
	}
	restarted := &Library{core: &Core{cfg: config{Folder: c.cfg.Folder}}}
	afterRestart, err := restarted.QuizHistory()
	if err != nil || !reflect.DeepEqual(afterRestart, history) {
		t.Fatalf("restart lost history: %+v %v", afterRestart, err)
	}
}

func TestQuizHistoryAttemptsSameDayAcrossDayAndBooks(t *testing.T) {
	folder := t.TempDir()
	zone := time.FixedZone("CST", 8*60*60)
	first := time.Date(2026, 11, 16, 9, 0, 0, 0, zone)
	for i, item := range []struct {
		book                   string
		at                     time.Time
		wantAttempt, wantDaily int
	}{
		{"list1.xlsx", first, 1, 1},
		{"list1.xlsx", first, 2, 2}, // Even identical timestamps remain independent.
		{"list2.xlsx", first, 1, 1},
		{"list1.xlsx", first.Add(24 * time.Hour), 3, 1},
	} {
		q := quizSession{recordID: newQuizRecordID(), folder: folder, book: item.book, score: 1, questions: []quizQuestion{{}}}
		changed, err := saveQuizRecord(&q, item.at)
		if err != nil || !changed || !q.historySaved {
			t.Fatalf("save %d: %v %v", i, changed, err)
		}
		// A re-delivery with the same stable ID is ignored after reading disk too.
		q.historySaved = false
		if changed, err := saveQuizRecord(&q, item.at); err != nil || changed {
			t.Fatalf("disk duplicate %d: %v %v", i, changed, err)
		}
		history, err := readQuizHistory(folder)
		if err != nil || len(history.Records) != i+1 {
			t.Fatalf("history %d: %+v %v", i, history, err)
		}
		r := history.Records[i]
		if r.Attempt != item.wantAttempt || r.DailyAttempt != item.wantDaily || r.Date != item.at.Format("2006-01-02") {
			t.Fatalf("numbering %d: %+v", i, r)
		}
	}
}

func TestQuizHistoryUnfinishedAndEmptyQuizzesNotRecorded(t *testing.T) {
	for _, reset := range []string{"restart", "browse", "book", "refresh"} {
		t.Run(reset, func(t *testing.T) {
			c := quizCore(t)
			l := &Library{c}
			s, _ := l.StartQuiz(englishToChinese)
			_, err := l.AnswerQuiz(s.Quiz.Token, (c.quiz.questions[0].correct+1)%4)
			if err != nil {
				t.Fatal(err)
			}
			switch reset {
			case "restart":
				_, _ = l.StartQuiz(englishToChinese)
			case "browse":
				_, _ = l.EndQuiz()
			case "book":
				_, _ = l.Select("list2.xlsx")
			case "refresh":
				_, _ = l.Refresh()
			}
			history, err := l.QuizHistory()
			if err != nil || len(history.Records) != 0 {
				t.Fatalf("unfinished quiz was recorded: %+v %v", history, err)
			}
		})
	}
	c := quizCore(t)
	c.words[c.cfg.Selected] = []wordbook.Word{{English: "unavailable", Meaning: "不可测试"}}
	l := &Library{c}
	s, _ := l.StartQuiz(englishToChinese)
	_, _ = l.AnswerQuiz(s.Quiz.Token, 0)
	_, _ = l.NextQuiz(s.Quiz.Token)
	history, err := l.QuizHistory()
	if err != nil || len(history.Records) != 0 {
		t.Fatalf("empty quiz recorded: %+v %v", history, err)
	}
}

func TestQuizHistoryCorruptFilePreservedAndLastAnswerRetry(t *testing.T) {
	c := quizCore(t)
	c.words[c.cfg.Selected] = c.words[c.cfg.Selected][:1]
	l := &Library{c}
	s, _ := l.StartQuiz(englishToChinese)
	choice := c.quiz.questions[0].correct
	folder := filepath.Join(c.cfg.Folder, quizHistoryDirectory)
	if err := os.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(folder, "records.json")
	for _, bad := range []string{"{broken", "{}", `{"version":99,"records":[]}`, `{"version":1,"records":[{"id":"broken"}]}`} {
		if err := os.WriteFile(path, []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		failed, err := l.AnswerQuiz(s.Quiz.Token, choice)
		if err == nil || failed.Quiz.Answered || failed.Quiz.Score != 0 {
			t.Fatalf("failed save accepted final answer: %+v %v", failed.Quiz, err)
		}
		data, _ := os.ReadFile(path)
		if string(data) != bad {
			t.Fatal("corrupt original overwritten")
		}
	}
	if err := os.WriteFile(path, []byte(`{"version":1,"records":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := l.AnswerQuiz(s.Quiz.Token, choice)
	if err != nil || !s.Quiz.Answered || s.Quiz.Score != 1 {
		t.Fatalf("retry failed: %+v %v", s.Quiz, err)
	}
	history, _ := l.QuizHistory()
	if len(history.Records) != 1 || history.Records[0].Score != 2 {
		t.Fatalf("retry result: %+v", history)
	}
}

func TestQuizHistoryUnavailableDestinationCanRetry(t *testing.T) {
	c := quizCore(t)
	c.words[c.cfg.Selected] = c.words[c.cfg.Selected][:1]
	l := &Library{c}
	s, _ := l.StartQuiz(englishToChinese)
	choice := c.quiz.questions[0].correct
	blocker := filepath.Join(c.cfg.Folder, quizHistoryDirectory)
	if err := os.WriteFile(blocker, []byte("blocked"), 0600); err != nil {
		t.Fatal(err)
	}
	failed, err := l.AnswerQuiz(s.Quiz.Token, choice)
	if err == nil || failed.Quiz.Answered || failed.Quiz.Score != 0 {
		t.Fatalf("blocked write accepted: %+v %v", failed.Quiz, err)
	}
	if err := os.Remove(blocker); err != nil {
		t.Fatal(err)
	}
	if _, err := l.AnswerQuiz(s.Quiz.Token, choice); err != nil {
		t.Fatalf("unblocked retry: %v", err)
	}
}

func TestQuizHistoryFolderSwitchUsesSeparateRecords(t *testing.T) {
	c := quizCore(t)
	l := &Library{c}
	finishQuizForHistory(t, l, 2)
	first, _ := l.QuizHistory()
	secondFolder := t.TempDir()
	if _, err := l.SetFolder(secondFolder); err != nil {
		t.Fatal(err)
	}
	second, err := l.QuizHistory()
	if err != nil || len(second.Records) != 0 || second.Folder == first.Folder {
		t.Fatalf("folder histories mixed: %+v %v", second, err)
	}
	if _, err := l.SetFolder(filepath.Dir(first.Folder)); err != nil {
		t.Fatal(err)
	}
	again, err := l.QuizHistory()
	if err != nil || !reflect.DeepEqual(first, again) {
		t.Fatalf("folder return lost history: %+v %v", again, err)
	}
}

func TestAtomicQuizHistoryWriteFailurePreservesTargetAndCleansTemp(t *testing.T) {
	folder := t.TempDir()
	target := filepath.Join(folder, "records.json")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(target, "preserved.txt")
	if err := os.WriteFile(marker, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := atomicWriteQuizHistory(folder, []byte("new")); err == nil {
		t.Fatal("expected destination replacement failure")
	}
	data, err := os.ReadFile(marker)
	if err != nil || string(data) != "original" {
		t.Fatalf("original target changed: %q %v", data, err)
	}
	entries, _ := os.ReadDir(folder)
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tmp") {
			t.Fatalf("temporary file leaked: %s", entry.Name())
		}
	}
}

func TestQuizHistoryJSONHasStablePublicFields(t *testing.T) {
	c := quizCore(t)
	l := &Library{c}
	finishQuizForHistory(t, l, 0)
	history, err := l.QuizHistory()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(history)
	for _, field := range []string{"folder", "records", "id", "book", "completedAt", "date", "correct", "wrong", "skipped", "total", "score", "attempt", "dailyAttempt"} {
		if !strings.Contains(string(data), `"`+field+`":`) {
			t.Fatalf("missing field %s: %s", field, data)
		}
	}
}
