package cli

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
)

// #295: kit stems and hook names have one canonical source each. Website
// tables must match; a miss names the file, the extra/missing item, and the
// fix. Update the table when you add or delete a component/hook — do not
// snapshot prose.

func TestMarkdownTableBackticks_readsFirstColumnUntilNextHeading(t *testing.T) {
	md := "## Shipped kit\n\n| Component |\n| --- |\n| `form` |\n| `input` |\n\n### Nested\n\n| Skip |\n| --- |\n| `nope` |\n"
	got := markdownTableBackticks(md, "## Shipped kit")
	want := []string{"form", "input"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReportContractDiff_namesFileItemAndFix(t *testing.T) {
	err := contractDiff("views-and-kit.md", "kit stem", []string{"form", "tooltip"}, []string{"form", "ghost"})
	if err == nil {
		t.Fatal("expected drift")
	}
	msg := err.Error()
	for _, needle := range []string{
		"views-and-kit.md",
		`missing kit stem "tooltip" — add a table row or remove it from the canonical source`,
		`extra kit stem "ghost" — remove the table row or add it to the canonical source`,
	} {
		if !strings.Contains(msg, needle) {
			t.Errorf("diff missing %q\n%s", needle, msg)
		}
	}
}

func TestDocs_viewsAndKitTablesMatchShippedKit(t *testing.T) {
	canonical := mapKeys(view.ShippedComponents())
	for _, spec := range []struct {
		path    string
		heading string
	}{
		{"../../website/src/content/docs/docs/reference/views-and-kit.md", "## Shipped kit"},
		{"../../website/src/content/docs/pt-br/docs/reference/views-and-kit.md", "## Kit incluso"},
	} {
		body, err := os.ReadFile(spec.path)
		if err != nil {
			t.Fatal(err)
		}
		documented := markdownTableBackticks(string(body), spec.heading)
		if documented == nil {
			t.Errorf("%s: missing heading %q", spec.path, spec.heading)
			continue
		}
		if diff := contractDiff(spec.path, "kit stem", canonical, documented); diff != nil {
			t.Error(diff)
		}
	}
}

func TestDocs_amarraJsHookTablesMatchRegister(t *testing.T) {
	src, err := os.ReadFile("../../pkg/amarra/js/hook.mjs")
	if err != nil {
		t.Fatal(err)
	}
	canonical := registeredHookNames(string(src))
	if len(canonical) == 0 {
		t.Fatal("hook.mjs: no register(\"name\", ...) calls")
	}
	for _, spec := range []struct {
		path    string
		heading string
	}{
		{"../../website/src/content/docs/docs/reference/amarra-js.md", "## Built-in hooks"},
		{"../../website/src/content/docs/pt-br/docs/reference/amarra-js.md", "## Hooks embutidos"},
	} {
		body, err := os.ReadFile(spec.path)
		if err != nil {
			t.Fatal(err)
		}
		documented := markdownTableBackticks(string(body), spec.heading)
		if documented == nil {
			t.Errorf("%s: missing heading %q", spec.path, spec.heading)
			continue
		}
		if diff := contractDiff(spec.path, "hook", canonical, documented); diff != nil {
			t.Error(diff)
		}
	}
}

func markdownTableBackticks(md, heading string) []string {
	idx := strings.Index(md, heading)
	if idx < 0 {
		return nil
	}
	rest := md[idx+len(heading):]
	if next := strings.Index(rest, "\n#"); next >= 0 {
		rest = rest[:next]
	}
	var names []string
	for _, line := range strings.Split(rest, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cells := strings.Split(line, "|")
		if len(cells) < 2 {
			continue
		}
		first := strings.TrimSpace(cells[1])
		if first == "" || strings.HasPrefix(first, "---") || !strings.HasPrefix(first, "`") {
			continue
		}
		name := strings.Trim(first, "`")
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

func contractDiff(file, kind string, canonical, documented []string) error {
	can := stringSet(canonical)
	doc := stringSet(documented)
	var b strings.Builder
	for _, name := range canonical {
		if !doc[name] {
			fmt.Fprintf(&b, "%s: missing %s %q — add a table row or remove it from the canonical source\n", file, kind, name)
		}
	}
	for _, name := range documented {
		if !can[name] {
			fmt.Fprintf(&b, "%s: extra %s %q — remove the table row or add it to the canonical source\n", file, kind, name)
		}
	}
	if b.Len() == 0 {
		return nil
	}
	return errors.New(strings.TrimSuffix(b.String(), "\n"))
}

func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func registeredHookNames(src string) []string {
	re := regexp.MustCompile(`(?m)^register\("([^"]+)",`)
	seen := map[string]bool{}
	var names []string
	for _, m := range re.FindAllStringSubmatch(src, -1) {
		if seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		names = append(names, m[1])
	}
	sort.Strings(names)
	return names
}

func stringSet(items []string) map[string]bool {
	out := make(map[string]bool, len(items))
	for _, s := range items {
		out[s] = true
	}
	return out
}
