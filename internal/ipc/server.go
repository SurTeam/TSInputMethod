package ipc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/SurTeam/TSInputMethod/internal/session"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

const (
	protocolMajor = 1
	protocolMinor = 0
)

type Server struct {
	listener  net.Listener
	manager   *session.Manager
	wg        sync.WaitGroup
	quit      chan struct{}
	closeOnce sync.Once
	connMu    sync.Mutex
	conns     map[*Conn]struct{}
}

type connState struct {
	negotiated bool
	sessions   map[string]struct{}
}

func ListenUnix(socketPath string, manager *session.Manager) (*Server, error) {
	if err := os.MkdirAll(filepath.Dir(socketPath), 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		_ = err
	}
	_ = os.Remove(socketPath)
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(socketPath, 0o600); err != nil {
		_ = ln.Close()
		return nil, err
	}
	return newServer(ln, manager), nil
}

func newServer(ln net.Listener, manager *session.Manager) *Server {
	return &Server{
		listener: ln,
		manager:  manager,
		quit:     make(chan struct{}),
		conns:    make(map[*Conn]struct{}),
	}
}

func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

func (s *Server) Serve() error {
	for {
		raw, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return nil
			default:
				return err
			}
		}
		conn := NewConn(raw)
		if !s.trackConn(conn) {
			_ = conn.Close()
			continue
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer s.untrackConn(conn)
			s.handleConn(conn)
		}()
	}
}

func (s *Server) Close() error {
	var err error
	s.closeOnce.Do(func() {
		s.connMu.Lock()
		close(s.quit)
		conns := make([]*Conn, 0, len(s.conns))
		for c := range s.conns {
			conns = append(conns, c)
		}
		s.connMu.Unlock()
		err = s.listener.Close()
		for _, c := range conns {
			_ = c.Close()
		}
		s.wg.Wait()
	})
	return err
}

func (s *Server) ConnCount() int {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	return len(s.conns)
}

func (s *Server) trackConn(conn *Conn) bool {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	select {
	case <-s.quit:
		return false
	default:
	}
	s.conns[conn] = struct{}{}
	return true
}

func (s *Server) untrackConn(conn *Conn) {
	s.connMu.Lock()
	delete(s.conns, conn)
	s.connMu.Unlock()
}

func (s *Server) handleConn(conn *Conn) {
	state := &connState{sessions: make(map[string]struct{})}
	defer func() {
		for id := range state.sessions {
			s.manager.Close(id)
		}
		_ = conn.Close()
	}()
	for {
		req, err := conn.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			return
		}
		resp := s.dispatch(req, state)
		if resp == nil {
			continue
		}
		resp.RequestId = req.RequestId
		resp.SequenceNumber = req.SequenceNumber
		if err := conn.Send(resp); err != nil {
			return
		}
	}
}

func (s *Server) dispatch(req *imev1.Envelope, state *connState) *imev1.Envelope {
	if hello, ok := req.Payload.(*imev1.Envelope_Hello); ok {
		return s.handleHello(hello.Hello, &state.negotiated)
	}
	if !state.negotiated {
		return errorEnvelope(1, "hello required")
	}
	switch payload := req.Payload.(type) {
	case *imev1.Envelope_OpenSession:
		resp := s.handleOpenSession(payload.OpenSession)
		if id := resp.GetOpenSessionAck().GetSessionId(); id != "" {
			state.sessions[id] = struct{}{}
		}
		return resp
	case *imev1.Envelope_CloseSession:
		delete(state.sessions, payload.CloseSession.GetSessionId())
		return s.handleCloseSession(payload.CloseSession)
	case *imev1.Envelope_InvalidateLearningContext:
		return s.handleInvalidateLearningContext(payload.InvalidateLearningContext)
	case *imev1.Envelope_KeyEvent:
		return s.handleKeyEvent(payload.KeyEvent, req.Generation)
	case *imev1.Envelope_SetInputMode:
		return s.handleSetInputMode(payload.SetInputMode)
	case *imev1.Envelope_GetInputMode:
		return s.handleGetInputMode(payload.GetInputMode)
	case *imev1.Envelope_SelectCandidate:
		return s.handleSelectCandidate(payload.SelectCandidate)
	default:
		return errorEnvelope(2, fmt.Sprintf("unsupported payload %T", req.Payload))
	}
}

func (s *Server) handleHello(hello *imev1.Hello, negotiated *bool) *imev1.Envelope {
	accepted := true
	message := "ok"
	if hello.GetClientVersion() != nil && hello.ClientVersion.Major != protocolMajor {
		accepted = false
		message = "incompatible major version"
	}
	if accepted {
		*negotiated = true
	}
	return &imev1.Envelope{
		Payload: &imev1.Envelope_HelloAck{
			HelloAck: &imev1.HelloAck{
				ServerVersion: &imev1.ProtocolVersion{Major: protocolMajor, Minor: protocolMinor},
				Accepted:      accepted,
				Message:       message,
			},
		},
	}
}

func (s *Server) handleOpenSession(req *imev1.OpenSession) *imev1.Envelope {
	sess := s.manager.Open(req.GetInitialMode())
	return &imev1.Envelope{
		Payload: &imev1.Envelope_OpenSessionAck{
			OpenSessionAck: &imev1.OpenSessionAck{
				SessionId: sess.ID(),
				Mode:      sess.Mode(),
			},
		},
	}
}

func (s *Server) handleCloseSession(req *imev1.CloseSession) *imev1.Envelope {
	ok := s.manager.Close(req.GetSessionId())
	return &imev1.Envelope{
		Payload: &imev1.Envelope_Status{
			Status: &imev1.Status{Ok: ok, Message: ternary(ok, "closed", "session not found")},
		},
	}
}

func (s *Server) handleInvalidateLearningContext(req *imev1.InvalidateLearningContext) *imev1.Envelope {
	sess, ok := s.manager.Get(req.GetSessionId())
	if !ok {
		return &imev1.Envelope{
			Payload: &imev1.Envelope_Status{
				Status: &imev1.Status{Ok: false, Message: "session not found"},
			},
		}
	}
	sess.InvalidateLearningContext()
	return &imev1.Envelope{
		Payload: &imev1.Envelope_Status{
			Status: &imev1.Status{Ok: true, Message: "learning context invalidated"},
		},
	}
}

func (s *Server) handleKeyEvent(ev *imev1.KeyEvent, envelopeGeneration uint64) *imev1.Envelope {
	sess, ok := s.manager.Get(ev.GetSessionId())
	if !ok {
		return errorEnvelope(3, "session not found")
	}
	result := sess.HandleKey(ev)
	if envelopeGeneration != 0 && result.Composition != nil && result.Composition.Generation < envelopeGeneration {
		return &imev1.Envelope{
			Generation: result.Composition.Generation,
			Payload: &imev1.Envelope_Status{
				Status: &imev1.Status{Ok: true, Message: "stale generation ignored"},
			},
		}
	}
	return &imev1.Envelope{
		Generation: valueOrZero(result),
		Payload: &imev1.Envelope_KeyEventResult{
			KeyEventResult: &imev1.KeyEventResult{
				Composition: result.Composition,
				Candidates:  result.Candidates,
				Commit:      result.Commit,
				Consumed:    result.Consumed,
			},
		},
	}
}

func (s *Server) handleSetInputMode(req *imev1.SetInputMode) *imev1.Envelope {
	sess, ok := s.manager.Get(req.GetSessionId())
	if !ok {
		return errorEnvelope(3, "session not found")
	}
	sess.SetMode(req.GetMode())
	return &imev1.Envelope{
		Payload: &imev1.Envelope_InputModeState{
			InputModeState: &imev1.InputModeState{
				SessionId: sess.ID(),
				Mode:      sess.Mode(),
			},
		},
	}
}

func (s *Server) handleGetInputMode(req *imev1.GetInputMode) *imev1.Envelope {
	sess, ok := s.manager.Get(req.GetSessionId())
	if !ok {
		return errorEnvelope(3, "session not found")
	}
	return &imev1.Envelope{
		Payload: &imev1.Envelope_InputModeState{
			InputModeState: &imev1.InputModeState{
				SessionId: sess.ID(),
				Mode:      sess.Mode(),
			},
		},
	}
}

func (s *Server) handleSelectCandidate(req *imev1.SelectCandidate) *imev1.Envelope {
	sess, ok := s.manager.Get(req.GetSessionId())
	if !ok {
		return errorEnvelope(3, "session not found")
	}
	result, accepted := sess.SelectCandidate(req.GetGeneration(), req.GetIndex())
	if !accepted {
		return &imev1.Envelope{
			Payload: &imev1.Envelope_Status{
				Status: &imev1.Status{Ok: true, Message: "stale candidate generation"},
			},
		}
	}
	return &imev1.Envelope{
		Generation: valueOrZero(result),
		Payload: &imev1.Envelope_KeyEventResult{
			KeyEventResult: &imev1.KeyEventResult{
				Composition: result.Composition,
				Candidates:  result.Candidates,
				Commit:      result.Commit,
				Consumed:    result.Consumed,
			},
		},
	}
}

func errorEnvelope(code int32, message string) *imev1.Envelope {
	return &imev1.Envelope{
		Payload: &imev1.Envelope_Error{
			Error: &imev1.Error{Code: code, Message: message},
		},
	}
}

func ternary(ok bool, a, b string) string {
	if ok {
		return a
	}
	return b
}

func valueOrZero(result *session.KeyResult) uint64 {
	if result == nil {
		return 0
	}
	if result.Composition != nil {
		return result.Composition.Generation
	}
	if result.Commit != nil {
		return result.Commit.Generation
	}
	return 0
}
