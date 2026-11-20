package pinyin

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
)

func TestUserDictionaryAppearsUnderXiaohe(t *testing.T) {
	dictionary.SetValidSyllableFunc(IsValidSyllable)
	system := dictionary.FromEntries([]dictionary.Entry{
		{Text: "星河", Syllables: []string{"xing", "he"}, Weight: 100},
		{Text: "工作室", Syllables: []string{"gong", "zuo", "shi"}, Weight: 90},
	})
	user := dictionary.NewUserDictionary()
	if _, err := user.AddEntry("星河工作室", []string{"xing", "he", "gong", "zuo", "shi"}, dictionary.UserSourceManual, 5000); err != nil {
		t.Fatalf("add user word: %v", err)
	}
	dec := NewDecoder(system, 96)
	dec.SetUserDictionary(user)

	code := "xkhegszoui"
	syls, rem, ok := ExpandXiaohe(code)
	if !ok || rem != "" {
		t.Fatalf("expand %s: ok=%v rem=%q syls=%v", code, ok, rem, syls)
	}
	want := []string{"xing", "he", "gong", "zuo", "shi"}
	if len(syls) != len(want) {
		t.Fatalf("expand syllables=%v want=%v", syls, want)
	}
	for i := range want {
		if syls[i] != want[i] {
			t.Fatalf("expand syllables=%v want=%v", syls, want)
		}
	}
	cands := dec.DecodeSyllables(syls, "")
	found := false
	for _, cand := range cands {
		if cand.Text == "星河工作室" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("user word missing under xiaohe code %s; cands=%v", code, summarize(cands))
	}
}

func summarize(cands []Candidate) []string {
	out := make([]string, 0, len(cands))
	for _, c := range cands {
		out = append(out, c.Text)
	}
	return out
}
