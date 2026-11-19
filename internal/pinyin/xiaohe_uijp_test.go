package pinyin

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
)

func TestExpandXiaoheUijpIsShiJie(t *testing.T) {
	syls, rem, ok := ExpandXiaohe("uijp")
	if !ok || rem != "" || len(syls) != 2 || syls[0] != "shi" || syls[1] != "jie" {
		t.Fatalf("expand uijp => %v rem=%q ok=%v", syls, rem, ok)
	}
	analyzed := AnalyzeInput(SchemeXiaohe, "uijp")
	if !analyzed.FixedSplit || analyzed.Joined != "shijie" {
		t.Fatalf("analyze = %#v", analyzed)
	}
}

func TestDecodeXiaoheUijpKeepsShiJieBoundary(t *testing.T) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "世界", Syllables: []string{"shi", "jie"}, Weight: 1000},
		{Text: "世", Syllables: []string{"shi"}, Weight: 100},
		{Text: "界", Syllables: []string{"jie"}, Weight: 90},
		{Text: "机", Syllables: []string{"ji"}, Weight: 80},
		{Text: "恶", Syllables: []string{"e"}, Weight: 70},
	})
	dec := NewDecoder(dict, 16)
	analyzed := AnalyzeInput(SchemeXiaohe, "uijp")
	cands := dec.DecodeSyllables(analyzed.Syllables, analyzed.Remainder)
	if len(cands) == 0 || cands[0].Text != "世界" || cands[0].Comment != "shi'jie" {
		t.Fatalf("xiaohe fixed decode = %#v", cands)
	}
	for _, cand := range cands {
		if cand.Comment == "shi'ji'e" || cand.Text == "世机恶" {
			t.Fatalf("quanpin re-split leaked into flypy decode: %#v", cands)
		}
	}
}
