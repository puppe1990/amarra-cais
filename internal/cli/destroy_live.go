package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func destroyLive(dir, name string, dryRun, force bool) error {
	data := dataForHandler(name)
	files := []string{
		filepath.Join("internal/handlers", data.Snake+"_live.go"),
		filepath.Join("internal/handlers", data.Snake+"_live_test.go"),
		filepath.Join("web/templates/pages", data.Snake+".html"),
	}
	if err := removeGeneratedFiles(dir, files, dryRun, force); err != nil {
		return err
	}
	return unpatchLive(dir, data, dryRun)
}

func unpatchLive(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	route := fmt.Sprintf(
		`r.Get("/live/%s", handlers.New%sPage(deps.Views, deps.Site, deps.Catalog, cfg).ServeHTTP)`,
		data.Snake, data.Pascal,
	)
	content = strings.Replace(content, "\n\t"+route, "", 1)
	content = strings.Replace(content, liveRegisterLine(data)+"\n\t", "", 1)
	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}
