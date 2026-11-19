package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/engine"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func TestMacCGDeleteIsBackspace(t *testing.T) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "你", Syllables: []string{"ni"}, Weight: 100},
		{Text: "泥", Syllables: []string{"ni"}, Weight: 10},
		{Text: "啊", Syllables: []string{"a"}, Weight: 50},
	})
	eng := engine.NewPinyinEngine(dict)
	m := NewManager(eng)

	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ni"})
	res1 := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyDelete})
	if !res1.Consumed || res1.Composition == nil || res1.Composition.Preedit != "n" {
		t.Fatalf("mac CG delete #1 = %#v", res1)
	}
	res2 := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyDelete})
	if !res2.Consumed || res2.Composition == nil || res2.Composition.Preedit != "" {
		t.Fatalf("mac CG delete #2 = %#v", res2)
	}
	if res2.Candidates == nil || len(res2.Candidates.Candidates) != 0 {
		t.Fatalf("after full clear candidates = %#v", res2.Candidates)
	}
	resNew := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "a"})
	if resNew.Composition == nil || resNew.Composition.Preedit != "a" {
		t.Fatalf("fresh input after clear = %#v", resNew.Composition)
	}
	if resNew.Candidates == nil || len(resNew.Candidates.Candidates) == 0 {
		t.Fatalf("fresh input should produce candidates = %#v", resNew.Candidates)
	}
}

func TestLegacyASCIIEditCodesAreIgnored(t *testing.T) {
	m := NewManager(engine.NewPinyinEngine(dictionary.FromEntries([]dictionary.Entry{
		{Text: "你", Syllables: []string{"ni"}, Weight: 100},
	})))
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ni"})

	if res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: 8}); res.Consumed {
		t.Fatalf("legacy ASCII backspace 8 must not edit, got %#v", res)
	}
	if res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: 27}); res.Consumed {
		t.Fatalf("legacy ASCII escape 27 must not cancel, got %#v", res)
	}
	if res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: 32}); res.Consumed {
		t.Fatalf("legacy ASCII space 32 must not commit, got %#v", res)
	}
	if res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "a"}); res.Composition == nil || res.Composition.Preedit != "nia" {
		t.Fatalf("preedit after ignored legacy codes = %#v", res.Composition)
	}
}
