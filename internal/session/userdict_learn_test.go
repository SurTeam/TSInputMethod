package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

type learnCapture struct {
	texts []string
	syls  [][]string
}

func (l *learnCapture) ObserveCommittedPhrase(text string, syllables []string) error {
	l.texts = append(l.texts, text)
	l.syls = append(l.syls, append([]string(nil), syllables...))
	return nil
}

func TestSoftConfirmCapturesSyllablesForAutoLearn(t *testing.T) {
	dictionary.SetValidSyllableFunc(pinyin.IsValidSyllable)
	capture := &learnCapture{}
	eng := &fakeDecodeEngine{lists: [][]engine.Candidate{
		{
			{Text: "星河", Score: 100, ConsumedKeys: 4, Syllables: []string{"xing", "he"}},
			{Text: "星河工作室", Score: 50, ConsumedKeys: 10, Syllables: []string{"xing", "he", "gong", "zuo", "shi"}},
		},
		{
			{Text: "工作室", Score: 90, ConsumedKeys: 6, Syllables: []string{"gong", "zuo", "shi"}},
		},
	}}
	m := NewManager(eng)
	m.SetPhraseObserver(capture)
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	first := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "xkhegszoui"})
	if first.Candidates == nil || len(first.Candidates.Candidates) == 0 {
		t.Fatalf("first decode empty: %#v", first)
	}
	soft := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "1"})
	if soft.Commit != nil {
		t.Fatalf("soft confirm should not final-commit, got %#v", soft.Commit)
	}
	if soft.Candidates == nil || len(soft.Candidates.Candidates) == 0 {
		t.Fatalf("remainder candidates empty: %#v", soft)
	}
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "1"})
	if res == nil || res.Commit == nil || res.Commit.Text != "星河工作室" {
		t.Fatalf("commit=%#v", res)
	}
	if len(capture.texts) != 1 || capture.texts[0] != "星河工作室" {
		t.Fatalf("observe texts=%v", capture.texts)
	}
	want := []string{"xing", "he", "gong", "zuo", "shi"}
	if len(capture.syls) != 1 || len(capture.syls[0]) != len(want) {
		t.Fatalf("observe syls=%v", capture.syls)
	}
	for i := range want {
		if capture.syls[0][i] != want[i] {
			t.Fatalf("observe syls=%v want=%v", capture.syls[0], want)
		}
	}
}
