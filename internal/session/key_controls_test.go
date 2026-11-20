package session

import (
	"testing"

	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func TestArrowKeysMoveSelectionAndPage(t *testing.T) {
	m := NewManager(manyEngine{})
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ni"})
	if res.Candidates == nil || res.Candidates.SelectedIndex != 0 {
		t.Fatalf("initial selected = %#v", res.Candidates)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyRightArrow})
	if res.Candidates == nil || res.Candidates.SelectedIndex != 1 || res.Candidates.Page != 0 {
		t.Fatalf("right arrow = %#v", res.Candidates)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyLeftArrow})
	if res.Candidates == nil || res.Candidates.SelectedIndex != 0 {
		t.Fatalf("left arrow = %#v", res.Candidates)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyDownArrow})
	if res.Candidates == nil || res.Candidates.Page != 1 || res.Candidates.SelectedIndex != 0 {
		t.Fatalf("down arrow page = %#v", res.Candidates)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyUpArrow})
	if res.Candidates == nil || res.Candidates.Page != 0 {
		t.Fatalf("up arrow page = %#v", res.Candidates)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyRightArrow})
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keySpace, Characters: " "})
	if res.Commit == nil || res.Commit.Text != "B" {
		t.Fatalf("space commits highlighted = %#v", res.Commit)
	}
}

func TestChineseDigitsAndPunctuationWithoutPreedit(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "3"})
	if res.Commit == nil || res.Commit.Text != "3" || !res.Consumed {
		t.Fatalf("digit commit = %#v", res)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: ","})
	if res.Commit == nil || res.Commit.Text != "，" {
		t.Fatalf("comma commit = %#v", res)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "."})
	if res.Commit == nil || res.Commit.Text != "。" {
		t.Fatalf("period commit = %#v", res)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "/"})
	if res.Commit == nil || res.Commit.Text != "/" {
		t.Fatalf("slash stays half-shape = %#v", res)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "$"})
	if res.Commit == nil || res.Commit.Text != "¥" {
		t.Fatalf("dollar maps to yen = %#v", res)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "\""})
	if res.Commit == nil || res.Commit.Text != "“" {
		t.Fatalf("opening quote = %#v", res)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "\""})
	if res.Commit == nil || res.Commit.Text != "”" {
		t.Fatalf("closing quote = %#v", res)
	}
}

func TestPunctuationIgnoredDuringPreedit(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ni"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: ","})
	if res.Commit != nil {
		t.Fatalf("unexpected commit = %#v", res.Commit)
	}
	if !res.Consumed || res.Composition == nil || res.Composition.Preedit != "ni" {
		t.Fatalf("punctuation during preedit = %#v", res)
	}
}

func TestShiftLetterCreatesLiteralCandidate(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "A", Modifiers: modShift})
	if res.Composition == nil || res.Composition.Preedit != "A" {
		t.Fatalf("shift letter preedit = %#v", res.Composition)
	}
	if res.Candidates == nil || len(res.Candidates.Candidates) != 1 || bareCandidateText(res.Candidates.Candidates[0].Text) != "A" {
		t.Fatalf("literal candidates = %#v", res.Candidates)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: " ", KeyCode: keySpace})
	if res.Commit == nil || res.Commit.Text != "A" {
		t.Fatalf("literal commit = %#v", res.Commit)
	}
}

func TestLiteralMixedCaseKeepsLeadingCapital(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "H", Modifiers: modShift})
	if res.Composition == nil || res.Composition.Preedit != "H" {
		t.Fatalf("leading capital preedit = %#v", res.Composition)
	}
	for _, ch := range []string{"e", "l", "l", "o"} {
		res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: ch})
	}
	if res.Composition == nil || res.Composition.Preedit != "Hello" {
		t.Fatalf("mixed case preedit = %#v", res.Composition)
	}
	if res.Candidates == nil || len(res.Candidates.Candidates) != 1 || bareCandidateText(res.Candidates.Candidates[0].Text) != "Hello" {
		t.Fatalf("mixed case candidates = %#v", res.Candidates)
	}
}

func TestLowercaseThenShiftLetterEntersLiteral(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "h"})
	if res.Composition == nil || res.Composition.Preedit != "h" {
		t.Fatalf("lowercase start preedit = %#v", res.Composition)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "E", Modifiers: modShift})
	if res.Composition == nil || res.Composition.Preedit != "hE" {
		t.Fatalf("shift after lowercase preedit = %#v", res.Composition)
	}
	if res.Candidates == nil || len(res.Candidates.Candidates) != 1 || bareCandidateText(res.Candidates.Candidates[0].Text) != "hE" {
		t.Fatalf("shift after lowercase candidates = %#v", res.Candidates)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "l"})
	if res.Composition == nil || res.Composition.Preedit != "hEl" {
		t.Fatalf("continue mixed preedit = %#v", res.Composition)
	}
	if res.Candidates == nil || len(res.Candidates.Candidates) != 1 || bareCandidateText(res.Candidates.Candidates[0].Text) != "hEl" {
		t.Fatalf("continue mixed candidates = %#v", res.Candidates)
	}
}

func TestAllCapsLiteralSingleCandidate(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	var res *KeyResult
	for _, ch := range []string{"O", "K"} {
		res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: ch, Modifiers: modShift})
	}
	if res.Composition == nil || res.Composition.Preedit != "OK" {
		t.Fatalf("all caps preedit = %#v", res.Composition)
	}
	if res.Candidates == nil || len(res.Candidates.Candidates) != 1 || bareCandidateText(res.Candidates.Candidates[0].Text) != "OK" {
		t.Fatalf("all caps candidates = %#v", res.Candidates)
	}
}

func TestAllLowercaseStaysPinyinCompose(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ni"})
	if res.Composition == nil || res.Composition.Preedit != "ni" {
		t.Fatalf("lowercase pinyin preedit = %#v", res.Composition)
	}
	if res.Candidates == nil || len(res.Candidates.Candidates) == 0 {
		t.Fatalf("lowercase pinyin should decode candidates = %#v", res.Candidates)
	}
	if bareCandidateText(res.Candidates.Candidates[0].Text) == "ni" && len(res.Candidates.Candidates) == 1 {
		t.Fatalf("lowercase pinyin should not collapse to single literal candidate = %#v", res.Candidates)
	}
}

func TestShiftSymbolUsesHalfShapeMap(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "!", Modifiers: modShift})
	if res.Commit == nil || res.Commit.Text != "！" {
		t.Fatalf("shift symbol = %#v", res)
	}
	res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "_", Modifiers: modShift})
	if res.Commit == nil || res.Commit.Text != "——" {
		t.Fatalf("shift underscore = %#v", res)
	}
}

func TestControlOptionCommandPassthrough(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ni"})
	for _, mod := range []uint32{modControl, modOption, modCommand} {
		res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "c", Modifiers: mod})
		if res.Consumed {
			t.Fatalf("modifier %#x should passthrough, got %#v", mod, res)
		}
	}
}

func TestEnglishModePassthrough(t *testing.T) {
	m := NewManager(testEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_ENGLISH)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "a"})
	if res.Consumed || res.Commit != nil {
		t.Fatalf("english passthrough = %#v", res)
	}
}
