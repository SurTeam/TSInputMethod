package ipc

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func TestLiveDaemonNihcCandidates(t *testing.T) {
	sock := filepath.Join(os.TempDir(), fmt.Sprintf("tsime-%d.sock", os.Getuid()))
	if _, err := os.Stat(sock); err != nil {
		t.Skip(err)
	}
	c, err := DialUnix(sock)
	if err != nil {
		t.Skip(err)
	}
	defer c.Close()
	if _, err := c.RoundTrip(&imev1.Envelope{Payload: &imev1.Envelope_Hello{Hello: &imev1.Hello{ClientVersion: &imev1.ProtocolVersion{Major: 1, Minor: 0}}}}); err != nil {
		t.Fatal(err)
	}
	open, err := c.RoundTrip(&imev1.Envelope{Payload: &imev1.Envelope_OpenSession{OpenSession: &imev1.OpenSession{InitialMode: imev1.InputMode_INPUT_MODE_CHINESE}}})
	if err != nil {
		t.Fatal(err)
	}
	sid := open.GetOpenSessionAck().GetSessionId()
	var last *imev1.Envelope
	for _, ch := range []string{"n", "i", "h", "c"} {
		last, err = c.RoundTrip(&imev1.Envelope{Payload: &imev1.Envelope_KeyEvent{KeyEvent: &imev1.KeyEvent{SessionId: sid, KeyDown: true, Characters: ch}}})
		if err != nil {
			t.Fatal(err)
		}
	}
	res := last.GetKeyEventResult()
	if res == nil || res.GetCandidates() == nil {
		t.Fatalf("nil result %#v", last)
	}
	cands := res.GetCandidates()
	if cands.PageSize != 9 || len(cands.Candidates) != 9 {
		t.Fatalf("pageSize=%d n=%d", cands.PageSize, len(cands.Candidates))
	}
	texts := make([]string, 0, len(cands.Candidates))
	for _, cand := range cands.Candidates {
		text := cand.Text
		if len(text) >= 2 && text[0] >= '1' && text[0] <= '9' && text[1] == '.' {
			text = text[2:]
		}
		texts = append(texts, text)
		t.Logf("%s (%s)", cand.Text, cand.Comment)
	}
	if texts[0] != "你好" {
		t.Fatalf("top=%v", texts)
	}
	foundNi := false
	for _, text := range texts {
		if text == "你" {
			foundNi = true
		}
		if text == "你好是接" || text == "尼好" || text == "拟好" {
			t.Fatalf("bad first-page candidate %q in %v", text, texts)
		}
	}
	if !foundNi {
		t.Fatalf("missing 你 on first page: %v", texts)
	}
}
