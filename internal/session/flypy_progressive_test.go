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

func TestFlypyProgressiveNihcWithRealDict(t *testing.T) {
	path := filepath.Join("..", "..", "data", "dicts", "rime-ice.tsdict")
	if _, err := os.Stat(path); err != nil {
		path = filepath.Join("data", "dicts", "rime-ice.tsdict")
	}
	dict, err := dictionary.LoadFile(path)
	if err != nil {
		t.Skip(err)
	}
	eng := engine.NewPinyinEngine(dict)
	eng.SetScheme(pinyin.SchemeXiaohe)
	m := NewManager(eng)
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	var last *KeyResult
	for _, ch := range []string{"n", "i", "h", "c"} {
		last = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: ch})
		t.Logf("key=%s pre=%v n=%d top=%v", ch, last.Composition, candN(last), topText(last))
	}
	if last == nil || last.Composition == nil || last.Composition.Preedit != "nihc" {
		t.Fatalf("preedit=%v", last)
	}
	if topBareText(last) != "你好" {
		t.Fatalf("want 你好 top=%q candidates=%v", topBareText(last), candidateBareTexts(last))
	}
}

func candN(r *KeyResult) int {
	if r == nil || r.Candidates == nil {
		return 0
	}
	return len(r.Candidates.Candidates)
}

func topText(r *KeyResult) string {
	return topBareText(r)
}

func candTexts(r *KeyResult) []string {
	return candidateBareTexts(r)
}
