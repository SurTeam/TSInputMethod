package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func TestAnPrefersAnFamilyNotA(t *testing.T) {
	eng := engine.NewPinyinEngine(loadIceDict(t))
	eng.SetScheme(pinyin.SchemeXiaohe)
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := typeKeys(s, "an")
	texts := candidateTexts(res)
	if len(texts) == 0 {
		t.Fatal("no candidates for an")
	}
	top := bareCandidateText(texts[0])
	if top == "啊" || top == "阿" {
		t.Fatalf("an top should be an-family, got %#v", texts)
	}
	if !containsText(texts, "安") && !containsText(texts, "按") {
		t.Fatalf("an missing 安/按: %#v", texts)
	}
	space := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keySpace, Characters: " "})
	if space.Commit == nil || space.Commit.Text == "" || space.Commit.Text == "啊" {
		t.Fatalf("an Space commit=%#v candidates were %#v", space.Commit, texts)
	}
	if space.Composition != nil && space.Composition.Preedit != "" {
		t.Fatalf("an Space should hard-commit, leftover %#v", space.Composition)
	}
}

func TestAnqSpaceCommitsAnquan(t *testing.T) {
	eng := engine.NewPinyinEngine(loadIceDict(t))
	eng.SetScheme(pinyin.SchemeXiaohe)
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := typeKeys(s, "anq")
	texts := candidateTexts(res)
	if !containsText(texts, "安全") {
		t.Fatalf("anq missing 安全: %#v", texts)
	}
	if bareCandidateText(texts[0]) != "安全" {
		t.Fatalf("anq top want 安全 got %#v", texts)
	}
	space := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keySpace, Characters: " "})
	if space.Commit == nil || space.Commit.Text != "安全" {
		t.Fatalf("anq Space commit=%#v leftover=%#v", space.Commit, space.Composition)
	}
}

func TestAnqrTopsAnquan(t *testing.T) {
	eng := engine.NewPinyinEngine(loadIceDict(t))
	eng.SetScheme(pinyin.SchemeXiaohe)
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := typeKeys(s, "anqr")
	texts := candidateTexts(res)
	if len(texts) == 0 || bareCandidateText(texts[0]) != "安全" {
		t.Fatalf("anqr top want 安全 got %#v", texts)
	}
	space := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keySpace, Characters: " "})
	if space.Commit == nil || space.Commit.Text != "安全" {
		t.Fatalf("anqr Space commit=%#v", space.Commit)
	}
}

func typeKeys(s *Session, keys string) *KeyResult {
	var res *KeyResult
	for _, ch := range keys {
		res = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: string(ch)})
	}
	return res
}
