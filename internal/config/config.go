// Package config reads and writes navi's config file, which says where
// notifications go: an ntfy server, a topic on it, and optionally an
// access token for servers that require one.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/dittofleet/navi/internal/xdg"
)

const (
	SchemaVersion = 1

	// DefaultServer is the public ntfy instance, which needs no account.
	DefaultServer = "https://ntfy.sh"
)

type Config struct {
	SchemaVersion int    `json:"schemaVersion"`
	Server        string `json:"server"`
	Topic         string `json:"topic"`
	Token         string `json:"token,omitempty"`
}

// topicPattern is what ntfy accepts as a topic name.
var topicPattern = regexp.MustCompile(`^[-_A-Za-z0-9]{1,64}$`)

func Path() string {
	return filepath.Join(xdg.ConfigDir(xdg.App), "config.json")
}

// ErrNotConfigured is returned by Load when neither the file nor the
// environment names a topic, so main can point at `navi setup`.
var ErrNotConfigured = errors.New("no topic configured")

// envFields maps each environment variable that overrides the file to the
// field it overrides, so the names are spelled in one place.
func (c *Config) envFields() map[string]*string {
	return map[string]*string{"NAVI_SERVER": &c.Server, "NAVI_TOPIC": &c.Topic, "NAVI_TOKEN": &c.Token}
}

// Read returns the config file with the NAVI_SERVER, NAVI_TOPIC and
// NAVI_TOKEN env vars laid over it, without requiring any of it to be
// set. A missing file reads as empty.
func Read() (*Config, error) {
	path := Path()
	cfg := &Config{SchemaVersion: SchemaVersion}

	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", path, err)
		}
		if cfg.SchemaVersion != SchemaVersion {
			return nil, fmt.Errorf("invalid %s:\n  - schemaVersion: expected %d, got %d", path, SchemaVersion, cfg.SchemaVersion)
		}
	case !errors.Is(err, fs.ErrNotExist):
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	for name, field := range cfg.envFields() {
		if v := os.Getenv(name); v != "" {
			*field = v
		}
	}
	cfg.Server = normalizeServer(cfg.Server)
	return cfg, nil
}

// Overridden names the environment variables that hold a value other than
// c's, which is what every send would use instead of c's.
func (c *Config) Overridden() []string {
	var names []string
	for name, field := range c.envFields() {
		v := os.Getenv(name)
		if v == "" {
			continue
		}
		if name == "NAVI_SERVER" {
			v = normalizeServer(v)
		}
		if v != *field {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

// normalizeServer drops a trailing slash, so the topic URL never doubles
// it, and fills in the public server for an empty value.
func normalizeServer(s string) string {
	s = strings.TrimRight(s, "/")
	if s == "" {
		return DefaultServer
	}
	return s
}

// Load is Read for sending: it fails unless the result names a topic
// and a server that can be posted to.
func Load() (*Config, error) {
	cfg, err := Read()
	if err != nil {
		return nil, err
	}
	if cfg.Topic == "" {
		return nil, fmt.Errorf("%w (looked in %s and NAVI_TOPIC)", ErrNotConfigured, Path())
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	u, err := url.Parse(c.Server)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("server must be an http(s) URL, got %q", c.Server)
	}
	if !topicPattern.MatchString(c.Topic) {
		return fmt.Errorf("topic must be 1-64 letters, digits, '-' or '_', got %q", c.Topic)
	}
	return nil
}

// Save validates the config and writes it owner-only, since the topic is
// as good as a password on a public server and the token is one outright.
// It goes through a temp file so an interrupted write cannot leave half a
// config.
func (c *Config) Save() error {
	c.Server = normalizeServer(c.Server)
	if err := c.Validate(); err != nil {
		return err
	}
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// RandomTopic returns an unguessable topic name. Anyone who knows a topic
// on a public server can read it and post to it, so its name is the only
// thing keeping strangers out.
func RandomTopic() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return "navi-" + hex.EncodeToString(b)
}
