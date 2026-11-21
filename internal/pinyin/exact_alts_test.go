package pinyin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
)

func iceDict(t *testing.T) *dictionary.Dictionary {
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

func candidateHas(cands []Candidate, want string) bool {
	for _, cand := range cands {
		if cand.Text == want {
			return true
		}
	}
	return false
}

func TestExactFullCodeKeepsSameSyllableAlternatives(t *testing.T) {
	dec := NewDecoder(iceDict(t), 96)
	analyzed := AnalyzeInput(SchemeXiaohe, "tsyi")
	if len(analyzed.Syllables) != 2 || analyzed.Syllables[0] != "tong" || analyzed.Syllables[1] != "yi" || analyzed.Remainder != "" {
		t.Fatalf("analyze tsyi = %#v", analyzed)
	}
	cands := dec.DecodeSyllables(analyzed.Syllables, analyzed.Remainder)
	for _, want := range []string{"同意", "统一", "同一", "同义", "通译"} {
		if !candidateHas(cands, want) {
			t.Fatalf("missing exact alternative %s: %#v", want, cands)
		}
	}
	if cands[0].Text != "同意" {
		t.Fatalf("top = %#v", cands[0])
	}
}

func TestHalfCodePrefixBroaderThanFullCodeExact(t *testing.T) {
	dec := NewDecoder(iceDict(t), 96)
	half := AnalyzeInput(SchemeXiaohe, "tsy")
	full := AnalyzeInput(SchemeXiaohe, "tsyi")
	halfCands := dec.DecodeSyllables(half.Syllables, half.Remainder)
	fullCands := dec.DecodeSyllables(full.Syllables, full.Remainder)
	if !candidateHas(halfCands, "同样") {
		t.Fatalf("half-code should keep incomplete prefix hit 同样: %#v", halfCands)
	}
	if candidateHas(fullCands, "同样") {
		t.Fatalf("full-code exact path should not include incomplete 同样: %#v", fullCands)
	}
	if !candidateHas(fullCands, "统一") {
		t.Fatalf("full-code should keep exact 统一: %#v", fullCands)
	}
}
