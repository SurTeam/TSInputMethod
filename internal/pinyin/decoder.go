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
}

type Decoder struct {
	dict      *dictionary.Dictionary
	segmenter *Segmenter
	limit     int
}

type latticeCell struct {
	text  string
	score uint64
}

func NewDecoder(dict *dictionary.Dictionary, limit int) *Decoder {
	if limit <= 0 {
		limit = 32
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
		d.collectExact(seg, seen)
		if len(seg.Syllables) > 0 {
			d.collectLattice(seg.Syllables, seen)
		}
		if seg.Remainder != "" {
			d.collectPrefix(seg, seen)
		}
	}
	out := make([]Candidate, 0, len(seen))
	for _, cand := range seen {
		out = append(out, cand)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Incomplete != out[j].Incomplete {
			return !out[i].Incomplete && out[j].Incomplete
		}
		if out[i].Score == out[j].Score {
			ri := len([]rune(out[i].Text))
			rj := len([]rune(out[j].Text))
			if ri == rj {
				return out[i].Text < out[j].Text
			}
			return ri > rj
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > d.limit {
		out = out[:d.limit]
	}
	return out
}

func (d *Decoder) collectExact(seg Segmentation, seen map[string]Candidate) {
	if seg.Remainder != "" || len(seg.Syllables) == 0 {
		return
	}
	for _, entry := range d.dict.Lookup(seg.Syllables) {
		addCandidate(seen, Candidate{
			Text:      entry.Text,
			Comment:   joinComment(seg.Syllables),
			Score:     int32(clampScore(entry.Weight)),
			Syllables: append([]string(nil), seg.Syllables...),
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

func (d *Decoder) collectLattice(syllables []string, seen map[string]Candidate) {
	n := len(syllables)
	best := make([][]latticeCell, n+1)
	best[0] = []latticeCell{{"", 0}}
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
			if len(limitEntries) > 8 {
				limitEntries = limitEntries[:8]
			}
			for _, prev := range best[start] {
				for _, entry := range limitEntries {
					nextScore := prev.score + uint64(entry.Weight) + 1
					nextText := prev.text + entry.Text
					best[end] = append(best[end], latticeCell{text: nextText, score: nextScore})
				}
			}
			if len(best[end]) > d.limit {
				sort.SliceStable(best[end], func(i, j int) bool {
					return best[end][i].score > best[end][j].score
				})
				best[end] = dedupeCells(best[end][:d.limit])
			}
		}
	}
	for _, item := range best[n] {
		if item.text == "" {
			continue
		}
		addCandidate(seen, Candidate{
			Text:      item.text,
			Comment:   joinComment(syllables),
			Score:     int32(clampScore(uint32(min64(item.score, 1<<31-1)))),
			Syllables: append([]string(nil), syllables...),
		})
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
	if !ok || cand.Score > prev.Score || (cand.Score == prev.Score && !cand.Incomplete && prev.Incomplete) {
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
