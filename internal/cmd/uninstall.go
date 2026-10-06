package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	clikit "github.com/dittofleet/go-cli-kit"
	"github.com/dittofleet/go-cli-kit/updatecheck"
	"github.com/dittofleet/navi/internal/config"
	"golang.org/x/term"
)

const uninstallUsage = "usage: navi uninstall [--yes]"

// Uninstall removes the navi binary, its config file, and its update
// cache. Order is cache → config → binary so a failure leaves a tool to
// retry with.
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

	if a.IsDev() {
		return errors.New("cannot uninstall a dev build")
	}

	binaryPath, err := clikit.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine binary path: %w", err)
	}

	configFile := config.Path()
	cacheFile := updatecheck.CachePath(a)

	fmt.Println("This will remove:")
	fmt.Printf("  - Binary:  %s\n", binaryPath)
	fmt.Printf("  - Config:  %s  (your topic, and token if any)\n", configFile)
	fmt.Printf("  - Cache:   %s\n", cacheFile)
	fmt.Println()

	if !yes {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return errors.New("refusing to uninstall non-interactively without --yes")
		}
		fmt.Print("Proceed? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		answer := strings.ToLower(strings.TrimSpace(line))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	steps := []struct {
		label string
		path  string
		fn    func(string) error
	}{
		{"cache", cacheFile, os.Remove},
		{"cache directory", filepath.Dir(cacheFile), removeIfEmpty},
		{"config", configFile, os.Remove},
		{"config directory", filepath.Dir(configFile), removeIfEmpty},
		{"binary", binaryPath, os.Remove},
	}
	var removed []string
	for _, s := range steps {
		err := s.fn(s.path)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			if len(removed) > 0 {
				fmt.Fprintf(os.Stderr, "Removed before failure: %s\n", strings.Join(removed, ", "))
			}
			return fmt.Errorf("failed to remove %s (%s): %w", s.label, s.path, err)
		}
		removed = append(removed, s.label)
	}

	fmt.Println("Uninstalled navi.")
	return nil
}

// removeIfEmpty removes a directory only when nothing else is left in it.
func removeIfEmpty(dir string) error {
	err := os.Remove(dir)
	if errors.Is(err, syscall.ENOTEMPTY) || errors.Is(err, syscall.EEXIST) {
		return nil
	}
	return err
}
