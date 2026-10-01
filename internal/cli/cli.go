package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CLI struct {
	Out io.Writer
}

func Main() int {
	c := &CLI{Out: os.Stdout}
	if err := c.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "amarra-cais: %v\n", err)
		return 1
	}
	return 0
}

func (c *CLI) Run(args []string) error {
	if len(args) == 0 {
		c.printHelp()
		return nil
	}

	switch args[0] {
	case "new":
		return c.cmdNew(args[1:])
	case "generate", "g":
		return c.cmdGenerate(args[1:])
	case "install", "i":
		return c.cmdInstall()
	case "css":
		return c.cmdCSS()
	case "dev":
		return c.cmdDev()
	case "build", "b":
		return c.cmdBuild(args[1:])
	case "server", "s":
		return c.cmdServer()
	case "test":
		return c.cmdTest()
	case "doctor":
		return c.cmdDoctor(args[1:])
	case "pwa":
		return c.cmdPWA(args[1:])
	case "link":
		return c.cmdLink(args[1:])
	case "upgrade":
		return c.cmdUpgrade(args[1:])
	case "console", "c":
		return c.cmdConsole()
	case "db":
		return c.cmdDB(args[1:])
	case "jobs":
		return c.cmdJobs(args[1:])
	case "routes":
		return c.cmdRoutes(args[1:])
	case "destroy", "d":
		return c.cmdDestroy(args[1:])
	case "version", "-v", "--version":
		return c.cmdVersion()
	case "help", "-h", "--help":
		c.printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command %q (run amarra-cais help)", args[0])
	}
}

func (c *CLI) cmdGenerate(args []string) error {
	dryRun := false
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--dry-run" {
			dryRun = true
			continue
		}
		filtered = append(filtered, arg)
	}
	args = filtered
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
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !isCaisApp(cwd) {
		if isCaisFramework(cwd) {
			return fmt.Errorf("you are inside the Cais framework directory — cd into your app first")
		}
		return fmt.Errorf("not a Cais app (missing go.mod with github.com/puppe1990/amarra-cais as a dependency)")
	}

	var genErr error
	switch kind {
	case "console":
		genErr = scaffoldConsole(cwd, dryRun)
	case "auth":
		genErr = scaffoldAuth(cwd, scaffoldData{AppName: filepath.Base(cwd), ModulePath: moduleFromDir(cwd)}, dryRun)
	case "ci":
		genErr = scaffoldCI(cwd, scaffoldData{AppName: filepath.Base(cwd), ModulePath: moduleFromDir(cwd)}, dryRun)
	case "job":
		if len(args) < 2 {
			return fmt.Errorf("usage: amarra-cais g job <name> [--cron \"0 3 * * *\"]")
		}
		if err := validateGeneratedName(kind, args[1]); err != nil {
			return err
		}
		opts, parseErr := parseJobOpts(args[2:])
		if parseErr != nil {
			return parseErr
		}
		opts.dryRun = dryRun
		genErr = scaffoldJob(cwd, args[1], opts)
	case "live":
		if len(args) < 2 {
			return fmt.Errorf("usage: amarra-cais g live <name>")
		}
		if err := validateGeneratedName(kind, args[1]); err != nil {
			return err
		}
		genErr = scaffoldLive(cwd, args[1], dryRun)
	case "stream":
		if len(args) < 2 || args[1] != "chat" {
			return fmt.Errorf("usage: amarra-cais g stream chat [--live]")
		}
		opts := streamOpts{dryRun: dryRun}
		for _, a := range args[2:] {
			if a == "--live" {
				opts.live = true
			}
		}
		genErr = scaffoldStreamChat(cwd, opts)
	case "component":
		if len(args) < 2 {
			return fmt.Errorf("usage: amarra-cais g component <name>|--list")
		}
		if err := validateGeneratedName(kind, args[1]); err != nil {
			return err
		}
		genErr = scaffoldComponent(cwd, args[1], dryRun)
	case "sitemap":
		genErr = scaffoldSitemap(cwd, dryRun)
	case "handler", "page", "migration", "resource", "model":
		if len(args) < 2 {
			return fmt.Errorf("usage: amarra-cais g %s <name>", kind)
		}
		name := args[1]
		if err := validateGeneratedName(kind, name); err != nil {
			return err
		}
		switch kind {
		case "handler":
			genErr = scaffoldHandler(cwd, name, dryRun)
		case "page":
			genErr = scaffoldPage(cwd, name, dryRun)
		case "migration":
			genErr = scaffoldMigration(cwd, name, dryRun)
		case "resource":
			opts, parseErr := parseResourceOpts(args[2:])
			if parseErr != nil {
				return parseErr
			}
			opts.dryRun = dryRun
			genErr = scaffoldResource(cwd, name, opts)
		case "model":
			opts, parseErr := parseModelOpts(args[2:])
			if parseErr != nil {
				return parseErr
			}
			opts.dryRun = dryRun
			genErr = scaffoldModel(cwd, name, opts)
		}
	default:
		return fmt.Errorf("unknown generator %q (use handler, page, component, migration, resource, model, stream, auth, ci, app, sitemap, or console)", kind)
	}
	if genErr != nil {
		return genErr
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
	case "app":
		_, _ = fmt.Fprintln(w, "=> Next: amarra-cais install && amarra-cais dev")
	default:
		_, _ = fmt.Fprintln(w, "=> Next: amarra-cais test")
	}
}

func (c *CLI) cmdVersion() error {
	_, _ = fmt.Fprintln(c.Out, frameworkVersion())
	return nil
}

func (c *CLI) cmdServer() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	if err := ensureStylesCSS(c.Out, dir); err != nil {
		// Soft: still start the server, but leave a clear signal that CSS is missing.
		_, _ = fmt.Fprintf(c.Out, "⚠ %v\n", err)
	}
	_, _ = fmt.Fprintln(c.Out, "→ go run ./cmd/server")
	return runCmd(dir, "go", "run", "./cmd/server")
}

func (c *CLI) cmdDoctor(args []string) error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	opts := doctorOptions{}
	for _, arg := range args {
		if arg == "--mobile" {
			opts.Mobile = true
		}
	}
	return runDoctor(c.Out, dir, opts)
}

func (c *CLI) cmdTest() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	return runCmd(dir, "go", "test", "./...", "-race", "-count=1")
}

func moduleName(app string) string {
	slug := strings.ToLower(strings.ReplaceAll(app, "-", ""))
	return "github.com/puppe1990/" + slug
}

func moduleFromDir(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return moduleName(filepath.Base(dir))
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return moduleName(filepath.Base(dir))
}

func isCaisApp(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return false
	}
	content := string(data)
	if strings.HasPrefix(content, "module github.com/puppe1990/amarra-cais") {
		return false
	}
	return strings.Contains(content, "github.com/puppe1990/amarra-cais")
}

func isCaisFramework(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return false
	}
	return strings.HasPrefix(string(data), "module github.com/puppe1990/amarra-cais")
}
