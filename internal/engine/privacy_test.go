package engine

import (
	"path/filepath"
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/ranking"
)

func TestPrivacyModeBlocksFeedbackAndUsage(t *testing.T) {
	eng := NewPinyinEngine(testDict())
	defer eng.Close()
	user := dictionary.NewUserDictionary()
	if _, err := user.AddEntry("你好", []string{"ni", "hao"}, dictionary.UserSourceManual, 100); err != nil {
		t.Fatalf("add: %v", err)
	}
	eng.SetUserDictionary(user)
	if !eng.EnqueueFeedback("ctx", "nihao", "你好", "泥") {
		t.Fatalf("expected feedback accepted before privacy")
	}
	eng.RecordSelectedUsage("你好")
	if user.UsageCountForText("你好") == 0 {
		t.Fatalf("expected usage recorded before privacy")
	}
	eng.Settings().SetPrivacyMode(true)
	eng.ApplySettings()
	if eng.EnqueueFeedback("ctx", "nihao", "你好", "泥") {
		t.Fatalf("privacy must block EnqueueFeedback")
	}
	before := user.UsageCountForText("你好")
	eng.RecordSelectedUsage("你好")
	if user.UsageCountForText("你好") != before {
		t.Fatalf("privacy must block RecordSelectedUsage")
	}
}

func TestClearPersonalWeightsPersistsEmptyModel(t *testing.T) {
	eng := NewPinyinEngine(testDict())
	defer eng.Close()
	path := filepath.Join(t.TempDir(), "personal_model.json")
	model := eng.PersonalModel()
	model.SetPath(path)
	model.PairwiseUpdate("前文", "nihao", "你好", "泥")
	if err := model.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	if model.WeightCount() == 0 || model.UpdateCount() == 0 {
		t.Fatalf("expected non-empty model before clear")
	}
	if err := eng.ClearPersonalWeights(); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if model.WeightCount() != 0 || model.UpdateCount() != 0 {
		t.Fatalf("expected empty in-memory model")
	}
	reloaded, err := ranking.OpenPersonalModel(path, ranking.DefaultSGDConfig())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if reloaded.WeightCount() != 0 || reloaded.UpdateCount() != 0 {
		t.Fatalf("expected empty persisted model updates=%d weights=%d", reloaded.UpdateCount(), reloaded.WeightCount())
	}
}
