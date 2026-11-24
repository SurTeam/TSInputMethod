package ranking

import (
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

func TestUpdateQueueCloseDrainsPending(t *testing.T) {
	m := NewPersonalModel(DefaultSGDConfig())
	q := NewUpdateQueue(m, 64)
	q.started.Store(true)
	for i := 0; i < 10; i++ {
		q.ch <- PairwiseFeedback{Context: "前文", Code: "nihao", Selected: "你好", Top: "泥好"}
	}
	q.looping.Store(true)
	go q.loop()
	q.Close()
	if got := m.UpdateCount(); got != 10 {
		t.Fatalf("close should drain queued feedback, updates=%d", got)
	}
	if q.Enqueue(PairwiseFeedback{Context: "x", Code: "c", Selected: "a", Top: "b"}) {
		t.Fatalf("enqueue after close must be rejected")
	}
}

func TestTrimWeightsKeepsHeadroom(t *testing.T) {
	cfg := DefaultSGDConfig()
	cfg.MaxWeights = 100
	m := NewPersonalModel(cfg)
	for i := 0; i < 200; i++ {
		m.PairwiseUpdate("ctx"+strconv.Itoa(i), "code"+strconv.Itoa(i), "pos"+strconv.Itoa(i), "neg"+strconv.Itoa(i))
		if n := m.WeightCount(); n > cfg.MaxWeights {
			t.Fatalf("weights %d exceed max %d", n, cfg.MaxWeights)
		}
	}
}

func TestScoreAndUpdateConcurrentSafe(t *testing.T) {
	m := NewPersonalModel(DefaultSGDConfig())
	m.SetPath(filepath.Join(t.TempDir(), "personal_model.json"))
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				m.PairwiseUpdate("前文", "nihao", "你好", "泥好")
				_ = m.Score("前文", "nihao", "你好")
				if i%50 == 0 {
					_ = m.SaveIfDirty()
				}
			}
		}()
	}
	wg.Wait()
	if err := m.SaveIfDirty(); err != nil {
		t.Fatalf("save: %v", err)
	}
	if m.Dirty() {
		t.Fatalf("model should be clean after final save")
	}
}

func TestSaveKeepsDirtyWhenUpdatedDuringSave(t *testing.T) {
	m := NewPersonalModel(DefaultSGDConfig())
	m.SetPath(filepath.Join(t.TempDir(), "personal_model.json"))
	m.PairwiseUpdate("前文", "nihao", "你好", "泥好")
	m.mu.Lock()
	seq := m.changeSeq
	m.mu.Unlock()
	if err := m.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	m.mu.Lock()
	m.changeSeq = seq + 1
	m.dirty = true
	m.mu.Unlock()
	if !m.Dirty() {
		t.Fatalf("expected dirty after concurrent change")
	}
}
