package pinyin

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
)

func progressiveDict() *dictionary.Dictionary {
	return dictionary.FromEntries([]dictionary.Entry{
		{Text: "中国人民", Syllables: []string{"zhong", "guo", "ren", "min"}, Weight: 5000},
		{Text: "中国人", Syllables: []string{"zhong", "guo", "ren"}, Weight: 4000},
		{Text: "中国", Syllables: []string{"zhong", "guo"}, Weight: 3000},
		{Text: "中", Syllables: []string{"zhong"}, Weight: 2000},
		{Text: "国", Syllables: []string{"guo"}, Weight: 900},
		{Text: "人", Syllables: []string{"ren"}, Weight: 800},
		{Text: "民", Syllables: []string{"min"}, Weight: 700},
		{Text: "人民", Syllables: []string{"ren", "min"}, Weight: 2500},
	})
}

func textsBySyllableLen(cands []Candidate) map[int][]string {
	out := map[int][]string{}
	for _, cand := range cands {
		out[len(cand.Syllables)] = append(out[len(cand.Syllables)], cand.Text)
	}
	return out
}

func hasText(cands []Candidate, want string) bool {
	for _, cand := range cands {
		if cand.Text == want {
			return true
		}
	}
	return false
}

func TestDecodeSyllablesProgressivePrefixes(t *testing.T) {
	dec := NewDecoder(progressiveDict(), 32)
	analyzed := AnalyzeInput(SchemeXiaohe, "vsgorfmb")
	if len(analyzed.Syllables) != 4 || analyzed.Remainder != "" {
		t.Fatalf("analyze = %#v", analyzed)
	}
	cands := dec.DecodeSyllables(analyzed.Syllables, analyzed.Remainder)
	byLen := textsBySyllableLen(cands)
	for _, n := range []int{4, 3, 2, 1} {
		if len(byLen[n]) == 0 {
			t.Fatalf("missing %d-syllable prefix candidates: %#v", n, cands)
		}
	}
	if !hasText(cands, "中国人民") || !hasText(cands, "中国人") || !hasText(cands, "中国") || !hasText(cands, "中") {
		t.Fatalf("expected progressive words, got %#v", cands)
	}
}

func TestDecodeSyllablesOddLengthKeepsCompletePrefixes(t *testing.T) {
	dec := NewDecoder(progressiveDict(), 32)
	analyzed := AnalyzeInput(SchemeXiaohe, "vsgorfm")
	if len(analyzed.Syllables) != 3 || analyzed.Remainder != "m" {
		t.Fatalf("analyze odd = %#v", analyzed)
	}
	cands := dec.DecodeSyllables(analyzed.Syllables, analyzed.Remainder)
	if !hasText(cands, "中国人") || !hasText(cands, "中国") || !hasText(cands, "中") {
		t.Fatalf("odd-length missing complete prefixes: %#v", cands)
	}
	for _, cand := range cands {
		if cand.Text == "中国人" && cand.Incomplete {
			t.Fatalf("complete-syllable hit should not be incomplete: %#v", cand)
		}
	}
}

func TestReservePrefixDiversityKeepsShortPrefixes(t *testing.T) {
	entries := []dictionary.Entry{
		{Text: "中", Syllables: []string{"zhong"}, Weight: 100},
		{Text: "国", Syllables: []string{"guo"}, Weight: 100},
		{Text: "人", Syllables: []string{"ren"}, Weight: 100},
		{Text: "民", Syllables: []string{"min"}, Weight: 100},
		{Text: "中国", Syllables: []string{"zhong", "guo"}, Weight: 100},
		{Text: "人民", Syllables: []string{"ren", "min"}, Weight: 100},
	}
	for i := 0; i < 40; i++ {
		entries = append(entries, dictionary.Entry{
			Text:      string(rune('A'+i)) + "词",
			Syllables: []string{"zhong", "guo", "ren", "min"},
			Weight:    uint32(9000 - i),
		})
	}
	dec := NewDecoder(dictionary.FromEntries(entries), 16)
	cands := dec.DecodeSyllables([]string{"zhong", "guo", "ren", "min"}, "")
	if len(cands) > 16 {
		t.Fatalf("limit broken: %d", len(cands))
	}
	if !hasText(cands, "中") || !hasText(cands, "中国") {
		t.Fatalf("short prefixes truncated away: %#v", cands)
	}
}
