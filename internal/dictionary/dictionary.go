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
	data           []byte
	closeData      func() error
	syllableList   []string
	syllables      map[string]struct{}
	maxSyllableLen int
	entryOff       []uint32
	keyBlob        []byte
	keyOff         []uint32
	postStart      []uint32
	postings       []uint32
	entries        []Entry
	byKey          map[string][]int
	sortedKeys     []string
}

func LoadFile(path string) (*Dictionary, error) {
	data, closer, err := mapDictFile(path)
	if err != nil {
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		closer = nil
	}
	dict, err := loadCompact(data, closer)
	if err != nil {
		if closer != nil {
			_ = closer()
		}
		return nil, err
	}
	return dict, nil
}

func Load(data []byte) (*Dictionary, error) {
	owned := append([]byte(nil), data...)
	return loadCompact(owned, nil)
}

func loadCompact(data []byte, closer func() error) (*Dictionary, error) {
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
	syllableList := make([]string, 0, syllableCount)
	for i := 0; i < syllableCount; i++ {
		if offset >= len(data) {
			return nil, fmt.Errorf("dictionary: truncated syllables")
		}
		n := int(data[offset])
		offset++
		if offset+n > len(data) {
			return nil, fmt.Errorf("dictionary: truncated syllable")
		}
		syllableList = append(syllableList, string(data[offset:offset+n]))
		offset += n
	}
	syllableSet := make(map[string]struct{}, syllableCount)
	maxLen := 0
	for _, s := range syllableList {
		syllableSet[s] = struct{}{}
		if len(s) > maxLen {
			maxLen = len(s)
		}
	}
	entryOff := make([]uint32, entryCount)
	byKey := make(map[string][]uint32, entryCount/2+1)
	for i := 0; i < entryCount; i++ {
		if offset+6 > len(data) {
			return nil, fmt.Errorf("dictionary: truncated entry header")
		}
		entryOff[i] = uint32(offset)
		textLen := int(binary.LittleEndian.Uint16(data[offset:]))
		offset += 2
		offset += 4
		if offset+textLen+1 > len(data) {
			return nil, fmt.Errorf("dictionary: truncated entry text")
		}
		offset += textLen
		nSyl := int(data[offset])
		offset++
		if offset+2*nSyl > len(data) {
			return nil, fmt.Errorf("dictionary: truncated entry syllables")
		}
		item := make([]string, nSyl)
		for j := 0; j < nSyl; j++ {
			idx := int(binary.LittleEndian.Uint16(data[offset:]))
			offset += 2
			if idx < 0 || idx >= len(syllableList) {
				return nil, fmt.Errorf("dictionary: syllable index out of range")
			}
			item[j] = syllableList[idx]
		}
		key := joinKey(item)
		byKey[key] = append(byKey[key], uint32(i))
	}
	keys := make([]string, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	keyBlob := make([]byte, 0, len(keys)*12)
	keyOff := make([]uint32, 0, len(keys))
	postings := make([]uint32, 0, entryCount)
	postStart := make([]uint32, 0, len(keys)+1)
	for _, key := range keys {
		idxs := byKey[key]
		sort.SliceStable(idxs, func(a, b int) bool {
			return betterCompactEntry(data, entryOff, idxs[a], idxs[b])
		})
		if len(key) > 65535 {
			return nil, fmt.Errorf("dictionary: key too long")
		}
		keyOff = append(keyOff, uint32(len(keyBlob)))
		var lenBuf [2]byte
		binary.LittleEndian.PutUint16(lenBuf[:], uint16(len(key)))
		keyBlob = append(keyBlob, lenBuf[:]...)
		keyBlob = append(keyBlob, key...)
		postStart = append(postStart, uint32(len(postings)))
		postings = append(postings, idxs...)
	}
	postStart = append(postStart, uint32(len(postings)))
	byKey = nil
	keys = nil
	return &Dictionary{
		data:           data,
		closeData:      closer,
		syllableList:   syllableList,
		syllables:      syllableSet,
		maxSyllableLen: maxLen,
		entryOff:       entryOff,
		keyBlob:        keyBlob,
		keyOff:         keyOff,
		postStart:      postStart,
		postings:       postings,
	}, nil
}

func betterCompactEntry(data []byte, entryOff []uint32, a, b uint32) bool {
	wa, ta := peekEntry(data, entryOff[a])
	wb, tb := peekEntry(data, entryOff[b])
	if wa != wb {
		return wa > wb
	}
	return ta < tb
}

func peekEntry(data []byte, off uint32) (weight uint32, text string) {
	o := int(off)
	textLen := int(binary.LittleEndian.Uint16(data[o:]))
	o += 2
	weight = binary.LittleEndian.Uint32(data[o:])
	o += 4
	text = string(data[o : o+textLen])
	return weight, text
}

func (d *Dictionary) entryAt(idx uint32) Entry {
	if d == nil || int(idx) >= len(d.entryOff) {
		return Entry{}
	}
	o := int(d.entryOff[idx])
	data := d.data
	textLen := int(binary.LittleEndian.Uint16(data[o:]))
	o += 2
	weight := binary.LittleEndian.Uint32(data[o:])
	o += 4
	text := string(data[o : o+textLen])
	o += textLen
	nSyl := int(data[o])
	o++
	syl := make([]string, nSyl)
	for j := 0; j < nSyl; j++ {
		sIdx := int(binary.LittleEndian.Uint16(data[o:]))
		o += 2
		syl[j] = d.syllableList[sIdx]
	}
	return Entry{Text: text, Syllables: syl, Weight: weight}
}

func (d *Dictionary) keyString(i int) string {
	o := int(d.keyOff[i])
	n := int(binary.LittleEndian.Uint16(d.keyBlob[o:]))
	o += 2
	return string(d.keyBlob[o : o+n])
}

func (d *Dictionary) findExactKey(key string) int {
	n := len(d.keyOff)
	i := sort.Search(n, func(i int) bool {
		return d.keyString(i) >= key
	})
	if i < n && d.keyString(i) == key {
		return i
	}
	return -1
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
	if d == nil {
		return false
	}
	_, ok := d.syllables[s]
	return ok
}

func (d *Dictionary) MaxSyllableLen() int {
	if d == nil {
		return 0
	}
	return d.maxSyllableLen
}

func (d *Dictionary) Lookup(syllables []string) []Entry {
	if d == nil || len(syllables) == 0 {
		return nil
	}
	if d.entries != nil {
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
	keyIdx := d.findExactKey(joinKey(syllables))
	if keyIdx < 0 {
		return nil
	}
	start := d.postStart[keyIdx]
	end := d.postStart[keyIdx+1]
	out := make([]Entry, 0, end-start)
	for _, idx := range d.postings[start:end] {
		out = append(out, d.entryAt(idx))
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
	if d.entries != nil {
		return d.lookupPrefixEager(prefix, incomplete)
	}
	return d.lookupPrefixCompact(prefix, incomplete)
}

func (d *Dictionary) lookupPrefixEager(prefix, incomplete string) []Entry {
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
	return flattenPrefixBuckets(buckets)
}

func (d *Dictionary) lookupPrefixCompact(prefix, incomplete string) []Entry {
	buckets := make([][]Entry, prefixMaxLens+1)
	nKeys := len(d.keyOff)
	i := sort.Search(nKeys, func(i int) bool {
		return d.keyString(i) >= prefix
	})
	for ; i < nKeys; i++ {
		key := d.keyString(i)
		if key != prefix && !strings.HasPrefix(key, prefix) {
			if len(key) < len(prefix) || key[:len(prefix)] > prefix {
				break
			}
			continue
		}
		if key != prefix && !strings.HasPrefix(key, prefix+"'") && !(incomplete != "" && strings.HasPrefix(key, prefix)) {
			continue
		}
		start := d.postStart[i]
		end := d.postStart[i+1]
		for _, idx := range d.postings[start:end] {
			entry := d.entryAt(idx)
			n := len(entry.Syllables)
			if n == 0 || n > prefixMaxLens {
				continue
			}
			buckets[n] = insertPrefixEntry(buckets[n], entry, prefixPerLen)
		}
	}
	return flattenPrefixBuckets(buckets)
}

func flattenPrefixBuckets(buckets [][]Entry) []Entry {
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
	if d.entries != nil {
		return len(d.entries)
	}
	return len(d.entryOff)
}

func (d *Dictionary) Close() error {
	if d == nil || d.closeData == nil {
		return nil
	}
	err := d.closeData()
	d.closeData = nil
	d.data = nil
	return err
}

func joinKey(syllables []string) string {
	return strings.Join(syllables, "'")
}
