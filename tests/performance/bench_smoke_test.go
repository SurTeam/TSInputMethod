package performance_test

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	"github.com/SurTeam/TSInputMethod/internal/ranking"
)

func BenchmarkRankingFastPath(b *testing.B) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "你好", Syllables: []string{"ni", "hao"}, Weight: 5000},
		{Text: "你", Syllables: []string{"ni"}, Weight: 1000},
		{Text: "好", Syllables: []string{"hao"}, Weight: 900},
	})
	dec := pinyin.NewDecoder(dict, 16)
	coord := ranking.NewCoordinator(ranking.NewFastRanker(), nil)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		raw := dec.Decode("nihao")
		req := ranking.Request{Pinyin: "nihao", Candidates: make([]ranking.Candidate, 0, len(raw))}
		for _, item := range raw {
			req.Candidates = append(req.Candidates, ranking.Candidate{
				Text:    item.Text,
				Comment: item.Comment,
				Score:   item.Score,
			})
		}
		_ = coord.RankSync(req)
	}
}
