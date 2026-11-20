package dictionary

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

var validSyllableFn = func(s string) bool {
	if s == "" || len(s) > 8 {
		return false
	}
	for _, r := range s {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return true
}

func SetValidSyllableFunc(fn func(string) bool) {
	if fn == nil {
		return
	}
	validSyllableFn = fn
}

const (
	UserSourceManual       = "manual"
	UserSourceLearned      = "learned"
	defaultUserWeight      = uint32(1000)
	maxUserTextRunes       = 16
	maxUserSyllables       = 12
	defaultMaxUserEntries  = -1
	maxUsageCount          = uint32(100000)
)

var (
	ErrInvalidUserEntry = errors.New("dictionary: invalid user entry")
	ErrUserEntryExists  = errors.New("dictionary: user entry exists")
	ErrUserEntryMissing = errors.New("dictionary: user entry missing")
	ErrUserDictFull     = errors.New("dictionary: user dictionary full")
)

type UserEntry struct {
	Text       string
	Syllables  []string
	Weight     uint32
	UsageCount uint32
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Source     string
	Enabled    bool
}

type UserDictionary struct {
	mu         sync.RWMutex
	entries    []UserEntry
	byKey      map[string][]int
	byText     map[string][]int
	version    uint64
	path       string
	maxEntries int
}

func NewUserDictionary() *UserDictionary {
	return &UserDictionary{
		byKey:      make(map[string][]int),
		byText:     make(map[string][]int),
		maxEntries: defaultMaxUserEntries,
	}
}

func (u *UserDictionary) MaxEntries() int {
	if u == nil {
		return defaultMaxUserEntries
	}
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.maxEntries
}

func (u *UserDictionary) SetMaxEntries(limit int) {
	if u == nil {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	if limit == 0 {
		limit = defaultMaxUserEntries
	}
	u.maxEntries = limit
}

func (u *UserDictionary) Version() uint64 {
	if u == nil {
		return 0
	}
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.version
}

func (u *UserDictionary) Path() string {
	if u == nil {
		return ""
	}
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.path
}

func (u *UserDictionary) SetPath(path string) {
	if u == nil {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	u.path = path
}

func (u *UserDictionary) EntryCount() int {
	if u == nil {
		return 0
	}
	u.mu.RLock()
	defer u.mu.RUnlock()
	return len(u.entries)
}

func ValidateUserSyllables(syllables []string) error {
	if len(syllables) == 0 || len(syllables) > maxUserSyllables {
		return ErrInvalidUserEntry
	}
	for _, syl := range syllables {
		syl = strings.ToLower(strings.TrimSpace(syl))
		if syl == "" || !validSyllableFn(syl) {
			return fmt.Errorf("%w: syllable %q", ErrInvalidUserEntry, syl)
		}
	}
	return nil
}

func normalizeUserText(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ErrInvalidUserEntry
	}
	runess := []rune(text)
	if len(runess) == 0 || len(runess) > maxUserTextRunes {
		return "", ErrInvalidUserEntry
	}
	for _, r := range runess {
		if unicode.IsControl(r) {
			return "", ErrInvalidUserEntry
		}
	}
	return text, nil
}

func normalizeUserSyllables(syllables []string) ([]string, error) {
	out := make([]string, 0, len(syllables))
	for _, syl := range syllables {
		out = append(out, strings.ToLower(strings.TrimSpace(syl)))
	}
	if err := ValidateUserSyllables(out); err != nil {
		return nil, err
	}
	return out, nil
}

func (u *UserDictionary) AddEntry(text string, syllables []string, source string, weight uint32) (UserEntry, error) {
	if u == nil {
		return UserEntry{}, ErrInvalidUserEntry
	}
	text, err := normalizeUserText(text)
	if err != nil {
		return UserEntry{}, err
	}
	syllables, err = normalizeUserSyllables(syllables)
	if err != nil {
		return UserEntry{}, err
	}
	if source != UserSourceManual && source != UserSourceLearned {
		source = UserSourceManual
	}
	if weight == 0 {
		weight = defaultUserWeight
	}
	now := time.Now().UTC()
	u.mu.Lock()
	defer u.mu.Unlock()
	key := joinKey(syllables)
	for _, idx := range u.byKey[key] {
		if u.entries[idx].Text == text {
			return UserEntry{}, ErrUserEntryExists
		}
	}
	if u.maxEntries > 0 && len(u.entries) >= u.maxEntries {
		return UserEntry{}, ErrUserDictFull
	}
	entry := UserEntry{
		Text:       text,
		Syllables:  append([]string(nil), syllables...),
		Weight:     weight,
		UsageCount: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
		Source:     source,
		Enabled:    true,
	}
	idx := len(u.entries)
	u.entries = append(u.entries, entry)
	u.byKey[key] = append(u.byKey[key], idx)
	u.byText[text] = append(u.byText[text], idx)
	u.version++
	return entry, nil
}

func (u *UserDictionary) RemoveEntry(text string, syllables []string) error {
	if u == nil {
		return ErrUserEntryMissing
	}
	text, err := normalizeUserText(text)
	if err != nil {
		return err
	}
	syllables, err = normalizeUserSyllables(syllables)
	if err != nil {
		return err
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	key := joinKey(syllables)
	idxs := u.byKey[key]
	found := -1
	for _, idx := range idxs {
		if u.entries[idx].Text == text {
			found = idx
			break
		}
	}
	if found < 0 {
		return ErrUserEntryMissing
	}
	u.deleteIndexLocked(found)
	u.version++
	return nil
}

func (u *UserDictionary) UpdateEntry(text string, syllables []string, weight uint32, enabled *bool) (UserEntry, error) {
	if u == nil {
		return UserEntry{}, ErrUserEntryMissing
	}
	text, err := normalizeUserText(text)
	if err != nil {
		return UserEntry{}, err
	}
	syllables, err = normalizeUserSyllables(syllables)
	if err != nil {
		return UserEntry{}, err
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	key := joinKey(syllables)
	for _, idx := range u.byKey[key] {
		if u.entries[idx].Text != text {
			continue
		}
		if weight > 0 {
			u.entries[idx].Weight = weight
		}
		if enabled != nil {
			u.entries[idx].Enabled = *enabled
		}
		u.entries[idx].UpdatedAt = time.Now().UTC()
		u.version++
		return cloneUserEntry(u.entries[idx]), nil
	}
	return UserEntry{}, ErrUserEntryMissing
}

func (u *UserDictionary) LookupEntry(text string, syllables []string) (UserEntry, bool) {
	if u == nil {
		return UserEntry{}, false
	}
	text = strings.TrimSpace(text)
	syllables, err := normalizeUserSyllables(syllables)
	if err != nil {
		return UserEntry{}, false
	}
	u.mu.RLock()
	defer u.mu.RUnlock()
	key := joinKey(syllables)
	for _, idx := range u.byKey[key] {
		if u.entries[idx].Text == text {
			return cloneUserEntry(u.entries[idx]), true
		}
	}
	return UserEntry{}, false
}

func (u *UserDictionary) ListEntries() []UserEntry {
	if u == nil {
		return nil
	}
	u.mu.RLock()
	defer u.mu.RUnlock()
	out := make([]UserEntry, 0, len(u.entries))
	for _, entry := range u.entries {
		out = append(out, cloneUserEntry(entry))
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Text == out[j].Text {
			return joinKey(out[i].Syllables) < joinKey(out[j].Syllables)
		}
		return out[i].Text < out[j].Text
	})
	return out
}

func (u *UserDictionary) ClearLearnedEntries() int {
	if u == nil {
		return 0
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	keep := make([]UserEntry, 0, len(u.entries))
	for _, entry := range u.entries {
		if entry.Source != UserSourceLearned {
			keep = append(keep, entry)
		}
	}
	removed := len(u.entries) - len(keep)
	u.replaceAllLocked(keep)
	if removed > 0 {
		u.version++
	}
	return removed
}

func (u *UserDictionary) ClearAllEntries() int {
	if u == nil {
		return 0
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	n := len(u.entries)
	u.replaceAllLocked(nil)
	if n > 0 {
		u.version++
	}
	return n
}

func (u *UserDictionary) Lookup(syllables []string) []Entry {
	if u == nil || len(syllables) == 0 {
		return nil
	}
	u.mu.RLock()
	defer u.mu.RUnlock()
	idxs := u.byKey[joinKey(syllables)]
	if len(idxs) == 0 {
		return nil
	}
	out := make([]Entry, 0, len(idxs))
	for _, idx := range idxs {
		entry := u.entries[idx]
		if !entry.Enabled {
			continue
		}
		out = append(out, Entry{
			Text:      entry.Text,
			Syllables: append([]string(nil), entry.Syllables...),
			Weight:    effectiveUserWeight(entry),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Weight == out[j].Weight {
			return out[i].Text < out[j].Text
		}
		return out[i].Weight > out[j].Weight
	})
	return out
}

func (u *UserDictionary) LookupPrefix(syllables []string, incomplete string) []Entry {
	if u == nil {
		return nil
	}
	prefix := joinKey(syllables)
	if incomplete != "" {
		if prefix != "" {
			prefix += "'"
		}
		prefix += incomplete
	}
	if prefix == "" {
		return nil
	}
	u.mu.RLock()
	defer u.mu.RUnlock()
	out := make([]Entry, 0)
	for key, idxs := range u.byKey {
		if key != prefix && !strings.HasPrefix(key, prefix) {
			continue
		}
		if key != prefix && !strings.HasPrefix(key, prefix+"'") && !(incomplete != "" && strings.HasPrefix(key, prefix)) {
			continue
		}
		for _, idx := range idxs {
			entry := u.entries[idx]
			if !entry.Enabled {
				continue
			}
			out = append(out, Entry{
				Text:      entry.Text,
				Syllables: append([]string(nil), entry.Syllables...),
				Weight:    effectiveUserWeight(entry),
			})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if len(out[i].Syllables) != len(out[j].Syllables) {
			return len(out[i].Syllables) < len(out[j].Syllables)
		}
		if out[i].Weight == out[j].Weight {
			return out[i].Text < out[j].Text
		}
		return out[i].Weight > out[j].Weight
	})
	return out
}

func (u *UserDictionary) RecordUsage(text string, syllables []string) {
	if u == nil {
		return
	}
	text = strings.TrimSpace(text)
	syllables, err := normalizeUserSyllables(syllables)
	if err != nil {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	key := joinKey(syllables)
	for _, idx := range u.byKey[key] {
		if u.entries[idx].Text != text {
			continue
		}
		if u.entries[idx].UsageCount < maxUsageCount {
			u.entries[idx].UsageCount++
		}
		u.entries[idx].UpdatedAt = time.Now().UTC()
		u.version++
		return
	}
}

func (u *UserDictionary) UsageCountForText(text string) uint32 {
	if u == nil {
		return 0
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	u.mu.RLock()
	defer u.mu.RUnlock()
	var best uint32
	for _, idx := range u.byText[text] {
		entry := u.entries[idx]
		if !entry.Enabled {
			continue
		}
		if entry.UsageCount > best {
			best = entry.UsageCount
		}
	}
	return best
}

func (u *UserDictionary) RecordUsageText(text string) {
	if u == nil || strings.TrimSpace(text) == "" {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	idxs := u.byText[text]
	if len(idxs) == 0 {
		return
	}
	for _, idx := range idxs {
		if !u.entries[idx].Enabled {
			continue
		}
		if u.entries[idx].UsageCount < maxUsageCount {
			u.entries[idx].UsageCount++
		}
		u.entries[idx].UpdatedAt = time.Now().UTC()
	}
	u.version++
}

func MergeEntries(system []Entry, user []Entry) []Entry {
	if len(user) == 0 {
		return system
	}
	if len(system) == 0 {
		return append([]Entry(nil), user...)
	}
	index := make(map[string]int, len(system))
	out := make([]Entry, len(system))
	copy(out, system)
	for i, entry := range out {
		index[entry.Text+"\x00"+joinKey(entry.Syllables)] = i
	}
	for _, entry := range user {
		key := entry.Text + "\x00" + joinKey(entry.Syllables)
		if i, ok := index[key]; ok {
			if entry.Weight > out[i].Weight {
				out[i].Weight = entry.Weight
			}
			continue
		}
		index[key] = len(out)
		out = append(out, entry)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Weight == out[j].Weight {
			return out[i].Text < out[j].Text
		}
		return out[i].Weight > out[j].Weight
	})
	return out
}

func effectiveUserWeight(entry UserEntry) uint32 {
	w := entry.Weight
	if w == 0 {
		w = defaultUserWeight
	}
	bonus := entry.UsageCount * 10
	if bonus > 50000 {
		bonus = 50000
	}
	if w > ^uint32(0)-bonus {
		return ^uint32(0)
	}
	return w + bonus
}

func cloneUserEntry(entry UserEntry) UserEntry {
	entry.Syllables = append([]string(nil), entry.Syllables...)
	return entry
}

func (u *UserDictionary) deleteIndexLocked(idx int) {
	last := len(u.entries) - 1
	removed := u.entries[idx]
	if idx != last {
		u.entries[idx] = u.entries[last]
	}
	u.entries = u.entries[:last]
	u.rebuildIndexesLocked()
	_ = removed
}

func (u *UserDictionary) replaceAllLocked(entries []UserEntry) {
	u.entries = append([]UserEntry(nil), entries...)
	for i := range u.entries {
		u.entries[i].Syllables = append([]string(nil), u.entries[i].Syllables...)
	}
	u.rebuildIndexesLocked()
}

func (u *UserDictionary) rebuildIndexesLocked() {
	u.byKey = make(map[string][]int, len(u.entries))
	u.byText = make(map[string][]int, len(u.entries))
	for i, entry := range u.entries {
		key := joinKey(entry.Syllables)
		u.byKey[key] = append(u.byKey[key], i)
		u.byText[entry.Text] = append(u.byText[entry.Text], i)
	}
}
