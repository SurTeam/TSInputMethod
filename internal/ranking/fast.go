package ranking

import "sort"

type Ranker interface {
	Rank(req Request) []Candidate
}

type FastRanker struct{}

func NewFastRanker() *FastRanker {
	return &FastRanker{}
}

func (r *FastRanker) Rank(req Request) []Candidate {
	out := make([]Candidate, len(req.Candidates))
	copy(out, req.Candidates)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			ri := len([]rune(out[i].Text))
			rj := len([]rune(out[j].Text))
			if ri == rj {
				return out[i].Text < out[j].Text
			}
			return ri > rj
		}
		return out[i].Score > out[j].Score
	})
	return out
}
