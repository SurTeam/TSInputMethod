package dictionary

import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Entry struct {
	Text      string
	Syllables []string
	Weight    uint32
}

type Dictionary struct {
	entries        []Entry
	byKey          map[string][]int
	sortedKeys     []string
	syllables      map[string]struct{}
	maxSyllableLen int
}

func LoadFile(path string) (*Dictionary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Load(data)
}

func Load(data []byte) (*Dictionary, error) {
	if len(data) < 16 || string(data[:8]) != magic {
		return nil, fmt.Errorf("dictionary: bad magic")
	}
	offset := 8
	version := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	offset += 4
	if version != formatVer {
		return nil, fmt.Errorf("dictionary: unsupported version %d", version)
	}
	if len(data) < offset+8 {
		return nil, fmt.Errorf("dictionary: truncated header")
	}
	syllableCount := int(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4
	entryCount := int(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4
	syllables := make([]string, 0, syllableCount)
	for i := 0; i < syllableCount; i++ {
		if offset >= len(data) {
			return nil, fmt.Errorf("dictionary: truncated syllables")
		}
		n := int(data[offset])
	offset++
		if offset+n > len(data) {
			return nil, fmt.Errorf("dictionary: truncated syllable")
		}
		syllables = append(syllables, string(data[offset:offset+n]))
	offset += n
	}
	entries := make([]Entry, 0, entryCount)
	byKey := make(map[string][]int, entryCount)
	syllableSet := make(map[string]struct{}, syllableCount)
	maxLen := 0
	for _, s := range syllables {
		syllableSet[s] = struct{}{}
		if len(s) > maxLen {
			maxLen = len(s)
		}
	}
	for i := 0; i < entryCount; i++ {
		if offset+6 > len(data) {
			return nil, fmt.Errorf("dictionary: truncated entry header")
		}
		textLen := int(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
		weight := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
		if offset+textLen+1 > len(data) {
			return nil, fmt.Errorf("dictionary: truncated entry text")
		}
		text := string(data[offset : offset+textLen])
	offset += textLen
		nSyl := int(data[offset])
	offset++
		if offset+2*nSyl > len(data) {
			return nil, fmt.Errorf("dictionary: truncated entry syllables")
		}
		itemSyllables := make([]string, nSyl)
		for j := 0; j < nSyl; j++ {
			idx := int(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
			if idx < 0 || idx >= len(syllables) {
				return nil, fmt.Errorf("dictionary: syllable index out of range")
			}
			itemSyllables[j] = syllables[idx]
		}
		entry := Entry{Text: text, Syllables: itemSyllables, Weight: weight}
		idx := len(entries)
		entries = append(entries, entry)
		key := joinKey(itemSyllables)
		byKey[key] = append(byKey[key], idx)
	}
	return finalize(entries, byKey, syllableSet, maxLen), nil
}

func FromEntries(entries []Entry) *Dictionary {
	byKey := make(map[string][]int, len(entries))
	syllableSet := make(map[string]struct{})
	maxLen := 0
	cloned := make([]Entry, len(entries))
	for i, entry := range entries {
		syl := append([]string(nil), entry.Syllables...)
		cloned[i] = Entry{Text: entry.Text, Syllables: syl, Weight: entry.Weight}
		for _, s := range syl {
			syllableSet[s] = struct{}{}
			if len(s) > maxLen {
				maxLen = len(s)
			}
		}
		key := joinKey(syl)
		byKey[key] = append(byKey[key], i)
	}
	return finalize(cloned, byKey, syllableSet, maxLen)
}

func finalize(entries []Entry, byKey map[string][]int, syllableSet map[string]struct{}, maxLen int) *Dictionary {
	keys := make([]string, 0, len(byKey))
	for key := range byKey {
		ints := byKey[key]
		sort.SliceStable(ints, func(i, j int) bool {
			if entries[ints[i]].Weight == entries[ints[j]].Weight {
				return entries[ints[i]].Text < entries[ints[j]].Text
			}
			return entries[ints[i]].Weight > entries[ints[j]].Weight
		})
		byKey[key] = ints
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return &Dictionary{
		entries:        entries,
		byKey:          byKey,
		sortedKeys:     keys,
		syllables:      syllableSet,
		maxSyllableLen: maxLen,
	}
}

func (d *Dictionary) HasSyllable(s string) bool {
	_, ok := d.syllables[s]
	return ok
}

func (d *Dictionary) MaxSyllableLen() int {
	return d.maxSyllableLen
}

func (d *Dictionary) Lookup(syllables []string) []Entry {
	if d == nil || len(syllables) == 0 {
		return nil
	}
	idxs := d.byKey[joinKey(syllables)]
	if len(idxs) == 0 {
		return nil
	}
	out := make([]Entry, len(idxs))
	for i, idx := range idxs {
		out[i] = d.entries[idx]
	}
	return out
}

func (d *Dictionary) LookupPrefix(syllables []string, incomplete string) []Entry {
	if d == nil {
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
	buckets := make([][]Entry, prefixMaxLens+1)
	i := sort.SearchStrings(d.sortedKeys, prefix)
	for ; i < len(d.sortedKeys); i++ {
		key := d.sortedKeys[i]
		if key != prefix && !strings.HasPrefix(key, prefix) {
			if len(key) < len(prefix) || key[:len(prefix)] > prefix {
				break
			}
			continue
		}
		if key != prefix && !strings.HasPrefix(key, prefix+"'") && !(incomplete != "" && strings.HasPrefix(key, prefix)) {
			continue
		}
		for _, idx := range d.byKey[key] {
			entry := d.entries[idx]
			n := len(entry.Syllables)
			if n == 0 || n > prefixMaxLens {
				continue
			}
			buckets[n] = insertPrefixEntry(buckets[n], entry, prefixPerLen)
		}
	}
	out := make([]Entry, 0, maxCandDef)
	for n := 1; n <= prefixMaxLens && len(out) < maxCandDef; n++ {
		for _, entry := range buckets[n] {
			out = append(out, entry)
			if len(out) >= maxCandDef {
				break
			}
		}
	}
	return out
}

func insertPrefixEntry(bucket []Entry, entry Entry, limit int) []Entry {
	if limit <= 0 {
		return bucket
	}
	if len(bucket) < limit {
		bucket = append(bucket, entry)
		for i := len(bucket) - 1; i > 0; i-- {
			if !betterPrefixEntry(bucket[i], bucket[i-1]) {
				break
			}
			bucket[i], bucket[i-1] = bucket[i-1], bucket[i]
		}
		return bucket
	}
	if !betterPrefixEntry(entry, bucket[limit-1]) {
		return bucket
	}
	bucket[limit-1] = entry
	for i := limit - 1; i > 0; i-- {
		if !betterPrefixEntry(bucket[i], bucket[i-1]) {
			break
		}
		bucket[i], bucket[i-1] = bucket[i-1], bucket[i]
	}
	return bucket
}

func betterPrefixEntry(a, b Entry) bool {
	if a.Weight != b.Weight {
		return a.Weight > b.Weight
	}
	return a.Text < b.Text
}

func (d *Dictionary) EntryCount() int {
	if d == nil {
		return 0
	}
	return len(d.entries)
}

func joinKey(syllables []string) string {
	return strings.Join(syllables, "'")
}
