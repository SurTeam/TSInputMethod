package emoji

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileLookup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "emoji.txt")
	content := "开心\t开心 😄\n舞者\t舞者 🧑‍🩰 🕺\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	table, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := table.Lookup("开心")
	if len(got) != 1 || got[0] != "😄" {
		t.Fatalf("开心 = %#v", got)
	}
	got = table.Lookup("舞者")
	if len(got) != 2 {
		t.Fatalf("舞者 = %#v", got)
	}
	if table.Lookup("不存在") != nil {
		t.Fatal("unexpected hit")
	}
}
