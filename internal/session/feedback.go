package session

import "github.com/SurTeam/TSInputMethod/internal/engine"

type candidateSnapshot struct {
	Preedit          string
	EffectiveContext string
	Generation       uint64
	Candidates       []engine.Candidate
	RankingVersion   uint64
	UserDictVersion  uint64
}

type pendingFeedback struct {
	SnapshotGeneration uint64
	Preedit            string
	EffectiveContext   string
	InputCode          string
	SelectedIndex      int
	Selected           engine.Candidate
	Top                engine.Candidate
	RankingVersion     uint64
	UserDictVersion    uint64
}

type FeedbackEvent struct {
	Preedit          string
	EffectiveContext string
	InputCode        string
	SelectedIndex    int
	Selected         engine.Candidate
	Top              engine.Candidate
	Generation       uint64
	RankingVersion   uint64
	UserDictVersion  uint64
}

func cloneCandidates(list []engine.Candidate) []engine.Candidate {
	if len(list) == 0 {
		return nil
	}
	out := make([]engine.Candidate, len(list))
	for i, item := range list {
		out[i] = item
		if len(item.Syllables) > 0 {
			out[i].Syllables = append([]string(nil), item.Syllables...)
		}
	}
	return out
}

func compatibleFeedbackPair(top, selected engine.Candidate) bool {
	if top.Text == "" || selected.Text == "" {
		return false
	}
	if top.Incomplete != selected.Incomplete {
		return false
	}
	if top.ConsumedKeys != selected.ConsumedKeys {
		return false
	}
	return true
}
