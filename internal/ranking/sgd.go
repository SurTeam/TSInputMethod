package ranking

import (
	"math"
	"sync"
	"time"
)

const (
	defaultLearningRate = 0.01
	defaultL2           = 1e-4
	defaultGradClip     = 1.0
	defaultWeightCap    = 5.0
	defaultMaxWeights   = 50000
)

type SGDConfig struct {
	Dimension    uint32
	LearningRate float64
	L2           float64
	GradClip     float64
	WeightCap    float64
	MaxWeights   int
}

func DefaultSGDConfig() SGDConfig {
	return SGDConfig{
		Dimension:    DefaultHashDimension,
		LearningRate: defaultLearningRate,
		L2:           defaultL2,
		GradClip:     defaultGradClip,
		WeightCap:    defaultWeightCap,
		MaxWeights:   defaultMaxWeights,
	}
}

type PersonalModel struct {
	mu          sync.RWMutex
	cfg         SGDConfig
	weights     map[uint32]float64
	updateCount uint64
	updatedAt   time.Time
	cache       *ShortTermCache
	enabled     bool
	dirty       bool
	path        string
}

func NewPersonalModel(cfg SGDConfig) *PersonalModel {
	if cfg.Dimension == 0 {
		cfg.Dimension = DefaultHashDimension
	}
	if cfg.LearningRate <= 0 {
		cfg.LearningRate = defaultLearningRate
	}
	if cfg.L2 < 0 {
		cfg.L2 = defaultL2
	}
	if cfg.GradClip <= 0 {
		cfg.GradClip = defaultGradClip
	}
	if cfg.WeightCap <= 0 {
		cfg.WeightCap = defaultWeightCap
	}
	if cfg.MaxWeights <= 0 {
		cfg.MaxWeights = defaultMaxWeights
	}
	return &PersonalModel{
		cfg:     cfg,
		weights: make(map[uint32]float64),
		cache:   NewShortTermCache(256, 5*time.Minute),
		enabled: true,
	}
}

func (m *PersonalModel) Enabled() bool {
	if m == nil {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

func (m *PersonalModel) SetEnabled(v bool) {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.enabled = v
	m.mu.Unlock()
}

func (m *PersonalModel) UpdateCount() uint64 {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.updateCount
}

func (m *PersonalModel) Version() uint64 {
	return m.UpdateCount()
}

func (m *PersonalModel) Score(context, code, candidate string) float64 {
	if m == nil || !m.Enabled() {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	feats := ExtractFeatures(context, code, candidate, m.cfg.Dimension)
	sgd := dotSparse(m.weights, feats)
	bias := 0.0
	if m.cache != nil {
		bias = m.cache.LookupLocked(context, code, candidate, time.Now())
	}
	return sgd + bias
}

func (m *PersonalModel) PairwiseUpdate(context, code, selected, top string) float64 {
	if m == nil || !m.Enabled() || selected == "" || top == "" || selected == top {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	pos := ExtractFeatures(context, code, selected, m.cfg.Dimension)
	neg := ExtractFeatures(context, code, top, m.cfg.Dimension)
	margin := dotSparse(m.weights, pos) - dotSparse(m.weights, neg)
	prob := stableSigmoid(-margin)
	gradScale := clipFloat(prob, -m.cfg.GradClip, m.cfg.GradClip)
	lr := m.cfg.LearningRate
	applyGrad := func(feats FeatureVector, sign float64) {
		for _, idx := range feats {
			w := m.weights[idx]
			g := sign*gradScale + m.cfg.L2*w
			w -= lr * g
			w = clipFloat(w, -m.cfg.WeightCap, m.cfg.WeightCap)
			if w == 0 {
				delete(m.weights, idx)
			} else {
				m.weights[idx] = w
			}
		}
	}
	applyGrad(pos, -1)
	applyGrad(neg, +1)
	m.trimWeightsLocked()
	m.updateCount++
	m.updatedAt = time.Now().UTC()
	m.dirty = true
	if m.cache != nil {
		m.cache.RememberLocked(context, code, selected, 0.5, time.Now())
	}
	return margin
}

func (m *PersonalModel) AcceptFeedback(context, code, selected, top string) {
	m.PairwiseUpdate(context, code, selected, top)
}

func (m *PersonalModel) ResetPersonalization() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.weights = make(map[uint32]float64)
	m.updateCount = 0
	m.updatedAt = time.Now().UTC()
	m.dirty = true
	if m.cache != nil {
		m.cache.ClearLocked()
	}
}

func (m *PersonalModel) WeightCount() int {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.weights)
}

func (m *PersonalModel) CacheLen() int {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.cache == nil {
		return 0
	}
	return m.cache.Len()
}

func (m *PersonalModel) trimWeightsLocked() {
	if len(m.weights) <= m.cfg.MaxWeights {
		return
	}
	type pair struct {
		idx uint32
		abs float64
	}
	items := make([]pair, 0, len(m.weights))
	for idx, w := range m.weights {
		items = append(items, pair{idx: idx, abs: math.Abs(w)})
	}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].abs < items[i].abs {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	drop := len(items) - m.cfg.MaxWeights
	for i := 0; i < drop; i++ {
		delete(m.weights, items[i].idx)
	}
}

func dotSparse(weights map[uint32]float64, feats FeatureVector) float64 {
	sum := 0.0
	for _, idx := range feats {
		sum += weights[idx]
	}
	return sum
}

func stableSigmoid(x float64) float64 {
	if x >= 0 {
		z := math.Exp(-x)
		return 1 / (1 + z)
	}
	z := math.Exp(x)
	return z / (1 + z)
}

func clipFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
