package dictionary

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUserDictionaryDirtyTracking(t *testing.T) {
	path := filepath.Join(t.TempDir(), "user_dictionary.json")
	u, err := OpenUserDictionary(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if u.Dirty() {
		t.Fatalf("fresh dictionary should be clean")
	}
	if _, err := u.AddEntry("你好", []string{"ni", "hao"}, UserSourceManual, 0); err != nil {
		t.Fatalf("add: %v", err)
	}
	if !u.Dirty() {
		t.Fatalf("add should mark dirty")
	}
	if err := u.SaveIfDirty(); err != nil {
		t.Fatalf("save: %v", err)
	}
	if u.Dirty() {
		t.Fatalf("save should clear dirty")
	}
	u.RecordUsageText("你好")
	if !u.Dirty() {
		t.Fatalf("usage should mark dirty")
	}
}

func TestUserDictionaryReloadsExternalChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "user_dictionary.json")
	daemon, err := OpenUserDictionary(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if reloaded, err := daemon.ReloadIfChangedOnDisk(); err != nil || reloaded {
		t.Fatalf("unchanged file should not reload: %v %v", reloaded, err)
	}
	cli, err := OpenUserDictionary(path)
	if err != nil {
		t.Fatalf("open cli: %v", err)
	}
	if _, err := cli.AddEntry("星河", []string{"xing", "he"}, UserSourceManual, 0); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := cli.SaveLocked(); err != nil {
		t.Fatalf("save: %v", err)
	}
	future := time.Now().Add(2 * time.Second)
	_ = os.Chtimes(path, future, future)
	reloaded, err := daemon.ReloadIfChangedOnDisk()
	if err != nil || !reloaded {
		t.Fatalf("expected reload after external save: %v %v", reloaded, err)
	}
	if len(daemon.Lookup([]string{"xing", "he"})) != 1 {
		t.Fatalf("external entry missing after reload")
	}
	if daemon.Dirty() {
		t.Fatalf("reloaded dictionary should be clean")
	}
}

func TestPhraseLearnerPendingBounded(t *testing.T) {
	u := NewUserDictionary()
	l := NewPhraseLearner(u, 3)
	syl := []string{"ni", "hao"}
	for i := 0; i < maxLearnPending+10; i++ {
		text := string([]rune{rune(0x4e00 + i%20000), rune(0x4e00 + i/20000 + 1)})
		_, _, _ = l.ObserveCommittedPhrase(text, syl)
	}
	if n := l.PendingCount(); n > maxLearnPending {
		t.Fatalf("pending %d exceeds bound %d", n, maxLearnPending)
	}
	l.Reset()
	if l.PendingCount() != 0 {
		t.Fatalf("reset should clear pending")
	}
}

func TestPhraseLearnerForgetsCountAfterPromotion(t *testing.T) {
	u := NewUserDictionary()
	l := NewPhraseLearner(u, 2)
	syl := []string{"xing", "he"}
	_, _, _ = l.ObserveCommittedPhrase("星河", syl)
	if _, added, _ := l.ObserveCommittedPhrase("星河", syl); !added {
		t.Fatalf("expected promotion")
	}
	if l.PendingCount() != 0 {
		t.Fatalf("promoted phrase should leave pending set")
	}
}
