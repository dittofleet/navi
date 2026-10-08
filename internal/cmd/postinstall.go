package cmd

import (
	"fmt"
	"os"

	clikit "github.com/dittofleet/go-cli-kit"
	"github.com/dittofleet/go-cli-kit/postinstall"
	"github.com/dittofleet/navi/internal/config"
)

// Postinstall is the install script's first-time setup. NAVI_TOPIC (with
// NAVI_SERVER and NAVI_TOKEN if needed) in the environment makes a fresh
// remote machine a single curl-pipe: setup saves whatever they hold.
// Without them a config already there, perhaps synced from another
// machine, is left as it is.
func Postinstall(a clikit.App) error {
	return postinstall.Run(a, func() error {
		if os.Getenv("NAVI_TOPIC") != "" {
			return Setup(nil)
		}
		if _, err := os.Stat(config.Path()); err != nil {
			fmt.Println("Run `navi setup` to pick a topic and see how to subscribe on your phone.")
		}
		return nil
	})
}
