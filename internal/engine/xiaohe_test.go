package engine

import (
	"os"
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/pinyin"
)

func TestPinyinEngineXiaoheScheme(t *testing.T) {
	eng := NewPinyinEngine(testDict())
	eng.SetScheme(pinyin.SchemeXiaohe)
	cands, err := eng.Decode("nihc")
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(cands) == 0 || cands[0].Text != "你好" {
		t.Fatalf("xiaohe engine decode = %#v", cands)
	}
}

func TestPinyinEngineDefaultRemainsQuanpin(t *testing.T) {
	_ = os.Unsetenv("TSIME_SCHEME")
	eng := NewPinyinEngine(testDict())
	if eng.Scheme() != pinyin.SchemeQuanpin {
		t.Fatalf("default scheme = %s", eng.Scheme())
	}
	cands, err := eng.Decode("nihao")
	if err != nil || len(cands) == 0 || cands[0].Text != "你好" {
		t.Fatalf("quanpin path broken: %#v err=%v", cands, err)
	}
}
