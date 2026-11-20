package ranking

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFeatureHashStableAcrossRestarts(t *testing.T) {
	a := HashFeature("code_candidate=bwjk|北京", DefaultHashDimension)
	b := HashFeature("code_candidate=bwjk|北京", DefaultHashDimension)
	if a != b {
		t.Fatalf("hash unstable: %d vs %d", a, b)
	}
	fa := ExtractFeatures("我明天要去", "bwjk", "北京", DefaultHashDimension)
	fb := ExtractFeatures("我明天要去", "bwjk", "北京", DefaultHashDimension)
	if len(fa) == 0 || len(fa) != len(fb) {
		t.Fatalf("features mismatch %v vs %v", fa, fb)
	}
	for i := range fa {
		if fa[i] != fb[i] {
			t.Fatalf("feature %d %d != %d", i, fa[i], fb[i])
		}
	}
}

func TestPairwiseSGDRaisesSelectedInSameContext(t *testing.T) {
	m := NewPersonalModel(DefaultSGDConfig())
	ctx := "我明天要去"
	code := "bwjk"
	beforeSel := m.Score(ctx, code, "北京")
	beforeTop := m.Score(ctx, code, "背景")
	for i := 0; i < 20; i++ {
		m.PairwiseUpdate(ctx, code, "北京", "背景")
	}
	afterSel := m.Score(ctx, code, "北京")
	afterTop := m.Score(ctx, code, "背景")
	if afterSel-afterTop <= beforeSel-beforeTop {
		t.Fatalf("expected selected margin to rise: before sel=%v top=%v after sel=%v top=%v", beforeSel, beforeTop, afterSel, afterTop)
	}
}

func TestDifferentContextNotGloballyOverwritten(t *testing.T) {
	m := NewPersonalModel(DefaultSGDConfig())
	for i := 0; i < 30; i++ {
		m.PairwiseUpdate("我明天要去", "bwjk", "北京", "背景")
	}
	marginA := m.Score("我明天要去", "bwjk", "北京") - m.Score("我明天要去", "bwjk", "背景")
	marginB := m.Score("这张图片的", "bwjk", "北京") - m.Score("这张图片的", "bwjk", "背景")
	if marginA <= 0 {
		t.Fatalf("trained context margin should be positive, got %v", marginA)
	}
	if marginB >= marginA {
		t.Fatalf("other context should not match or exceed trained margin: a=%v b=%v", marginA, marginB)
	}
}

func TestWeightAmplitudeCaps(t *testing.T) {
	cfg := DefaultSGDConfig()
	cfg.WeightCap = 0.2
	cfg.LearningRate = 1.0
	m := NewPersonalModel(cfg)
	for i := 0; i < 100; i++ {
		m.PairwiseUpdate("ctx", "code", "pos", "neg")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for idx, w := range m.weights {
		if w > cfg.WeightCap || w < -cfg.WeightCap {
			t.Fatalf("weight %d=%v exceeds cap", idx, w)
		}
	}
}

func TestShortTermCacheBoundedAndDecays(t *testing.T) {
	c := NewShortTermCache(2, time.Second)
	now := time.Now()
	c.RememberLocked("c", "k", "a", 1, now)
	c.RememberLocked("c", "k", "b", 1, now)
	c.RememberLocked("c", "k", "d", 1, now)
	if c.Len() != 2 {
		t.Fatalf("capacity exceeded: %d", c.Len())
	}
	v := c.LookupLocked("c", "k", "d", now)
	if v <= 0 {
		t.Fatalf("expected live bias, got %v", v)
	}
	decayed := c.LookupLocked("c", "k", "d", now.Add(900*time.Millisecond))
	if decayed >= v {
		t.Fatalf("expected decay %v -> %v", v, decayed)
	}
	if c.LookupLocked("c", "k", "d", now.Add(2*time.Second)) != 0 {
		t.Fatalf("expired entry should be zero")
	}
}

func TestUpdateQueueDropsWhenFullWithoutBlocking(t *testing.T) {
	m := NewPersonalModel(DefaultSGDConfig())
	q := NewUpdateQueue(m, 1)
	defer q.Close()
	q.started.Store(true)
	q.ch <- PairwiseFeedback{Context: "x", Code: "c", Selected: "a", Top: "b"}
	start := time.Now()
	ok := q.Enqueue(PairwiseFeedback{Context: "x", Code: "c", Selected: "a", Top: "b"})
	if ok {
		t.Fatalf("expected drop when full")
	}
	if time.Since(start) > 200*time.Millisecond {
		t.Fatalf("enqueue blocked on full queue")
	}
	if q.Dropped() == 0 {
		t.Fatalf("expected dropped counter")
	}
}

func TestPersonalModelPersistResetAndIncompatible(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "personal_model.json")
	m := NewPersonalModel(DefaultSGDConfig())
	m.SetPath(path)
	m.PairwiseUpdate("我明天要去", "bwjk", "北京", "背景")
	if err := m.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	m2, err := OpenPersonalModel(path, DefaultSGDConfig())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if m2.UpdateCount() == 0 || m2.WeightCount() == 0 {
		t.Fatalf("expected restored weights, updates=%d weights=%d", m2.UpdateCount(), m2.WeightCount())
	}
	m2.ResetPersonalization()
	if m2.UpdateCount() != 0 || m2.WeightCount() != 0 {
		t.Fatalf("reset failed")
	}
	bad := personalModelFile{
		FormatVersion:        personalModelFormatVersion,
		HashDimension:        128,
		HashAlgorithm:        HashAlgorithmFNV1a64,
		FeatureSchemaVersion: FeatureSchemaVersion,
		Weights:              map[string]float64{"1": 0.1},
	}
	if err := savePersonalModelFile(path, bad); err != nil {
		t.Fatalf("write bad: %v", err)
	}
	if err := NewPersonalModel(DefaultSGDConfig()).LoadFile(path); err == nil {
		t.Fatalf("expected incompatible dimension error")
	}
	_ = os.WriteFile(path, []byte("{bad"), 0o600)
	if err := NewPersonalModel(DefaultSGDConfig()).LoadFile(path); err == nil {
		t.Fatalf("expected corrupt error")
	}
}

func TestSameFeaturesUsedForTrainAndScore(t *testing.T) {
	m := NewPersonalModel(DefaultSGDConfig())
	ctx, code, cand := "前文", "nihc", "你好"
	feats := ExtractFeatures(ctx, code, cand, m.cfg.Dimension)
	m.mu.Lock()
	for _, idx := range feats {
		m.weights[idx] = 0.25
	}
	m.mu.Unlock()
	got := m.Score(ctx, code, cand)
	if got < 0.25 {
		t.Fatalf("score should use extracted features, got %v", got)
	}
}
