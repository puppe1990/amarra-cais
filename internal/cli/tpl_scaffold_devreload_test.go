package cli

import (
	"strings"
	"testing"
)

// #340: the scaffold has to carry the dev reload wiring — disk templates, the
// in-process swap, the SSE hub and the air config that stops rebuilding HTML.
func TestScaffoldMain_loadsDiskTemplatesAndWatchesThem(t *testing.T) {
	for _, token := range []string{
		"devreload.Watch(ctx, cfg, views.ReloadFromDir)",
		"func templateFS(cfg cais.Config) (fs.FS, error) {",
		"os.DirFS(dir)",
		"DevReload: devReload,",
		`fs.Sub(web.Templates, "templates")`,
	} {
		if !strings.Contains(tplMain, token) {
			t.Errorf("scaffold cmd/server/main.go missing %q (#340)", token)
		}
	}
}

func TestScaffoldApp_registersDevReloadRoute(t *testing.T) {
	for _, blob := range []struct {
		name string
		body string
	}{{"full", tplApp}, {"blank", tplAppBlank}} {
		for _, token := range []string{
			"DevReload *devreload.Hub",
			"deps.DevReload.Register(r, cfg.Env)",
		} {
			if !strings.Contains(blob.body, token) {
				t.Errorf("%s scaffold internal/app/app.go missing %q (#340)", blob.name, token)
			}
		}
	}
}

func TestScaffoldLayout_subscribesToDevReload(t *testing.T) {
	for name, layout := range map[string]string{"full": tplLayout, "minimal": tplLayoutMinimal} {
		for _, token := range []string{
			`new EventSource("/_dev/reload")`,
			`es.addEventListener("reload"`,
			`es.addEventListener("boot"`,
			`/^(localhost|127\.0\.0\.1|\[::1\])$/.test(location.hostname)`,
		} {
			if !strings.Contains(layout, token) {
				t.Errorf("%s layout missing dev reload token %q (#340)", name, token)
			}
		}
	}
}

// air must not rebuild on HTML/CSS anymore: that restart was the slow path (#340).
func TestScaffoldAir_watchesGoOnly(t *testing.T) {
	if !strings.Contains(tplAir, `include_ext = ["go"]`) {
		t.Error("air should watch .go only")
	}
	if !strings.Contains(tplAir, `"web"`) {
		t.Error("air should exclude the web/ tree")
	}
}
