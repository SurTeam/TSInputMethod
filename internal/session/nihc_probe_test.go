package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func loadIceDict(t *testing.T) *dictionary.Dictionary {
	t.Helper()
	path := filepath.Join("..", "..", "data", "dicts", "rime-ice.tsdict")
	if _, err := os.Stat(path); err != nil {
		path = filepath.Join("data", "dicts", "rime-ice.tsdict")
	}
	dict, err := dictionary.LoadFile(path)
	if err != nil {
		t.Skip(err)
	}
	return dict
}

func TestRealDictNihcProgressivePrefixesOnFirstPage(t *testing.T) {
	eng := engine.NewPinyinEngine(loadIceDict(t))
	eng.SetScheme(pinyin.SchemeXiaohe)
	m := NewManager(eng)
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	var last *KeyResult
	for _, ch := range "nihc" {
		last = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: string(ch)})
	}
	if last == nil || last.Candidates == nil {
		t.Fatal("nil candidates")
	}
	if last.Candidates.PageSize != 9 {
		t.Fatalf("pageSize=%d", last.Candidates.PageSize)
	}
	if len(last.Candidates.Candidates) != 9 {
		t.Fatalf("want 9 visible candidates, got %d %#v", len(last.Candidates.Candidates), candidateTexts(last))
	}
	texts := candidateTexts(last)
	if texts[0] != "你好" {
		t.Fatalf("top=%q texts=%v", texts[0], texts)
	}
	foundNi := false
	for _, text := range texts {
		if text == "你" {
			foundNi = true
		}
		if text == "拟好" || text == "尼好" || text == "泥好" {
			t.Fatalf("weak same-length phrase crowded first page: %v", texts)
		}
		if len([]rune(text)) > 2 {
			t.Fatalf("unexpected composed long candidate on first page: %v", texts)
		}
	}
	if !foundNi {
		t.Fatalf("first page missing single-char 你: %v", texts)
	}
}

func TestRealDictNihcuijpShowsPrefixesNotOnlyComposed(t *testing.T) {
	eng := engine.NewPinyinEngine(loadIceDict(t))
	eng.SetScheme(pinyin.SchemeXiaohe)
	m := NewManager(eng)
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	var last *KeyResult
	for _, ch := range "nihcuijp" {
		last = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: string(ch)})
	}
	texts := candidateTexts(last)
	if len(texts) != 9 {
		t.Fatalf("want 9 candidates, got %v", texts)
	}
	if !containsText(texts, "你好") || !containsText(texts, "你") {
		t.Fatalf("missing progressive prefixes: %v", texts)
	}
	for _, text := range texts {
		if text == "你好是接" || text == "你好是解" || text == "尼好是接" {
			t.Fatalf("composed lattice leaked onto first page: %v", texts)
		}
	}
}
