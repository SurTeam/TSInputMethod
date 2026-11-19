package engine

import (
	"context"
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/inference"
	"github.com/SurTeam/TSInputMethod/internal/ranking"
)

func testDict() *dictionary.Dictionary {
	return dictionary.FromEntries([]dictionary.Entry{
		{Text: "你", Syllables: []string{"ni"}, Weight: 100},
		{Text: "泥", Syllables: []string{"ni"}, Weight: 10},
		{Text: "你好", Syllables: []string{"ni", "hao"}, Weight: 1000},
		{Text: "好", Syllables: []string{"hao"}, Weight: 90},
	})
}

func TestPinyinEngineUsesFastRanking(t *testing.T) {
	eng := NewPinyinEngine(testDict())
	cands, err := eng.Decode("nihao")
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(cands) == 0 || cands[0].Text != "你好" {
		t.Fatalf("expected frequency top candidate, got %#v", cands)
	}
}

func TestPinyinEngineDegradesWhenBackendUnavailable(t *testing.T) {
	coord := ranking.NewCoordinator(ranking.NewFastRanker(), inference.UnavailableBackend{})
	eng := NewPinyinEngineWithRanking(testDict(), coord)
	eng.SetInferenceBackend(inference.UnavailableBackend{})
	cands, err := eng.Decode("nihao")
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(cands) == 0 || cands[0].Text != "你好" {
		t.Fatalf("no-model path must still return base candidates, got %#v", cands)
	}
	reordered := coord.RankWithContext(context.Background(), ranking.Request{
		Pinyin: "nihao",
		Candidates: []ranking.Candidate{
			{Text: "泥", Score: 10},
			{Text: "你", Score: 100},
		},
		Generation: 1,
	})
	if len(reordered) == 0 || reordered[0].Text != "你" {
		t.Fatalf("RankWithContext should degrade to FastRanker order, got %#v", reordered)
	}
}
