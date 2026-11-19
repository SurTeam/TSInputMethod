package ranking

import (
	"context"
	"sync"
	"time"

	"github.com/SurTeam/TSInputMethod/internal/inference"
)

type Update struct {
	Generation uint64
	Candidates []Candidate
}

type Coordinator struct {
	mu       sync.Mutex
	ranker   Ranker
	backend  inference.Backend
	timeout  time.Duration
	onUpdate func(Update)
}

func NewCoordinator(ranker Ranker, backend inference.Backend) *Coordinator {
	if ranker == nil {
		ranker = NewFastRanker()
	}
	return &Coordinator{
		ranker:  ranker,
		backend: backend,
		timeout: 50 * time.Millisecond,
	}
}

func (c *Coordinator) SetBackend(backend inference.Backend) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.backend = backend
}

func (c *Coordinator) SetTimeout(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if d > 0 {
		c.timeout = d
	}
}

func (c *Coordinator) SetUpdateHandler(handler func(Update)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onUpdate = handler
}

func (c *Coordinator) RankSync(req Request) []Candidate {
	c.mu.Lock()
	ranker := c.ranker
	c.mu.Unlock()
	if ranker == nil {
		ranker = NewFastRanker()
	}
	return ranker.Rank(req)
}

func (c *Coordinator) Rank(req Request) []Candidate {
	base := c.RankSync(req)
	c.mu.Lock()
	backend := c.backend
	timeout := c.timeout
	onUpdate := c.onUpdate
	c.mu.Unlock()
	if backend == nil || onUpdate == nil {
		return base
	}
	go c.rerankAsync(backend, timeout, req, onUpdate)
	return base
}

func (c *Coordinator) RankWithContext(ctx context.Context, req Request) []Candidate {
	base := c.RankSync(req)
	c.mu.Lock()
	backend := c.backend
	timeout := c.timeout
	c.mu.Unlock()
	if backend == nil {
		return base
	}
	if ctx == nil {
		ctx = context.Background()
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ranked, err := backend.Rank(runCtx, toInferenceRequest(req))
	if err != nil || len(ranked) == 0 {
		return base
	}
	return fromInference(ranked)
}

func (c *Coordinator) rerankAsync(backend inference.Backend, timeout time.Duration, req Request, onUpdate func(Update)) {
	runCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ranked, err := backend.Rank(runCtx, toInferenceRequest(req))
	if err != nil || len(ranked) == 0 {
		return
	}
	onUpdate(Update{
		Generation: req.Generation,
		Candidates: fromInference(ranked),
	})
}

func ApplyUpdate(currentGeneration uint64, update Update) ([]Candidate, bool) {
	if update.Generation != currentGeneration {
		return nil, false
	}
	if len(update.Candidates) == 0 {
		return nil, false
	}
	out := make([]Candidate, len(update.Candidates))
	copy(out, update.Candidates)
	return out, true
}

func toInferenceRequest(req Request) inference.Request {
	cands := make([]inference.Candidate, len(req.Candidates))
	for i, item := range req.Candidates {
		cands[i] = inference.Candidate{
			Text:    item.Text,
			Comment: item.Comment,
			Score:   item.Score,
		}
	}
	return inference.Request{
		Pinyin:     req.Pinyin,
		Candidates: cands,
		Context:    req.Context,
		SessionID:  req.SessionID,
		Generation: req.Generation,
	}
}

func fromInference(in []inference.Candidate) []Candidate {
	out := make([]Candidate, len(in))
	for i, item := range in {
		out[i] = Candidate{
			Text:    item.Text,
			Comment: item.Comment,
			Score:   item.Score,
		}
	}
	return out
}
