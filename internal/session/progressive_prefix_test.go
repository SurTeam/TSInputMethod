package session

import (
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

func progressiveSessionDict() *dictionary.Dictionary {
	return dictionary.FromEntries([]dictionary.Entry{
		{Text: "中国人民", Syllables: []string{"zhong", "guo", "ren", "min"}, Weight: 5000},
		{Text: "中国人", Syllables: []string{"zhong", "guo", "ren"}, Weight: 4000},
		{Text: "中国", Syllables: []string{"zhong", "guo"}, Weight: 3000},
		{Text: "中", Syllables: []string{"zhong"}, Weight: 2000},
		{Text: "国", Syllables: []string{"guo"}, Weight: 900},
		{Text: "人", Syllables: []string{"ren"}, Weight: 800},
		{Text: "民", Syllables: []string{"min"}, Weight: 700},
		{Text: "人民", Syllables: []string{"ren", "min"}, Weight: 2500},
	})
}

func openProgressiveSession(t *testing.T) *Session {
	t.Helper()
	eng := engine.NewPinyinEngine(progressiveSessionDict())
	eng.SetScheme(pinyin.SchemeXiaohe)
	m := NewManager(eng)
	return m.Open(imev1.InputMode_INPUT_MODE_CHINESE)
}

func candidateTexts(res *KeyResult) []string {
	return candidateBareTexts(res)
}

func containsText(texts []string, want string) bool {
	return containsBareText(texts, want)
}

func TestFlypyLongInputShowsProgressivePrefixes(t *testing.T) {
	s := openProgressiveSession(t)
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "vsgorfmb"})
	if res.Composition == nil || res.Composition.Preedit != "vsgorfmb" {
		t.Fatalf("preedit=%#v", res.Composition)
	}
	if res.Candidates == nil || res.Candidates.PageSize != 9 {
		t.Fatalf("pageSize=%#v", res.Candidates)
	}
	texts := candidateTexts(res)
	for _, want := range []string{"中国人民", "中国人", "中国", "中"} {
		if !containsText(texts, want) {
			t.Fatalf("missing %s in %#v", want, texts)
		}
	}
}

func TestFlypyOddLengthPrefixThenRemainder(t *testing.T) {
	s := openProgressiveSession(t)
	first := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "vsgorfm"})
	texts := candidateTexts(first)
	if !containsText(texts, "中国") {
		t.Fatalf("odd input missing 中国: %#v", texts)
	}
	idx := -1
	for i, text := range texts {
		if bareCandidateText(text) == "中国" {
			idx = i
			break
		}
	}
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: string(rune('1' + idx))})
	if res.Commit == nil || res.Commit.Text != "中国" {
		t.Fatalf("commit=%#v", res.Commit)
	}
	if res.Composition == nil || res.Composition.Preedit != "rfm" {
		t.Fatalf("remainder preedit=%#v", res.Composition)
	}
}

func TestFlypyPartialCommitContinuesMatching(t *testing.T) {
	s := openProgressiveSession(t)
	first := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "vsgorfmb"})
	idx := -1
	for i, text := range candidateTexts(first) {
		if bareCandidateText(text) == "中国" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("missing 中国: %#v", candidateTexts(first))
	}
	res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: string(rune('1' + idx))})
	if res.Commit == nil || res.Commit.Text != "中国" {
		t.Fatalf("commit=%#v", res.Commit)
	}
	if res.Composition == nil || res.Composition.Preedit != "rfmb" {
		t.Fatalf("remainder=%#v", res.Composition)
	}
	if !containsText(candidateTexts(res), "人民") && !containsText(candidateTexts(res), "人") {
		t.Fatalf("remainder candidates=%#v", candidateTexts(res))
	}
}

func TestCandidatePageSizeAlwaysNine(t *testing.T) {
	s := openProgressiveSession(t)
res := s.HandleKey(&imev1.KeyEvent{KeyDown: true, Characters: "vsgorfmb"})
	if res.Candidates == nil {
		t.Fatal("nil candidates")
	}
	if res.Candidates.PageSize != 9 {
		t.Fatalf("pageSize=%d", res.Candidates.PageSize)
	}
	if len(res.Candidates.Candidates) > int(res.Candidates.PageSize) {
		t.Fatalf("page overflow: %d > %d", len(res.Candidates.Candidates), res.Candidates.PageSize)
	}
}
