package cmd

import (
	"fmt"

	clikit "github.com/dittofleet/go-cli-kit"
	"github.com/dittofleet/go-cli-kit/uninstall"
	"github.com/dittofleet/navi/internal/config"
)

const uninstallUsage = "usage: navi uninstall [--yes]"

// Uninstall removes the navi binary, its config file, and its update
// cache.
//
// It removes navi's own files rather than its directories: ~/.config/navi
// and ~/.local/share/navi are also where the unrelated navi cheatsheet
// tool (github.com/denisidoro/navi) keeps its config and cheatsheets, so
// a directory is only removed once it is empty.
func Uninstall(args []string, a clikit.App) error {
	var yes bool
	args, err := flags{bools: yesFlag(&yes)}.parse(args, uninstallUsage)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %v\n%s", args, uninstallUsage)
	}
	return uninstall.Run(a, yes, uninstall.Plan{Items: []uninstall.Item{
		{Label: "Config", Path: config.Path(), Note: "your topic, and token if any", Remove: uninstall.FileAndEmptyDir},
	}})
}
