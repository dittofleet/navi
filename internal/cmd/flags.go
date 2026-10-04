package cmd

import (
	"fmt"
	"strings"
)

// flags declares what a command accepts. Keys are flag names without
// dashes. Registering several keys against the same pointer gives a flag
// its aliases (e.g. "t" and "title"). A flag in lists collects every
// value it is given instead of keeping only the last.
type flags struct {
	bools  map[string]*bool
	values map[string]*string
	lists  map[string]*[]string
}

func yesFlag(target *bool) map[string]*bool {
	return map[string]*bool{"y": target, "yes": target}
}

// declares reports whether arg is one of this command's flags, or the "--"
// that ends them. Text that merely starts with a dash is not: a
// message can begin with one.
func (f flags) declares(arg string) bool {
	if arg == "--" {
		return true
	}
	if len(arg) < 2 || !strings.HasPrefix(arg, "-") {
		return false
	}
	name, _, _ := strings.Cut(strings.TrimLeft(arg, "-"), "=")
	return f.bools[name] != nil || f.values[name] != nil || f.lists[name] != nil
}

// parse pulls the declared flags out of args and returns the positionals.
// Flags may appear anywhere, and a bare "--" ends flag parsing so a
// message starting with a dash can still be written literally.
func (f flags) parse(args []string, usage string) ([]string, error) {
	positional := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if len(arg) < 2 || !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			continue
		}

		name, inline, hasInline := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		if target, isBool := f.bools[name]; isBool {
			if hasInline {
				return nil, fmt.Errorf("flag takes no value: %s\n%s", arg, usage)
			}
			*target = true
			continue
		}

		single, list := f.values[name], f.lists[name]
		if single == nil && list == nil {
			return nil, fmt.Errorf("unknown flag: %s\n%s", arg, usage)
		}
		value := inline
		// Another flag is never the value, so an unset variable in
		// `send -t $TITLE --click ...` is an error rather than a title
		// reading "--click".
		if !hasInline && i+1 < len(args) && !f.declares(args[i+1]) {
			i++
			value = args[i]
		}
		// An empty value is rejected rather than ignored: a script passing
		// an unset variable should not quietly fall back to the default.
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("flag needs a value: %s\n%s", arg, usage)
		}
		if list != nil {
			*list = append(*list, value)
			continue
		}
		*single = value
	}
	return positional, nil
}
