package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/engine"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func TestSessionIsolation(t *testing.T) {
	m := NewManager(engine.NewStubEngine())
	a := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	b := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	if a.ID() == b.ID() {
		t.Fatal("session ids must differ")
	}
	resA := a.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ni"})
	if resA.Composition == nil || resA.Composition.Preedit != "ni" {
		t.Fatalf("session a preedit = %#v", resA.Composition)
	}
	resB := b.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "hao"})
	if resB.Composition == nil || resB.Composition.Preedit != "hao" {
		t.Fatalf("session b preedit = %#v", resB.Composition)
	}
	if resA.Composition.Preedit == resB.Composition.Preedit {
		t.Fatal("sessions leaked composition state")
	}
}

func TestModeSwitchAndCancel(t *testing.T) {
	m := NewManager(engine.NewStubEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "abc"})
	s.SetMode(imev1.InputMode_INPUT_MODE_ENGLISH)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "x"})
	if res.Commit == nil || res.Commit.Text != "x" {
		t.Fatalf("english commit = %#v", res.Commit)
	}
	s.SetMode(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "wo"})
	cancel := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyEscape})
	if cancel.Composition == nil || cancel.Composition.Preedit != "" {
		t.Fatalf("cancel left preedit %#v", cancel.Composition)
	}
}

func TestBackspace(t *testing.T) {
	m := NewManager(engine.NewStubEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ab"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyDelete})
	if res.Composition == nil || res.Composition.Preedit != "a" {
		t.Fatalf("backspace preedit = %#v", res.Composition)
	}
}
