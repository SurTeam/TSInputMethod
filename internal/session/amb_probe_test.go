package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func TestAmbiguousXiaoheKeysSurfaceAllReadings(t *testing.T) {
	eng := engine.NewPinyinEngine(loadIceDict(t))
	eng.SetScheme(pinyin.SchemeXiaohe)
	cases := []struct {
		input string
		want  []string
	}{
		{"lo", []string{"落", "咯"}},
		{"lt", []string{"略"}},
		{"nt", []string{"虐"}},
		{"dv", []string{"对", "度"}},
		{"lv", []string{"绿", "路"}},
		{"uv", []string{"水", "书"}},
		{"wo", []string{"我"}},
	}
	for _, tc := range cases {
		s := NewManager(eng).Open(imev1.InputMode_INPUT_MODE_CHINESE)
		res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: tc.input})
		texts := candidateBareTexts(res)
		if len(texts) == 0 {
			t.Fatalf("%s produced no candidates", tc.input)
		}
		for _, want := range tc.want {
			if !containsText(texts, want) {
				t.Fatalf("%s missing %s: %#v", tc.input, want, texts)
			}
		}
	}
}
