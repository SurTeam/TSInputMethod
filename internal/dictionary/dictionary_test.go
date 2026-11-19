package dictionary

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFromEntriesLookupAndWeightOrder(t *testing.T) {
	dict := FromEntries([]Entry{
		{Text: "你好", Syllables: []string{"ni", "hao"}, Weight: 100},
		{Text: "逆好", Syllables: []string{"ni", "hao"}, Weight: 10},
		{Text: "你", Syllables: []string{"ni"}, Weight: 50},
		{Text: "好", Syllables: []string{"hao"}, Weight: 40},
	})
	got := dict.Lookup([]string{"ni", "hao"})
	if len(got) != 2 || got[0].Text != "你好" || got[1].Text != "逆好" {
		t.Fatalf("lookup order = %#v", got)
	}
	if dict.EntryCount() != 4 {
		t.Fatalf("entry count = %d", dict.EntryCount())
	}
}

func TestLoadRoundTripFile(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "dicts", "fixture.tsdict")
	if _, err := os.Stat(path); err != nil {
		t.Skip("compiled fixture missing; run tools/dictBuilder first")
	}
	dict, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if dict.EntryCount() == 0 {
		t.Fatal("empty dictionary")
	}
	if !dict.HasSyllable("ni") {
		t.Fatal("missing syllable ni")
	}
}
