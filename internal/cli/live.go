package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const liveViewsMarker = "// cais:live-views"

func scaffoldLive(dir, name string, dryRun bool) error {
	data := dataForHandler(name)
	data.ModulePath = moduleFromDir(dir)
	files := map[string]string{
		filepath.Join("internal/handlers", data.Snake+"_live.go"):      tplLiveView,
		filepath.Join("internal/handlers", data.Snake+"_live_test.go"): tplLiveTest,
		filepath.Join("web/templates/pages", data.Snake+".html"):       tplLivePage,
	}
	for path, tpl := range files {
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); err == nil {
			return fmt.Errorf("%s already exists", path)
		}
		if err := writeScaffoldTemplate(full, tpl, data, path, dryRun); err != nil {
			return err
		}
	}
	if err := patchRoutesForLive(dir, data, dryRun); err != nil {
		return err
	}
	return patchLiveViewsRegister(dir, data, dryRun)
}

func patchRoutesForLive(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	route := fmt.Sprintf(`r.Get("/live/%s", handlers.New%sPage(deps.Views, deps.Site, deps.Catalog, cfg).ServeHTTP)`, data.Snake, data.Pascal)
	if strings.Contains(content, route) {
		return nil
	}
	needle := "\n}"
	idx := strings.LastIndex(content, "func registerRoutes")
	if idx < 0 {
		return fmt.Errorf("routes.go: missing registerRoutes")
	}
	rest := content[idx:]
	closeAt := strings.Index(rest, needle)
	if closeAt < 0 {
		return fmt.Errorf("routes.go: could not close registerRoutes")
	}
	insertAt := idx + closeAt
	content = content[:insertAt] + "\n\t" + route + content[insertAt:]
	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}

func liveRegisterLine(data scaffoldData) string {
	if data.Snake == "chat" {
		return fmt.Sprintf(`hub.Register(%q, func() live.View { return handlers.New%sLive(deps.Store) })`, data.Snake, data.Pascal)
	}
	return fmt.Sprintf(`hub.Register(%q, func() live.View { return handlers.New%sLive() })`, data.Snake, data.Pascal)
}

func patchLiveViewsRegister(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	reg := liveRegisterLine(data)
	if strings.Contains(content, reg) {
		return nil
	}
	if !strings.Contains(content, liveViewsMarker) {
		return fmt.Errorf("routes.go: missing %s", liveViewsMarker)
	}
	content = strings.Replace(content, liveViewsMarker, reg+"\n\t"+liveViewsMarker, 1)
	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}
