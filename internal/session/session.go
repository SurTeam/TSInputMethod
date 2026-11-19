package session

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/SurTeam/TSInputMethod/internal/engine"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

const (
	defaultPageSize = 9
	keyDelete       = 51
	keyEscape       = 53
	keySpace        = 49
	keyForwardDel   = 117
	keyReturn       = 36
	keypadEnter     = 76
)

type confirmedSegment struct {
	Text string
	Keys string
}

type Session struct {
	mu         sync.Mutex
	id         string
	mode       imev1.InputMode
	preedit    []rune
	cursor     int
	confirmed  []confirmedSegment
	generation uint64
	page       uint32
	pageSize   uint32
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
		id:       id,
		mode:     mode,
		engine:   eng,
		pageSize: defaultPageSize,
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

func (s *Session) SetMode(mode imev1.InputMode) *KeyResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	var commit *KeyResult
	if len(s.confirmed) > 0 || len(s.preedit) > 0 {
		commit = s.commitCompositionLocked()
	}
	s.mode = mode
	s.clearCompositionLocked()
	if commit == nil {
		return &KeyResult{Consumed: false}
	}
	return commit
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
	switch ev.Characters {
	case "-", "_":
		if len(s.preedit) == 0 {
			return &KeyResult{Consumed: false}
		}
		return s.pageLocked(-1)
	case "=", "+":
		if len(s.preedit) == 0 {
			return &KeyResult{Consumed: false}
		}
		return s.pageLocked(1)
case " ":
			return s.commitTopCandidateLocked()
		case "\r", "\n":
			return s.commitRawPreeditLocked()
		}
	if hasLetterOrDigit(ev.Characters) {
		return s.composeCharactersLocked(ev.Characters)
	}
switch ev.KeyCode {
	case keyDelete, keyForwardDel:
		return s.backspaceLocked()
	case keyEscape:
		return s.cancelLocked()
	case keySpace:
		return s.commitTopCandidateLocked()
	case keyReturn, keypadEnter:
		return s.commitRawPreeditLocked()
	}
	if ev.Characters == "" {
		return &KeyResult{Consumed: false}
	}
	return s.composeCharactersLocked(ev.Characters)
}

func (s *Session) composeCharactersLocked(characters string) *KeyResult {
	if characters == "" {
		return &KeyResult{Consumed: false}
	}
	for _, r := range characters {
		if unicode.IsDigit(r) {
			return s.selectByDigitLocked(r)
		}
		if r >= 'a' && r <= 'z' {
			s.preedit = append(s.preedit[:s.cursor], append([]rune{r}, s.preedit[s.cursor:]...)...)
			s.cursor++
			s.page = 0
			continue
		}
		if r >= 'A' && r <= 'Z' {
			lower := unicode.ToLower(r)
			s.preedit = append(s.preedit[:s.cursor], append([]rune{lower}, s.preedit[s.cursor:]...)...)
			s.cursor++
			s.page = 0
			continue
		}
		return &KeyResult{Consumed: false}
	}
	return s.refreshLocked()
}

func hasLetterOrDigit(characters string) bool {
	for _, r := range characters {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func (s *Session) backspaceLocked() *KeyResult {
	if len(s.confirmed) > 0 {
		last := s.confirmed[len(s.confirmed)-1]
		s.confirmed = s.confirmed[:len(s.confirmed)-1]
		s.preedit = append([]rune(last.Keys), s.preedit...)
		s.cursor = len(s.preedit)
		s.page = 0
		return s.refreshLocked()
	}
	if len(s.preedit) > 0 && s.cursor > 0 {
		s.preedit = append(s.preedit[:s.cursor-1], s.preedit[s.cursor:]...)
		s.cursor--
		s.page = 0
		return s.refreshLocked()
	}
	if len(s.preedit) == 0 {
		return &KeyResult{Consumed: false}
	}
	return s.refreshLocked()
}

func (s *Session) cancelLocked() *KeyResult {
	if len(s.preedit) == 0 && len(s.confirmed) == 0 {
		return &KeyResult{Consumed: false}
	}
	s.clearCompositionLocked()
	return s.refreshLocked()
}

func (s *Session) commitTopCandidateLocked() *KeyResult {
	if len(s.preedit) == 0 {
		if len(s.confirmed) == 0 {
			return &KeyResult{Consumed: false}
		}
		return s.flushConfirmedLocked()
	}
	candidates := s.decodeLocked()
	if len(candidates) == 0 {
		return s.commitCompositionLocked()
	}
	return s.commitCandidateLocked(candidates[0])
}

func (s *Session) commitRawPreeditLocked() *KeyResult {
	return s.commitCompositionLocked()
}

func (s *Session) commitCompositionLocked() *KeyResult {
	text := s.confirmedTextLocked() + string(s.preedit)
	if text == "" {
		return &KeyResult{Consumed: false}
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

func (s *Session) flushConfirmedLocked() *KeyResult {
	text := s.confirmedTextLocked()
	if text == "" {
		return &KeyResult{Consumed: false}
	}
	s.confirmed = nil
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

func (s *Session) confirmedTextLocked() string {
	if len(s.confirmed) == 0 {
		return ""
	}
	var b strings.Builder
	for _, seg := range s.confirmed {
		b.WriteString(seg.Text)
	}
	return b.String()
}

func (s *Session) selectByDigitLocked(digit rune) *KeyResult {
	candidates := s.decodeLocked()
	if len(candidates) == 0 {
		return s.refreshLocked()
	}
	if digit < '1' || digit > '9' {
		return s.refreshLocked()
	}
	index := int(digit - '1')
	if index < 0 || index >= int(s.pageSizeLocked()) {
		return s.refreshLocked()
	}
	abs := int(s.page)*int(s.pageSizeLocked()) + index
	if abs < 0 || abs >= len(candidates) {
		return s.refreshLocked()
	}
	return s.commitCandidateLocked(candidates[abs])
}

func (s *Session) pageLocked(delta int) *KeyResult {
	list := s.decodeLocked()
	pageSize := int(s.pageSizeLocked())
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	maxPage := 0
	if len(list) > 0 {
		maxPage = (len(list) - 1) / pageSize
	}
	next := int(s.page) + delta
	if next < 0 {
		next = 0
	}
	if next > maxPage {
		next = maxPage
	}
	s.page = uint32(next)
	return s.refreshLocked()
}

func (s *Session) refreshLocked() *KeyResult {
	s.generation++
	list := s.decodeLocked()
	return &KeyResult{
		Composition: s.compositionLocked(),
		Candidates:  s.candidateListLocked(list),
		Consumed:    true,
	}
}

func (s *Session) decodeLocked() []engine.Candidate {
	pinyin := string(s.preedit)
	if pinyin == "" {
		return nil
	}
	decoded, err := s.engine.Decode(pinyin)
	if err != nil {
		return nil
	}
	return decoded
}

func (s *Session) compositionLocked() *imev1.CompositionState {
	confirmed := s.confirmedTextLocked()
	preedit := confirmed + string(s.preedit)
	cursor := len([]rune(confirmed)) + s.cursor
	return &imev1.CompositionState{
		SessionId:  s.id,
		Preedit:    preedit,
		Cursor:     uint32(cursor),
		Generation: s.generation,
	}
}

func (s *Session) candidateListLocked(list []engine.Candidate) *imev1.CandidateList {
	pageSize := s.pageSizeLocked()
	out := &imev1.CandidateList{
		SessionId:  s.id,
		Generation: s.generation,
		Page:       s.page,
		PageSize:   pageSize,
	}
	if len(list) == 0 {
		return out
	}
	start := int(s.page) * int(pageSize)
	if start >= len(list) {
		s.page = 0
		start = 0
		out.Page = 0
	}
	end := start + int(pageSize)
	if end > len(list) {
		end = len(list)
	}
for i, item := range list[start:end] {
			out.Candidates = append(out.Candidates, &imev1.Candidate{
				Text:    fmt.Sprintf("%d.%s", i+1, item.Text),
				Comment: item.Comment,
				Score:   item.Score,
			})
		}
	return out
}

func (s *Session) pageSizeLocked() uint32 {
	if s.pageSize == 0 {
		return defaultPageSize
	}
	return s.pageSize
}

func (s *Session) clearCompositionLocked() {
	s.preedit = nil
	s.cursor = 0
	s.page = 0
	s.confirmed = nil
}

func (s *Session) SelectCandidate(generation uint64, index uint32) (*KeyResult, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if generation != 0 && generation != s.generation {
		return nil, false
	}
	candidates := s.decodeLocked()
	abs := int(s.page)*int(s.pageSizeLocked()) + int(index)
	if abs < 0 || abs >= len(candidates) {
		return s.refreshLocked(), true
	}
	return s.commitCandidateLocked(candidates[abs]), true
}

func (s *Session) commitCandidateLocked(cand engine.Candidate) *KeyResult {
	consumedKeys := cand.ConsumedKeys
	if consumedKeys <= 0 || consumedKeys > len(s.preedit) {
		if cand.Incomplete {
			consumedKeys = 0
		} else {
			consumedKeys = len(s.preedit)
		}
	}
	if consumedKeys <= 0 {
		return s.refreshLocked()
	}
	covered := string(s.preedit[:consumedKeys])
	remainder := ""
	if consumedKeys < len(s.preedit) {
		remainder = string(s.preedit[consumedKeys:])
	}
	s.confirmed = append(s.confirmed, confirmedSegment{Text: cand.Text, Keys: covered})
	s.preedit = []rune(remainder)
	s.cursor = len(s.preedit)
	s.page = 0
	s.generation++
	if remainder != "" {
		list := s.decodeLocked()
		return &KeyResult{
			Composition: s.compositionLocked(),
			Candidates:  s.candidateListLocked(list),
			Consumed:    true,
		}
	}
	text := s.confirmedTextLocked()
	s.confirmed = nil
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

func IsLetterInput(characters string) bool {
	return strings.IndexFunc(characters, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'))
	}) < 0 && characters != ""
}
