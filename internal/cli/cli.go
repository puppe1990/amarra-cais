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

func stripFlag(args []string, name string) (bool, []string) {
	out := make([]string, 0, len(args))
	found := false
	for _, arg := range args {
		if arg == name {
			found = true
			continue
		}
		out = append(out, arg)
	}
	return found, out
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
