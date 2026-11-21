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

func TestFastRankerKeepsSameCoverageExactContiguous(t *testing.T) {
	out := NewFastRanker().Rank(Request{
		Candidates: []Candidate{
			{Text: "同意", Score: 500560, Coverage: 2},
			{Text: "同", Score: 2614683, Coverage: 1},
			{Text: "统一", Score: 5006, Coverage: 2},
			{Text: "通", Score: 1186869, Coverage: 1},
			{Text: "同一", Score: 5003, Coverage: 2},
		},
	})
	seenLower := false
	for _, cand := range out {
		if cand.Coverage == 2 {
			if seenLower {
				t.Fatalf("coverage-2 exacts split by singles: %#v", out)
			}
			continue
		}
		if cand.Coverage < 2 {
			seenLower = true
		}
	}
	if out[0].Text != "同意" || out[1].Text != "统一" || out[2].Text != "同一" {
		t.Fatalf("expected contiguous exacts by score then singles, got %#v", out)
	}
}

func TestFusedRankerCanPromoteWeakerExactInsideContiguousGroup(t *testing.T) {
	model := NewPersonalModel(DefaultSGDConfig())
	for i := 0; i < 40; i++ {
		model.PairwiseUpdate("国家", "tsyi", "统一", "同意")
	}
	fused := NewFusedRanker(NewFastRanker(), model, nil, DefaultFusionConfig())
	out := fused.Rank(Request{
		Pinyin:  "tsyi",
		Context: "国家",
		Candidates: []Candidate{
			{Text: "同意", Score: 500560, Coverage: 2},
			{Text: "统一", Score: 5006, Coverage: 2},
			{Text: "同", Score: 2614683, Coverage: 1},
		},
	})
	if out[0].Text != "统一" {
		t.Fatalf("context fusion should promote 统一 inside coverage-2 group, got %#v", out)
	}
	if out[2].Coverage != 1 {
		t.Fatalf("single should remain after exact group: %#v", out)
	}
}
