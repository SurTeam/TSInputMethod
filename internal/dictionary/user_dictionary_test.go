package dictionary

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func init() {
	SetValidSyllableFunc(func(s string) bool {
		switch s {
		case "xing", "he", "gong", "zuo", "shi", "bei", "jing", "ni", "hao", "jie", "wo", "xiang":
			return true
		default:
			return len(s) >= 1 && len(s) <= 6
		}
	})
}

func TestUserDictionaryCRUDAndLookup(t *testing.T) {
	u := NewUserDictionary()
	entry, err := u.AddEntry("星河工作室", []string{"xing", "he", "gong", "zuo", "shi"}, UserSourceManual, 2000)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if entry.Text != "星河工作室" || entry.Weight != 2000 {
		t.Fatalf("unexpected entry %#v", entry)
	}
	if _, err := u.AddEntry("星河工作室", []string{"xing", "he", "gong", "zuo", "shi"}, UserSourceManual, 0); err != ErrUserEntryExists {
		t.Fatalf("expected exists, got %v", err)
	}
	if got := u.Lookup([]string{"xing", "he", "gong", "zuo", "shi"}); len(got) != 1 || got[0].Text != "星河工作室" {
		t.Fatalf("exact lookup: %#v", got)
	}
	prefix := u.LookupPrefix([]string{"xing", "he"}, "")
	if len(prefix) == 0 || prefix[0].Text != "星河工作室" {
		t.Fatalf("prefix lookup: %#v", prefix)
	}
	enabled := false
	if _, err := u.UpdateEntry("星河工作室", []string{"xing", "he", "gong", "zuo", "shi"}, 3000, &enabled); err != nil {
		t.Fatalf("update: %v", err)
	}
	if got := u.Lookup([]string{"xing", "he", "gong", "zuo", "shi"}); len(got) != 0 {
		t.Fatalf("disabled entry should hide: %#v", got)
	}
	enabled = true
	if _, err := u.UpdateEntry("星河工作室", []string{"xing", "he", "gong", "zuo", "shi"}, 0, &enabled); err != nil {
		t.Fatalf("re-enable: %v", err)
	}
	if err := u.RemoveEntry("星河工作室", []string{"xing", "he", "gong", "zuo", "shi"}); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if u.EntryCount() != 0 {
		t.Fatalf("expected empty after remove")
	}
}

func TestMergeEntriesDedupAndFrequencyFusion(t *testing.T) {
	system := []Entry{{
		Text: "北京", Syllables: []string{"bei", "jing"}, Weight: 100,
	}}
	user := []Entry{{
		Text: "北京", Syllables: []string{"bei", "jing"}, Weight: 5000,
	}, {
		Text: "星河工作室", Syllables: []string{"xing", "he", "gong", "zuo", "shi"}, Weight: 2000,
	}}
	out := MergeEntries(system, user)
	if len(out) != 2 {
		t.Fatalf("expected 2 merged, got %d %#v", len(out), out)
	}
	if out[0].Text != "北京" || out[0].Weight != 5000 {
		t.Fatalf("expected fused beijing weight, got %#v", out[0])
	}
	if out[1].Text != "星河工作室" {
		t.Fatalf("expected user-only word, got %#v", out[1])
	}
}

func TestUserDictionaryPersistAndCorruptSafety(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "user_dictionary.json")
	u, err := OpenUserDictionary(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, err := u.AddEntry("你好", []string{"ni", "hao"}, UserSourceManual, 0); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := u.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	u2, err := OpenUserDictionary(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if u2.EntryCount() != 1 {
		t.Fatalf("expected 1 entry after restart, got %d", u2.EntryCount())
	}
	if err := os.WriteFile(path, []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("corrupt write: %v", err)
	}
	if _, err := OpenUserDictionary(path); err == nil {
		t.Fatalf("expected corrupt file error")
	}
}

func TestUserDictionaryConcurrency(t *testing.T) {
	u := NewUserDictionary()
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			text := "词" + string(rune('一'+i%10))
			_, _ = u.AddEntry(text, []string{"ni", "hao"}, UserSourceManual, 0)
			_ = u.Lookup([]string{"ni", "hao"})
			_ = u.ListEntries()
		}(i)
	}
	wg.Wait()
}

func TestPhraseLearnerAutoLearnAndExclusions(t *testing.T) {
	u := NewUserDictionary()
	l := NewPhraseLearner(u, 2)
	syl := []string{"xing", "he", "gong", "zuo", "shi"}
	if _, added, err := l.ObserveCommittedPhrase("星河工作室", syl); err != nil || added {
		t.Fatalf("first observe should not add: added=%v err=%v", added, err)
	}
	if _, added, err := l.ObserveCommittedPhrase("星河工作室", syl); err != nil || !added {
		t.Fatalf("second observe should add: added=%v err=%v", added, err)
	}
	if got := u.Lookup(syl); len(got) != 1 {
		t.Fatalf("learned lookup: %#v", got)
	}
	if IsAutoLearnablePhrase("ok", []string{"ni", "hao"}) {
		t.Fatalf("non-han should be excluded")
	}
	if IsAutoLearnablePhrase("星", []string{"xing"}) {
		t.Fatalf("too short should be excluded")
	}
	if IsAutoLearnablePhrase("星河工作室", []string{"xing", "he"}) {
		t.Fatalf("syllable/text length mismatch should be excluded")
	}
}
