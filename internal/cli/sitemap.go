// cais g sitemap: blog (posts resource) plus a dynamic /sitemap.xml.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// sitemapPostFields is the blog schema: public posts with slug URLs and a
// published flag the sitemap filters on.
const sitemapPostFields = "title:string,slug:string,body:text,published:bool"

func scaffoldSitemap(dir string, dryRun bool) error {
	postModel := filepath.Join(dir, "internal/models/post.go")
	if _, err := os.Stat(postModel); os.IsNotExist(err) {
		if err := scaffoldResource(dir, "post", resourceOpts{
			Fields:    sitemapPostFields,
			Public:    true,
			Paginate:  true,
			Seed:      true,
			AdminAuth: "session",
			dryRun:    dryRun,
		}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if err := checkSitemapPostModel(postModel); err != nil {
		return err
	}

	routesPath := filepath.Join(dir, "internal/app/routes.go")
	routesBody, err := os.ReadFile(routesPath)
	if err != nil {
		return err
	}
	if strings.Contains(string(routesBody), "/sitemap.xml") {
		return nil
	}

	modulePath := readModulePath(dir)
	files := map[string]string{
		filepath.Join("internal/handlers", "sitemap.go"):      buildSitemapHandler(modulePath),
		filepath.Join("internal/handlers", "sitemap_test.go"): buildSitemapTest(modulePath),
	}
	for rel, content := range files {
		if err := writeScaffoldFile(filepath.Join(dir, rel), []byte(content), 0o644, rel, dryRun); err != nil {
			return err
		}
	}

	content, err := insertBeforeFunctionEnd(string(routesBody), "registerRoutes", sitemapRouteInsert)
	if err != nil {
		return fmt.Errorf("could not patch routes.go: %w", err)
	}
	if err := updateScaffoldFile(routesPath, []byte(content), "internal/app/routes.go", dryRun); err != nil {
		return err
	}
	if dryRun {
		return nil
	}
	return gofmtGoFiles(dir)
}

const sitemapRouteInsert = "\tsitemap := handlers.NewSitemapHandler(deps.Store, deps.Site, cfg)\n\tr.Get(\"/sitemap.xml\", sitemap.ServeHTTP)\n"

// checkSitemapPostModel ensures an existing posts table has the slug +
// published columns the sitemap handler queries.
func checkSitemapPostModel(path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(body)
	// gofmt aligns struct fields (Slug      string), so match name + type loosely.
	hasSlug := regexp.MustCompile(`(?m)^\tSlug\s+string\b`).MatchString(src)
	hasPublished := regexp.MustCompile(`(?m)^\tPublished\s+bool\b`).MatchString(src)
	if !hasSlug || !hasPublished {
		return fmt.Errorf("g sitemap needs posts with slug:string and published:bool (models/post.go already exists without them)")
	}
	return nil
}

func buildSitemapHandler(modulePath string) string {
	return fmt.Sprintf(`package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"%s/pkg/cais"
	"%s/pkg/cais/meta"

	"%s/internal/store"
)

type SitemapHandler struct {
	store store.Store
	site  meta.Site
	cfg   cais.Config
}

func NewSitemapHandler(s store.Store, site meta.Site, cfg cais.Config) *SitemapHandler {
	return &SitemapHandler{store: s, site: site, cfg: cfg}
}

// ServeHTTP renders /sitemap.xml: static pages plus published posts.
func (h *SitemapHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	posts, err := h.store.ListAllPosts("", "id", "desc")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	base := strings.TrimSuffix(strings.TrimSpace(h.site.AppURL), "/")
	var b strings.Builder
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	b.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
	writeSitemapURL(&b, base+"/")
	writeSitemapURL(&b, base+"/posts")
	for _, p := range posts {
		if !p.Published || p.Slug == "" {
			continue
		}
		writeSitemapURL(&b, base+"/posts/"+url.PathEscape(p.Slug), p.CreatedAt.UTC().Format("2006-01-02"))
	}
	b.WriteString("</urlset>")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write([]byte(b.String()))
}

func writeSitemapURL(b *strings.Builder, loc string, lastmod ...string) {
	b.WriteString("  <url><loc>")
	_ = xml.EscapeText(b, []byte(loc))
	b.WriteString("</loc>")
	if len(lastmod) > 0 && lastmod[0] != "" {
		fmt.Fprintf(b, "<lastmod>%%s</lastmod>", lastmod[0])
	}
	b.WriteString("</url>\n")
}
`, frameworkModule, frameworkModule, modulePath)
}

func buildSitemapTest(modulePath string) string {
	return fmt.Sprintf(`package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"%s/pkg/cais"
	"%s/internal/models"
)

func TestSitemapHandler_ServesXML(t *testing.T) {
	s := setupTestStore(t)
	if _, err := s.InsertPost(models.Post{Title: "Hello", Slug: "hello", Body: "x", Published: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.InsertPost(models.Post{Title: "Draft", Slug: "draft", Body: "x"}); err != nil {
		t.Fatal(err)
	}
	h := NewSitemapHandler(s, testSite(), cais.Config{})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %%d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/xml; charset=utf-8" {
		t.Errorf("Content-Type = %%q", ct)
	}
	body := rr.Body.String()
	for _, want := range []string{"<urlset", "https://cais.example.com/posts/hello", "<lastmod>"} {
		if !strings.Contains(body, want) {
			t.Errorf("sitemap missing %%q:\n%%s", want, body)
		}
	}
	if strings.Contains(body, "/posts/draft") {
		t.Errorf("sitemap must not list unpublished posts:\n%%s", body)
	}
}
`, frameworkModule, modulePath)
}
