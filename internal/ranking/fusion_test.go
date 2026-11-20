package ranking

import "testing"

func TestFusedRankerReordersWithinCompatibleGroup(t *testing.T) {
	model := NewPersonalModel(DefaultSGDConfig())
	for i := 0; i < 25; i++ {
		model.PairwiseUpdate("我明天要去", "bwjk", "北京", "背景")
	}
	fused := NewFusedRanker(NewFastRanker(), model, nil, DefaultFusionConfig())
	out := fused.Rank(Request{
		Pinyin:  "bwjk",
		Context: "我明天要去",
		Candidates: []Candidate{
			{Text: "背景", Score: 1000, Coverage: 2},
			{Text: "北京", Score: 900, Coverage: 2},
		},
	})
	if out[0].Text != "北京" {
		t.Fatalf("expected 北京 first after fusion, got %#v", out)
	}
}

func TestFusedRankerKeepsIncompatibleGroupsSeparated(t *testing.T) {
	model := NewPersonalModel(DefaultSGDConfig())
	for i := 0; i < 40; i++ {
		model.PairwiseUpdate("ctx", "code", "短", "长词")
	}
	fused := NewFusedRanker(NewFastRanker(), model, nil, DefaultFusionConfig())
	out := fused.Rank(Request{
		Pinyin:  "code",
		Context: "ctx",
		Candidates: []Candidate{
			{Text: "长词", Score: 100, Coverage: 2},
			{Text: "短", Score: 10, Coverage: 1},
		},
	})
	if out[0].Coverage != 2 {
		t.Fatalf("coverage group order broken: %#v", out)
	}
}

func TestFusedRankerDisabledRestoresFastRanker(t *testing.T) {
	model := NewPersonalModel(DefaultSGDConfig())
	for i := 0; i < 25; i++ {
		model.PairwiseUpdate("我明天要去", "bwjk", "北京", "背景")
	}
	cfg := DefaultFusionConfig()
	cfg.Enabled = false
	fused := NewFusedRanker(NewFastRanker(), model, nil, cfg)
	out := fused.Rank(Request{
		Pinyin:  "bwjk",
		Context: "我明天要去",
		Candidates: []Candidate{
			{Text: "背景", Score: 1000, Coverage: 2},
			{Text: "北京", Score: 900, Coverage: 2},
		},
	})
	if out[0].Text != "背景" {
		t.Fatalf("disabled fusion should keep FastRanker order, got %#v", out)
	}
}

func TestUserFrequencyContributesInsideGroup(t *testing.T) {
	usage := func(text string) uint32 {
		if text == "北京" {
			return 100
		}
		return 0
	}
	fused := NewFusedRanker(NewFastRanker(), nil, usage, DefaultFusionConfig())
	out := fused.Rank(Request{
		Pinyin: "bwjk",
		Candidates: []Candidate{
			{Text: "背景", Score: 100, Coverage: 2},
			{Text: "北京", Score: 99, Coverage: 2},
		},
	})
	if out[0].Text != "北京" {
		t.Fatalf("user frequency should lift 北京, got %#v", out)
	}
}
