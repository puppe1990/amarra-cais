package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// g sitemap scaffolds a blog (posts resource) plus a dynamic /sitemap.xml.
func TestScaffoldSitemap_BlogPlusXML(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "sitemapapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "sitemapapp",
		ModulePath: "github.com/puppe1990/sitemapapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldSitemap(appDir, false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/handlers/sitemap.go",
		"internal/handlers/sitemap_test.go",
		"internal/models/post.go",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	handler, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/sitemap.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"sitemap.xml",
		"ListAllPosts",
		"Published",
		"application/xml",
	} {
		if !strings.Contains(string(handler), want) {
			t.Errorf("sitemap handler missing %q", want)
		}
	}

	routes, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(routes), "/sitemap.xml") != 1 {
		t.Errorf("routes.go should register /sitemap.xml exactly once")
	}

	// Rerun reuses posts instead of duplicating migration/handler/routes.
	if err := scaffoldSitemap(appDir, false); err != nil {
		t.Fatalf("rerun: %v", err)
	}
	routes, err = os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(routes), "/sitemap.xml") != 1 {
		t.Error("rerun duplicated the /sitemap.xml route")
	}
}

func TestScaffoldSitemap_GeneratedTestsPass(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	caisDir := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	t.Setenv("CAIS_REPLACE", caisDir)

	appDir := filepath.Join(t.TempDir(), "sitemaprun")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "sitemaprun",
		ModulePath: "github.com/puppe1990/sitemaprun",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldSitemap(appDir, false); err != nil {
		t.Fatal(err)
	}

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = appDir
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}

	cmd := exec.Command("go", "test", "./internal/handlers/", "-run", "TestSitemapHandler", "-count=1")
	cmd.Dir = appDir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("sitemap handler test failed: %v\n%s", err, out)
	}
}
