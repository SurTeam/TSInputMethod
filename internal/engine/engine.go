package engine

import (
	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/inference"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	"github.com/SurTeam/TSInputMethod/internal/ranking"
)

type Candidate struct {
	Text         string
	Comment      string
	Score        int32
	Syllables    []string
	Incomplete   bool
	ConsumedKeys int
}

type Engine interface {
	Decode(pinyinInput string) ([]Candidate, error)
}

type StubEngine struct{}

func NewStubEngine() *StubEngine {
	return &StubEngine{}
}

func (e *StubEngine) Decode(pinyinInput string) ([]Candidate, error) {
	if pinyinInput == "" {
		return nil, nil
	}
	return []Candidate{{
		Text:    pinyinInput,
		Comment: "stub",
		Score:   0,
	}}, nil
}

type PinyinEngine struct {
	decoder     *pinyin.Decoder
	coordinator *ranking.Coordinator
	scheme      pinyin.Scheme
}

func NewPinyinEngine(dict *dictionary.Dictionary) *PinyinEngine {
	return NewPinyinEngineWithRanking(dict, ranking.NewCoordinator(ranking.NewFastRanker(), nil))
}

func NewPinyinEngineWithRanking(dict *dictionary.Dictionary, coordinator *ranking.Coordinator) *PinyinEngine {
	if coordinator == nil {
		coordinator = ranking.NewCoordinator(ranking.NewFastRanker(), nil)
	}
	return &PinyinEngine{
		decoder:     pinyin.NewDecoder(dict, 96),
		coordinator: coordinator,
		scheme:      pinyin.SchemeFromEnv(),
	}
}

func (e *PinyinEngine) SetScheme(scheme pinyin.Scheme) {
	if e == nil {
		return
	}
	e.scheme = scheme
}

func (e *PinyinEngine) Scheme() pinyin.Scheme {
	if e == nil {
		return pinyin.SchemeQuanpin
	}
	return e.scheme
}

func (e *PinyinEngine) SetInferenceBackend(backend inference.Backend) {
	if e == nil || e.coordinator == nil {
		return
	}
	e.coordinator.SetBackend(backend)
}

func (e *PinyinEngine) Coordinator() *ranking.Coordinator {
	if e == nil {
		return nil
	}
	return e.coordinator
}

func (e *PinyinEngine) Decode(pinyinInput string) ([]Candidate, error) {
	if e == nil || e.decoder == nil {
		return nil, nil
	}
	analyzed := pinyin.AnalyzeInput(e.scheme, pinyinInput)
	if analyzed.FixedSplit && len(analyzed.SyllableAlts) > 1 {
		lists := make([][]Candidate, 0, len(analyzed.SyllableAlts))
		for _, path := range analyzed.SyllableAlts {
			raw := e.decoder.DecodeSyllables(path, analyzed.Remainder)
			lists = append(lists, e.rankRaw(analyzed, raw))
		}
		return interleaveEngineCandidates(lists, 96), nil
	}
	var raw []pinyin.Candidate
	if analyzed.FixedSplit {
		if len(analyzed.SyllableAlts) > 0 {
			raw = e.decoder.DecodeSyllablePaths(analyzed.SyllableAlts, analyzed.Remainder)
		} else {
			raw = e.decoder.DecodeSyllables(analyzed.Syllables, analyzed.Remainder)
		}
	} else {
		raw = e.decoder.Decode(analyzed.Joined)
	}
	return e.rankRaw(analyzed, raw), nil
}

func (e *PinyinEngine) rankRaw(analyzed pinyin.NormalizedInput, raw []pinyin.Candidate) []Candidate {
	req := ranking.Request{
		Pinyin:     analyzed.Joined,
		Candidates: make([]ranking.Candidate, 0, len(raw)),
	}
	meta := make([]pinyin.Candidate, len(raw))
	copy(meta, raw)
	for _, item := range raw {
		req.Candidates = append(req.Candidates, ranking.Candidate{
			Text:       item.Text,
			Comment:    item.Comment,
			Score:      item.Score,
			Coverage:   len(item.Syllables),
			Incomplete: item.Incomplete,
			Composed:   item.Composed,
		})
	}
	ranked := e.coordinator.RankSync(req)
	out := make([]Candidate, 0, len(ranked))
	for _, item := range ranked {
		cand := Candidate{
			Text:    item.Text,
			Comment: item.Comment,
			Score:   item.Score,
		}
		for _, src := range meta {
			if src.Text == item.Text {
				cand.Syllables = append([]string(nil), src.Syllables...)
				cand.Incomplete = src.Incomplete
				break
			}
		}
		cand.ConsumedKeys = consumedKeysForCandidate(analyzed, cand)
		cand.Comment = rawCommentForCandidate(analyzed, cand)
		out = append(out, cand)
	}
	return out
}

func interleaveEngineCandidates(lists [][]Candidate, limit int) []Candidate {
	if limit <= 0 {
		limit = 96
	}
	indexes := make([]int, len(lists))
	seen := map[string]struct{}{}
	out := make([]Candidate, 0, limit)
	for len(out) < limit {
		progressed := false
		for i, list := range lists {
			for indexes[i] < len(list) {
				cand := list[indexes[i]]
				indexes[i]++
				if _, ok := seen[cand.Text]; ok {
					continue
				}
				seen[cand.Text] = struct{}{}
				out = append(out, cand)
				progressed = true
				break
			}
			if len(out) >= limit {
				break
			}
		}
		if !progressed {
			break
		}
	}
	return out
}

func rawCommentForCandidate(analyzed pinyin.NormalizedInput, cand Candidate) string {
	preedit := analyzed.Raw
	if preedit == "" {
		return cand.Comment
	}
	n := cand.ConsumedKeys
	if n <= 0 {
		if analyzed.FixedSplit && analyzed.Remainder != "" && len(cand.Syllables) == 0 {
			return analyzed.Remainder
		}
		return cand.Comment
	}
	if n > len(preedit) {
		n = len(preedit)
	}
	if n <= 0 {
		return cand.Comment
	}
	return preedit[:n]
}

func consumedKeysForCandidate(analyzed pinyin.NormalizedInput, cand Candidate) int {
	preedit := analyzed.Raw
	if preedit == "" {
		return 0
	}
	if len(cand.Syllables) == 0 {
		if cand.Incomplete {
			return 0
		}
		return len(preedit)
	}
	if analyzed.FixedSplit {
		if cand.Incomplete && len(cand.Syllables) > len(analyzed.Syllables) {
			return len(preedit)
		}
		n := len(cand.Syllables) * 2
		if n > len(preedit) {
			return len(preedit)
		}
		if n == 0 && cand.Incomplete {
			return 0
		}
		return n
	}
	covered := ""
	for _, syl := range cand.Syllables {
		covered += syl
	}
	if len(covered) <= len(preedit) && preedit[:len(covered)] == covered {
		return len(covered)
	}
	if cand.Incomplete {
		return 0
	}
	return len(preedit)
}
