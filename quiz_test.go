package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"fuckenglish/internal/wordbook"
)

func quizCore(t *testing.T) *Core {
	c := testCore(t)
	c.words[c.cfg.Selected] = []wordbook.Word{
		{English: "aircraft", Meaning: "n. 航空器；飞行器", ChineseDistractors: []string{"n. 机场", "n. 飞行员", "n. 跑道"}, EnglishDistractors: []string{"airport", "pilot", "runway"}},
		{English: "aisle", Meaning: "n. 过道", ChineseDistractors: []string{"n. 岛屿", "n. 出口", "n. 座位"}, EnglishDistractors: []string{"island", "exit", "seat"}},
		{English: "old", Meaning: "旧的"},
		{English: "alert", Meaning: "警觉的", ChineseDistractors: []string{"安静的", "迟钝的", "疲倦的"}, EnglishDistractors: []string{"ALERT", "tired", "quiet"}},
	}
	return c
}

func TestQuizDirectionsAndEligibility(t *testing.T) {
	for _, direction := range []string{englishToChinese, chineseToEnglish} {
		c := quizCore(t)
		l := &Library{c}
		s, err := l.StartQuiz(direction)
		if err != nil {
			t.Fatal(err)
		}
		want := 3
		if direction == chineseToEnglish {
			want = 2
		}
		if s.Quiz.Total != want || s.Quiz.Skipped != 4-want {
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
				if direction == englishToChinese && w.English == q.Question || direction == chineseToEnglish && w.Meaning == q.Question {
					word = w
				}
			}
			_, choices, ok := quizChoices(word, direction)
			if !ok {
				t.Fatalf("unknown prompt: %+v", q)
			}
			correct := -1
			for index, option := range q.Options {
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

func TestInvalidDistractorsRemainBrowsable(t *testing.T) {
	for _, choices := range [][]string{nil, {"a", "b"}, {"a", "b", " "}, {"a", "b", " a "}, {"x", "b", "c"}, {"a  b", "a b", "c"}} {
		w := wordbook.Word{English: "word", Meaning: "x", ChineseDistractors: choices}
		if _, _, ok := quizChoices(w, englishToChinese); ok {
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
