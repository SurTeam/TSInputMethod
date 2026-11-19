package engine

import (
	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/inference"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	"github.com/SurTeam/TSInputMethod/internal/ranking"
)

type Candidate struct {
	Text    string
	Comment string
	Score   int32
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
		decoder:     pinyin.NewDecoder(dict, 32),
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
	normalized := pinyin.NormalizeInput(e.scheme, pinyinInput)
	raw := e.decoder.Decode(normalized)
	req := ranking.Request{
		Pinyin:     normalized,
		Candidates: make([]ranking.Candidate, 0, len(raw)),
	}
	for _, item := range raw {
		req.Candidates = append(req.Candidates, ranking.Candidate{
			Text:    item.Text,
			Comment: item.Comment,
			Score:   item.Score,
		})
	}
	ranked := e.coordinator.RankSync(req)
	out := make([]Candidate, 0, len(ranked))
	for _, item := range ranked {
		out = append(out, Candidate{
			Text:    item.Text,
			Comment: item.Comment,
			Score:   item.Score,
		})
	}
	return out, nil
}
