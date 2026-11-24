package engine

import (
	"path/filepath"
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
)

func TestObserveCommittedPhraseRespectsSettings(t *testing.T) {
	dictionary.SetValidSyllableFunc(func(string) bool { return true })
	eng := NewPinyinEngine(testDict())
	defer eng.Close()
	user := dictionary.NewUserDictionary()
	eng.SetUserDictionary(user)
	syl := []string{"xing", "he"}
	eng.Settings().SetAutoLearnEnabled(false)
	eng.ApplySettings()
	for i := 0; i < 3; i++ {
		_ = eng.ObserveCommittedPhrase("星河", syl)
	}
	if user.EntryCount() != 0 {
		t.Fatalf("auto-learn off must not add entries")
	}
	eng.Settings().SetAutoLearnEnabled(true)
	eng.Settings().SetPrivacyMode(true)
	eng.ApplySettings()
	for i := 0; i < 3; i++ {
		_ = eng.ObserveCommittedPhrase("星河", syl)
	}
	if user.EntryCount() != 0 {
		t.Fatalf("privacy mode must not add entries")
	}
	eng.Settings().SetPrivacyMode(false)
	eng.ApplySettings()
	for i := 0; i < 2; i++ {
		_ = eng.ObserveCommittedPhrase("星河", syl)
	}
	if user.EntryCount() != 1 {
		t.Fatalf("expected learned entry, got %d", user.EntryCount())
	}
}

func TestFeedbackCodeMatchesRankingCode(t *testing.T) {
	eng := NewPinyinEngine(testDict())
	raw := "nihc"
	normalized := pinyin.NormalizeInput(eng.Scheme(), raw)
	if normalized == raw {
		t.Fatalf("fixture expects double-pinyin code to normalize, got %q", normalized)
	}
	if !eng.EnqueueFeedback("前文", raw, "泥", "你") {
		t.Fatalf("enqueue failed")
	}
	eng.UpdateQueue().Close()
	model := eng.PersonalModel()
	if model.Score("前文", normalized, "泥") <= model.Score("前文", raw, "泥") {
		t.Fatalf("training must use the same code key that ranking scores with")
	}
}

func TestCloseFlushesPersonalData(t *testing.T) {
	dir := t.TempDir()
	eng := NewPinyinEngine(testDict())
	model := eng.PersonalModel()
	model.SetPath(filepath.Join(dir, "personal_model.json"))
	user := dictionary.NewUserDictionary()
	user.SetPath(filepath.Join(dir, "user_dictionary.json"))
	eng.SetUserDictionary(user)
	if !eng.EnqueueFeedback("前文", "nihao", "泥", "你") {
		t.Fatalf("enqueue failed")
	}
	if _, err := user.AddEntry("你好", []string{"ni", "hao"}, dictionary.UserSourceManual, 0); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := eng.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if model.UpdateCount() != 1 || model.Dirty() || user.Dirty() {
		t.Fatalf("close should drain and flush: updates=%d modelDirty=%v userDirty=%v", model.UpdateCount(), model.Dirty(), user.Dirty())
	}
	reloaded, err := dictionary.OpenUserDictionary(user.Path())
	if err != nil || reloaded.EntryCount() != 1 {
		t.Fatalf("user dictionary not persisted: %v", err)
	}
}
