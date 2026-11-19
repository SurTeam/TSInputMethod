package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/emoji"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
)

func TestPinyinEngineExpandsEmojiAfterChinese(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "emoji.txt")
	if err := os.WriteFile(path, []byte("你好\t你好 👋\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	table, err := emoji.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	eng := NewPinyinEngine(testDict())
	eng.SetScheme(pinyin.SchemeQuanpin)
	eng.SetEmojiTable(table)
	cands, err := eng.Decode("nihao")
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(cands) < 2 || cands[0].Text != "你好" || cands[1].Text != "👋" {
		t.Fatalf("emoji expand = %#v", cands)
	}
	if cands[1].ConsumedKeys != cands[0].ConsumedKeys {
		t.Fatalf("emoji consumed keys = %d want %d", cands[1].ConsumedKeys, cands[0].ConsumedKeys)
	}
}
