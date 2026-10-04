package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dittofleet/navi/internal/config"
	"github.com/dittofleet/navi/internal/ntfy"
)

const sendUsage = `usage: navi send "<message>" [-t <title>] [-p <priority>] [--tag <tag>]... [--click <url>]`

const sendTimeout = 15 * time.Second

func Send(args []string) error {
	var msg ntfy.Message
	var priority string
	var tags []string
	args, err := flags{
		values: map[string]*string{
			"t": &msg.Title, "title": &msg.Title,
			"p": &priority, "priority": &priority,
			"click": &msg.Click,
		},
		lists: map[string]*[]string{"tag": &tags},
	}.parse(args, sendUsage)
	if err != nil {
		return err
	}
	if len(args) != 1 {
		return fmt.Errorf("give the message as one quoted argument\n%s", sendUsage)
	}
	msg.Message = strings.TrimSpace(args[0])
	if msg.Message == "" {
		return fmt.Errorf("the message is empty\n%s", sendUsage)
	}
	msg.Tags = splitTags(tags)
	if msg.Title == "" {
		msg.Title = defaultTitle()
	}
	if priority != "" {
		if msg.Priority, err = ntfy.ParsePriority(priority); err != nil {
			return err
		}
	}
	if msg.Click != "" {
		if u, err := url.Parse(msg.Click); err != nil || u.Scheme == "" {
			return fmt.Errorf("--click must be a URL, got %q", msg.Click)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	msg.Topic = cfg.Topic

	ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()
	if err := ntfy.Publish(ctx, cfg.Server, cfg.Token, msg); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("%s did not answer within %s", cfg.Server, sendTimeout)
		}
		return err
	}
	fmt.Println("Sent")
	return nil
}

// splitTags flattens repeated and comma-separated tags into one list.
func splitTags(raw []string) []string {
	var tags []string
	for _, r := range raw {
		for _, t := range strings.Split(r, ",") {
			if t = strings.TrimSpace(t); t != "" {
				tags = append(tags, t)
			}
		}
	}
	return tags
}

// defaultTitle says where a notification came from, "<project> on
// <machine>", since a phone collecting them from agents in several repos
// on several machines is otherwise left guessing.
func defaultTitle() string {
	host, _ := os.Hostname()
	// macOS reports "Name.local", and a fully qualified name says more
	// than a notification title has room for.
	host, _, _ = strings.Cut(host, ".")

	project := projectName()
	switch {
	case project != "" && host != "":
		return project + " on " + host
	case host != "":
		return host
	default:
		return project
	}
}

// projectName is the repo name from the origin remote, so a worktree
// carries the name of its repo rather than of its directory. A repo
// without an origin goes by its directory, and outside a repo there is
// no project to name.
func projectName() string {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if name := repoName(string(out)); err == nil && name != "" {
		return name
	}
	out, err = exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if top := strings.TrimSpace(string(out)); err == nil && top != "" {
		return filepath.Base(top)
	}
	return ""
}

// repoName takes the repo's name off a remote URL. Both
// https://host/owner/repo.git and git@host:owner/repo end in it, after the
// last separator.
func repoName(remote string) string {
	remote = strings.TrimSuffix(strings.TrimRight(strings.TrimSpace(remote), "/"), ".git")
	if i := strings.LastIndexAny(remote, "/:"); i >= 0 {
		remote = remote[i+1:]
	}
	return remote
}
