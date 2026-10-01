// amarra-cais new argument parsing, split from command dispatch (#288).
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const newUsage = "usage: amarra-cais new <app> [dir] [--minimal] [--blank] [--module <path>] [--no-git]"

func (c *CLI) cmdNew(args []string) error {
	if newArgsWantHelp(args) {
		_, _ = fmt.Fprintln(c.Out, newUsage)
		return nil
	}
	if len(args) == 0 {
		return fmt.Errorf("%s", newUsage)
	}

	opts, err := parseNewArgs(args)
	if err != nil {
		return err
	}

	abs, err := filepath.Abs(opts.dir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(abs); err == nil {
		return fmt.Errorf("directory %s already exists", abs)
	}

	module := opts.module
	if module == "" {
		module = moduleName(opts.name)
	}
	if err := scaffoldNewApp(abs, scaffoldData{
		AppName:     opts.name,
		ModulePath:  module,
		CaisVersion: scaffoldModuleVersion(),
	}, opts.minimal, opts.blank); err != nil {
		return err
	}

	if replace := os.Getenv("CAIS_REPLACE"); replace != "" {
		printLinkMessage(c.Out, replace)
	}

	if err := initAppRepo(c.Out, abs, opts.noGit); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(c.Out, "Created app %q at %s\n\nNext steps:\n  cd %s\n  amarra-cais install\n  amarra-cais dev\n", opts.name, abs, abs)
	return nil
}

type newOpts struct {
	name    string
	dir     string
	minimal bool
	blank   bool
	module  string
	noGit   bool
}

func parseNewArgs(args []string) (newOpts, error) {
	opts := newOpts{}
	positional := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--minimal":
			opts.minimal = true
		case "--blank":
			opts.blank = true
		case "--no-git":
			opts.noGit = true
		case "--module":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--module requires a value")
			}
			i++
			opts.module = args[i]
		default:
			if strings.HasPrefix(args[i], "-") {
				return opts, fmt.Errorf("unknown flag %s", args[i])
			}
			positional = append(positional, args[i])
		}
	}
	if len(positional) == 0 {
		return opts, fmt.Errorf("%s", newUsage)
	}

	opts.name = positional[0]
	opts.dir = opts.name
	if len(positional) > 1 {
		opts.dir = positional[1]
	}
	return opts, nil
}

func newArgsWantHelp(args []string) bool {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return true
		}
	}
	return false
}
