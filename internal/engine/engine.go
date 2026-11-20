package engine

import (
	"fmt"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/emoji"
	"github.com/SurTeam/TSInputMethod/internal/inference"
	"github.com/SurTeam/TSInputMethod/internal/personal"
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

type ContextDecoder interface {
	DecodeWithContext(pinyinInput string, effectiveContext string) ([]Candidate, error)
}

type StubEngine struct{}

func NewStubEngine() *StubEngine {
	return &StubEngine{}
}

func (e *StubEngine) Decode(pinyinInput string) ([]Candidate, error) {
	return e.DecodeWithContext(pinyinInput, "")
}

func (e *StubEngine) DecodeWithContext(pinyinInput string, _ string) ([]Candidate, error) {
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
	emoji       *emoji.Table
	user        *dictionary.UserDictionary
	settings    *personal.Settings
	model       *ranking.PersonalModel
	queue       *ranking.UpdateQueue
	fused       *ranking.FusedRanker
}

func NewPinyinEngine(dict *dictionary.Dictionary) *PinyinEngine {
	settings := personal.DefaultSettings()
	model := ranking.NewPersonalModel(ranking.DefaultSGDConfig())
	fused := ranking.NewFusedRanker(ranking.NewFastRanker(), model, nil, ranking.FusionConfig{
		Enabled:             settings.PersonalizationEnabled,
		SGDWeight:           settings.SGDWeight,
		UserFrequencyWeight: settings.UserFrequencyWeight,
	})
	coord := ranking.NewCoordinator(fused, nil)
	dictionary.SetValidSyllableFunc(pinyin.IsValidSyllable)
	eng := &PinyinEngine{
		decoder:     pinyin.NewDecoder(dict, 96),
		coordinator: coord,
		scheme:      pinyin.SchemeFromEnv(),
		settings:    settings,
		model:       model,
		fused:       fused,
		queue:       ranking.NewUpdateQueue(model, 128),
	}
	eng.queue.Start()
	return eng
}

func NewPinyinEngineWithRanking(dict *dictionary.Dictionary, coordinator *ranking.Coordinator) *PinyinEngine {
	if coordinator == nil {
		coordinator = ranking.NewCoordinator(ranking.NewFastRanker(), nil)
	}
	dictionary.SetValidSyllableFunc(pinyin.IsValidSyllable)
	return &PinyinEngine{
		decoder:     pinyin.NewDecoder(dict, 96),
		coordinator: coordinator,
		scheme:      pinyin.SchemeFromEnv(),
		settings:    personal.DefaultSettings(),
	}
}

func (e *PinyinEngine) SetUserDictionary(user *dictionary.UserDictionary) {
	if e == nil {
		return
	}
	e.user = user
	if e.decoder != nil {
		if e.settings == nil || e.settings.Snapshot().UserDictionaryEnabled {
			e.decoder.SetUserDictionary(user)
		} else {
			e.decoder.SetUserDictionary(nil)
		}
	}
	if e.fused != nil {
		if user == nil {
			e.fused.SetUsageLookup(nil)
		} else {
			e.fused.SetUsageLookup(user.UsageCountForText)
		}
	}
}

func (e *PinyinEngine) UserDictionary() *dictionary.UserDictionary {
	if e == nil {
		return nil
	}
	return e.user
}

func (e *PinyinEngine) Settings() *personal.Settings {
	if e == nil {
		return nil
	}
	return e.settings
}

func (e *PinyinEngine) PersonalModel() *ranking.PersonalModel {
	if e == nil {
		return nil
	}
	return e.model
}

func (e *PinyinEngine) UpdateQueue() *ranking.UpdateQueue {
	if e == nil {
		return nil
	}
	return e.queue
}

func (e *PinyinEngine) ApplySettings() {
	if e == nil || e.settings == nil {
		return
	}
	snap := e.settings.Snapshot()
	if e.model != nil {
		e.model.SetEnabled(snap.PersonalizationEnabled && !snap.PrivacyMode)
	}
	if e.fused != nil {
		e.fused.SetConfig(ranking.FusionConfig{
			Enabled:             snap.PersonalizationEnabled && !snap.PrivacyMode,
			SGDWeight:           snap.SGDWeight,
			UserFrequencyWeight: snap.UserFrequencyWeight,
		})
	}
	if e.decoder != nil {
		if snap.UserDictionaryEnabled {
			e.decoder.SetUserDictionary(e.user)
		} else {
			e.decoder.SetUserDictionary(nil)
		}
	}
}

func (e *PinyinEngine) EnqueueFeedback(context, code, selected, top string) bool {
	if e == nil || e.queue == nil {
		return false
	}
	if e.settings != nil {
		snap := e.settings.Snapshot()
		if !snap.PersonalizationEnabled || snap.PrivacyMode {
			return false
		}
	}
	return e.queue.Enqueue(ranking.PairwiseFeedback{
		Context:  context,
		Code:     code,
		Selected: selected,
		Top:      top,
	})
}

func (e *PinyinEngine) RecordSelectedUsage(text string) {
	if e == nil || e.user == nil || text == "" {
		return
	}
	if e.settings != nil {
		snap := e.settings.Snapshot()
		if !snap.UserDictionaryEnabled || snap.PrivacyMode {
			return
		}
	}
	e.user.RecordUsageText(text)
}

func (e *PinyinEngine) ClearPersonalWeights() error {
	if e == nil || e.model == nil {
		return nil
	}
	e.model.ResetPersonalization()
	if e.model.Path() == "" {
		return nil
	}
	return e.model.SaveIfDirty()
}

func (e *PinyinEngine) ClearUserDictionary(mode string) error {
	if e == nil || e.user == nil {
		return nil
	}
	switch mode {
	case "learned":
		_ = e.user.ClearLearnedEntries()
	case "all":
		_ = e.user.ClearAllEntries()
	default:
		return fmt.Errorf("engine: clear user dictionary expects learned|all")
	}
	if e.user.Path() == "" {
		return nil
	}
	return e.user.Save()
}

func (e *PinyinEngine) ClearPersonalData(clearWeights, clearUserDictLearned, clearUserDictAll bool) error {
	if e == nil {
		return nil
	}
	var first error
	if clearWeights {
		if err := e.ClearPersonalWeights(); err != nil && first == nil {
			first = err
		}
	}
	if clearUserDictAll {
		if err := e.ClearUserDictionary("all"); err != nil && first == nil {
			first = err
		}
	} else if clearUserDictLearned {
		if err := e.ClearUserDictionary("learned"); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (e *PinyinEngine) Close() error {
	if e == nil {
		return nil
	}
	if e.queue != nil {
		e.queue.Close()
	}
	var first error
	if e.model != nil {
		if err := e.model.SaveIfDirty(); err != nil {
			first = err
		}
	}
	if e.user != nil && e.user.Path() != "" {
		if err := e.user.Save(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (e *PinyinEngine) SetEmojiTable(table *emoji.Table) {
	if e == nil {
		return
	}
	e.emoji = table
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
	return e.DecodeWithContext(pinyinInput, "")
}

func (e *PinyinEngine) DecodeWithContext(pinyinInput string, effectiveContext string) ([]Candidate, error) {
	if e == nil || e.decoder == nil {
		return nil, nil
	}
	analyzed := pinyin.AnalyzeInput(e.scheme, pinyinInput)
	if analyzed.FixedSplit && len(analyzed.SyllableAlts) > 1 {
		lists := make([][]Candidate, 0, len(analyzed.SyllableAlts))
		for _, path := range analyzed.SyllableAlts {
			raw := e.decoder.DecodeSyllables(path, analyzed.Remainder)
			lists = append(lists, e.expandEmoji(e.rankRaw(analyzed, raw, effectiveContext)))
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
	return e.expandEmoji(e.rankRaw(analyzed, raw, effectiveContext)), nil
}

func (e *PinyinEngine) expandEmoji(cands []Candidate) []Candidate {
	if e == nil || e.emoji == nil || len(cands) == 0 {
		return cands
	}
	out := make([]Candidate, 0, len(cands)+8)
	seen := make(map[string]struct{}, len(cands)+8)
	for _, cand := range cands {
		if _, ok := seen[cand.Text]; ok {
			continue
		}
		seen[cand.Text] = struct{}{}
		out = append(out, cand)
		for _, emojiText := range e.emoji.Lookup(cand.Text) {
			if _, ok := seen[emojiText]; ok {
				continue
			}
			seen[emojiText] = struct{}{}
			out = append(out, Candidate{
				Text:         emojiText,
				Comment:      cand.Comment,
				Score:        cand.Score,
				Syllables:    append([]string(nil), cand.Syllables...),
				Incomplete:   cand.Incomplete,
				ConsumedKeys: cand.ConsumedKeys,
			})
		}
	}
	return out
}

func (e *PinyinEngine) rankRaw(analyzed pinyin.NormalizedInput, raw []pinyin.Candidate, effectiveContext string) []Candidate {
	req := ranking.Request{
		Pinyin:     analyzed.Joined,
		Candidates: make([]ranking.Candidate, 0, len(raw)),
		Context:    effectiveContext,
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
