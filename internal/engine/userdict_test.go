package engine

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
)

func TestPinyinEngineSurfacesUserWord(t *testing.T) {
	dictionary.SetValidSyllableFunc(pinyin.IsValidSyllable)
	dict := dictionary.FromEntries([]dictionary.Entry{
		{Text: "星河", Syllables: []string{"xing", "he"}, Weight: 100},
	})
	eng := NewPinyinEngine(dict)
	eng.SetScheme(pinyin.SchemeXiaohe)
	user := dictionary.NewUserDictionary()
	if _, err := user.AddEntry("星河工作室", []string{"xing", "he", "gong", "zuo", "shi"}, dictionary.UserSourceManual, 8000); err != nil {
		t.Fatalf("add: %v", err)
	}
	eng.SetUserDictionary(user)
	cands, err := eng.Decode("xkhegszoui")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	found := false
	for _, c := range cands {
		if c.Text == "星河工作室" {
			found = true
			break
		}
	}
	if !found {
		texts := make([]string, 0, len(cands))
		for _, c := range cands {
			texts = append(texts, c.Text)
		}
		t.Fatalf("user word missing: %v", texts)
	}
}
