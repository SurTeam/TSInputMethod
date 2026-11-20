package ranking

import "testing"

func BenchmarkPersonalScore(b *testing.B) {
	m := NewPersonalModel(DefaultSGDConfig())
	for i := 0; i < 50; i++ {
		m.PairwiseUpdate("我明天要去", "bwjk", "北京", "背景")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Score("我明天要去", "bwjk", "北京")
	}
}

func BenchmarkFusedRanker(b *testing.B) {
	m := NewPersonalModel(DefaultSGDConfig())
	for i := 0; i < 50; i++ {
		m.PairwiseUpdate("我明天要去", "bwjk", "北京", "背景")
	}
	fused := NewFusedRanker(NewFastRanker(), m, func(text string) uint32 {
		if text == "北京" {
			return 20
		}
		return 0
	}, DefaultFusionConfig())
	req := Request{
		Pinyin:  "bwjk",
		Context: "我明天要去",
		Candidates: []Candidate{
			{Text: "背景", Score: 1000, Coverage: 2},
			{Text: "北京", Score: 900, Coverage: 2},
			{Text: "背井", Score: 100, Coverage: 2},
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fused.Rank(req)
	}
}
