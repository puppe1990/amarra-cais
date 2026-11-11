package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const generateKindsUsage = "handler, page, component, migration, resource, model, stream, live, job, console, auth, ci, or sitemap"

type generatorRun func(cwd string, rest []string, dryRun bool) error

func withGeneratedName(kind, usage string, run func(cwd, name string, rest []string, dryRun bool) error) generatorRun {
	return func(cwd string, rest []string, dryRun bool) error {
		if len(rest) < 1 {
			return fmt.Errorf("usage: amarra-cais g %s", usage)
		}
		if err := validateGeneratedName(kind, rest[0]); err != nil {
			return err
		}
		return run(cwd, rest[0], rest[1:], dryRun)
	}
}

var generateCmds = map[string]generatorRun{
	"console": func(cwd string, _ []string, dryRun bool) error {
		return scaffoldConsole(cwd, dryRun)
	},
	"auth": func(cwd string, _ []string, dryRun bool) error {
		return scaffoldAuth(cwd, scaffoldData{AppName: filepath.Base(cwd), ModulePath: moduleFromDir(cwd)}, dryRun)
	},
	"ci": func(cwd string, _ []string, dryRun bool) error {
		return scaffoldCI(cwd, scaffoldData{AppName: filepath.Base(cwd), ModulePath: moduleFromDir(cwd)}, dryRun)
	},
	"sitemap": func(cwd string, _ []string, dryRun bool) error {
		return scaffoldSitemap(cwd, dryRun)
	},
	"job": withGeneratedName("job", `job <name> [--cron "0 3 * * *"]`, func(cwd, name string, rest []string, dryRun bool) error {
		opts, err := parseJobOpts(rest)
		if err != nil {
			return err
		}
		opts.dryRun = dryRun
		return scaffoldJob(cwd, name, opts)
	}),
	"live": withGeneratedName("live", "live <name>", func(cwd, name string, _ []string, dryRun bool) error {
		return scaffoldLive(cwd, name, dryRun)
	}),
	"component": withGeneratedName("component", "component <name>|--list", func(cwd, name string, _ []string, dryRun bool) error {
		return scaffoldComponent(cwd, name, dryRun)
	}),
	"handler": withGeneratedName("handler", "handler <name>", func(cwd, name string, _ []string, dryRun bool) error {
		return scaffoldHandler(cwd, name, dryRun)
	}),
	"page": withGeneratedName("page", "page <name>", func(cwd, name string, _ []string, dryRun bool) error {
		return scaffoldPage(cwd, name, dryRun)
	}),
	"migration": withGeneratedName("migration", "migration <name>", func(cwd, name string, _ []string, dryRun bool) error {
		return scaffoldMigration(cwd, name, dryRun)
	}),
	"resource": withGeneratedName("resource", "resource <name>", func(cwd, name string, rest []string, dryRun bool) error {
		opts, err := parseResourceOpts(rest)
		if err != nil {
			return err
		}
		opts.dryRun = dryRun
		return scaffoldResource(cwd, name, opts)
	}),
	"model": withGeneratedName("model", "model <name>", func(cwd, name string, rest []string, dryRun bool) error {
		opts, err := parseModelOpts(rest)
		if err != nil {
			return err
		}
		opts.dryRun = dryRun
		return scaffoldModel(cwd, name, opts)
	}),
	"stream": func(cwd string, rest []string, dryRun bool) error {
		if len(rest) < 1 || rest[0] != "chat" {
			return fmt.Errorf("usage: amarra-cais g stream chat [--live]")
		}
		opts := streamOpts{dryRun: dryRun}
		for _, a := range rest[1:] {
			if a == "--live" {
				opts.live = true
			}
		}
		return scaffoldStreamChat(cwd, opts)
	},
}

func generateAppDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if isCaisApp(cwd) {
		return cwd, nil
	}
	if isCaisFramework(cwd) {
		return "", fmt.Errorf("you are inside the Cais framework directory — cd into your app first")
	}
	return "", fmt.Errorf("not a Cais app (missing go.mod with github.com/puppe1990/amarra-cais as a dependency)")
}

func (c *CLI) cmdGenerate(args []string) error {
	dryRun, args := stripFlag(args, "--dry-run")
	setScaffoldOut(c.Out)

	if len(args) < 1 {
		return fmt.Errorf("usage: amarra-cais g [--dry-run] <handler|page|component|migration|resource|model|stream|live|job|console|auth|ci|sitemap> [name]")
	}

	kind := args[0]
	// --list is discovery and needs no app: the shipped kit is framework-side (#63).
	if kind == "component" && len(args) >= 2 && args[1] == "--list" {
		printShippedComponents(c.Out)
		return nil
	}
	cwd, err := generateAppDir()
	if err != nil {
		return err
	}

	run, ok := generateCmds[kind]
	if !ok {
		return fmt.Errorf("unknown generator %q (use %s)", kind, generateKindsUsage)
	}
	if err := run(cwd, args[1:], dryRun); err != nil {
		return err
	}
	if !dryRun {
		printGenerateNextSteps(c.Out, kind)
	}
	return nil
}

func printGenerateNextSteps(w io.Writer, kind string) {
	_, _ = fmt.Fprintln(w)
	switch kind {
	case "resource", "model", "migration", "auth", "stream", "sitemap":
		_, _ = fmt.Fprintln(w, "=> Next: amarra-cais db migrate && amarra-cais test")
	default:
		_, _ = fmt.Fprintln(w, "=> Next: amarra-cais test")
	}
}
