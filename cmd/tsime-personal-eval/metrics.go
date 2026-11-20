package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/SurTeam/TSInputMethod/internal/ranking"
)

type ModeName string

const (
	ModeFast        ModeName = "Fast"
	ModeUserFreq    ModeName = "Fast+UserFreq"
	ModeSGD         ModeName = "Fast+SGD"
	ModeUserFreqSGD ModeName = "Fast+UserFreq+SGD"
)

type ModeResult struct {
	Mode               ModeName
	Samples            int
	Top1               float64
	Top5               float64
	MRR                float64
	AvgRankDelta       float64
	FalsePromotions    int
	WeightCount        int
	EstModelBytes      int64
	RuntimeHeapAllocB  uint64
	AvgLatencyNS       float64
}

func rankOf(cands []ranking.Candidate, text string) int {
	for i, c := range cands {
		if c.Text == text {
			return i
		}
	}
	return len(cands)
}

func runMode(name ModeName, samples []Sample, useSGD, useFreq bool) ModeResult {
	var model *ranking.PersonalModel
	if useSGD {
		model = ranking.NewPersonalModel(ranking.DefaultSGDConfig())
	}
	usageMap := map[string]uint32{}
	if useFreq {
		usageMap["北京"] = 80
	}
	usage := func(text string) uint32 { return usageMap[text] }
	var usageLookup ranking.UsageLookup
	if useFreq {
		usageLookup = usage
	}
	fused := ranking.NewFusedRanker(ranking.NewFastRanker(), model, usageLookup, ranking.FusionConfig{
		Enabled:             useSGD || useFreq,
		SGDWeight:           40,
		UserFrequencyWeight: 8,
	})

	var top1, top5, mrrSum float64
	var rankDeltaSum float64
	var rankDeltaN int
	var falsePromo int
	var latencySum time.Duration
	firstSeen := map[string]int{}

	for i, sample := range samples {
		if sample.SelectedIndex < 0 || sample.SelectedIndex >= len(sample.Candidates) {
			continue
		}
		selected := sample.Candidates[sample.SelectedIndex].Text
		req := ranking.Request{
			Pinyin:     sample.InputCode,
			Context:    sample.Context,
			Candidates: cloneCands(sample.Candidates),
		}
		start := time.Now()
		out := fused.Rank(req)
		latencySum += time.Since(start)

		rank := rankOf(out, selected)
		if rank == 0 {
			top1++
		}
		if rank < 5 {
			top5++
		}
		if rank < len(out) {
			mrrSum += 1.0 / float64(rank+1)
		}

		key := sample.Context + "\x00" + sample.InputCode
		if prev, ok := firstSeen[key]; ok {
			rankDeltaSum += float64(prev - rank)
			rankDeltaN++
		} else {
			firstSeen[key] = rank
		}

		if useSGD && model != nil && rank > 0 && len(out) > 0 {
			topText := out[0].Text
			model.PairwiseUpdate(sample.Context, sample.InputCode, selected, topText)
			if useFreq {
				usageMap[selected]++
			}
		} else if useFreq && !useSGD {
			usageMap[selected]++
		}

		_ = i
	}

	if useSGD && model != nil {
		for _, sample := range samples {
			if sample.SelectedIndex < 0 || sample.SelectedIndex >= len(sample.Candidates) {
				continue
			}
			selected := sample.Candidates[sample.SelectedIndex].Text
			req := ranking.Request{
				Pinyin:     sample.InputCode,
				Context:    sample.Context,
				Candidates: cloneCands(sample.Candidates),
			}
			out := fused.Rank(req)
			if len(out) == 0 {
				continue
			}
			if out[0].Text != selected {
				for _, other := range samples {
					if other.Context == sample.Context || other.InputCode != sample.InputCode {
						continue
					}
					if other.SelectedIndex < 0 || other.SelectedIndex >= len(other.Candidates) {
						continue
					}
					want := other.Candidates[other.SelectedIndex].Text
					if out[0].Text == want {
						falsePromo++
						break
					}
				}
			}
		}
	}

n := float64(len(samples))
	if n == 0 {
		n = 1
	}
	var ms runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&ms)
	res := ModeResult{
		Mode:              name,
		Samples:           len(samples),
		Top1:              top1 / n,
		Top5:              top5 / n,
		MRR:               mrrSum / n,
		FalsePromotions:   falsePromo,
		AvgLatencyNS:      float64(latencySum) / n,
		RuntimeHeapAllocB: ms.HeapAlloc,
	}
	if rankDeltaN > 0 {
		res.AvgRankDelta = rankDeltaSum / float64(rankDeltaN)
	}
	if model != nil {
		res.WeightCount = model.WeightCount()
		res.EstModelBytes = int64(res.WeightCount)*16 + int64(model.CacheLen())*64
	}
	res.EstModelBytes += int64(len(usageMap)) * 24
	return res
}

func formatResult(r ModeResult) string {
	return fmt.Sprintf(
		"%-18s samples=%d top1=%.3f top5=%.3f mrr=%.3f avg_rank_delta=%+.3f false_promo=%d weights=%d est_model_bytes=%d heap_alloc_b=%d avg_latency_ns=%.0f",
		r.Mode, r.Samples, r.Top1, r.Top5, r.MRR, r.AvgRankDelta, r.FalsePromotions, r.WeightCount, r.EstModelBytes, r.RuntimeHeapAllocB, r.AvgLatencyNS,
	)
}
