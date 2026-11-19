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
	return SchemeQuanpin
}

func NormalizeInput(scheme Scheme, input string) string {
	switch scheme {
	case SchemeXiaohe:
		return FlypyToFull(input)
	default:
		return input
	}
}
