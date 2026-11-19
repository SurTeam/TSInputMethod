package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func testEngine() *engine.PinyinEngine {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "你", Syllables: []string{"ni"}, Weight: 100},
		{Text: "好", Syllables: []string{"hao"}, Weight: 90},
		{Text: "你好", Syllables: []string{"ni", "hao"}, Weight: 1000},
		{Text: "泥", Syllables: []string{"ni"}, Weight: 10},
	})
	eng := engine.NewPinyinEngine(dict)
	eng.SetScheme(pinyin.SchemeXiaohe)
	return eng
}

func TestSpaceCommitsTopCandidate(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nihc"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keySpace, Characters: " "})
	if res.Commit == nil || res.Commit.Text != "你好" {
		t.Fatalf("space commit = %#v", res.Commit)
	}
	if res.Composition == nil || res.Composition.Preedit != "" {
		t.Fatalf("preedit after commit = %#v", res.Composition)
	}
}

func TestReturnCommitsRawPreeditAndClearsCandidates(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nihc"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyReturn, Characters: "\r"})
	if res.Commit == nil || res.Commit.Text != "nihc" {
		t.Fatalf("return commit = %#v", res.Commit)
	}
	if res.Composition == nil || res.Composition.Preedit != "" {
		t.Fatalf("preedit after return = %#v", res.Composition)
	}
	if res.Candidates == nil || len(res.Candidates.Candidates) != 0 {
		t.Fatalf("candidates after return = %#v", res.Candidates)
	}
}

func TestKeypadEnterCommitsRawPreedit(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ni"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keypadEnter, Characters: "\n"})
	if res.Commit == nil || res.Commit.Text != "ni" {
		t.Fatalf("keypad enter commit = %#v", res.Commit)
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
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nihc"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyEscape})
	if res.Composition == nil || res.Composition.Preedit != "" {
		t.Fatalf("esc preedit = %#v", res.Composition)
	}
}

func TestBackspaceEditsPreedit(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nihc"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyDelete})
	if res.Composition == nil || res.Composition.Preedit != "nih" {
		t.Fatalf("backspace preedit = %#v", res.Composition)
	}
}
