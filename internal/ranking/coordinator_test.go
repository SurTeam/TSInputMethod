package ranking

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/SurTeam/TSInputMethod/internal/inference"
)

func sampleCandidates() []Candidate {
	return []Candidate{
		{Text: "泥", Score: 10},
		{Text: "你", Score: 100},
		{Text: "你好", Score: 1000},
	}
}

func TestFastRankerOrdersByScore(t *testing.T) {
	ranker := NewFastRanker()
	out := ranker.Rank(Request{Candidates: sampleCandidates()})
	if len(out) != 3 || out[0].Text != "你好" || out[1].Text != "你" {
		t.Fatalf("unexpected order %#v", out)
	}
}

func TestCoordinatorNilBackendReturnsBase(t *testing.T) {
	c := NewCoordinator(NewFastRanker(), nil)
	out := c.Rank(Request{
		Pinyin:     "ni",
		Candidates: sampleCandidates(),
		Generation: 1,
	})
	if len(out) == 0 || out[0].Text != "你好" {
		t.Fatalf("base ranking failed: %#v", out)
	}
}

func TestCoordinatorUnavailableBackendDegrades(t *testing.T) {
	c := NewCoordinator(NewFastRanker(), inference.UnavailableBackend{})
	out := c.RankWithContext(context.Background(), Request{
		Pinyin:     "ni",
		Candidates: sampleCandidates(),
		Generation: 2,
	})
	if len(out) == 0 || out[0].Text != "你好" {
		t.Fatalf("expected frequency base after backend failure, got %#v", out)
	}
}

func TestCoordinatorInjectedBackendReorders(t *testing.T) {
	backend := inference.FuncBackend(func(_ context.Context, req inference.Request) ([]inference.Candidate, error) {
		out := make([]inference.Candidate, len(req.Candidates))
		copy(out, req.Candidates)
		if len(out) >= 2 {
			out[0], out[1] = out[1], out[0]
		}
		return out, nil
	})
	c := NewCoordinator(NewFastRanker(), backend)
	base := sampleCandidates()
	out := c.RankWithContext(context.Background(), Request{
		Pinyin:     "ni",
		Candidates: base,
		Generation: 3,
	})
	if len(out) < 2 || out[0].Text != "你" {
		t.Fatalf("injected backend did not reorder: %#v", out)
	}
}

func TestApplyUpdateDropsStaleGeneration(t *testing.T) {
	update := Update{
		Generation: 1,
		Candidates: []Candidate{{Text: "旧", Score: 1}},
	}
	if _, ok := ApplyUpdate(2, update); ok {
		t.Fatal("stale generation must be discarded")
	}
	applied, ok := ApplyUpdate(1, update)
	if !ok || len(applied) != 1 || applied[0].Text != "旧" {
		t.Fatalf("current generation apply failed: ok=%v %#v", ok, applied)
	}
}

func TestAsyncRerankDropsStaleUpdate(t *testing.T) {
	var (
		mu       sync.Mutex
		releases []Update
	)
	started := make(chan struct{})
	backend := inference.FuncBackend(func(ctx context.Context, req inference.Request) ([]inference.Candidate, error) {
		close(started)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
		return []inference.Candidate{{Text: "异步", Score: 9999}}, nil
	})
	c := NewCoordinator(NewFastRanker(), backend)
	c.SetTimeout(200 * time.Millisecond)
	c.SetUpdateHandler(func(u Update) {
		mu.Lock()
		defer mu.Unlock()
		releases = append(releases, u)
	})
	base := c.Rank(Request{
		Pinyin:     "ni",
		Candidates: sampleCandidates(),
		Generation: 7,
	})
	if len(base) == 0 || base[0].Text != "你好" {
		t.Fatalf("sync base missing: %#v", base)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("async backend did not start")
	}
	deadline := time.After(time.Second)
	for {
		mu.Lock()
		n := len(releases)
		var latest Update
		if n > 0 {
			latest = releases[n-1]
		}
		mu.Unlock()
		if n > 0 {
			if _, ok := ApplyUpdate(8, latest); ok {
				t.Fatalf("stale async update incorrectly applied: %#v", latest)
			}
			if applied, ok := ApplyUpdate(7, latest); !ok || applied[0].Text != "异步" {
				t.Fatalf("matching generation apply failed: %#v", latest)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for async update")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}
