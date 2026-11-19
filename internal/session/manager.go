package session

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/SurTeam/TSInputMethod/internal/engine"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	nextID   atomic.Uint64
	engine   engine.Engine
}

func NewManager(eng engine.Engine) *Manager {
	if eng == nil {
		eng = engine.NewStubEngine()
	}
	return &Manager{
		sessions: make(map[string]*Session),
		engine:   eng,
	}
}

func (m *Manager) Open(initialMode imev1.InputMode) *Session {
	id := fmt.Sprintf("s-%d", m.nextID.Add(1))
	sess := newSession(id, initialMode, m.engine)
	m.mu.Lock()
	m.sessions[id] = sess
	m.mu.Unlock()
	return sess
}

func (m *Manager) Get(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sess, ok := m.sessions[id]
	return sess, ok
}

func (m *Manager) Close(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sessions[id]; !ok {
		return false
	}
	delete(m.sessions, id)
	return true
}

func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}
