package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func TestSoftConfirmBackspaceRestoresKeys(t *testing.T) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "世界", Syllables: []string{"shi", "jie"}, Weight: 50},
		{Text: "世", Syllables: []string{"shi"}, Weight: 1000},
		{Text: "界", Syllables: []string{"jie"}, Weight: 900},
	})
	eng := engine.NewPinyinEngine(dict)
	eng.SetScheme(pinyin.SchemeXiaohe)
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
	first := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "uijp"})
	idx := -1
	for i, text := range candidateTexts(first) {
		if bareCandidateText(text) == "世" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("missing 世: %#v", candidateTexts(first))
	}
	soft := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: string(rune('1' + idx))})
	if soft.Composition == nil || soft.Composition.Preedit != "世jp" {
		t.Fatalf("soft preedit=%#v", soft.Composition)
	}
	back := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyDelete})
	if back.Composition == nil || back.Composition.Preedit != "uijp" {
		t.Fatalf("backspace restore=%#v", back.Composition)
	}
	if back.Composition.Cursor != uint32(len([]rune("uijp"))) {
		t.Fatalf("cursor after restore=%d want end", back.Composition.Cursor)
	}
	if topBareText(back) != "世界" && !containsText(candidateTexts(back), "世界") {
		t.Fatalf("restored candidates=%#v", candidateTexts(back))
	}
	trim := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyDelete})
	if trim.Composition == nil || trim.Composition.Preedit != "uij" {
		t.Fatalf("second backspace should trim end: %#v", trim.Composition)
	}
	if trim.Composition.Cursor != uint32(len([]rune("uij"))) {
		t.Fatalf("cursor after trim=%d", trim.Composition.Cursor)
	}
}

func TestSoftConfirmThenFinishHardCommitsPhrase(t *testing.T) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "你好", Syllables: []string{"ni", "hao"}, Weight: 1000},
		{Text: "你", Syllables: []string{"ni"}, Weight: 100},
		{Text: "好", Syllables: []string{"hao"}, Weight: 90},
		{Text: "世界", Syllables: []string{"shi", "jie"}, Weight: 800},
		{Text: "世", Syllables: []string{"shi"}, Weight: 70},
		{Text: "界", Syllables: []string{"jie"}, Weight: 60},
	})
	eng := engine.NewPinyinEngine(dict)
	eng.SetScheme(pinyin.SchemeXiaohe)
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nihcuijp"})
	soft := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: " ", KeyCode: keySpace})
	if soft.Composition == nil || soft.Composition.Preedit != "你好uijp" {
		t.Fatalf("soft=%#v", soft.Composition)
	}
	fin := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: " ", KeyCode: keySpace})
	if fin.Commit == nil || fin.Commit.Text != "你好世界" {
		t.Fatalf("final commit=%#v", fin.Commit)
	}
	if fin.Composition == nil || fin.Composition.Preedit != "" {
		t.Fatalf("final preedit=%#v", fin.Composition)
	}
}
