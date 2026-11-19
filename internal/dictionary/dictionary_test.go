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

func TestLookupPrefixPrefersShortWeightAcrossLexOrder(t *testing.T) {
	dict := FromEntries([]Entry{
		{Text: "你还", Syllables: []string{"ni", "hai"}, Weight: 10},
		{Text: "你还不", Syllables: []string{"ni", "hai", "bu"}, Weight: 1000},
		{Text: "你好", Syllables: []string{"ni", "hao"}, Weight: 500},
	})
	got := dict.LookupPrefix([]string{"ni"}, "h")
	if len(got) < 2 {
		t.Fatalf("prefix hits = %#v", got)
	}
	found := false
	for _, entry := range got {
		if entry.Text == "你好" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("你好 missing from %#v", got)
	}
	if got[0].Text != "你还" && got[0].Text != "你好" {
		t.Fatalf("unexpected shortest-first top %#v", got)
	}
}
