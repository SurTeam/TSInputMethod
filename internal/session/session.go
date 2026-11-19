package session

import (
	"strings"
	"sync"
	"unicode"

	"github.com/SurTeam/TSInputMethod/internal/engine"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

type Session struct {
	mu         sync.Mutex
	id         string
	mode       imev1.InputMode
	preedit    []rune
	cursor     int
	generation uint64
	engine     engine.Engine
}

type KeyResult struct {
	Composition *imev1.CompositionState
	Candidates  *imev1.CandidateList
	Commit      *imev1.CommitText
	Consumed    bool
}

func newSession(id string, mode imev1.InputMode, eng engine.Engine) *Session {
	if mode == imev1.InputMode_INPUT_MODE_UNSPECIFIED {
		mode = imev1.InputMode_INPUT_MODE_CHINESE
	}
	if eng == nil {
		eng = engine.NewStubEngine()
	}
	return &Session{
		id:     id,
		mode:   mode,
		engine: eng,
	}
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) Mode() imev1.InputMode {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mode
}

func (s *Session) SetMode(mode imev1.InputMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode
	s.clearCompositionLocked()
}

func (s *Session) HandleKey(ev *imev1.KeyEvent) *KeyResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ev == nil || !ev.KeyDown {
		return &KeyResult{Consumed: false}
	}
	if s.mode == imev1.InputMode_INPUT_MODE_ENGLISH {
		return s.handleEnglishLocked(ev)
	}
	return s.handleChineseLocked(ev)
}

func (s *Session) handleEnglishLocked(ev *imev1.KeyEvent) *KeyResult {
	if ev.Characters == "" {
		return &KeyResult{Consumed: false}
	}
	s.generation++
	return &KeyResult{
		Commit: &imev1.CommitText{
			SessionId:  s.id,
			Text:       ev.Characters,
			Generation: s.generation,
		},
		Consumed: true,
	}
}

func (s *Session) handleChineseLocked(ev *imev1.KeyEvent) *KeyResult {
	switch ev.KeyCode {
	case 8, 127:
		return s.backspaceLocked()
	case 27:
		return s.cancelLocked()
	case 32:
		return s.commitPreeditLocked()
	}
	if ev.Characters == "" {
		return &KeyResult{Consumed: false}
	}
	for _, r := range ev.Characters {
		if unicode.IsDigit(r) {
			return s.selectByDigitLocked(r)
		}
		if r >= 'a' && r <= 'z' {
			s.preedit = append(s.preedit[:s.cursor], append([]rune{r}, s.preedit[s.cursor:]...)...)
			s.cursor++
			continue
		}
		if r >= 'A' && r <= 'Z' {
			lower := unicode.ToLower(r)
			s.preedit = append(s.preedit[:s.cursor], append([]rune{lower}, s.preedit[s.cursor:]...)...)
			s.cursor++
			continue
		}
		return &KeyResult{Consumed: false}
	}
	return s.refreshLocked()
}

func (s *Session) backspaceLocked() *KeyResult {
	if len(s.preedit) == 0 || s.cursor == 0 {
		if len(s.preedit) == 0 {
			return &KeyResult{Consumed: false}
		}
		return s.refreshLocked()
	}
	s.preedit = append(s.preedit[:s.cursor-1], s.preedit[s.cursor:]...)
	s.cursor--
	return s.refreshLocked()
}

func (s *Session) cancelLocked() *KeyResult {
	if len(s.preedit) == 0 {
		return &KeyResult{Consumed: false}
	}
	s.clearCompositionLocked()
	return s.refreshLocked()
}

func (s *Session) commitPreeditLocked() *KeyResult {
	if len(s.preedit) == 0 {
		return &KeyResult{Consumed: false}
	}
	pinyinInput := string(s.preedit)
	candidates, err := s.engine.Decode(pinyinInput)
	text := pinyinInput
	if err == nil && len(candidates) > 0 {
		text = candidates[0].Text
	}
	s.clearCompositionLocked()
	s.generation++
	return &KeyResult{
		Composition: s.compositionLocked(),
		Candidates:  s.candidateListLocked(nil),
		Commit: &imev1.CommitText{
			SessionId:  s.id,
			Text:       text,
			Generation: s.generation,
		},
		Consumed: true,
	}
}

func (s *Session) selectByDigitLocked(digit rune) *KeyResult {
	pinyin := string(s.preedit)
	candidates, err := s.engine.Decode(pinyin)
	if err != nil || len(candidates) == 0 {
		return s.refreshLocked()
	}
	index := int(digit - '1')
	if digit == '0' {
		index = 9
	}
	if index < 0 || index >= len(candidates) {
		return s.refreshLocked()
	}
	text := candidates[index].Text
	s.clearCompositionLocked()
	s.generation++
	return &KeyResult{
		Composition: s.compositionLocked(),
		Candidates:  s.candidateListLocked(nil),
		Commit: &imev1.CommitText{
			SessionId:  s.id,
			Text:       text,
			Generation: s.generation,
		},
		Consumed: true,
	}
}

func (s *Session) refreshLocked() *KeyResult {
	s.generation++
	pinyin := string(s.preedit)
	var list []engine.Candidate
	if pinyin != "" {
		decoded, err := s.engine.Decode(pinyin)
		if err == nil {
			list = decoded
		}
	}
	return &KeyResult{
		Composition: s.compositionLocked(),
		Candidates:  s.candidateListLocked(list),
		Consumed:    true,
	}
}

func (s *Session) compositionLocked() *imev1.CompositionState {
	return &imev1.CompositionState{
		SessionId:  s.id,
		Preedit:    string(s.preedit),
		Cursor:     uint32(s.cursor),
		Generation: s.generation,
	}
}

func (s *Session) candidateListLocked(list []engine.Candidate) *imev1.CandidateList {
	out := &imev1.CandidateList{
		SessionId:  s.id,
		Generation: s.generation,
		Page:       0,
		PageSize:   9,
	}
	for _, item := range list {
		out.Candidates = append(out.Candidates, &imev1.Candidate{
			Text:    item.Text,
			Comment: item.Comment,
			Score:   item.Score,
		})
	}
	return out
}

func (s *Session) clearCompositionLocked() {
	s.preedit = nil
	s.cursor = 0
}

func (s *Session) SelectCandidate(generation uint64, index uint32) (*KeyResult, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if generation != 0 && generation != s.generation {
		return nil, false
	}
	pinyin := string(s.preedit)
	candidates, err := s.engine.Decode(pinyin)
	if err != nil || int(index) >= len(candidates) {
		return s.refreshLocked(), true
	}
	text := candidates[index].Text
	s.clearCompositionLocked()
	s.generation++
	return &KeyResult{
		Composition: s.compositionLocked(),
		Candidates:  s.candidateListLocked(nil),
		Commit: &imev1.CommitText{
			SessionId:  s.id,
			Text:       text,
			Generation: s.generation,
		},
		Consumed: true,
	}, true
}

func IsLetterInput(characters string) bool {
	return strings.IndexFunc(characters, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'))
	}) < 0 && characters != ""
}
