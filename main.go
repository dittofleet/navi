package main

import (
	"errors"
	"fmt"
	"os"

	clikit "github.com/dittofleet/go-cli-kit"
	"github.com/dittofleet/go-cli-kit/selfupdate"
	"github.com/dittofleet/go-cli-kit/updatecheck"
	"github.com/dittofleet/navi/internal/app"
	"github.com/dittofleet/navi/internal/cmd"
	"github.com/dittofleet/navi/internal/config"
)

var errUnknownCommand = errors.New("unknown command")

var version = "dev"

const usage = `Usage: navi <command>

Commands:
  send "<message>"         Send a notification to your phone
  setup                    Pick where notifications go, and show how to subscribe
  update                   Download and install the latest version
  uninstall [--yes]        Remove the binary, config, and cache
  version                  Print the installed version
  help                     Print this help message

Flags:
  -t, --title <text>       (send) Title, defaulting to "<repo> on <machine>"
  -p, --priority <level>   (send) min, low, default, high, urgent, or 1-5
      --tag <tag>          (send) Tag or emoji shortcode, e.g. warning or
                           white_check_mark. Repeat it or separate with commas
      --click <url>        (send) Open this URL when the notification is tapped
      --topic <name>       (setup) Topic to post to, generated if never set
      --rotate             (setup) Replace the topic with a new random one
      --server <url>       (setup) ntfy server, defaulting to https://ntfy.sh
      --token <token>      (setup) Access token, for servers that require one
      --                   Everything after this is the message, not flags

Notifications go through ntfy (https://ntfy.sh): navi posts to a topic, and
the ntfy app on your phone, subscribed to the same topic, shows them.

The config lives in ~/.config/navi/config.json. NAVI_TOPIC, NAVI_SERVER and
NAVI_TOKEN override it, and setup saves what they hold.
`

func printUsage() {
	fmt.Print(usage)
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		os.Exit(0)
	}

	navi := app.New(version)
	if err := dispatch(navi, args); err != nil {
		switch {
		case errors.Is(err, errUnknownCommand):
			// Naming it catches the common slip of putting a flag before
			// the command, where a bare usage dump explains nothing.
			fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n\n", args[0])
			printUsage()
		case errors.Is(err, config.ErrNotConfigured):
			fmt.Fprintln(os.Stderr, "Error:", err)
			fmt.Fprintln(os.Stderr, "Run `navi setup` to pick a topic.")
		default:
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
		os.Exit(1)
	}

	updatecheck.MaybeCheck(navi, args[0])
}

func dispatch(navi clikit.App, args []string) error {
	switch args[0] {
	case "send":
		return cmd.Send(args[1:])
	case "setup":
		return cmd.Setup(args[1:])
	case "update":
		_, err := selfupdate.Run(navi)
		return err
	case "uninstall":
		return cmd.Uninstall(args[1:], navi)
	case "postinstall":
		return cmd.Postinstall(navi)
	case "version", "--version", "-v":
		fmt.Println(version)
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return errUnknownCommand
	}
}
