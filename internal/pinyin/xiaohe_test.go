package pinyin

import "testing"

func TestExpandXiaoheNihao(t *testing.T) {
	syls, rem, ok := ExpandXiaohe("nihc")
	if !ok || rem != "" {
		t.Fatalf("expand nihc: ok=%v rem=%q syls=%v", ok, rem, syls)
	}
	if len(syls) != 2 || syls[0] != "ni" || syls[1] != "hao" {
		t.Fatalf("expected ni+hao, got %v", syls)
	}
	if got := FlypyToFull("nihc"); got != "nihao" {
		t.Fatalf("FlypyToFull=nihc -> %q", got)
	}
}

func TestExpandXiaoheZeroInitial(t *testing.T) {
	cases := map[string]string{
		"aa": "a",
		"oo": "o",
		"ee": "e",
		"aj": "an",
		"ah": "ang",
		"er": "er",
	}
	for in, want := range cases {
		syls, rem, ok := ExpandXiaohe(in)
		if !ok || rem != "" || len(syls) != 1 || syls[0] != want {
			t.Fatalf("%s => %v rem=%q ok=%v want %s", in, syls, rem, ok, want)
		}
	}
}

func TestExpandXiaoheZhChSh(t *testing.T) {
	if got := FlypyToFull("vh"); got != "zhang" {
		t.Fatalf("vh => %q want zhang", got)
	}
	if got := FlypyToFull("ii"); got != "chi" {
		t.Fatalf("ii => %q want chi", got)
	}
	if got := FlypyToFull("uu"); got != "shu" {
		t.Fatalf("uu => %q want shu", got)
	}
}

func TestExpandXiaoheRemainder(t *testing.T) {
	syls, rem, ok := ExpandXiaohe("nih")
	if !ok || rem != "h" || len(syls) != 1 || syls[0] != "ni" {
		t.Fatalf("got syls=%v rem=%q ok=%v", syls, rem, ok)
	}
}

func TestExpandXiaoheWOYO(t *testing.T) {
	cases := map[string]string{
		"wo": "wo",
		"yo": "yo",
		"bo": "bo",
		"po": "po",
		"mo": "mo",
		"fo": "fo",
		"lo": "luo",
		"go": "guo",
	}
	for in, want := range cases {
		if got := FlypyToFull(in); got != want {
			t.Fatalf("%s => %q want %q", in, got, want)
		}
	}
}

func TestExpandXiaoheAmbiguousPaths(t *testing.T) {
	cases := map[string][]string{
		"lo": {"luo", "lo"},
		"lt": {"lve", "lue"},
		"nt": {"nve", "nue"},
		"lv": {"lv", "lu"},
		"dv": {"dui", "du"},
		"uv": {"shui", "shu"},
	}
	for in, want := range cases {
		paths, rem, ok := ExpandXiaohePaths(in)
		if !ok || rem != "" || len(paths) != len(want) {
			t.Fatalf("%s paths=%v rem=%q ok=%v want %v", in, paths, rem, ok, want)
		}
		got := make([]string, 0, len(paths))
		for _, path := range paths {
			if len(path) != 1 {
				t.Fatalf("%s path=%v", in, path)
			}
			got = append(got, path[0])
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%s got %v want %v", in, got, want)
			}
		}
	}
}

func TestNormalizeInputSchemes(t *testing.T) {
	if got := NormalizeInput(SchemeQuanpin, "nihao"); got != "nihao" {
		t.Fatalf("quanpin normalize: %q", got)
	}
	if got := NormalizeInput(SchemeXiaohe, "nihc"); got != "nihao" {
		t.Fatalf("xiaohe normalize: %q", got)
	}
}

func TestDecodeXiaoheThroughQuanpin(t *testing.T) {
	dec := NewDecoder(fixtureDict(), 16)
	quanpin := FlypyToFull("nihc")
	cands := dec.Decode(quanpin)
	if len(cands) == 0 || cands[0].Text != "你好" {
		t.Fatalf("xiaohe nihc -> %q candidates %#v", quanpin, cands)
	}
}
