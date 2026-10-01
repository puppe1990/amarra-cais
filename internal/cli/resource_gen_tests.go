// Generated handler tests for amarra-cais g resource (#288).
package cli

import (
	"fmt"
	"strconv"
	"strings"
)

func buildResourceAdminTest(data scaffoldData) string {
	first := data.Fields[0]
	formBody := buildAdminTestFormBody(data.Fields)
	setup, refsLiteral, refVars := adminTestRefSetup(data)
	readerExpr := "strings.NewReader(fmt.Sprintf(" + strconv.Quote(formBody) + ", " + refVars + "))"
	return fmt.Sprintf(`package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"%s/pkg/cais"
	"%s/pkg/cais/testutil"
	"%s/internal/models"
)

func TestAdmin%sHandler_Show(t *testing.T) {
	s := setupTestStore(t)
%s	id, err := s.Insert%s(models.%s{%s: "show-me"%s%s})
	if err != nil {
		t.Fatal(err)
	}
	h := NewAdmin%sHandler(setupTestViews(t), s, testSite(), cais.Config{})
	rr := httptest.NewRecorder()
	h.Show(rr, testutil.NewRequest(http.MethodGet, "/admin/%s/1", testutil.PathValue("id", "1")), id)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "show-me") {
		t.Error("missing item detail in show page")
	}
}

func TestAdmin%sHandler_Index(t *testing.T) {
	s := setupTestStore(t)
	h := NewAdmin%sHandler(setupTestViews(t), s, testSite(), cais.Config{})
	rr := httptest.NewRecorder()
	h.Index(rr, httptest.NewRequest(http.MethodGet, "/admin/%s", nil))
	if rr.Code != http.StatusOK {
		t.Errorf("status = %%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "admin-%s") {
		t.Error("missing admin table")
	}
}

func TestAdmin%sHandler_Create(t *testing.T) {
	s := setupTestStore(t)
%s	h := NewAdmin%sHandler(setupTestViews(t), s, testSite(), cais.Config{})
	req := httptest.NewRequest(http.MethodPost, "/admin/%s", %s)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %%d", rr.Code)
	}
}

func TestAdmin%sHandler_Delete(t *testing.T) {
	s := setupTestStore(t)
%s	id, err := s.Insert%s(models.%s{%s: "x"%s%s})
	if err != nil {
		t.Fatal(err)
	}
	h := NewAdmin%sHandler(setupTestViews(t), s, testSite(), cais.Config{})
	rr := httptest.NewRecorder()
	h.Delete(rr, testutil.NewRequest(http.MethodDelete, "/admin/%s/1", testutil.PathValue("id", "1")), id)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %%d", rr.Code)
	}
}
`,
		frameworkModule, frameworkModule, data.ModulePath,
		data.PluralPascal, setup, data.Pascal, data.Pascal, first.Pascal, urlFieldTestExtra(data), refsLiteral,
		data.PluralPascal, data.Plural,
		data.PluralPascal, data.PluralPascal, data.Plural, data.Plural,
		data.PluralPascal, setup, data.PluralPascal, data.Plural, readerExpr,
		data.PluralPascal, setup, data.Pascal, data.Pascal, first.Pascal, urlFieldTestExtra(data), refsLiteral,
		data.PluralPascal, data.Plural,
	)
}

// adminTestRefSetup seeds one parent row per required reference field — FK
// constraints fail when the referenced table is empty.
func adminTestRefSetup(data scaffoldData) (setup, literal, vars string) {
	var varNames []string
	for _, f := range data.Fields {
		if f.RefTable == "" || !f.Required {
			continue
		}
		varName := strings.ToLower(f.RefPascal) + "ID"
		setup += fmt.Sprintf("%s, err := s.Insert%s(%s)\n\tif err != nil {\n\t\tt.Fatal(err)\n\t}\n\t", varName, f.RefPascal, parentSampleLiteral(data.AppDir, f))
		literal += fmt.Sprintf(", %s: %s", f.Pascal, varName)
		varNames = append(varNames, varName)
	}
	return setup, literal, strings.Join(varNames, ", ")
}

func buildAdminTestFormBody(fields []FieldDef) string {
	var parts []string
	for _, f := range fields {
		if !f.Required || f.GoType == "bool" {
			continue
		}
		if f.RefTable != "" {
			// The test seeds the parent and passes its id via Sprintf.
			parts = append(parts, f.Name+"=%d")
			continue
		}
		val := "Demo"
		switch f.GoType {
		case "int64":
			val = "30"
		case "float64":
			val = "25.49"
		default:
			if f.HTMLType == "url" {
				val = "https://example.com"
			}
			if f.Widget == "textarea" {
				val = "Sample " + f.Pascal
			}
		}
		parts = append(parts, f.Name+"="+val)
	}
	if len(parts) == 0 && len(fields) > 0 {
		return fields[0].Name + "=Demo"
	}
	return strings.Join(parts, "&")
}

func urlFieldTestExtra(data scaffoldData) string {
	for _, f := range data.Fields {
		if f.HTMLType == "url" {
			return fmt.Sprintf(", %s: \"https://example.com\"", f.Pascal)
		}
	}
	return ""
}

func buildResourcePublicTest(data scaffoldData) string {
	seedCall := ""
	if data.Seed {
		seedCall = `	if err := s.SeedDemo` + data.PluralPascal + `(); err != nil {
		t.Fatal(err)
	}
`
	}
	return fmt.Sprintf(`package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"%s/pkg/cais"
)

func Test%sHandler_List(t *testing.T) {
	s := setupTestStore(t)
%s
	h := New%sHandler(setupTestViews(t), s, testSite(), cais.Config{})
	rr := httptest.NewRecorder()
	h.List(rr, httptest.NewRequest(http.MethodGet, "/%s", nil))
	if rr.Code != http.StatusOK {
		t.Errorf("status = %%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "%s-list") {
		t.Error("missing public list")
	}
}
`, frameworkModule, data.PluralPascal, seedCall, data.PluralPascal, data.Plural, data.Plural)
}
