package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveIfEmptyLeavesOtherToolsFiles(t *testing.T) {
	dir := t.TempDir()
	shared := filepath.Join(dir, "navi")
	cheat := filepath.Join(shared, "cheats", "git.cheat")
	if err := os.MkdirAll(filepath.Dir(cheat), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cheat, []byte("% git"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := removeIfEmpty(shared); err != nil {
		t.Fatalf("removeIfEmpty on a shared directory: %v", err)
	}
	if _, err := os.Stat(cheat); err != nil {
		t.Fatalf("another tool's file is gone: %v", err)
	}

	empty := filepath.Join(dir, "empty")
	if err := os.Mkdir(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := removeIfEmpty(empty); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(empty); !os.IsNotExist(err) {
		t.Fatalf("empty directory still there: %v", err)
	}
}
