package session

import (
	"strings"
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	"github.com/SurTeam/TSInputMethod/internal/ranking"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func TestUnicodeTailAndWindowClamp(t *testing.T) {
	if got := unicodeTail("你好世界", 2); got != "世界" {
		t.Fatalf("tail=%q", got)
	}
	if got := clampContextWindow(0); got != defaultContextWindow {
		t.Fatalf("default window=%d", got)
	}
	if got := clampContextWindow(100); got != contextMax {
		t.Fatalf("max window=%d", got)
	}
	runess := appendCommittedRunes(nil, strings.Repeat("汉", contextMax+8), contextMax)
	if len(runess) != contextMax {
		t.Fatalf("committed len=%d", len(runess))
	}
}

func TestEffectiveContextSoftConfirmBackspaceEscape(t *testing.T) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "世界", Syllables: []string{"shi", "jie"}, Weight: 50},
		{Text: "世", Syllables: []string{"shi"}, Weight: 1000},
		{Text: "界", Syllables: []string{"jie"}, Weight: 900},
	})
	eng := engine.NewPinyinEngine(dict)
	eng.SetScheme(pinyin.SchemeXiaohe)
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)

	first := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "uijp"})
	idx := -1
	for i, text := range candidateTexts(first) {
		if bareCandidateText(text) == "世" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("missing 世: %#v", candidateTexts(first))
	}
	soft := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: string(rune('1' + idx))})
	if soft.Composition == nil || soft.Composition.Preedit != "世jp" {
		t.Fatalf("soft preedit=%#v", soft.Composition)
	}
	if s.ConfirmedComposition() != "世" {
		t.Fatalf("confirmed=%q", s.ConfirmedComposition())
	}
	if s.EffectiveContext() != "世" {
		t.Fatalf("effective after soft=%q", s.EffectiveContext())
	}
	if s.CommittedContext() != "" {
		t.Fatalf("committed should stay empty during soft confirm: %q", s.CommittedContext())
	}

	back := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyDelete})
	if back.Composition == nil || back.Composition.Preedit != "uijp" {
		t.Fatalf("backspace restore=%#v", back.Composition)
	}
	if s.ConfirmedComposition() != "" || s.EffectiveContext() != "" {
		t.Fatalf("context after undo confirmed=%q effective=%q", s.ConfirmedComposition(), s.EffectiveContext())
	}

	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: string(rune('1' + idx))})
	cancel := s.HandleKey(&imev1.KeyEvent{KeyDown: true, KeyCode: keyEscape})
	if cancel.Composition == nil || cancel.Composition.Preedit != "" {
		t.Fatalf("escape composition=%#v", cancel.Composition)
	}
	if s.ConfirmedComposition() != "" || s.EffectiveContext() != "" {
		t.Fatalf("context after escape confirmed=%q effective=%q", s.ConfirmedComposition(), s.EffectiveContext())
	}
}

func TestFinalCommitAppendsCommittedOnce(t *testing.T) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "你好", Syllables: []string{"ni", "hao"}, Weight: 1000},
		{Text: "你", Syllables: []string{"ni"}, Weight: 100},
		{Text: "好", Syllables: []string{"hao"}, Weight: 90},
		{Text: "世界", Syllables: []string{"shi", "jie"}, Weight: 800},
		{Text: "世", Syllables: []string{"shi"}, Weight: 70},
		{Text: "界", Syllables: []string{"jie"}, Weight: 60},
	})
	eng := engine.NewPinyinEngine(dict)
	eng.SetScheme(pinyin.SchemeXiaohe)
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)

	_ = s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "nihcuijp"})
	soft := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: " ", KeyCode: keySpace})
	if soft.Composition == nil || soft.Composition.Preedit != "你好uijp" {
		t.Fatalf("soft=%#v", soft.Composition)
	}
	if s.EffectiveContext() != "你好" {
		t.Fatalf("effective during soft=%q", s.EffectiveContext())
	}
	if s.CommittedContext() != "" {
		t.Fatalf("committed during soft=%q", s.CommittedContext())
	}

	fin := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: " ", KeyCode: keySpace})
	if fin.Commit == nil || fin.Commit.Text != "你好世界" {
		t.Fatalf("final commit=%#v", fin.Commit)
	}
	if s.ConfirmedComposition() != "" {
		t.Fatalf("confirmed after commit=%q", s.ConfirmedComposition())
	}
	if s.CommittedContext() != "你好世界" {
		t.Fatalf("committed after final=%q", s.CommittedContext())
	}
	if s.EffectiveContext() != "你好世界" {
		t.Fatalf("effective after final=%q", s.EffectiveContext())
	}
}

func TestCommittedContextWindowMaxAndIsolation(t *testing.T) {
	m := NewManager(engine.NewStubEngine())
	a := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	b := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	a.SetContextWindow(contextMax)

	overlong := strings.Repeat("甲", contextMax+5)
	a.mu.Lock()
	a.appendCommittedLocked(overlong)
	a.mu.Unlock()
	if got := []rune(a.CommittedContext()); len(got) != contextMax {
		t.Fatalf("committed len=%d", len(got))
	}
	if b.CommittedContext() != "" {
		t.Fatalf("session b leaked context=%q", b.CommittedContext())
	}

	a.ClearLearningContext()
	if a.CommittedContext() != "" {
		t.Fatalf("clear failed: %q", a.CommittedContext())
	}

	a.mu.Lock()
	a.appendCommittedLocked("残留")
	a.mu.Unlock()
	if a.CommittedContext() != "残留" {
		t.Fatalf("seed committed=%q", a.CommittedContext())
	}
	if !m.Close(a.ID()) {
		t.Fatalf("close failed")
	}
	if _, ok := m.Get(a.ID()); ok {
		t.Fatalf("session still present after close")
	}
	if b.CommittedContext() != "" {
		t.Fatalf("session b changed after a close: %q", b.CommittedContext())
	}
}

func TestSoftConfirmFeedsEffectiveContextIntoRanking(t *testing.T) {
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "世界", Syllables: []string{"shi", "jie"}, Weight: 50},
		{Text: "世", Syllables: []string{"shi"}, Weight: 1000},
		{Text: "界", Syllables: []string{"jie"}, Weight: 900},
	})
	capture := &contextCaptureRanker{inner: ranking.NewFastRanker()}
	eng := engine.NewPinyinEngineWithRanking(dict, ranking.NewCoordinator(capture, nil))
	eng.SetScheme(pinyin.SchemeXiaohe)
	s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)

	first := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "uijp"})
	idx := -1
	for i, text := range candidateTexts(first) {
		if bareCandidateText(text) == "世" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("missing 世: %#v", candidateTexts(first))
	}
	capture.lastContext = "unset"
	soft := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: string(rune('1' + idx))})
	if soft.Composition == nil || soft.Composition.Preedit != "世jp" {
		t.Fatalf("soft=%#v", soft.Composition)
	}
	if capture.lastContext != "世" {
		t.Fatalf("ranking context=%q want 世", capture.lastContext)
	}
	if soft.Candidates == nil || len(soft.Candidates.Candidates) == 0 || bareCandidateText(soft.Candidates.Candidates[0].Text) != "界" {
		t.Fatalf("remainder candidates=%#v", soft.Candidates)
	}
}

type contextCaptureRanker struct {
	inner       ranking.Ranker
	lastContext string
}

func (r *contextCaptureRanker) Rank(req ranking.Request) []ranking.Candidate {
	r.lastContext = req.Context
	if r.inner == nil {
		return ranking.NewFastRanker().Rank(req)
	}
	return r.inner.Rank(req)
}
