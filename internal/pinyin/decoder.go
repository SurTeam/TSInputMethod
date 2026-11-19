package pinyin

import (
	"sort"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
)

type Candidate struct {
	Text       string
	Comment    string
	Score      int32
	Syllables  []string
	Incomplete bool
	Composed   bool
}

type Decoder struct {
	dict      *dictionary.Dictionary
	segmenter *Segmenter
	limit     int
}

type latticeCell struct {
	text     string
	score    uint64
	segments int
}

func NewDecoder(dict *dictionary.Dictionary, limit int) *Decoder {
	if limit <= 0 {
		limit = 96
	}
	return &Decoder{
		dict:      dict,
		segmenter: NewSegmenter(nil),
		limit:     limit,
	}
}

func (d *Decoder) Decode(input string) []Candidate {
	if d == nil || d.dict == nil || input == "" {
		return nil
	}
	segs := d.segmenter.Segment(input)
	seen := map[string]Candidate{}
	for _, seg := range segs {
		d.collectExact(seg.Syllables, seen)
		if len(seg.Syllables) > 0 {
			d.collectPartialWords(seg.Syllables, seen)
			d.collectLattice(seg.Syllables, seen)
		}
		if seg.Remainder != "" {
			d.collectPrefix(seg, seen)
		}
	}
	return d.finish(seen)
}

func (d *Decoder) DecodeSyllables(syllables []string, remainder string) []Candidate {
	if d == nil || d.dict == nil {
		return nil
	}
	if len(syllables) == 0 && remainder == "" {
		return nil
	}
	seen := map[string]Candidate{}
	if len(syllables) > 0 {
		d.collectExact(syllables, seen)
		d.collectPartialWords(syllables, seen)
	}
	if remainder != "" {
		d.collectPrefix(Segmentation{Syllables: syllables, Remainder: remainder}, seen)
	}
	return d.finish(seen)
}

func (d *Decoder) finish(seen map[string]Candidate) []Candidate {
	if len(seen) == 0 {
		return nil
	}
	byLen := map[int][]Candidate{}
	maxLen := 0
	for _, cand := range seen {
		n := len(cand.Syllables)
		if n > maxLen {
			maxLen = n
		}
		byLen[n] = append(byLen[n], cand)
	}
	for n, group := range byLen {
		sort.SliceStable(group, func(i, j int) bool {
			return betterCandidate(group[i], group[j])
		})
		byLen[n] = group
	}
	out := make([]Candidate, 0, d.limit)
	used := map[string]struct{}{}
	appendUnique := func(cands []Candidate, max int) {
		added := 0
		for _, cand := range cands {
			if max > 0 && added >= max {
				break
			}
			if _, ok := used[cand.Text]; ok {
				continue
			}
			used[cand.Text] = struct{}{}
			out = append(out, cand)
			added++
		}
	}
	filtered := map[int][]Candidate{}
	for n := maxLen; n >= 1; n-- {
		group := byLen[n]
		exact := make([]Candidate, 0, len(group))
		composed := make([]Candidate, 0, len(group))
		for _, cand := range group {
			if cand.Composed {
				composed = append(composed, cand)
			} else {
				exact = append(exact, cand)
			}
		}
		if n > 1 {
			exact = keepCompetitive(exact)
		}
		filtered[n] = exact
		appendUnique(exact, exactQuota(n, maxLen))
		if len(exact) == 0 {
			appendUnique(composed, composedQuota(n, maxLen))
		}
	}
	if remainder := byLen[0]; len(remainder) > 0 {
		appendUnique(remainder, 4)
	}
	if len(out) < d.limit {
		for n := 1; n <= maxLen && len(out) < d.limit; n++ {
			appendUnique(filtered[n], d.limit-len(out))
		}
		if remainder := byLen[0]; len(remainder) > 0 && len(out) < d.limit {
			appendUnique(remainder, d.limit-len(out))
		}
	}
	if len(out) > d.limit {
		out = out[:d.limit]
	}
	return out
}

func keepCompetitive(cands []Candidate) []Candidate {
	if len(cands) <= 1 {
		return cands
	}
	best := cands[0].Score
	if best <= 0 {
		return cands
	}
	out := cands[:0]
	for _, cand := range cands {
		if cand.Score*20 >= best {
			out = append(out, cand)
		}
	}
	if len(out) == 0 {
		return cands[:1]
	}
	return out
}

func exactQuota(n, maxLen int) int {
	if n <= 1 {
		return 12
	}
	if n == maxLen {
		return 2
	}
	if n == 2 {
		return 2
	}
	return 2
}

func composedQuota(n, maxLen int) int {
	if n == maxLen {
		return 2
	}
	return 0
}

func betterCandidate(a, b Candidate) bool {
	if a.Incomplete != b.Incomplete {
		return !a.Incomplete && b.Incomplete
	}
	si := len(a.Syllables)
	sj := len(b.Syllables)
	if si != sj {
		return si > sj
	}
	if a.Composed != b.Composed {
		return !a.Composed && b.Composed
	}
	if a.Score != b.Score {
		return a.Score > b.Score
	}
	ri := len([]rune(a.Text))
	rj := len([]rune(b.Text))
	if ri != rj {
		return ri > rj
	}
	return a.Text < b.Text
}

func (d *Decoder) collectExact(syllables []string, seen map[string]Candidate) {
	if len(syllables) == 0 {
		return
	}
	for _, entry := range d.dict.Lookup(syllables) {
		addCandidate(seen, Candidate{
			Text:      entry.Text,
			Comment:   joinComment(syllables),
			Score:     int32(clampScore(entry.Weight)),
			Syllables: append([]string(nil), syllables...),
		})
	}
}

func (d *Decoder) collectPrefix(seg Segmentation, seen map[string]Candidate) {
	entries := d.dict.LookupPrefix(seg.Syllables, seg.Remainder)
	for i, entry := range entries {
		if i >= d.limit {
			break
		}
		addCandidate(seen, Candidate{
			Text:       entry.Text,
			Comment:    joinComment(entry.Syllables),
			Score:      int32(clampScore(entry.Weight / 2)),
			Syllables:  append([]string(nil), entry.Syllables...),
			Incomplete: true,
		})
	}
}

func (d *Decoder) collectPartialWords(syllables []string, seen map[string]Candidate) {
	n := len(syllables)
	if n <= 1 {
		return
	}
	for end := n - 1; end >= 1; end-- {
		piece := syllables[:end]
		entries := d.dict.Lookup(piece)
		limitEntries := entries
		if len(limitEntries) > exactLookupCap(end) {
			limitEntries = limitEntries[:exactLookupCap(end)]
		}
		for _, entry := range limitEntries {
			addCandidate(seen, Candidate{
				Text:      entry.Text,
				Comment:   joinComment(piece),
				Score:     int32(clampScore(entry.Weight)),
				Syllables: append([]string(nil), piece...),
			})
		}
	}
}

func exactLookupCap(syllableLen int) int {
	if syllableLen <= 1 {
		return 12
	}
	if syllableLen == 2 {
		return 10
	}
	return 8
}

func (d *Decoder) collectLattice(syllables []string, seen map[string]Candidate) {
	n := len(syllables)
	if n <= 1 {
		return
	}
	hasExact := len(d.dict.Lookup(syllables)) > 0
	best := make([][]latticeCell, n+1)
	best[0] = []latticeCell{{"", 0, 0}}
	for start := 0; start < n; start++ {
		if len(best[start]) == 0 {
			continue
		}
		for end := start + 1; end <= n; end++ {
			piece := syllables[start:end]
			entries := d.dict.Lookup(piece)
			if len(entries) == 0 {
				continue
			}
			limitEntries := entries
			if len(limitEntries) > 4 {
				limitEntries = limitEntries[:4]
			}
			for _, prev := range best[start] {
				for _, entry := range limitEntries {
					nextScore := prev.score + uint64(entry.Weight) + 1
					nextText := prev.text + entry.Text
					best[end] = append(best[end], latticeCell{
						text:     nextText,
						score:    nextScore,
						segments: prev.segments + 1,
					})
				}
			}
			if len(best[end]) > 16 {
				sort.SliceStable(best[end], func(i, j int) bool {
					return best[end][i].score > best[end][j].score
				})
				best[end] = dedupeCells(best[end][:16])
			}
		}
	}
	limitComposed := 4
	if hasExact {
		limitComposed = 2
	}
	cells := best[n]
	sort.SliceStable(cells, func(i, j int) bool {
		return cells[i].score > cells[j].score
	})
	added := 0
	for _, item := range cells {
		if item.text == "" || item.segments <= 1 {
			continue
		}
		if _, exists := seen[item.text]; exists {
			continue
		}
		score := item.score
		if item.segments > 1 {
			score = score / uint64(item.segments*item.segments)
		}
		addCandidate(seen, Candidate{
			Text:      item.text,
			Comment:   joinComment(syllables),
			Score:     int32(clampScore(uint32(min64(score, 1<<31-1)))),
			Syllables: append([]string(nil), syllables...),
			Composed:  true,
		})
		added++
		if added >= limitComposed {
			break
		}
	}
}

func dedupeCells(in []latticeCell) []latticeCell {
	seen := map[string]struct{}{}
	out := in[:0]
	for _, item := range in {
		if _, ok := seen[item.text]; ok {
			continue
		}
		seen[item.text] = struct{}{}
		out = append(out, item)
	}
	return out
}

func addCandidate(seen map[string]Candidate, cand Candidate) {
	prev, ok := seen[cand.Text]
	if !ok {
		seen[cand.Text] = cand
		return
	}
	if betterCandidate(cand, prev) {
		seen[cand.Text] = cand
	}
}

func joinComment(syllables []string) string {
	if len(syllables) == 0 {
		return ""
	}
	out := syllables[0]
	for i := 1; i < len(syllables); i++ {
		out += "'" + syllables[i]
	}
	return out
}

func clampScore(v uint32) uint32 {
	if v > 1<<31-1 {
		return 1<<31 - 1
	}
	return v
}

func min64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
