package ranking

import (
	"math"
	"sort"
)

type UsageLookup func(text string) uint32

type FusionConfig struct {
	Enabled             bool
	SGDWeight           float64
	UserFrequencyWeight float64
}

func DefaultFusionConfig() FusionConfig {
	return FusionConfig{
		Enabled:             true,
		SGDWeight:           40,
		UserFrequencyWeight: 8,
	}
}

type FusedRanker struct {
	base   Ranker
	model  *PersonalModel
	usage  UsageLookup
	cfg    FusionConfig
}

func NewFusedRanker(base Ranker, model *PersonalModel, usage UsageLookup, cfg FusionConfig) *FusedRanker {
	if base == nil {
		base = NewFastRanker()
	}
	if cfg.SGDWeight == 0 {
		cfg.SGDWeight = DefaultFusionConfig().SGDWeight
	}
	if cfg.UserFrequencyWeight == 0 {
		cfg.UserFrequencyWeight = DefaultFusionConfig().UserFrequencyWeight
	}
	return &FusedRanker{base: base, model: model, usage: usage, cfg: cfg}
}

func (r *FusedRanker) SetConfig(cfg FusionConfig) {
	if r == nil {
		return
	}
	r.cfg = cfg
}

func (r *FusedRanker) SetModel(model *PersonalModel) {
	if r == nil {
		return
	}
	r.model = model
}

func (r *FusedRanker) SetUsageLookup(usage UsageLookup) {
	if r == nil {
		return
	}
	r.usage = usage
}

func (r *FusedRanker) Rank(req Request) []Candidate {
	if r == nil {
		return NewFastRanker().Rank(req)
	}
	base := r.base.Rank(req)
	if !r.cfg.Enabled || (r.model == nil && r.usage == nil) {
		return base
	}
	if r.model != nil && !r.model.Enabled() && r.usage == nil {
		return base
	}
	return fuseWithinGroups(base, req, r.model, r.usage, r.cfg)
}

func fuseWithinGroups(base []Candidate, req Request, model *PersonalModel, usage UsageLookup, cfg FusionConfig) []Candidate {
	if len(base) <= 1 {
		return base
	}
	out := make([]Candidate, len(base))
	copy(out, base)
	i := 0
	for i < len(out) {
		j := i + 1
		for j < len(out) && sameConsumptionGroup(out[i], out[j]) {
			j++
		}
		if j-i > 1 {
			sort.SliceStable(out[i:j], func(a, b int) bool {
				ca := out[i+a]
				cb := out[i+b]
				sa := finalScore(req, ca, model, usage, cfg)
				sb := finalScore(req, cb, model, usage, cfg)
				if sa == sb {
					return better(ca, cb)
				}
				return sa > sb
			})
		}
		i = j
	}
	return out
}

func sameConsumptionGroup(a, b Candidate) bool {
	return a.Coverage == b.Coverage && a.Incomplete == b.Incomplete && a.Composed == b.Composed
}

func finalScore(req Request, cand Candidate, model *PersonalModel, usage UsageLookup, cfg FusionConfig) float64 {
	base := calibrateBase(cand.Score)
	userFreq := 0.0
	if usage != nil && cfg.UserFrequencyWeight != 0 {
		userFreq = cfg.UserFrequencyWeight * math.Log1p(float64(usage(cand.Text)))
	}
	sgd := 0.0
	if model != nil && model.Enabled() && cfg.SGDWeight != 0 {
		sgd = cfg.SGDWeight * model.Score(req.Context, req.Pinyin, cand.Text)
	}
	return base + userFreq + sgd
}

func calibrateBase(score int32) float64 {
	if score <= 0 {
		return 0
	}
	return math.Log1p(float64(score))
}
