package inference

import (
	"context"
	"errors"
)

var ErrUnavailable = errors.New("inference backend unavailable")

type Candidate struct {
	Text    string
	Comment string
	Score   int32
}

type Request struct {
	Pinyin     string
	Candidates []Candidate
	Context    string
	SessionID  string
	Generation uint64
}

type Backend interface {
	Rank(ctx context.Context, req Request) ([]Candidate, error)
}

type NoopBackend struct{}

func (NoopBackend) Rank(_ context.Context, req Request) ([]Candidate, error) {
	out := make([]Candidate, len(req.Candidates))
	copy(out, req.Candidates)
	return out, nil
}

type UnavailableBackend struct{}

func (UnavailableBackend) Rank(_ context.Context, _ Request) ([]Candidate, error) {
	return nil, ErrUnavailable
}

type FuncBackend func(ctx context.Context, req Request) ([]Candidate, error)

func (f FuncBackend) Rank(ctx context.Context, req Request) ([]Candidate, error) {
	if f == nil {
		return nil, ErrUnavailable
	}
	return f(ctx, req)
}
