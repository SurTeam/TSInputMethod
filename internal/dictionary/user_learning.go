package dictionary

import (
	"strings"
	"sync"
	"unicode"
)

const (
	defaultLearnMinRepeats = 2
	maxLearnTextRunes      = 12
)

type PhraseLearner struct {
	mu         sync.Mutex
	user       *UserDictionary
	counts     map[string]int
	minRepeats int
}

func NewPhraseLearner(user *UserDictionary, minRepeats int) *PhraseLearner {
	if minRepeats <= 0 {
		minRepeats = defaultLearnMinRepeats
	}
	return &PhraseLearner{
		user:       user,
		counts:     make(map[string]int),
		minRepeats: minRepeats,
	}
}

func (l *PhraseLearner) SetMinRepeats(n int) {
	if l == nil {
		return
	}
	if n <= 0 {
		n = defaultLearnMinRepeats
	}
	l.mu.Lock()
	l.minRepeats = n
	l.mu.Unlock()
}

func (l *PhraseLearner) ObserveCommittedPhrase(text string, syllables []string) (UserEntry, bool, error) {
	if l == nil || l.user == nil {
		return UserEntry{}, false, nil
	}
	text = strings.TrimSpace(text)
	if !IsAutoLearnablePhrase(text, syllables) {
		return UserEntry{}, false, nil
	}
	syllables, err := normalizeUserSyllables(syllables)
	if err != nil {
		return UserEntry{}, false, nil
	}
	if _, ok := l.user.LookupEntry(text, syllables); ok {
		l.user.RecordUsage(text, syllables)
		return UserEntry{}, false, nil
	}
	key := text + "\x00" + joinKey(syllables)
	l.mu.Lock()
	l.counts[key]++
	count := l.counts[key]
	need := l.minRepeats
	l.mu.Unlock()
	if count < need {
		return UserEntry{}, false, nil
	}
	entry, err := l.user.AddEntry(text, syllables, UserSourceLearned, defaultUserWeight)
	if err != nil {
		if err == ErrUserEntryExists {
			l.user.RecordUsage(text, syllables)
			return UserEntry{}, false, nil
		}
		return UserEntry{}, false, err
	}
	return entry, true, nil
}

func IsAutoLearnablePhrase(text string, syllables []string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	runess := []rune(text)
	if len(runess) < 2 || len(runess) > maxLearnTextRunes {
		return false
	}
	if len(syllables) < 2 || len(syllables) > maxUserSyllables {
		return false
	}
	if len(runess) != len(syllables) {
		return false
	}
	for _, r := range runess {
		if !unicode.In(r, unicode.Han) {
			return false
		}
	}
	if err := ValidateUserSyllables(syllables); err != nil {
		return false
	}
	return true
}
