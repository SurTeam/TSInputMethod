package pinyin

import "testing"

func TestFlypyHCIsHao(t *testing.T) {
	if got := FlypyToFull("hc"); got != "hao" {
		t.Fatalf("hc => %q", got)
	}
	syls, rem, ok := ExpandXiaohe("hc")
	if !ok || rem != "" || len(syls) != 1 || syls[0] != "hao" {
		t.Fatalf("expand hc => %v rem=%q ok=%v", syls, rem, ok)
	}
}

func TestFlypyKeymapCoreMappings(t *testing.T) {
	// from https://github.com/boomker/rime-fast-xhup/blob/main/flypy_keymap.txt finals
	cases := map[string]string{
		"hc":   "hao",
		"nihc": "nihao",
		"go":   "gou", // o => uo after g? wait g+o => guo in flypy
		"bo":   "bo",
		"vh":   "zhang",
		"ul":   "shuang",
		"jx":   "jia",
	}
	// fix go expectation: g+o => guo
	cases["go"] = "guo"
	for in, want := range cases {
		if got := FlypyToFull(in); got != want {
			t.Fatalf("%s => %q want %q", in, got, want)
		}
	}
}
