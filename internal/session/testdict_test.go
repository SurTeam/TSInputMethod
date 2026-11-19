package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
)

func loadIceDict(t *testing.T) *dictionary.Dictionary {
	t.Helper()
	path := filepath.Join("..", "..", "data", "dicts", "rime-ice.tsdict")
	if _, err := os.Stat(path); err != nil {
		path = filepath.Join("data", "dicts", "rime-ice.tsdict")
	}
	dict, err := dictionary.LoadFile(path)
	if err != nil {
		t.Skip(err)
	}
	return dict
}
