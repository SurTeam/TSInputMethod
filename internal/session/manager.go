package session

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/SurTeam/TSInputMethod/internal/engine"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

type Manager struct {
	mu             sync.RWMutex
	sessions       map[string]*Session
	nextID         atomic.Uint64
	engine         engine.Engine
	phraseObserver PhraseObserver
	feedbackSink   FeedbackSink
	userDictVer    uint64
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

func (m *Manager) SetPhraseObserver(observer PhraseObserver) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.phraseObserver = observer
	for _, sess := range m.sessions {
		sess.SetPhraseObserver(observer)
	}
}

func (m *Manager) SetUserDictVersion(version uint64) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userDictVer = version
	for _, sess := range m.sessions {
		sess.SetUserDictVersion(version)
	}
}

func (m *Manager) SetFeedbackSink(sink FeedbackSink) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.feedbackSink = sink
	for _, sess := range m.sessions {
		sess.SetFeedbackSink(sink)
	}
}

func (m *Manager) Open(initialMode imev1.InputMode) *Session {
	id := fmt.Sprintf("s-%d", m.nextID.Add(1))
	sess := newSession(id, initialMode, m.engine)
	m.mu.Lock()
	sess.SetPhraseObserver(m.phraseObserver)
	sess.SetFeedbackSink(m.feedbackSink)
	sess.SetUserDictVersion(m.userDictVer)
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
	sess, ok := m.sessions[id]
	if !ok {
		return false
	}
	sess.ClearLearningContext()
	delete(m.sessions, id)
	return true
}

func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

func (m *Manager) ClearAllLearningContext() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, sess := range m.sessions {
		sess.ClearLearningContext()
	}
}
