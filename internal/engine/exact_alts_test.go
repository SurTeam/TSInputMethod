package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
)

func iceEngineDict(t *testing.T) *dictionary.Dictionary {
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

func engineHas(cands []Candidate, want string) bool {
	for _, cand := range cands {
		if cand.Text == want {
			return true
		}
	}
	return false
}

func TestEngineExactFullCodeKeepsSameSyllableAlternatives(t *testing.T) {
	eng := NewPinyinEngine(iceEngineDict(t))
	eng.SetScheme(pinyin.SchemeXiaohe)
	cands, err := eng.Decode("tsyi")
	if err != nil {
		t.Fatal(err)
	}
	if len(cands) == 0 || cands[0].Text != "同意" {
		t.Fatalf("top = %#v", cands)
	}
	for _, want := range []string{"统一", "同一", "同义", "通译"} {
		if !engineHas(cands, want) {
			t.Fatalf("missing %s: %#v", want, cands)
		}
	}
	seenLower := false
	for _, cand := range cands {
		n := len(cand.Syllables)
		if n == 2 && !cand.Incomplete {
			if seenLower {
				t.Fatalf("same-coverage exacts are not contiguous: %#v", cands)
			}
			continue
		}
		if n < 2 && !cand.Incomplete {
			seenLower = true
		}
	}
}

func TestEngineCoverageFirstExactGroupBeforeSingles(t *testing.T) {
	eng := NewPinyinEngine(iceEngineDict(t))
	eng.SetScheme(pinyin.SchemeXiaohe)
	cands, err := eng.Decode("nihc")
	if err != nil {
		t.Fatal(err)
	}
	if len(cands) == 0 || cands[0].Text != "你好" {
		t.Fatalf("top = %#v", cands)
	}
	if !engineHas(cands, "你") {
		t.Fatalf("missing 你: %#v", cands)
	}
	firstSingle := -1
	lastExact2 := -1
	for i, cand := range cands {
		if !cand.Incomplete && len(cand.Syllables) == 2 {
			lastExact2 = i
		}
		if !cand.Incomplete && len(cand.Syllables) == 1 && firstSingle < 0 {
			firstSingle = i
		}
	}
	if firstSingle < 0 || lastExact2 < 0 {
		t.Fatalf("missing exact/single groups: %#v", cands)
	}
	if lastExact2 > firstSingle {
		t.Fatalf("exact coverage group split by singles: %#v", cands)
	}
}
