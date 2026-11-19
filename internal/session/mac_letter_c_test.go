package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func TestMacANSIKeyCIsNotBackspace(t *testing.T) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "你", Syllables: []string{"ni"}, Weight: 100},
		{Text: "好", Syllables: []string{"hao"}, Weight: 90},
		{Text: "你好", Syllables: []string{"ni", "hao"}, Weight: 1000},
	})
	eng := engine.NewPinyinEngine(dict)
	eng.SetScheme(pinyin.SchemeXiaohe)
	m := NewManager(eng)
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ni"})
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "h"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: 8, Characters: "c"})
	if res.Composition == nil || res.Composition.Preedit != "nihc" {
		t.Fatalf("letter c with CG keycode 8 became backspace: %#v", res.Composition)
	}
	if res.Candidates == nil || len(res.Candidates.Candidates) == 0 || bareCandidateText(res.Candidates.Candidates[0].Text) != "你好" {
		t.Fatalf("flypy nihc candidates = %#v", res.Candidates)
	}
}

func TestMacCGDeleteStillWorksWithEmptyCharacters(t *testing.T) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "你", Syllables: []string{"ni"}, Weight: 100},
	})
	eng := engine.NewPinyinEngine(dict)
	eng.SetScheme(pinyin.SchemeXiaohe)
	m := NewManager(eng)
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ni"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyDelete})
	if res.Composition == nil || res.Composition.Preedit != "n" {
		t.Fatalf("mac CG delete empty chars = %#v", res.Composition)
	}
}
