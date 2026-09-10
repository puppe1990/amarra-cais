package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParentModelFilename_matchesGeneratedModelPath(t *testing.T) {
	fields, err := parseFields("category_id:references")
	if err != nil {
		t.Fatal(err)
	}
	got := parentModelFilename(fields[0])
	if got != "category.go" {
		t.Errorf("parentModelFilename = %q, want category.go (toSnake(RefPascal)=%q; Linux CI is case-sensitive)", got, toSnake(fields[0].RefPascal)+".go")
	}
}

func TestParentDisplayField_readsTitleFromSnakeCaseFile(t *testing.T) {
	dir := t.TempDir()
	modelDir := filepath.Join(dir, "internal/models")
	if err := os.MkdirAll(modelDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := "package models\n\ntype Category struct {\n\tID int64\n\tTitle string\n}\n"
	if err := os.WriteFile(filepath.Join(modelDir, "category.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	f := FieldDef{Name: "category_id", RefPascal: "Category"}
	if got := parentDisplayField(dir, f); got != "Title" {
		t.Errorf("parentDisplayField = %q, want Title", got)
	}
}
