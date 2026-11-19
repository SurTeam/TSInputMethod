package inference

import (
	"context"
	"errors"
	"testing"
)

func TestNoopBackendPreservesOrder(t *testing.T) {
	req := Request{
		Candidates: []Candidate{
			{Text: "你", Score: 1},
			{Text: "泥", Score: 2},
		},
	}
	out, err := (NoopBackend{}).Rank(context.Background(), req)
	if err != nil {
		t.Fatalf("noop: %v", err)
	}
	if len(out) != 2 || out[0].Text != "你" || out[1].Text != "泥" {
		t.Fatalf("unexpected %#v", out)
	}
}

func TestUnavailableBackendFails(t *testing.T) {
	_, err := (UnavailableBackend{}).Rank(context.Background(), Request{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
}

func TestFuncBackendNilIsUnavailable(t *testing.T) {
	var backend FuncBackend
	_, err := backend.Rank(context.Background(), Request{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
}
