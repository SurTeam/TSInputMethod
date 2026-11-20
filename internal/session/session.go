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
	keyLeftArrow    = 123
	keyRightArrow   = 124
	keyDownArrow    = 125
	keyUpArrow      = 126

	modShift   = uint32(1 << 17)
	modControl = uint32(1 << 18)
	modOption  = uint32(1 << 19)
	modCommand = uint32(1 << 20)
)

type confirmedSegment struct {
	Text      string
	Keys      string
	Syllables []string
}

type PhraseObserver interface {
	ObserveCommittedPhrase(text string, syllables []string) error
}

type FeedbackSink interface {
	AcceptFeedbackEvents([]FeedbackEvent)
}

type Session struct {
	mu               sync.Mutex
	id               string
	mode             imev1.InputMode
	preedit          []rune
	cursor           int
	confirmed        []confirmedSegment
	committedContext []rune
	contextWindow    int
	generation       uint64
	page             uint32
	pageSize         uint32
	selectedIndex    uint32
	literal          bool
	quoteOpen        bool
	singleOpen       bool
	engine           engine.Engine
	snapshot         *candidateSnapshot
	pendingFeedback  []pendingFeedback
	acceptedFeedback []FeedbackEvent
	rankingVersion   uint64
	userDictVersion  uint64
	phraseObserver   PhraseObserver
	feedbackSink     FeedbackSink
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
		id:            id,
		mode:          mode,
		engine:        eng,
		pageSize:      defaultPageSize,
		contextWindow: defaultContextWindow,
	}
}

func (s *Session) SetPhraseObserver(observer PhraseObserver) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.phraseObserver = observer
}

func (s *Session) SetUserDictVersion(version uint64) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userDictVersion = version
}

func (s *Session) SetFeedbackSink(sink FeedbackSink) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.feedbackSink = sink
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) Mode() imev1.InputMode {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mode
}

func (s *Session) CommittedContext() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return string(s.committedContext)
}

func (s *Session) ConfirmedComposition() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.confirmedTextLocked()
}

func (s *Session) EffectiveContext() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.effectiveContextLocked()
}

func (s *Session) ClearLearningContext() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.committedContext = nil
	s.pendingFeedback = nil
}

func (s *Session) TakeAcceptedFeedback() []FeedbackEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.acceptedFeedback) == 0 {
		return nil
	}
	out := append([]FeedbackEvent(nil), s.acceptedFeedback...)
	s.acceptedFeedback = nil
	return out
}

func (s *Session) PendingFeedbackCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.pendingFeedback)
}

func (s *Session) SnapshotGeneration() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.snapshot == nil {
		return 0
	}
	return s.snapshot.Generation
}

func (s *Session) SnapshotCandidateTexts() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.snapshot == nil {
		return nil
	}
	out := make([]string, len(s.snapshot.Candidates))
	for i, cand := range s.snapshot.Candidates {
		out[i] = cand.Text
	}
	return out
}

func (s *Session) InvalidateLearningContext() {
	s.ClearLearningContext()
}

func (s *Session) SetContextWindow(window int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contextWindow = clampContextWindow(window)
	if len(s.committedContext) > contextMax {
		s.committedContext = append([]rune(nil), s.committedContext[len(s.committedContext)-contextMax:]...)
	}
}

func (s *Session) effectiveContextLocked() string {
	return unicodeTail(joinCommittedAndConfirmed(s.committedContext, s.confirmedTextLocked()), s.contextWindow)
}

func (s *Session) appendCommittedLocked(text string) {
	if text == "" {
		return
	}
	s.committedContext = appendCommittedRunes(s.committedContext, text, contextMax)
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
	s.committedContext = nil
	s.pendingFeedback = nil
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
	if hasPassthroughModifier(ev.Modifiers) {
		return &KeyResult{Consumed: false}
	}
	if s.mode == imev1.InputMode_INPUT_MODE_ENGLISH {
		return &KeyResult{Consumed: false}
	}
	return s.handleChineseLocked(ev)
}

func hasPassthroughModifier(modifiers uint32) bool {
	return modifiers&(modControl|modOption|modCommand) != 0
}

func hasShift(modifiers uint32) bool {
	return modifiers&modShift != 0
}

func (s *Session) handleChineseLocked(ev *imev1.KeyEvent) *KeyResult {
	switch ev.KeyCode {
	case keyLeftArrow:
		return s.moveSelectionLocked(-1)
	case keyRightArrow:
		return s.moveSelectionLocked(1)
	case keyUpArrow:
		return s.pageIfComposingLocked(-1)
	case keyDownArrow:
		return s.pageIfComposingLocked(1)
	case keyDelete, keyForwardDel:
		return s.backspaceLocked()
	case keyEscape:
		return s.cancelLocked()
	case keySpace:
		return s.commitSelectedCandidateLocked()
	case keyReturn, keypadEnter:
		return s.commitRawPreeditLocked()
	}

	switch ev.Characters {
	case "-", "_":
		if s.hasCompositionLocked() {
			return s.pageLocked(-1)
		}
	case "=", "+":
		if s.hasCompositionLocked() {
			return s.pageLocked(1)
		}
	case " ":
		return s.commitSelectedCandidateLocked()
	case "\r", "\n":
		return s.commitRawPreeditLocked()
	}

	if ev.Characters == "" {
		return &KeyResult{Consumed: false}
	}

	if hasShift(ev.Modifiers) {
		return s.handleShiftInputLocked(ev.Characters)
	}
	return s.composeCharactersLocked(ev.Characters)
}

func (s *Session) handleShiftInputLocked(characters string) *KeyResult {
	if characters == "" {
		return &KeyResult{Consumed: false}
	}
	var letters []rune
	var symbols []rune
	for _, r := range characters {
		switch {
		case r >= 'A' && r <= 'Z':
			letters = append(letters, r)
		case r >= 'a' && r <= 'z':
			letters = append(letters, unicode.ToUpper(r))
		default:
			symbols = append(symbols, r)
		}
	}
	if len(letters) > 0 {
		if s.hasCompositionLocked() && !s.literal {
			return &KeyResult{Consumed: true, Composition: s.compositionLocked(), Candidates: s.candidateListLocked(s.currentCandidatesLocked())}
		}
		for _, r := range letters {
			s.preedit = append(s.preedit[:s.cursor], append([]rune{r}, s.preedit[s.cursor:]...)...)
			s.cursor++
		}
		s.literal = true
		s.page = 0
		s.selectedIndex = 0
		return s.refreshLocked()
	}
	if len(symbols) == 0 {
		return &KeyResult{Consumed: false}
	}
	if s.hasCompositionLocked() {
		return &KeyResult{Consumed: true, Composition: s.compositionLocked(), Candidates: s.candidateListLocked(s.currentCandidatesLocked())}
	}
	var mapped []rune
	for _, r := range symbols {
		if isPunctuationOrSymbol(r) {
			mapped = append(mapped, []rune(s.mapChinesePunctuationLocked(r))...)
			continue
		}
		mapped = append(mapped, r)
	}
	return s.commitTextLocked(string(mapped))
}

func (s *Session) composeCharactersLocked(characters string) *KeyResult {
	if characters == "" {
		return &KeyResult{Consumed: false}
	}
	for _, r := range characters {
		if unicode.IsDigit(r) {
			if !s.hasCompositionLocked() {
				return s.commitTextLocked(string(r))
			}
			return s.selectByDigitLocked(r)
		}
		if r >= 'a' && r <= 'z' {
			if s.literal {
				s.literal = false
				s.preedit = nil
				s.cursor = 0
				s.confirmed = nil
				s.page = 0
				s.selectedIndex = 0
			}
			s.preedit = append(s.preedit[:s.cursor], append([]rune{r}, s.preedit[s.cursor:]...)...)
			s.cursor++
			s.page = 0
			s.selectedIndex = 0
			continue
		}
		if r >= 'A' && r <= 'Z' {
			lower := unicode.ToLower(r)
			if s.literal {
				s.literal = false
				s.preedit = nil
				s.cursor = 0
				s.confirmed = nil
				s.page = 0
				s.selectedIndex = 0
			}
			s.preedit = append(s.preedit[:s.cursor], append([]rune{lower}, s.preedit[s.cursor:]...)...)
			s.cursor++
			s.page = 0
			s.selectedIndex = 0
			continue
		}
			if isPunctuationOrSymbol(r) {
				if s.hasCompositionLocked() {
					return &KeyResult{Consumed: true, Composition: s.compositionLocked(), Candidates: s.candidateListLocked(s.currentCandidatesLocked())}
				}
				return s.commitTextLocked(s.mapChinesePunctuationLocked(r))
			}
		return &KeyResult{Consumed: false}
	}
	return s.refreshLocked()
}

func (s *Session) hasCompositionLocked() bool {
	return len(s.preedit) > 0 || len(s.confirmed) > 0
}

func isPunctuationOrSymbol(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
		return false
	}
	if unicode.IsControl(r) {
		return false
	}
	return unicode.IsPunct(r) || unicode.IsSymbol(r) || (r >= 0x21 && r <= 0x2F) || (r >= 0x3A && r <= 0x40) || (r >= 0x5B && r <= 0x60) || (r >= 0x7B && r <= 0x7E)
}

func (s *Session) mapChinesePunctuationLocked(r rune) string {
	switch r {
	case '"':
		if !s.quoteOpen {
			s.quoteOpen = true
			return "“"
		}
		s.quoteOpen = false
		return "”"
	case '\'':
		if !s.singleOpen {
			s.singleOpen = true
			return "‘"
		}
		s.singleOpen = false
		return "’"
	}
	if mapped, ok := halfShapePunctuation[r]; ok {
		return mapped
	}
	return string(r)
}

var halfShapePunctuation = map[rune]string{
	',':  "，",
	'.':  "。",
	'<':  "《",
	'>':  "》",
	'/':  "/",
	'?':  "？",
	';':  "；",
	':':  "：",
	'\\': "、",
	'|':  "|",
	'`':  "·",
	'~':  "~",
	'!':  "！",
	'@':  "@",
	'#':  "#",
	'%':  "%",
	'$':  "¥",
	'^':  "……",
	'&':  "&",
	'*':  "*",
	'(':  "（",
	')':  "）",
	'-':  "-",
	'_':  "——",
	'+':  "+",
	'=':  "=",
	'[':  "【",
	']':  "】",
	'{':  "「",
	'}':  "」",
}

func (s *Session) backspaceLocked() *KeyResult {
	if len(s.confirmed) > 0 {
		last := s.confirmed[len(s.confirmed)-1]
		s.confirmed = s.confirmed[:len(s.confirmed)-1]
		s.preedit = append([]rune(last.Keys), s.preedit...)
		s.cursor = len(s.preedit)
		s.page = 0
		s.selectedIndex = 0
		s.literal = false
		s.dropLastPendingLocked()
		return s.refreshLocked()
	}
	if len(s.preedit) > 0 && s.cursor > 0 {
		s.preedit = append(s.preedit[:s.cursor-1], s.preedit[s.cursor:]...)
		s.cursor--
		s.page = 0
		s.selectedIndex = 0
		if len(s.preedit) == 0 {
			s.literal = false
		}
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
	s.pendingFeedback = nil
	s.clearCompositionLocked()
	return s.refreshLocked()
}

func (s *Session) commitSelectedCandidateLocked() *KeyResult {
	if len(s.preedit) == 0 {
		if len(s.confirmed) == 0 {
			return &KeyResult{Consumed: false}
		}
		return s.flushConfirmedLocked()
	}
	if s.literal {
		return s.commitCompositionLocked()
	}
	candidates := s.snapshotCandidatesLocked()
	if len(candidates) == 0 {
		return s.commitCompositionLocked()
	}
	abs := int(s.page)*int(s.pageSizeLocked()) + int(s.selectedIndex)
	if abs < 0 || abs >= len(candidates) {
		abs = int(s.page) * int(s.pageSizeLocked())
	}
	if abs < 0 || abs >= len(candidates) {
		return s.commitCompositionLocked()
	}
	return s.commitCandidateAtLocked(abs)
}

func (s *Session) commitTopCandidateLocked() *KeyResult {
	return s.commitSelectedCandidateLocked()
}

func (s *Session) commitRawPreeditLocked() *KeyResult {
	return s.commitCompositionLocked()
}

func (s *Session) commitTextLocked(text string) *KeyResult {
	if text == "" {
		return &KeyResult{Consumed: false}
	}
	s.clearCompositionLocked()
	s.appendCommittedLocked(text)
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

func (s *Session) commitCompositionLocked() *KeyResult {
	confirmedText := s.confirmedTextLocked()
	syllables := s.confirmedSyllablesLocked()
	text := confirmedText + string(s.preedit)
	if text == "" {
		return &KeyResult{Consumed: false}
	}
	s.acceptPendingFeedbackLocked()
	s.clearCompositionLocked()
	s.appendCommittedLocked(text)
	if confirmedText != "" && len(syllables) > 0 && text == confirmedText {
		s.observeCommittedLocked(text, syllables)
	}
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
	syllables := s.confirmedSyllablesLocked()
	if text == "" {
		return &KeyResult{Consumed: false}
	}
	s.confirmed = nil
	s.selectedIndex = 0
	s.acceptPendingFeedbackLocked()
	s.appendCommittedLocked(text)
	s.observeCommittedLocked(text, syllables)
	s.snapshot = nil
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
	candidates := s.snapshotCandidatesLocked()
	if len(candidates) == 0 {
		if !s.hasCompositionLocked() {
			return s.commitTextLocked(string(digit))
		}
		return s.refreshLocked()
	}
	if digit < '1' || digit > '9' {
		if !s.hasCompositionLocked() {
			return s.commitTextLocked(string(digit))
		}
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
	if s.literal {
		if index == 0 {
			return s.commitCompositionLocked()
		}
		return s.refreshLocked()
	}
	return s.commitCandidateAtLocked(abs)
}

func (s *Session) pageIfComposingLocked(delta int) *KeyResult {
	if !s.hasCompositionLocked() {
		return &KeyResult{Consumed: false}
	}
	return s.pageLocked(delta)
}

func (s *Session) pageLocked(delta int) *KeyResult {
	list := s.snapshotCandidatesLocked()
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
	s.selectedIndex = 0
	return s.bumpSnapshotPresentationLocked()
}

func (s *Session) moveSelectionLocked(delta int) *KeyResult {
	if !s.hasCompositionLocked() {
		return &KeyResult{Consumed: false}
	}
	list := s.snapshotCandidatesLocked()
	pageSize := int(s.pageSizeLocked())
	start := int(s.page) * pageSize
	if start >= len(list) {
		start = 0
		s.page = 0
	}
	end := start + pageSize
	if end > len(list) {
		end = len(list)
	}
	count := end - start
	if count <= 0 {
		return s.refreshLocked()
	}
	next := int(s.selectedIndex) + delta
	if next < 0 {
		next = 0
	}
	if next >= count {
		next = count - 1
	}
	s.selectedIndex = uint32(next)
	return s.bumpSnapshotPresentationLocked()
}

func (s *Session) refreshLocked() *KeyResult {
	s.generation++
	list := s.decodeCandidatesLocked()
	s.storeSnapshotLocked(list)
	return &KeyResult{
		Composition: s.compositionLocked(),
		Candidates:  s.candidateListLocked(list),
		Consumed:    true,
	}
}

func (s *Session) bumpSnapshotPresentationLocked() *KeyResult {
	list := s.snapshotCandidatesLocked()
	s.generation++
	if s.snapshot != nil {
		s.snapshot.Generation = s.generation
	}
	return &KeyResult{
		Composition: s.compositionLocked(),
		Candidates:  s.candidateListLocked(list),
		Consumed:    true,
	}
}

func (s *Session) currentCandidatesLocked() []engine.Candidate {
	return s.snapshotCandidatesLocked()
}

func (s *Session) decodeCandidatesLocked() []engine.Candidate {
	if s.literal {
		text := string(s.preedit)
		if text == "" {
			return nil
		}
		return []engine.Candidate{{
			Text:         text,
			ConsumedKeys: len(s.preedit),
		}}
	}
	return s.decodeLocked()
}

func (s *Session) snapshotCandidatesLocked() []engine.Candidate {
	if s.snapshot != nil && s.snapshot.Generation == s.generation && s.snapshot.Preedit == string(s.preedit) {
		return s.snapshot.Candidates
	}
	list := s.decodeCandidatesLocked()
	s.storeSnapshotLocked(list)
	return list
}

func (s *Session) storeSnapshotLocked(list []engine.Candidate) {
	s.snapshot = &candidateSnapshot{
		Preedit:          string(s.preedit),
		EffectiveContext: s.effectiveContextLocked(),
		Generation:       s.generation,
		Candidates:       cloneCandidates(list),
		RankingVersion:   s.rankingVersion,
		UserDictVersion:  s.userDictVersion,
	}
}

func (s *Session) decodeLocked() []engine.Candidate {
	pinyin := string(s.preedit)
	if pinyin == "" {
		return nil
	}
	ctx := s.effectiveContextLocked()
	if withCtx, ok := s.engine.(engine.ContextDecoder); ok {
		decoded, err := withCtx.DecodeWithContext(pinyin, ctx)
		if err != nil {
			return nil
		}
		return decoded
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
		SessionId:     s.id,
		Generation:    s.generation,
		Page:          s.page,
		PageSize:      pageSize,
		SelectedIndex: s.selectedIndex,
	}
	if len(list) == 0 {
		out.SelectedIndex = 0
		s.selectedIndex = 0
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
	pageCount := end - start
	if pageCount <= 0 {
		out.SelectedIndex = 0
		s.selectedIndex = 0
		return out
	}
	if int(s.selectedIndex) >= pageCount {
		s.selectedIndex = uint32(pageCount - 1)
		out.SelectedIndex = s.selectedIndex
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
	s.selectedIndex = 0
	s.confirmed = nil
	s.literal = false
	s.snapshot = nil
}

func (s *Session) SelectCandidate(generation uint64, index uint32) (*KeyResult, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if generation != 0 && generation != s.generation {
		return nil, false
	}
	candidates := s.snapshotCandidatesLocked()
	abs := int(s.page)*int(s.pageSizeLocked()) + int(index)
	if abs < 0 || abs >= len(candidates) {
		return s.refreshLocked(), true
	}
	s.selectedIndex = index
	if s.literal {
		return s.commitCompositionLocked(), true
	}
	return s.commitCandidateAtLocked(abs), true
}

func (s *Session) commitCandidateAtLocked(index int) *KeyResult {
	candidates := s.snapshotCandidatesLocked()
	if index < 0 || index >= len(candidates) {
		return s.refreshLocked()
	}
	s.recordPendingFromSnapshotLocked(index)
	return s.commitCandidateLocked(candidates[index])
}

func (s *Session) recordPendingFromSnapshotLocked(selectedIndex int) {
	if s.snapshot == nil || selectedIndex <= 0 || selectedIndex >= len(s.snapshot.Candidates) {
		return
	}
	if len(s.snapshot.Candidates) < 2 {
		return
	}
	top := s.snapshot.Candidates[0]
	selected := s.snapshot.Candidates[selectedIndex]
	if !compatibleFeedbackPair(top, selected) {
		return
	}
	s.pendingFeedback = append(s.pendingFeedback, pendingFeedback{
		SnapshotGeneration: s.snapshot.Generation,
		Preedit:            s.snapshot.Preedit,
		EffectiveContext:   s.snapshot.EffectiveContext,
		InputCode:          s.snapshot.Preedit,
		SelectedIndex:      selectedIndex,
		Selected:           selected,
		Top:                top,
		RankingVersion:     s.snapshot.RankingVersion,
		UserDictVersion:    s.snapshot.UserDictVersion,
	})
}

func (s *Session) dropLastPendingLocked() {
	if len(s.pendingFeedback) == 0 {
		return
	}
	s.pendingFeedback = s.pendingFeedback[:len(s.pendingFeedback)-1]
}

func (s *Session) acceptPendingFeedbackLocked() {
	if len(s.pendingFeedback) == 0 {
		return
	}
	batch := make([]FeedbackEvent, 0, len(s.pendingFeedback))
	for _, pending := range s.pendingFeedback {
		evt := FeedbackEvent{
			Preedit:          pending.Preedit,
			EffectiveContext: pending.EffectiveContext,
			InputCode:        pending.InputCode,
			SelectedIndex:    pending.SelectedIndex,
			Selected:         pending.Selected,
			Top:              pending.Top,
			Generation:       pending.SnapshotGeneration,
			RankingVersion:   pending.RankingVersion,
			UserDictVersion:  pending.UserDictVersion,
		}
		s.acceptedFeedback = append(s.acceptedFeedback, evt)
		batch = append(batch, evt)
	}
	s.pendingFeedback = nil
	if s.feedbackSink != nil && len(batch) > 0 {
		s.feedbackSink.AcceptFeedbackEvents(batch)
	}
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
		s.confirmed = append(s.confirmed, confirmedSegment{
			Text:      cand.Text,
			Keys:      covered,
			Syllables: append([]string(nil), cand.Syllables...),
		})
		s.preedit = []rune(remainder)
		s.cursor = len(s.preedit)
		s.page = 0
		s.selectedIndex = 0
		s.literal = false
		s.generation++
		if remainder != "" {
			list := s.decodeCandidatesLocked()
			s.storeSnapshotLocked(list)
			return &KeyResult{
				Composition: s.compositionLocked(),
				Candidates:  s.candidateListLocked(list),
				Consumed:    true,
			}
		}
		text := s.confirmedTextLocked()
		syllables := s.confirmedSyllablesLocked()
		s.confirmed = nil
		s.acceptPendingFeedbackLocked()
		s.appendCommittedLocked(text)
		s.observeCommittedLocked(text, syllables)
		s.snapshot = nil
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

func (s *Session) confirmedSyllablesLocked() []string {
	if len(s.confirmed) == 0 {
		return nil
	}
	var out []string
	for _, seg := range s.confirmed {
		out = append(out, seg.Syllables...)
	}
	return out
}

func (s *Session) observeCommittedLocked(text string, syllables []string) {
	if s.phraseObserver == nil || text == "" || len(syllables) == 0 {
		return
	}
	_ = s.phraseObserver.ObserveCommittedPhrase(text, syllables)
}

func IsLetterInput(characters string) bool {
	return strings.IndexFunc(characters, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'))
	}) < 0 && characters != ""
}
