package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"fuckenglish/internal/wordbook"
)

func quizCore(t *testing.T) *Core {
	c := testCore(t)
	c.words[c.cfg.Selected] = []wordbook.Word{
		{English: "aircraft", Phonetic: "[ˈeəkrɑːft]", Meaning: "n. 航空器；飞行器", ChineseDistractors: []string{"n. 机场", "n. 飞行员", "n. 跑道"}},
		{English: "aisle", Meaning: "n. 过道", ChineseDistractors: []string{"n. 岛屿", "n. 出口", "n. 座位"}},
		{English: "old", Meaning: "旧的"},
		{English: "alert", Meaning: "警觉的", ChineseDistractors: []string{"安静的", "迟钝的", "疲倦的"}},
	}
	return c
}

func TestQuizEnglishToChineseAndEligibility(t *testing.T) {
	c := quizCore(t)
	l := &Library{c}
	s, err := l.StartQuiz(englishToChinese)
	if err != nil {
		t.Fatal(err)
	}
	const want = 3
	if s.Quiz.Total != want || s.Quiz.Skipped != 1 {
		t.Fatalf("eligibility: %+v", s.Quiz)
	}
	seen := map[string]bool{}
	for i := 0; i < want; i++ {
		q := s.Quiz
		if len(q.Options) != 4 || q.Answered || q.Correct != -1 || seen[q.Question] {
			t.Fatalf("question: %+v", q)
		}
		seen[q.Question] = true
		var word wordbook.Word
		for _, w := range c.words[c.cfg.Selected] {
			if w.English == q.Question {
				word = w
			}
		}
		_, choices, ok := quizChoices(word)
		if q.Phonetic != word.Phonetic {
			t.Fatalf("phonetic does not match shuffled question: %+v", q)
		}
		if !ok {
			t.Fatalf("unknown prompt: %+v", q)
		}
		correct := -1
		uniqueOptions := map[string]bool{}
		for index, option := range q.Options {
			if uniqueOptions[option] {
				t.Fatalf("duplicate option: %q", option)
			}
			uniqueOptions[option] = true
			found := false
			for _, choice := range choices {
				if choice == option {
					found = true
				}
			}
			if !found {
				t.Fatalf("unprovided distractor: %q", option)
			}
			if option == choices[0] {
				correct = index
			}
		}
		if correct < 0 {
			t.Fatal("correct answer missing")
		}
		if _, err := l.NextQuiz(q.Token); err == nil {
			t.Fatal("advanced unanswered question")
		}
		s, err = l.AnswerQuiz(q.Token, correct)
		if err != nil || !s.Quiz.Answered || s.Quiz.Correct != correct || s.Quiz.Score != i+1 {
			t.Fatalf("answer: %+v %v", s.Quiz, err)
		}
		duplicate, err := l.AnswerQuiz(q.Token, (correct+1)%4)
		if err != nil || duplicate.Quiz.Choice != correct || duplicate.Quiz.Score != i+1 {
			t.Fatal("duplicate answer changed score")
		}
		s, err = l.NextQuiz(q.Token)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := l.NextQuiz(q.Token); err == nil {
			t.Fatal("stale next token accepted")
		}
		if _, err := l.AnswerQuiz(q.Token, correct); err == nil {
			t.Fatal("stale answer accepted")
		}
	}
	if !s.Quiz.Completed || s.Quiz.Score != want || len(s.Quiz.Options) != 0 {
		t.Fatalf("result: %+v", s.Quiz)
	}
}

func TestQuizRemovedDirectionsPreserveCurrentSession(t *testing.T) {
	c := quizCore(t)
	l := &Library{c}
	s, err := l.StartQuiz(englishToChinese)
	if err != nil {
		t.Fatal(err)
	}
	wrong := (c.quiz.questions[0].correct + 1) % 4
	s, err = l.AnswerQuiz(s.Quiz.Token, wrong)
	if err != nil {
		t.Fatal(err)
	}
	for _, direction := range []string{"chinese-english", "unknown", ""} {
		got, err := l.StartQuiz(direction)
		if err == nil {
			t.Fatalf("unsupported direction accepted: %q", direction)
		}
		if !reflect.DeepEqual(got.Quiz, s.Quiz) || !reflect.DeepEqual(l.Snapshot().Quiz, s.Quiz) {
			t.Fatalf("rejected direction changed current quiz: %q", direction)
		}
	}
}

func TestQuizWrongAnswersEmptyAndReset(t *testing.T) {
	c := quizCore(t)
	l := &Library{c}
	s, _ := l.StartQuiz(englishToChinese)
	if _, err := l.AnswerQuiz(s.Quiz.Token, -1); err == nil {
		t.Fatal("invalid choice accepted")
	}
	wrong := (c.quiz.questions[0].correct + 1) % 4
	s, err := l.AnswerQuiz(s.Quiz.Token, wrong)
	if err != nil || s.Quiz.Score != 0 || !s.Quiz.Answered || s.Quiz.Choice != wrong {
		t.Fatalf("wrong answer: %+v %v", s.Quiz, err)
	}
	if _, err := l.Page(0); err != nil || !l.Snapshot().Quiz.Answered {
		t.Fatal("navigation erased quiz")
	}
	if _, err := l.StartQuiz("unknown"); err == nil {
		t.Fatal("invalid direction accepted")
	}
	if _, err := l.Select("list2.xlsx"); err != nil || l.Snapshot().Quiz.Direction != englishToChinese || l.Snapshot().Quiz.Score != 0 {
		t.Fatal("book switch did not start a fresh quiz in the same direction")
	}
	s, _ = l.StartQuiz(englishToChinese)
	if s.Quiz.Total != 0 || s.Quiz.Skipped != 50 || s.Quiz.Completed {
		t.Fatalf("old book: %+v", s.Quiz)
	}
	s, _ = l.EndQuiz()
	if s.Quiz.Direction != "" {
		t.Fatal("did not return to browsing")
	}
}

func TestQuizAutoAdvanceCorrectOnlyAndFinishManually(t *testing.T) {
	c := quizCore(t)
	l := &Library{c}
	s, _ := l.StartQuiz(englishToChinese)
	if s.Quiz.Total != 3 {
		t.Fatalf("need three questions: %+v", s.Quiz)
	}
	correct := c.quiz.questions[0].correct
	answered, err := l.AnswerQuiz(s.Quiz.Token, correct)
	if err != nil || !answered.Quiz.Answered || answered.Quiz.Index != 0 {
		t.Fatalf("correct feedback missing: %+v %v", answered.Quiz, err)
	}
	// Repeated delivery must neither add points nor schedule another skip.
	_, _ = l.AnswerQuiz(s.Quiz.Token, correct)
	deadline := time.Now().Add(3 * time.Second)
	for l.Snapshot().Quiz.Index == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	s = l.Snapshot()
	if s.Quiz.Index != 1 || s.Quiz.Answered || s.Quiz.Score != 1 {
		t.Fatalf("correct answer did not advance once: %+v", s.Quiz)
	}
	c.mu.Lock()
	wrong := (c.quiz.questions[1].correct + 1) % 4
	c.mu.Unlock()
	s, _ = l.AnswerQuiz(s.Quiz.Token, wrong)
	time.Sleep(quizAutoAdvanceDelay + 100*time.Millisecond)
	if got := l.Snapshot().Quiz; got.Index != 1 || !got.Answered || got.Score != 1 {
		t.Fatalf("wrong answer advanced automatically: %+v", got)
	}
	s, _ = l.NextQuiz(s.Quiz.Token)
	c.mu.Lock()
	correct = c.quiz.questions[2].correct
	c.mu.Unlock()
	s, _ = l.AnswerQuiz(s.Quiz.Token, correct)
	time.Sleep(quizAutoAdvanceDelay + 100*time.Millisecond)
	if got := l.Snapshot().Quiz; got.Index != 2 || got.Completed || !got.Answered {
		t.Fatalf("last question completed without confirmation: %+v", got)
	}
	s, err = l.NextQuiz(s.Quiz.Token)
	if err != nil || !s.Quiz.Completed || s.Quiz.Score != 2 {
		t.Fatalf("manual finish failed: %+v %v", s.Quiz, err)
	}
}

func TestQuizPendingAutoAdvanceCannotChangeNewSession(t *testing.T) {
	for _, reset := range []string{"restart", "book", "refresh", "browse"} {
		t.Run(reset, func(t *testing.T) {
			c := quizCore(t)
			l := &Library{c}
			s, _ := l.StartQuiz(englishToChinese)
			_, _ = l.AnswerQuiz(s.Quiz.Token, c.quiz.questions[0].correct)
			switch reset {
			case "restart":
				_, _ = l.StartQuiz(englishToChinese)
			case "book":
				_, _ = l.Select("list2.xlsx")
			case "refresh":
				_, _ = l.Refresh()
			case "browse":
				_, _ = l.EndQuiz()
			}
			before := l.Snapshot().Quiz
			c.autoAdvanceQuiz(s.Quiz.Token)
			if got := l.Snapshot().Quiz; !reflect.DeepEqual(got, before) {
				t.Fatalf("pending advance changed %s: %+v", reset, got)
			}
		})
	}
}

func TestQuizFinalWrongAndSingleQuestionWaitForFinish(t *testing.T) {
	for _, correctAnswer := range []bool{false, true} {
		c := quizCore(t)
		c.words[c.cfg.Selected] = c.words[c.cfg.Selected][:1]
		l := &Library{c}
		s, _ := l.StartQuiz(englishToChinese)
		choice := c.quiz.questions[0].correct
		if !correctAnswer {
			choice = (choice + 1) % 4
		}
		s, _ = l.AnswerQuiz(s.Quiz.Token, choice)
		c.autoAdvanceQuiz(s.Quiz.Token)
		if got := l.Snapshot().Quiz; got.Completed || got.Index != 0 || !got.Answered {
			t.Fatalf("single/final question completed automatically: %+v", got)
		}
		finished, err := l.NextQuiz(s.Quiz.Token)
		if err != nil || !finished.Quiz.Completed {
			t.Fatalf("single/final question could not finish: %+v %v", finished.Quiz, err)
		}
	}
}

func TestQuizManualNextInvalidatesPendingAutoAdvance(t *testing.T) {
	c := quizCore(t)
	l := &Library{c}
	s, _ := l.StartQuiz(englishToChinese)
	_, _ = l.AnswerQuiz(s.Quiz.Token, c.quiz.questions[0].correct)
	next, err := l.NextQuiz(s.Quiz.Token)
	if err != nil {
		t.Fatal(err)
	}
	c.autoAdvanceQuiz(s.Quiz.Token)
	if got := l.Snapshot().Quiz; !reflect.DeepEqual(got, next.Quiz) {
		t.Fatalf("pending timer skipped the new question: %+v", got)
	}
}

func TestInvalidDistractorsRemainBrowsable(t *testing.T) {
	for _, choices := range [][]string{nil, {"a", "b"}, {"a", "b", " "}, {"a", "b", " a "}, {"x", "b", "c"}, {"a  b", "a b", "c"}} {
		w := wordbook.Word{English: "word", Meaning: "x", ChineseDistractors: choices}
		if _, _, ok := quizChoices(w); ok {
			t.Fatalf("invalid accepted: %v", choices)
		}
	}
}

func TestQuizShufflesWithoutChangingAuthoredData(t *testing.T) {
	c := quizCore(t)
	before, _ := json.Marshal(c.words)
	orders, positions := map[string]bool{}, map[int]bool{}
	for i := 0; i < 40; i++ {
		if err := c.startQuizLocked(englishToChinese); err != nil {
			t.Fatal(err)
		}
		order := ""
		for _, q := range c.quiz.questions {
			order += q.prompt + ","
			if q.prompt == "aircraft" {
				positions[q.correct] = true
			}
		}
		orders[order] = true
	}
	if len(orders) < 2 || len(positions) < 2 {
		t.Fatal("questions or choices were not shuffled")
	}
	after, _ := json.Marshal(c.words)
	if string(before) != string(after) {
		t.Fatal("quiz mutated wordbook")
	}
}

func TestBrowseRandomOrderAndResume(t *testing.T) {
	c := testCore(t)
	l := &Library{c}
	_, _ = l.Page(2)
	random, err := l.SetBrowseOrder("random")
	if err != nil || random.Page != 0 || random.BrowseOrder != "random" {
		t.Fatalf("random: %+v %v", random, err)
	}
	first, _ := json.Marshal(random.Words)
	seen := map[string]bool{}
	for i := 0; i < random.Pages; i++ {
		s, err := l.Page(i)
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range s.Words {
			if seen[w.English] {
				t.Fatalf("repeated word: %s", w.English)
			}
			seen[w.English] = true
		}
	}
	if len(seen) != 50 {
		t.Fatalf("missing words: %d", len(seen))
	}
	_, _ = l.Page(0)
	again, _ := json.Marshal(l.Snapshot().Words)
	if string(first) != string(again) {
		t.Fatal("page change reshuffled")
	}
	_, _ = l.Page(3)
	sequential, _ := l.SetBrowseOrder("sequential")
	if sequential.Page != 2 || sequential.Words[0].English != c.words[c.cfg.Selected][10].English {
		t.Fatal("sequential progress changed")
	}
	_, _ = l.SetBrowseOrder("random")
	if l.Snapshot().Page != 3 {
		t.Fatal("random page not restored")
	}
	_, _ = l.Select("list2.xlsx")
	_, _ = l.Select("list1.xlsx")
	if l.Snapshot().Page != 3 {
		t.Fatal("switching books lost progress")
	}
	var cfg config
	data, _ := os.ReadFile(c.path)
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	restarted := &Core{cfg: cfg}
	restarted.scanLocked()
	if !reflect.DeepEqual(restarted.snapshot().Words, l.Snapshot().Words) {
		t.Fatal("restart changed shuffle")
	}
	if _, err := l.SetBrowseOrder("unknown"); err == nil {
		t.Fatal("invalid order accepted")
	}
	c.path += "/missing/settings.json"
	if _, err := l.SetBrowseOrder("sequential"); err == nil {
		t.Fatal("expected write failure")
	}
	if l.Snapshot().BrowseOrder != "random" {
		t.Fatal("failed save changed order")
	}
}
