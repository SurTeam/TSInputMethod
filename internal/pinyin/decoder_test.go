package pinyin

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
)

func fixtureDict() *dictionary.Dictionary {
	return dictionary.FromEntries([]dictionary.Entry{
		{Text: "你", Syllables: []string{"ni"}, Weight: 1000},
		{Text: "泥", Syllables: []string{"ni"}, Weight: 100},
		{Text: "好", Syllables: []string{"hao"}, Weight: 900},
		{Text: "号", Syllables: []string{"hao"}, Weight: 80},
		{Text: "你好", Syllables: []string{"ni", "hao"}, Weight: 5000},
		{Text: "世界", Syllables: []string{"shi", "jie"}, Weight: 4000},
		{Text: "是", Syllables: []string{"shi"}, Weight: 800},
		{Text: "市", Syllables: []string{"shi"}, Weight: 200},
		{Text: "界", Syllables: []string{"jie"}, Weight: 700},
		{Text: "西安", Syllables: []string{"xi", "an"}, Weight: 3000},
		{Text: "先", Syllables: []string{"xian"}, Weight: 2500},
		{Text: "西", Syllables: []string{"xi"}, Weight: 600},
		{Text: "安", Syllables: []string{"an"}, Weight: 500},
	})
}

func TestDecodeExactPhraseAndWeight(t *testing.T) {
	dec := NewDecoder(fixtureDict(), 16)
	cands := dec.Decode("nihao")
	if len(cands) == 0 {
		t.Fatal("no candidates")
	}
	if cands[0].Text != "你好" {
		t.Fatalf("top = %#v", cands)
	}
}

func TestDecodeLatticeComposition(t *testing.T) {
	dec := NewDecoder(fixtureDict(), 16)
	cands := dec.Decode("shijie")
	found := false
	for _, cand := range cands {
		if cand.Text == "世界" || cand.Text == "是界" {
			found = true
		}
	}
	if !found {
		t.Fatalf("lattice missing phrase, got %#v", cands)
	}
}

func TestDecodeAmbiguousSegmentation(t *testing.T) {
	dec := NewDecoder(fixtureDict(), 16)
	cands := dec.Decode("xian")
	hasXian := false
	hasXiAn := false
	for _, cand := range cands {
		if cand.Text == "先" {
			hasXian = true
		}
		if cand.Text == "西安" {
			hasXiAn = true
		}
	}
	if !hasXian || !hasXiAn {
		t.Fatalf("expected 先 and 西安, got %#v", cands)
	}
}
