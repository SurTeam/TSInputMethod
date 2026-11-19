package pinyin

type Segmenter struct {
	maxLen    int
	syllables map[string]struct{}
}

func NewSegmenter(extra map[string]struct{}) *Segmenter {
	syllables := AllValidSyllables()
	maxLen := 0
	for s := range syllables {
		if len(s) > maxLen {
			maxLen = len(s)
		}
	}
	for s := range extra {
		syllables[s] = struct{}{}
		if len(s) > maxLen {
			maxLen = len(s)
		}
	}
	return &Segmenter{maxLen: maxLen, syllables: syllables}
}

func (s *Segmenter) Has(syllable string) bool {
	_, ok := s.syllables[syllable]
	return ok
}

type Span struct {
	Start int
	End   int
	Text  string
}

func (s *Segmenter) Spans(input string) [][]Span {
	n := len(input)
	spans := make([][]Span, n)
	for i := 0; i < n; i++ {
		max := s.maxLen
		if i+max > n {
			max = n - i
		}
		for length := max; length >= 1; length-- {
			cand := input[i : i+length]
			if s.Has(cand) {
				spans[i] = append(spans[i], Span{Start: i, End: i + length, Text: cand})
			}
		}
	}
	return spans
}

type Segmentation struct {
	Syllables  []string
	Remainder  string
	CoveredLen int
}

func (s *Segmenter) Segment(input string) []Segmentation {
	if input == "" {
		return nil
	}
	const maxSegs = 64
	spans := s.Spans(input)
	var out []Segmentation
	var walk func(pos int, path []string)
	walk = func(pos int, path []string) {
		if len(out) >= maxSegs {
			return
		}
		if pos == len(input) {
			copied := append([]string(nil), path...)
			out = append(out, Segmentation{Syllables: copied, CoveredLen: pos})
			return
		}
		if len(spans[pos]) == 0 {
			out = append(out, Segmentation{
				Syllables:  append([]string(nil), path...),
				Remainder:  input[pos:],
				CoveredLen: pos,
			})
			return
		}
		for _, span := range spans[pos] {
			walk(span.End, append(path, span.Text))
			if len(out) >= maxSegs {
				return
			}
		}
	}
	walk(0, nil)
	if len(out) == 0 {
		out = append(out, Segmentation{Remainder: input, CoveredLen: 0})
	}
	return out
}

func (s *Segmenter) PreferLongest(input string) Segmentation {
	segs := s.Segment(input)
	if len(segs) == 0 {
		return Segmentation{Remainder: input}
	}
	best := segs[0]
	for _, seg := range segs[1:] {
		if betterSegmentation(seg, best) {
			best = seg
		}
	}
	return best
}

func betterSegmentation(a, b Segmentation) bool {
	if a.CoveredLen != b.CoveredLen {
		return a.CoveredLen > b.CoveredLen
	}
	if len(a.Remainder) != len(b.Remainder) {
		return len(a.Remainder) < len(b.Remainder)
	}
	if len(a.Syllables) != len(b.Syllables) {
		return len(a.Syllables) < len(b.Syllables)
	}
	return false
}
