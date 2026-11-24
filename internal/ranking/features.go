package ranking

import (
	"hash/fnv"
	"strconv"
	"unicode/utf8"
)

const (
	DefaultHashDimension = 1 << 18
	FeatureSchemaVersion = 1
	HashAlgorithmFNV1a64 = "fnv1a64"
)

type FeatureVector []uint32

func HashFeature(name string, dim uint32) uint32 {
	if dim == 0 {
		dim = DefaultHashDimension
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return uint32(h.Sum64() % uint64(dim))
}

func ExtractFeatures(context, code, candidate string, dim uint32) FeatureVector {
	if dim == 0 {
		dim = DefaultHashDimension
	}
	out := make(FeatureVector, 0, 8)
	add := func(name string) {
		out = append(out, HashFeature(name, dim))
	}
	add("code_candidate=" + code + "|" + candidate)
	tail := unicodeTailString(context, 4)
	runess := []rune(tail)
	for n := 1; n <= len(runess) && n <= 4; n++ {
		ng := string(runess[len(runess)-n:])
		add("last" + strconv.Itoa(n) + "_candidate=" + ng + "|" + candidate)
	}
	add("cand_len=" + strconv.Itoa(utf8.RuneCountInString(candidate)))
	add("cand=" + candidate)
	return dedupeFeatures(out)
}

func unicodeTailString(s string, n int) string {
	if n <= 0 || s == "" {
		return ""
	}
	runess := []rune(s)
	if len(runess) <= n {
		return s
	}
	return string(runess[len(runess)-n:])
}

func dedupeFeatures(in FeatureVector) FeatureVector {
	if len(in) <= 1 {
		return in
	}
	seen := make(map[uint32]struct{}, len(in))
	out := make(FeatureVector, 0, len(in))
	for _, idx := range in {
		if _, ok := seen[idx]; ok {
			continue
		}
		seen[idx] = struct{}{}
		out = append(out, idx)
	}
	return out
}
