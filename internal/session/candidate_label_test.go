package session

import (
	"regexp"
	"testing"

	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

var numberedCandidate = regexp.MustCompile(`^[1-9]\.`)

func bareCandidateText(text string) string {
	if numberedCandidate.MatchString(text) {
		return text[2:]
	}
	return text
}

func candidateBareTexts(res *KeyResult) []string {
	if res == nil || res.Candidates == nil {
		return nil
	}
	out := make([]string, 0, len(res.Candidates.Candidates))
	for _, cand := range res.Candidates.Candidates {
		out = append(out, bareCandidateText(cand.Text))
	}
	return out
}

func containsBareText(texts []string, want string) bool {
	for _, text := range texts {
		if bareCandidateText(text) == want {
			return true
		}
	}
	return false
}

func topBareText(res *KeyResult) string {
	if res == nil || res.Candidates == nil || len(res.Candidates.Candidates) == 0 {
		return ""
	}
	return bareCandidateText(res.Candidates.Candidates[0].Text)
}

func TestCandidateListUsesNumberedLabels(t *testing.T) {
	m := NewManager(manyEngine{})
	s := m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "ni"})
	if res.Candidates == nil || len(res.Candidates.Candidates) != 9 {
		t.Fatalf("candidates = %#v", res.Candidates)
	}
	for i, cand := range res.Candidates.Candidates {
		want := string(rune('1'+i)) + "." + string(rune('A'+i))
		if cand.Text != want {
			t.Fatalf("label[%d]=%q want %q", i, cand.Text, want)
		}
	}
}
