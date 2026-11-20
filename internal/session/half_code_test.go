package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func TestSingleKeyHPrefersShortPrefixHits(t *testing.T) {
	eng := engine.NewPinyinEngine(loadIceDict(t))
	eng.SetScheme(pinyin.SchemeXiaohe)
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "h"})
	texts := candidateTexts(res)
	if len(texts) == 0 {
		t.Fatal("no candidates")
	}
	if len([]rune(texts[0])) > 2 {
		t.Fatalf("single-key top should be short, got %q in %#v", texts[0], texts)
	}
	for _, text := range texts {
		if len([]rune(text)) > 4 {
			t.Fatalf("long phrase crowded first page for h: %#v", texts)
		}
	}
}

func TestSingleKeyMappedInitialsSurfaceCandidates(t *testing.T) {
	eng := engine.NewPinyinEngine(loadIceDict(t))
	eng.SetScheme(pinyin.SchemeXiaohe)
	for _, key := range []string{"i", "u", "v"} {
		s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
		res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: key})
		texts := candidateTexts(res)
		if len(texts) == 0 {
			t.Fatalf("%s produced no candidates", key)
		}
		if res.Composition == nil || res.Composition.Preedit != key {
			t.Fatalf("%s preedit = %#v", key, res.Composition)
		}
	}
}

func TestOddNihSurfacesNihao(t *testing.T) {
	eng := engine.NewPinyinEngine(loadIceDict(t))
	eng.SetScheme(pinyin.SchemeXiaohe)
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nih"})
	texts := candidateTexts(res)
	if !containsText(texts, "你好") {
		t.Fatalf("nih missing 你好: %#v", texts)
	}
	if len(texts) == 0 || len([]rune(texts[0])) > 2 {
		t.Fatalf("unexpected top for nih: %#v", texts)
	}
	for _, text := range texts {
		if len([]rune(text)) > 4 {
			t.Fatalf("long phrase crowded first page for nih: %#v", texts)
		}
	}
}
