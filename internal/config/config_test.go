package config

import (
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
)

func isolate(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NAVI_SERVER", "")
	t.Setenv("NAVI_TOPIC", "")
	t.Setenv("NAVI_TOKEN", "")
}

func TestLoadWithoutTopicIsNotConfigured(t *testing.T) {
	isolate(t)
	if _, err := Load(); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err = %v, want NotConfiguredError", err)
	}
}

func TestEnvAloneIsEnough(t *testing.T) {
	isolate(t)
	t.Setenv("NAVI_TOPIC", "from-env")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Topic != "from-env" || cfg.Server != DefaultServer {
		t.Errorf("got %+v", cfg)
	}
}

func TestSaveThenEnvOverrides(t *testing.T) {
	isolate(t)
	if err := (&Config{SchemaVersion: SchemaVersion, Server: "https://ntfy.example.com", Topic: "saved", Token: "tk"}).Save(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(Path())
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("config mode = %v, want 0600", info.Mode().Perm())
	}

	t.Setenv("NAVI_SERVER", "https://other.example.com/")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Topic != "saved" || cfg.Token != "tk" || cfg.Server != "https://other.example.com" {
		t.Errorf("got %+v", cfg)
	}
}

func TestValidate(t *testing.T) {
	bad := []Config{
		{Server: "ntfy.sh", Topic: "ok"},
		{Server: "ftp://ntfy.sh", Topic: "ok"},
		{Server: DefaultServer, Topic: "has space"},
		{Server: DefaultServer, Topic: "a/b"},
		{Server: DefaultServer, Topic: strings.Repeat("x", 65)},
	}
	for _, c := range bad {
		if err := c.Validate(); err == nil {
			t.Errorf("Validate(%+v) succeeded, want an error", c)
		}
	}
	good := Config{Server: DefaultServer, Topic: RandomTopic()}
	if err := good.Validate(); err != nil {
		t.Errorf("random topic rejected: %v", err)
	}
}

func TestRandomTopicsDiffer(t *testing.T) {
	if RandomTopic() == RandomTopic() {
		t.Error("two random topics matched")
	}
}

func TestSaveNormalizesAndValidates(t *testing.T) {
	isolate(t)
	c := &Config{SchemaVersion: SchemaVersion, Server: "https://ntfy.example.com/", Topic: "ok"}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	if c.Server != "https://ntfy.example.com" {
		t.Errorf("server saved as %q", c.Server)
	}
	if err := (&Config{SchemaVersion: SchemaVersion, Server: DefaultServer, Topic: "bad topic"}).Save(); err == nil {
		t.Error("saved an invalid topic")
	}
}

func TestOverridden(t *testing.T) {
	isolate(t)
	c := &Config{Server: DefaultServer, Topic: "saved", Token: "tk"}
	if got := c.Overridden(); len(got) != 0 {
		t.Errorf("nothing set, got %v", got)
	}
	t.Setenv("NAVI_SERVER", DefaultServer+"/")
	t.Setenv("NAVI_TOPIC", "other")
	t.Setenv("NAVI_TOKEN", "tk")
	if got := c.Overridden(); !slices.Equal(got, []string{"NAVI_TOPIC"}) {
		t.Errorf("got %v, want only NAVI_TOPIC", got)
	}
}
