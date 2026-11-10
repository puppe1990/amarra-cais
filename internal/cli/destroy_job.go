package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func destroyJob(dir, name string, dryRun, force bool) error {
	data := dataForHandler(name)
	files := []string{
		filepath.Join("internal/jobs", data.Snake+".go"),
		filepath.Join("internal/jobs", data.Snake+"_test.go"),
	}
	if err := removeGeneratedFiles(dir, files, dryRun, force); err != nil {
		return err
	}
	if err := unpatchJobRegistry(dir, data, dryRun); err != nil {
		return err
	}
	return unpatchJobRecurringSeed(dir, data, dryRun)
}

func unpatchJobRegistry(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/jobs/registry.go")
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	line := fmt.Sprintf("\treg.Register(%q, Perform%s(s))\n", data.Pascal, data.Pascal)
	content := strings.Replace(string(body), line, "", 1)
	return updateScaffoldFile(path, []byte(content), "internal/jobs/registry.go", dryRun)
}

func unpatchJobRecurringSeed(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/db/seeds.go")
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	re := regexp.MustCompile(
		`\n\tif err := jobStore\.UpsertRecurring\(context\.Background\(\), caisjobs\.RecurringOptions\{\n\t\tKind: "` +
			regexp.QuoteMeta(data.Pascal) +
			`",\n\t\tCron: "[^"]*",\n\t\}\); err != nil \{\n\t\treturn err\n\t\}\n`,
	)
	content := re.ReplaceAllString(string(body), "\n")
	return updateScaffoldFile(path, []byte(content), "internal/db/seeds.go", dryRun)
}
