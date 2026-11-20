package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/engine"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

type fakeDecodeEngine struct {
	calls int
	lists [][]engine.Candidate
}

func (e *fakeDecodeEngine) Decode(input string) ([]engine.Candidate, error) {
	return e.DecodeWithContext(input, "")
}

func (e *fakeDecodeEngine) DecodeWithContext(input string, _ string) ([]engine.Candidate, error) {
	if input == "" {
		return nil, nil
	}
	e.calls++
	if len(e.lists) == 0 {
		return nil, nil
	}
	idx := e.calls - 1
	if idx >= len(e.lists) {
		idx = len(e.lists) - 1
	}
	return cloneCandidates(e.lists[idx]), nil
}

func TestSnapshotBindsSelectAgainstReDecode(t *testing.T) {
	eng := &fakeDecodeEngine{lists: [][]engine.Candidate{
		{
			{Text: "背景", Score: 100, ConsumedKeys: 4, Syllables: []string{"bei", "jing"}},
			{Text: "北京", Score: 90, ConsumedKeys: 4, Syllables: []string{"bei", "jing"}},
		},
		{
			{Text: "北京", Score: 100, ConsumedKeys: 4, Syllables: []string{"bei", "jing"}},
			{Text: "背景", Score: 90, ConsumedKeys: 4, Syllables: []string{"bei", "jing"}},
		},
	}}
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
	first := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "bwjk"})
	if first.Candidates == nil || bareCandidateText(first.Candidates.Candidates[0].Text) != "背景" {
		t.Fatalf("display top=%#v", first.Candidates)
	}
	shownGen := first.Candidates.Generation
	texts := s.SnapshotCandidateTexts()
	if len(texts) < 2 || texts[0] != "背景" || texts[1] != "北京" {
		t.Fatalf("snapshot texts=%#v", texts)
	}
	res, ok := s.SelectCandidate(shownGen, 1)
	if !ok || res.Commit == nil || res.Commit.Text != "北京" {
		t.Fatalf("select second got ok=%v res=%#v", ok, res)
	}
	if eng.calls != 1 {
		t.Fatalf("select should reuse snapshot without re-decode, calls=%d", eng.calls)
	}
}

func TestPendingFeedbackNonTopAcceptOnFinalCommit(t *testing.T) {
	eng := &fakeDecodeEngine{lists: [][]engine.Candidate{{
		{Text: "背景", Score: 100, ConsumedKeys: 4, Syllables: []string{"bei", "jing"}},
		{Text: "北京", Score: 90, ConsumedKeys: 4, Syllables: []string{"bei", "jing"}},
	}}}
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "bwjk"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "2"})
	if res.Commit == nil || res.Commit.Text != "北京" {
		t.Fatalf("commit=%#v", res.Commit)
	}
	if s.PendingFeedbackCount() != 0 {
		t.Fatalf("pending should accept on final commit, pending=%d", s.PendingFeedbackCount())
	}
	events := s.TakeAcceptedFeedback()
	if len(events) != 1 {
		t.Fatalf("accepted=%#v", events)
	}
	if events[0].Selected.Text != "北京" || events[0].Top.Text != "背景" || events[0].SelectedIndex != 1 {
		t.Fatalf("event=%#v", events[0])
	}
	if events[0].InputCode != "bwjk" {
		t.Fatalf("input code=%q", events[0].InputCode)
	}
}

func TestTopSelectionSkipsPairwisePending(t *testing.T) {
	eng := &fakeDecodeEngine{lists: [][]engine.Candidate{{
		{Text: "背景", Score: 100, ConsumedKeys: 4},
		{Text: "北京", Score: 90, ConsumedKeys: 4},
	}}}
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "bwjk"})
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "1"})
	if res.Commit == nil || res.Commit.Text != "背景" {
		t.Fatalf("commit=%#v", res.Commit)
	}
	if got := s.TakeAcceptedFeedback(); len(got) != 0 {
		t.Fatalf("top select should not create pairwise feedback: %#v", got)
	}
}

func TestSoftConfirmPendingDroppedOnBackspaceAndEscape(t *testing.T) {
	eng := &fakeDecodeEngine{lists: [][]engine.Candidate{
		{
			{Text: "泥", Score: 100, ConsumedKeys: 2, Syllables: []string{"ni"}},
			{Text: "你", Score: 90, ConsumedKeys: 2, Syllables: []string{"ni"}},
		},
		{
			{Text: "好", Score: 90, ConsumedKeys: 2, Syllables: []string{"hao"}},
		},
	}}
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
	first := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nihc"})
	if bareCandidateText(first.Candidates.Candidates[0].Text) != "泥" {
		t.Fatalf("top=%#v", first.Candidates)
	}
	soft := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "2"})
	if soft.Commit != nil && soft.Commit.Text != "" {
		t.Fatalf("expected soft confirm, got commit %#v", soft.Commit)
	}
	if soft.Composition == nil || soft.Composition.Preedit != "你hc" {
		t.Fatalf("soft preedit=%#v", soft.Composition)
	}
	if s.PendingFeedbackCount() != 1 {
		t.Fatalf("pending after soft non-top=%d", s.PendingFeedbackCount())
	}
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyDelete})
	if s.PendingFeedbackCount() != 0 {
		t.Fatalf("backspace should drop pending, pending=%d", s.PendingFeedbackCount())
	}

	eng.calls = 0
	eng.lists = [][]engine.Candidate{{
		{Text: "泥", Score: 100, ConsumedKeys: 2, Syllables: []string{"ni"}},
		{Text: "你", Score: 90, ConsumedKeys: 2, Syllables: []string{"ni"}},
	}}
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nihc"})
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "2"})
	if s.PendingFeedbackCount() != 1 {
		t.Fatalf("pending before escape=%d", s.PendingFeedbackCount())
	}
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyEscape})
	if s.PendingFeedbackCount() != 0 {
		t.Fatalf("escape should drop pending, pending=%d", s.PendingFeedbackCount())
	}
	if got := s.TakeAcceptedFeedback(); len(got) != 0 {
		t.Fatalf("cancelled soft confirm must not accept: %#v", got)
	}
}

func TestIncompatibleConsumedKeysSkipPending(t *testing.T) {
	eng := &fakeDecodeEngine{lists: [][]engine.Candidate{{
		{Text: "你好", Score: 50, ConsumedKeys: 4},
		{Text: "你", Score: 100, ConsumedKeys: 2},
	}}}
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nihc"})
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "2"})
	if s.PendingFeedbackCount() != 0 {
		t.Fatalf("incompatible pair should skip pending")
	}
}

func TestStaleSelectCandidateRejected(t *testing.T) {
	eng := &fakeDecodeEngine{lists: [][]engine.Candidate{{
		{Text: "背景", Score: 100, ConsumedKeys: 4},
		{Text: "北京", Score: 90, ConsumedKeys: 4},
	}}}
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
	first := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "bwjk"})
	oldGen := first.Candidates.Generation
	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "x"})
	res, ok := s.SelectCandidate(oldGen, 1)
	if ok || res != nil {
		t.Fatalf("stale generation should reject select, ok=%v res=%#v", ok, res)
	}
}
