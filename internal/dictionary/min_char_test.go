package dictionary

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompiledDictContainsMinGlyph(t *testing.T) {
	path := filepath.Join("..", "..", "data", "dicts", "rime-ice.tsdict")
	if _, err := os.Stat(path); err != nil {
		t.Skip("compiled dictionary missing")
	}
	dict, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := dict.Lookup([]string{"min"})
	for _, entry := range got {
		if entry.Text == "慜" {
			return
		}
	}
	t.Fatalf("慜 missing from min lookup (%d hits)", len(got))
}
