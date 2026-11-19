package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

type manyEngine struct{}

func (manyEngine) Decode(input string) ([]engine.Candidate, error) {
	out := make([]engine.Candidate, 0, 20)
	for i := 0; i < 20; i++ {
		out = append(out, engine.Candidate{Text: string(rune('A' + i)), Score: int32(100 - i)})
	}
	return out, nil
}

func TestCandidatePagingMinusEquals(t *testing.T) {
	m := NewManager(manyEngine{})
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ni"})
	if res.Candidates == nil || len(res.Candidates.Candidates) != 9 || res.Candidates.Page != 0 {
		t.Fatalf("page0 = %#v", res.Candidates)
	}
	if bareCandidateText(res.Candidates.Candidates[0].Text) != "A" {
		t.Fatalf("first = %s", res.Candidates.Candidates[0].Text)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "=", KeyCode: 0})
	if res.Candidates == nil || res.Candidates.Page != 1 || len(res.Candidates.Candidates) != 9 {
		t.Fatalf("page1 = %#v", res.Candidates)
	}
	if bareCandidateText(res.Candidates.Candidates[0].Text) != "J" {
		t.Fatalf("page1 first = %s", res.Candidates.Candidates[0].Text)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "-", KeyCode: 0})
	if res.Candidates == nil || res.Candidates.Page != 0 {
		t.Fatalf("back to page0 = %#v", res.Candidates)
	}
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "="})
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "2"})
	if res.Commit == nil || res.Commit.Text != "K" {
		t.Fatalf("digit on page1 = %#v", res.Commit)
	}
}

func TestXiaoheMultiSyllableThroughSession(t *testing.T) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "你", Syllables: []string{"ni"}, Weight: 100},
		{Text: "好", Syllables: []string{"hao"}, Weight: 90},
		{Text: "你好", Syllables: []string{"ni", "hao"}, Weight: 1000},
	})
	eng := engine.NewPinyinEngine(dict)
	eng.SetScheme(pinyin.SchemeXiaohe)
	m := NewManager(eng)
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nihc"})
	if res.Composition == nil || res.Composition.Preedit != "nihc" {
		t.Fatalf("preedit = %#v", res.Composition)
	}
	if res.Candidates == nil || len(res.Candidates.Candidates) == 0 || bareCandidateText(res.Candidates.Candidates[0].Text) != "你好" {
		t.Fatalf("candidates = %#v", res.Candidates)
	}
}
