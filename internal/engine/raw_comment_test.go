package engine

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
)

func TestXiaoheRawCommentUsesCoveredKeys(t *testing.T) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "世界", Syllables: []string{"shi", "jie"}, Weight: 1000},
		{Text: "世", Syllables: []string{"shi"}, Weight: 100},
		{Text: "界", Syllables: []string{"jie"}, Weight: 90},
	})
	eng := NewPinyinEngine(dict)
	eng.SetScheme(pinyin.SchemeXiaohe)
	cands, err := eng.Decode("uijp")
	if err != nil || len(cands) == 0 {
		t.Fatalf("decode = %#v err=%v", cands, err)
	}
	if cands[0].Text != "世界" || cands[0].Comment != "uijp" {
		t.Fatalf("top=%#v", cands[0])
	}
	foundShi := false
	for _, cand := range cands {
		if cand.Text == "世" {
			foundShi = true
			if cand.Comment != "ui" {
				t.Fatalf("prefix comment=%#v", cand)
			}
		}
		if cand.Comment == "shi'jie" || cand.Comment == "shi" {
			t.Fatalf("quanpin comment leaked: %#v", cand)
		}
	}
	if !foundShi {
		t.Fatalf("missing 世 in %#v", cands)
	}
}
