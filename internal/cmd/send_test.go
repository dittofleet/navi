package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/dittofleet/navi/internal/config"
	"github.com/dittofleet/navi/internal/ntfy"
)

// isolate points config at an empty temp dir and clears the NAVI_*
// overrides, so a developer's own setup never leaks into a test.
func isolate(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NAVI_SERVER", "")
	t.Setenv("NAVI_TOPIC", "")
	t.Setenv("NAVI_TOKEN", "")
}

func TestRepoName(t *testing.T) {
	for in, want := range map[string]string{
		"https://github.com/dittofleet/lichen.git\n": "lichen",
		"https://github.com/dittofleet/lichen":       "lichen",
		"git@github.com:dittofleet/port-pool.git":    "port-pool",
		"ssh://git@host:22/owner/repo/":              "repo",
	} {
		if got := repoName(in); got != want {
			t.Errorf("repoName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSplitTags(t *testing.T) {
	got := splitTags([]string{"warning, skull", "", "tada"})
	if !slices.Equal(got, []string{"warning", "skull", "tada"}) {
		t.Errorf("got %v", got)
	}
}

func TestSendPostsToConfiguredTopic(t *testing.T) {
	var got ntfy.Message
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&got)
	}))
	defer srv.Close()
	isolate(t)
	t.Setenv("NAVI_SERVER", srv.URL)
	t.Setenv("NAVI_TOPIC", "test-topic")

	err := Send([]string{"-p", "high", "--tag", "warning,skull", "--", "-3 tests failing"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Topic != "test-topic" || got.Message != "-3 tests failing" || got.Priority != 4 || len(got.Tags) != 2 {
		t.Errorf("server got %+v", got)
	}
	if got.Title == "" {
		t.Error("no default title")
	}
}

func TestSendRefusesBadInputBeforePosting(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("posted despite bad input")
	}))
	defer srv.Close()
	isolate(t)
	t.Setenv("NAVI_SERVER", srv.URL)
	t.Setenv("NAVI_TOPIC", "test-topic")

	for _, args := range [][]string{
		{},
		{"  "},
		{"two", "words"},
		{"msg", "-p", "loud"},
		{"msg", "--click", "not a url"},
	} {
		if err := Send(args); err == nil {
			t.Errorf("Send(%q) succeeded, want an error", args)
		}
	}
}

func TestSetupRotateReplacesTopic(t *testing.T) {
	isolate(t)

	if err := Setup([]string{"--topic", "old-topic", "--token", "tk_keep"}); err != nil {
		t.Fatal(err)
	}
	if err := Setup(nil); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil || cfg.Topic != "old-topic" {
		t.Fatalf("bare rerun changed the topic: %+v, %v", cfg, err)
	}

	if err := Setup([]string{"--rotate"}); err != nil {
		t.Fatal(err)
	}
	cfg, err = config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Topic == "old-topic" || !strings.HasPrefix(cfg.Topic, "navi-") || cfg.Token != "tk_keep" {
		t.Errorf("after --rotate: %+v", cfg)
	}

	if err := Setup([]string{"--rotate", "--topic", "x"}); err == nil {
		t.Error("--rotate with --topic succeeded")
	}
}
