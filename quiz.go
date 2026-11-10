package main

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"fuckenglish/internal/wordbook"
)

const englishToChinese = "english-chinese"
const chineseToEnglish = "chinese-english"

type QuizState struct {
	Direction string   `json:"direction"`
	Token     string   `json:"token"`
	Question  string   `json:"question"`
	Phonetic  string   `json:"phonetic"`
	Options   []string `json:"options"`
	Index     int      `json:"index"`
	Total     int      `json:"total"`
	Skipped   int      `json:"skipped"`
	Answered  bool     `json:"answered"`
	Choice    int      `json:"choice"`
	Correct   int      `json:"correct"`
	Score     int      `json:"score"`
	Completed bool     `json:"completed"`
}

type quizQuestion struct {
	prompt   string
	phonetic string
	options  []string
	correct  int
}

type quizSession struct {
	direction                     string
	id                            uint64
	questions                     []quizQuestion
	index, skipped, choice, score int
	answered                      bool
}

// Only the authored distractors from this row are candidates. Missing or
// duplicate choices disable this direction without blocking wordbook browsing.
func quizChoices(w wordbook.Word, direction string) (string, []string, bool) {
	prompt, answer, distractors := w.English, w.Meaning, w.ChineseDistractors
	if direction == chineseToEnglish {
		prompt, answer, distractors = w.Meaning, w.English, w.EnglishDistractors
	}
	if strings.TrimSpace(prompt) == "" || strings.TrimSpace(answer) == "" || len(distractors) != 3 {
		return "", nil, false
	}
	options := append([]string{strings.TrimSpace(answer)}, distractors...)
	seen := map[string]bool{}
	for i, option := range options {
		option = strings.TrimSpace(option)
		key := strings.ToLower(strings.Join(strings.Fields(option), " "))
		if key == "" || seen[key] {
			return "", nil, false
		}
		seen[key] = true
		options[i] = option
	}
	return strings.TrimSpace(prompt), options, true
}

func (c *Core) startQuizLocked(direction string) error {
	if direction != englishToChinese && direction != chineseToEnglish {
		return fmt.Errorf("请选择英选中或中选英")
	}
	q := quizSession{direction: direction, id: rand.Uint64(), choice: -1}
	for _, w := range c.words[c.cfg.Selected] {
		prompt, choices, ok := quizChoices(w, direction)
		if !ok {
			q.skipped++
			continue
		}
		question := quizQuestion{prompt: prompt, options: make([]string, 4)}
		if direction == englishToChinese {
			question.phonetic = strings.TrimSpace(w.Phonetic)
		}
		for pos, source := range rand.Perm(4) {
			question.options[pos] = choices[source]
			if source == 0 {
				question.correct = pos
			}
		}
		q.questions = append(q.questions, question)
	}
	rand.Shuffle(len(q.questions), func(i, j int) { q.questions[i], q.questions[j] = q.questions[j], q.questions[i] })
	c.quiz = q
	return nil
}

func (c *Core) quizSnapshotLocked() QuizState {
	q := &c.quiz
	s := QuizState{Direction: q.direction, Index: q.index, Total: len(q.questions), Skipped: q.skipped, Choice: -1, Correct: -1, Score: q.score, Options: []string{}}
	if q.direction == "" {
		return s
	}
	s.Completed = len(q.questions) > 0 && q.index >= len(q.questions)
	if len(q.questions) == 0 || s.Completed {
		return s
	}
	question := q.questions[q.index]
	s.Token = fmt.Sprintf("%x-%d", q.id, q.index)
	s.Question = question.prompt
	s.Phonetic = question.phonetic
	s.Options = append([]string{}, question.options...)
	s.Answered = q.answered
	if q.answered {
		s.Choice = q.choice
		s.Correct = question.correct
	}
	return s
}

// The token prevents an old/double click from answering or advancing a new question.
func (c *Core) quizChange(run func() error) (LibraryState, error) {
	c.mu.Lock()
	err := run()
	s := c.snapshotLocked()
	c.mu.Unlock()
	if err == nil {
		_ = LibraryEvent.Broadcast(s)
	}
	return s, err
}

func (l *Library) StartQuiz(direction string) (LibraryState, error) {
	return l.core.quizChange(func() error { return l.core.startQuizLocked(direction) })
}

func (l *Library) EndQuiz() (LibraryState, error) {
	return l.core.quizChange(func() error { l.core.quiz = quizSession{}; return nil })
}

func (l *Library) AnswerQuiz(token string, choice int) (LibraryState, error) {
	return l.core.quizChange(func() error {
		s := l.core.quizSnapshotLocked()
		if token == "" || s.Token != token || choice < 0 || choice >= len(s.Options) {
			return fmt.Errorf("题目已更新，请重新选择")
		}
		q := &l.core.quiz
		if q.answered {
			return nil
		}
		q.choice, q.answered = choice, true
		if choice == q.questions[q.index].correct {
			q.score++
		}
		return nil
	})
}

func (l *Library) NextQuiz(token string) (LibraryState, error) {
	return l.core.quizChange(func() error {
		s := l.core.quizSnapshotLocked()
		if token == "" || s.Token != token || !s.Answered {
			return fmt.Errorf("请先回答当前题目")
		}
		q := &l.core.quiz
		q.index++
		q.choice, q.answered = -1, false
		return nil
	})
}
