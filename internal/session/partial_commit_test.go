package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func TestXiaoheUijpTopCandidateIsShijie(t *testing.T) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "世界", Syllables: []string{"shi", "jie"}, Weight: 1000},
		{Text: "世", Syllables: []string{"shi"}, Weight: 100},
		{Text: "界", Syllables: []string{"jie"}, Weight: 90},
		{Text: "机", Syllables: []string{"ji"}, Weight: 80},
		{Text: "恶", Syllables: []string{"e"}, Weight: 70},
	})
	eng := engine.NewPinyinEngine(dict)
	eng.SetScheme(pinyin.SchemeXiaohe)
	m := NewManager(eng)
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "uijp"})
	if res.Candidates == nil || len(res.Candidates.Candidates) == 0 {
		t.Fatalf("no candidates: %#v", res)
	}
	if bareCandidateText(res.Candidates.Candidates[0].Text) != "世界" {
		t.Fatalf("top=%#v", res.Candidates.Candidates[0])
	}
	if res.Candidates.Candidates[0].Comment != "shi'jie" {
		t.Fatalf("comment=%q", res.Candidates.Candidates[0].Comment)
	}
	for _, cand := range res.Candidates.Candidates {
		if cand.Comment == "shi'ji'e" || bareCandidateText(cand.Text) == "世机恶" {
			t.Fatalf("quanpin re-split leaked: %#v", res.Candidates.Candidates)
		}
	}
}

func TestPartialWordCommitLeavesRemainder(t *testing.T) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "世界", Syllables: []string{"shi", "jie"}, Weight: 50},
		{Text: "世", Syllables: []string{"shi"}, Weight: 1000},
		{Text: "界", Syllables: []string{"jie"}, Weight: 900},
	})
	eng := engine.NewPinyinEngine(dict)
	eng.SetScheme(pinyin.SchemeXiaohe)
	m := NewManager(eng)
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	first := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "uijp"})
	if first.Candidates == nil {
		t.Fatalf("no candidates")
	}
	idx := -1
	for i, cand := range first.Candidates.Candidates {
		if bareCandidateText(cand.Text) == "世" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("missing partial candidate 世 in %#v", first.Candidates.Candidates)
	}
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: string(rune('1' + idx))})
	if res.Commit == nil || res.Commit.Text != "世" {
		t.Fatalf("partial commit = %#v", res.Commit)
	}
	if res.Composition == nil || res.Composition.Preedit != "jp" {
		t.Fatalf("remainder preedit = %#v", res.Composition)
	}
	if res.Candidates == nil || len(res.Candidates.Candidates) == 0 || bareCandidateText(res.Candidates.Candidates[0].Text) != "界" {
		t.Fatalf("remainder candidates = %#v", res.Candidates)
	}
}
