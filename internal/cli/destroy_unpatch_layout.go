// Destroy unpatch for layout nav and leftover seed wiring (#288).
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func unpatchStoreTestForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store_test.go")
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read generated file: %w", err)
	}
	content, err := removeDeclsByName(string(body), map[string]bool{"TestStore_Insert" + data.Pascal: true})
	if err != nil {
		return err
	}
	return updateScaffoldFile(path, []byte(content), "internal/store/store_test.go", dryRun)
}

func unpatchSeedsForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/db/seeds.go")
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read generated file: %w", err)
	}
	block := fmt.Sprintf("\tif err := s.SeedDemo%s(); err != nil {\n\t\treturn err\n\t}\n", data.PluralPascal)
	content := strings.Replace(string(body), block, "", 1)
	return updateScaffoldFile(path, []byte(content), "internal/db/seeds.go", dryRun)
}

func unpatchMainForSeed(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "cmd/server/main.go")
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read generated file: %w", err)
	}
	block := fmt.Sprintf(`
	if err := s.SeedDemo%s(); err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("seed: %%w", err)
	}
`, data.PluralPascal)
	if !strings.Contains(string(body), block) {
		return nil
	}
	content := strings.Replace(string(body), block, "", 1)
	return updateScaffoldFile(path, []byte(content), "cmd/server/main.go", dryRun)
}

func unpatchLayoutNavForResource(dir string, data scaffoldData, dryRun bool) error {
	path, inertia := layoutNavFile(dir)
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read layout: %w", err)
	}
	link := adminNavLink(data)
	if data.Public {
		link += publicNavLink(data, inertia)
	}
	content := strings.Replace(string(body), link, "", 1)
	rel := strings.TrimPrefix(path, dir+string(os.PathSeparator))
	return updateScaffoldFile(path, []byte(content), rel, dryRun)
}
