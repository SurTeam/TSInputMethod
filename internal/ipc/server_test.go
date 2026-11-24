package ipc

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/session"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func startTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "tsime.sock")
	manager := session.NewManager(engine.NewStubEngine())
	server, err := ListenUnix(socketPath, manager)
	if err != nil {
		t.Fatalf("ListenUnix: %v", err)
	}
	go func() {
		_ = server.Serve()
	}()
	t.Cleanup(func() {
		_ = server.Close()
	})
	time.Sleep(20 * time.Millisecond)
	return server, socketPath
}

func TestHelloOpenSessionKeyEvent(t *testing.T) {
	_, socketPath := startTestServer(t)
	client, err := DialUnix(socketPath)
	if err != nil {
		t.Fatalf("DialUnix: %v", err)
	}
	defer client.Close()

	hello, err := client.RoundTrip(&imev1.Envelope{
		RequestId: 1,
		Payload: &imev1.Envelope_Hello{
			Hello: &imev1.Hello{
				ClientVersion: &imev1.ProtocolVersion{Major: 1, Minor: 0},
				ClientName:    "test",
			},
		},
	})
	if err != nil {
		t.Fatalf("hello: %v", err)
	}
	ack := hello.GetHelloAck()
	if ack == nil || !ack.Accepted {
		t.Fatalf("hello ack = %#v", hello)
	}

	open, err := client.RoundTrip(&imev1.Envelope{
		RequestId: 2,
		Payload: &imev1.Envelope_OpenSession{
			OpenSession: &imev1.OpenSession{InitialMode: imev1.InputMode_INPUT_MODE_CHINESE},
		},
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	sessionID := open.GetOpenSessionAck().GetSessionId()
	if sessionID == "" {
		t.Fatalf("missing session id: %#v", open)
	}

	key, err := client.RoundTrip(&imev1.Envelope{
		RequestId: 3,
		Payload: &imev1.Envelope_KeyEvent{
			KeyEvent: &imev1.KeyEvent{
				SessionId:  sessionID,
				Characters: "ni",
				KeyDown:    true,
			},
		},
	})
	if err != nil {
		t.Fatalf("key: %v", err)
	}
	result := key.GetKeyEventResult()
	if result == nil || result.Composition == nil || result.Composition.Preedit != "ni" {
		t.Fatalf("key result = %#v", key)
	}
	if result.Candidates == nil || len(result.Candidates.Candidates) == 0 {
		t.Fatalf("expected stub candidates, got %#v", result.Candidates)
	}
}

func TestInvalidateLearningContextRPC(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "i.sock")
	manager := session.NewManager(engine.NewStubEngine())
	server, err := ListenUnix(socketPath, manager)
	if err != nil {
		t.Fatalf("ListenUnix: %v", err)
	}
	go func() { _ = server.Serve() }()
	t.Cleanup(func() { _ = server.Close() })
	time.Sleep(20 * time.Millisecond)

	client, err := DialUnix(socketPath)
	if err != nil {
		t.Fatalf("DialUnix: %v", err)
	}
	defer client.Close()

	_, _ = client.RoundTrip(&imev1.Envelope{
		RequestId: 1,
		Payload: &imev1.Envelope_Hello{
			Hello: &imev1.Hello{ClientVersion: &imev1.ProtocolVersion{Major: 1, Minor: 0}},
		},
	})
	open, err := client.RoundTrip(&imev1.Envelope{
		RequestId: 2,
		Payload: &imev1.Envelope_OpenSession{
			OpenSession: &imev1.OpenSession{InitialMode: imev1.InputMode_INPUT_MODE_CHINESE},
		},
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	sessionID := open.GetOpenSessionAck().GetSessionId()
	sess, ok := server.manager.Get(sessionID)
	if !ok {
		t.Fatalf("session missing")
	}
	_ = sess.HandleKey(&imev1.KeyEvent{SessionId: sessionID, Characters: "ni", KeyDown: true})
	_ = sess.HandleKey(&imev1.KeyEvent{SessionId: sessionID, KeyCode: 49, KeyDown: true})
	if sess.CommittedContext() == "" && sess.EffectiveContext() == "" {
		t.Fatalf("expected stub commit to seed learning context")
	}
	resp, err := client.RoundTrip(&imev1.Envelope{
		RequestId: 3,
		Payload: &imev1.Envelope_InvalidateLearningContext{
			InvalidateLearningContext: &imev1.InvalidateLearningContext{SessionId: sessionID},
		},
	})
	if err != nil {
		t.Fatalf("invalidate: %v", err)
	}
	status := resp.GetStatus()
	if status == nil || !status.Ok {
		t.Fatalf("invalidate status = %#v", resp)
	}
	if sess.CommittedContext() != "" || sess.EffectiveContext() != "" {
		t.Fatalf("context remained committed=%q effective=%q", sess.CommittedContext(), sess.EffectiveContext())
	}
}

func TestStaleCandidateGeneration(t *testing.T) {
	_, socketPath := startTestServer(t)
	client, err := DialUnix(socketPath)
	if err != nil {
		t.Fatalf("DialUnix: %v", err)
	}
	defer client.Close()

	_, _ = client.RoundTrip(&imev1.Envelope{
		RequestId: 1,
		Payload: &imev1.Envelope_Hello{
			Hello: &imev1.Hello{ClientVersion: &imev1.ProtocolVersion{Major: 1, Minor: 0}},
		},
	})
	open, _ := client.RoundTrip(&imev1.Envelope{
		RequestId: 2,
		Payload: &imev1.Envelope_OpenSession{
			OpenSession: &imev1.OpenSession{InitialMode: imev1.InputMode_INPUT_MODE_CHINESE},
		},
	})
	sessionID := open.GetOpenSessionAck().GetSessionId()
	key, _ := client.RoundTrip(&imev1.Envelope{
		RequestId: 3,
		Payload: &imev1.Envelope_KeyEvent{
			KeyEvent: &imev1.KeyEvent{SessionId: sessionID, Characters: "wo", KeyDown: true},
		},
	})
	gen := key.GetKeyEventResult().GetComposition().GetGeneration()
	if gen < 2 {
		key, _ = client.RoundTrip(&imev1.Envelope{
			RequestId: 4,
			Payload: &imev1.Envelope_KeyEvent{
				KeyEvent: &imev1.KeyEvent{SessionId: sessionID, Characters: "men", KeyDown: true},
			},
		})
		gen = key.GetKeyEventResult().GetComposition().GetGeneration()
	}
	stale, err := client.RoundTrip(&imev1.Envelope{
		RequestId: 5,
		Payload: &imev1.Envelope_SelectCandidate{
			SelectCandidate: &imev1.SelectCandidate{
				SessionId:  sessionID,
				Generation: 1,
				Index:      0,
			},
		},
	})
	_ = gen
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if stale.GetStatus() == nil || stale.GetStatus().Message != "stale candidate generation" {
		t.Fatalf("expected stale status, got %#v", stale)
	}
}
