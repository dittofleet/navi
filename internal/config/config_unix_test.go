//go:build !windows

package config

import (
	"os"
	"testing"
)

// Windows has no such mode to check: there the user's own folder keeps
// others out.
func TestSavedConfigIsOwnerOnly(t *testing.T) {
	isolate(t)
	if err := (&Config{SchemaVersion: SchemaVersion, Server: DefaultServer, Topic: "saved"}).Save(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(Path())
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("config mode = %v, want 0600", info.Mode().Perm())
	}
}
