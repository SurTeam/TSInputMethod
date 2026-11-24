package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/engine"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func TestManagerClearAllLearningContext(t *testing.T) {
	m := NewManager(engine.NewStubEngine())
	a := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	b := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	a.mu.Lock()
	a.appendCommittedLocked("甲乙")
	a.mu.Unlock()
	b.mu.Lock()
	b.appendCommittedLocked("丙丁")
	b.mu.Unlock()
	if a.CommittedContext() == "" || b.CommittedContext() == "" {
		t.Fatalf("expected seeded context")
	}
	m.ClearAllLearningContext()
	if a.CommittedContext() != "" {
		t.Fatalf("session a still has context=%q", a.CommittedContext())
	}
	if b.CommittedContext() != "" {
		t.Fatalf("session b still has context=%q", b.CommittedContext())
	}
}

func TestInvalidateLearningContextClearsEffectiveContext(t *testing.T) {
	m := NewManager(engine.NewStubEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	s.mu.Lock()
	s.appendCommittedLocked("我明天要去")
	s.mu.Unlock()
	if s.EffectiveContext() == "" {
		t.Fatalf("expected seeded effective context")
	}
	s.InvalidateLearningContext()
	if s.CommittedContext() != "" {
		t.Fatalf("committed context after invalidate=%q", s.CommittedContext())
	}
	if s.EffectiveContext() != "" {
		t.Fatalf("effective context after invalidate=%q", s.EffectiveContext())
	}
}

func TestSetModeClearsLearningContext(t *testing.T) {
	m := NewManager(engine.NewStubEngine())
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	s.mu.Lock()
	s.appendCommittedLocked("这张图片的")
	s.mu.Unlock()
	if s.EffectiveContext() == "" {
		t.Fatalf("expected seeded effective context")
	}
	_ = s.SetMode(imev1.InputMode_INPUT_MODE_ENGLISH)
	if s.CommittedContext() != "" {
		t.Fatalf("committed context after SetMode=%q", s.CommittedContext())
	}
	if s.EffectiveContext() != "" {
		t.Fatalf("effective context after SetMode=%q", s.EffectiveContext())
	}
	_ = s.SetMode(imev1.InputMode_INPUT_MODE_CHINESE)
	if s.EffectiveContext() != "" {
		t.Fatalf("effective context after return to Chinese=%q", s.EffectiveContext())
	}
}
