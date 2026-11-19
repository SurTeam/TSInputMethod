package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/engine"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func testEngine() engine.Engine {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "你", Syllables: []string{"ni"}, Weight: 100},
		{Text: "好", Syllables: []string{"hao"}, Weight: 90},
		{Text: "你好", Syllables: []string{"ni", "hao"}, Weight: 1000},
		{Text: "泥", Syllables: []string{"ni"}, Weight: 10},
	})
	return engine.NewPinyinEngine(dict)
}

func TestSpaceCommitsTopCandidate(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nihao"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: 32, Characters: " "})
	if res.Commit == nil || res.Commit.Text != "你好" {
		t.Fatalf("space commit = %#v", res.Commit)
	}
	if res.Composition == nil || res.Composition.Preedit != "" {
		t.Fatalf("preedit after commit = %#v", res.Composition)
	}
}

func TestDigitSelectsCandidate(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ni"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "2"})
	if res.Commit == nil || res.Commit.Text == "" {
		t.Fatalf("digit commit = %#v", res)
	}
}

func TestEscClearsPreedit(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nihao"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: 27})
	if res.Composition == nil || res.Composition.Preedit != "" {
		t.Fatalf("esc preedit = %#v", res.Composition)
	}
}

func TestBackspaceEditsPreedit(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nihao"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: 8})
	if res.Composition == nil || res.Composition.Preedit != "niha" {
		t.Fatalf("backspace preedit = %#v", res.Composition)
	}
}
