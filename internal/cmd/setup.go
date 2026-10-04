package cmd

import (
	"fmt"
	"strings"

	"github.com/dittofleet/navi/internal/config"
)

const setupUsage = "usage: navi setup [--topic <name> | --rotate] [--server <url>] [--token <token>]"

// Setup saves where notifications go and tells the user how to subscribe.
// Each value comes from its flag, then the NAVI_* environment, then the
// config already saved, so a bare rerun changes nothing and just shows
// the subscription again. A topic nobody named is generated, and so is
// one replacing a topic that leaked (--rotate).
func Setup(args []string) error {
	var topic, server, token string
	var rotate bool
	args, err := flags{
		values: map[string]*string{
			"topic":  &topic,
			"server": &server,
			"token":  &token,
		},
		bools: map[string]*bool{"rotate": &rotate},
	}.parse(args, setupUsage)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %v\n%s", args, setupUsage)
	}
	if rotate && topic != "" {
		return fmt.Errorf("--topic and --rotate both pick the topic, give one\n%s", setupUsage)
	}

	cfg, err := config.Read()
	if err != nil {
		return err
	}
	if topic != "" {
		cfg.Topic = topic
	}
	if server != "" {
		cfg.Server = server
	}
	if token != "" {
		cfg.Token = token
	}
	generated := rotate || cfg.Topic == ""
	if generated {
		cfg.Topic = config.RandomTopic()
	}
	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Saved %s\n\n", config.Path())
	if generated {
		fmt.Println("Generated a new topic. Anyone who knows its name can read and post to it,")
		fmt.Println("so keep it to yourself.")
		fmt.Println()
	}
	if rotate {
		fmt.Println("The old topic gets nothing more from this machine. Give every other machine")
		fmt.Println("the new topic too, and unsubscribe from the old one in the app.")
		fmt.Println()
	}
	// The environment wins over the file on every send, so a value saved
	// here would be silently ignored.
	if names := cfg.Overridden(); len(names) > 0 {
		list := strings.Join(names, " and ")
		fmt.Printf("Note: your environment overrides what was saved here (%s).\n", list)
		fmt.Printf("Unset %s, or sends from this shell keep using the old values.\n", list)
		fmt.Println()
	}
	fmt.Println("To get notifications on your phone, install the ntfy app (iOS or Android),")
	fmt.Println("tap +, and subscribe to:")
	fmt.Println()
	fmt.Printf("  topic:   %s\n", cfg.Topic)
	if cfg.Server != config.DefaultServer {
		fmt.Printf("  server:  %s  (turn on \"Use another server\")\n", cfg.Server)
	}
	fmt.Println()
	fmt.Printf("Or follow along in a browser at %s/%s\n\n", cfg.Server, cfg.Topic)
	fmt.Println(`Then try: navi send "hello from navi"`)
	return nil
}
