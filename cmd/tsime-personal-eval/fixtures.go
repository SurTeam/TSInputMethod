package main

import "github.com/SurTeam/TSInputMethod/internal/ranking"

type Sample struct {
	Context       string
	InputCode     string
	Candidates    []ranking.Candidate
	SelectedIndex int
}

func deterministicFixtures() []Sample {
	bwjk := []ranking.Candidate{
		{Text: "背景", Score: 1000, Coverage: 2},
		{Text: "北京", Score: 900, Coverage: 2},
		{Text: "背井", Score: 100, Coverage: 2},
	}
	return []Sample{
		{
			Context:       "我明天要去",
			InputCode:     "bwjk",
			Candidates:    cloneCands(bwjk),
			SelectedIndex: 1,
		},
		{
			Context:       "这张图片的",
			InputCode:     "bwjk",
			Candidates:    cloneCands(bwjk),
			SelectedIndex: 0,
		},
		{
			Context:       "我明天要去",
			InputCode:     "bwjk",
			Candidates:    cloneCands(bwjk),
			SelectedIndex: 1,
		},
		{
			Context:       "这张图片的",
			InputCode:     "bwjk",
			Candidates:    cloneCands(bwjk),
			SelectedIndex: 0,
		},
		{
			Context:       "我明天要去",
			InputCode:     "bwjk",
			Candidates:    cloneCands(bwjk),
			SelectedIndex: 1,
		},
		{
			Context:       "公司安排去",
			InputCode:     "bwjk",
			Candidates:    cloneCands(bwjk),
			SelectedIndex: 1,
		},
		{
			Context:       "这张图片的",
			InputCode:     "bwjk",
			Candidates:    cloneCands(bwjk),
			SelectedIndex: 0,
		},
		{
			Context:       "我明天要去",
			InputCode:     "bwjk",
			Candidates:    cloneCands(bwjk),
			SelectedIndex: 1,
		},
	}
}

func cloneCands(in []ranking.Candidate) []ranking.Candidate {
	out := make([]ranking.Candidate, len(in))
	copy(out, in)
	return out
}
