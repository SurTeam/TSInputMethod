package ipc

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/session"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func helloAndOpen(t *testing.T, client *Client) string {
	t.Helper()
	if _, err := client.RoundTrip(&imev1.Envelope{
		RequestId: 1,
		Payload: &imev1.Envelope_Hello{Hello: &imev1.Hello{
			ClientVersion: &imev1.ProtocolVersion{Major: 1},
			ClientName:    "lifecycle",
		}},
	}); err != nil {
		t.Fatalf("hello: %v", err)
	}
	open, err := client.RoundTrip(&imev1.Envelope{
		RequestId: 2,
		Payload: &imev1.Envelope_OpenSession{OpenSession: &imev1.OpenSession{
			InitialMode: imev1.InputMode_INPUT_MODE_CHINESE,
		}},
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	return open.GetOpenSessionAck().GetSessionId()
}

func shortSocketPath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "tsl")
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return filepath.Join(dir, "s.sock")
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("condition not reached")
}

func TestDisconnectClosesOwnedSessions(t *testing.T) {
	socketPath := shortSocketPath(t)
	manager := session.NewManager(engine.NewStubEngine())
	server, err := ListenUnix(socketPath, manager)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() { _ = server.Serve() }()
	defer server.Close()

	client, err := DialUnix(socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	if id := helloAndOpen(t, client); id == "" {
		t.Fatalf("missing session id")
	}
	if manager.Count() != 1 {
		t.Fatalf("expected 1 session, got %d", manager.Count())
	}
	_ = client.Close()
	waitFor(t, func() bool { return manager.Count() == 0 && server.ConnCount() == 0 })
}

func TestCloseDoesNotHangOnIdleConnections(t *testing.T) {
	socketPath := shortSocketPath(t)
	manager := session.NewManager(engine.NewStubEngine())
	server, err := ListenUnix(socketPath, manager)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() { _ = server.Serve() }()
	client, err := DialUnix(socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()
	helloAndOpen(t, client)
	done := make(chan struct{})
	go func() {
		_ = server.Close()
		_ = server.Close()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("server close blocked on idle connection")
	}
	if manager.Count() != 0 {
		t.Fatalf("shutdown should close sessions, got %d", manager.Count())
	}
}
