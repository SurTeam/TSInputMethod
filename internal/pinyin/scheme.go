package pinyin

import (
	"os"
	"strings"
)

type Scheme string

const (
	SchemeQuanpin Scheme = "quanpin"
	SchemeXiaohe  Scheme = "xiaohe"
)

type NormalizedInput struct {
	Raw          string
	Syllables    []string
	SyllableAlts [][]string
	Remainder    string
	Joined       string
	FixedSplit   bool
}

func ParseScheme(name string) Scheme {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "xiaohe", "flypy", "double", "shuangpin":
		return SchemeXiaohe
	default:
		return SchemeQuanpin
	}
}

func SchemeFromEnv() Scheme {
	if v := os.Getenv("TSIME_SCHEME"); v != "" {
		return ParseScheme(v)
	}
	return SchemeXiaohe
}

func NormalizeInput(scheme Scheme, input string) string {
	return AnalyzeInput(scheme, input).Joined
}

func AnalyzeInput(scheme Scheme, input string) NormalizedInput {
	input = strings.ToLower(strings.TrimSpace(input))
	out := NormalizedInput{Raw: input, Joined: input}
	if input == "" {
		return out
	}
	switch scheme {
	case SchemeXiaohe:
		paths, rem, ok := ExpandXiaohePaths(input)
		if !ok && (len(paths) == 0 || (len(paths) == 1 && len(paths[0]) == 0)) {
			return out
		}
		out.SyllableAlts = paths
		if len(paths) > 0 {
			out.Syllables = append([]string(nil), paths[0]...)
		}
		out.Remainder = rem
		out.FixedSplit = true
		out.Joined = strings.Join(out.Syllables, "")
		if rem != "" {
			out.Joined += rem
		}
		return out
	default:
		return out
	}
}
