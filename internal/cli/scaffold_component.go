package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

const tplComponent = `<div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
  {{"{{"}} .Inner {{"}}"}}
</div>
`

func scaffoldComponent(dir, name string, dryRun bool) error {
	data := dataForHandler(name)
	rel := filepath.Join("web/templates/components", data.Snake+".html")
	path := filepath.Join(dir, rel)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", rel)
	}
	return writeScaffoldTemplate(path, tplComponent, data, rel, dryRun)
}

func destroyComponent(dir, name string, dryRun, force bool) error {
	data := dataForHandler(name)
	files := []string{
		filepath.Join("web/templates/components", data.Snake+".html"),
	}
	return removeGeneratedFiles(dir, files, dryRun, force)
}
