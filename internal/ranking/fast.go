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
		return better(out[i], out[j])
	})
	return out
}

func better(a, b Candidate) bool {
	if a.Incomplete != b.Incomplete {
		return !a.Incomplete && b.Incomplete
	}
	if a.Coverage != b.Coverage {
		return a.Coverage > b.Coverage
	}
	if a.Composed != b.Composed {
		return !a.Composed && b.Composed
	}
	if a.Score != b.Score {
		return a.Score > b.Score
	}
	ri := len([]rune(a.Text))
	rj := len([]rune(b.Text))
	if ri != rj {
		return ri > rj
	}
	return a.Text < b.Text
}
