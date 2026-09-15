package view

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func expandAndRun(t *testing.T, src string, components map[string]string, data any) string {
	t.Helper()
	got, err := ExpandAll(src, components)
	if err != nil {
		t.Fatalf("expand: %v", err)
	}
	tmpl, err := template.New("t").Parse(got)
	if err != nil {
		t.Fatalf("parse: %v\n%s", err, got)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("execute: %v\n%s", err, got)
	}
	return buf.String()
}

// A kit attribute that mixes literal text with an action must interpolate like
// plain markup instead of printing the action to the user (#72).
func TestExpandAll_interpolatesActionInAttrValue(t *testing.T) {
	components := map[string]string{
		"stat": `<div class="stat"><span>{{ .Label }}</span><b>{{ .Value }}</b></div>`,
	}
	out := expandAndRun(t, `<.stat label="Potência instalada" value="{{ .Power }} kWp" />`, components, map[string]any{"Power": 5})

	if strings.Contains(out, "{{") {
		t.Fatalf("action rendered literally: %q", out)
	}
	if !strings.Contains(out, "<b>5 kWp</b>") {
		t.Errorf("got %q", out)
	}
}

func TestExpandAll_interpolatesMultipleActionsInAttrValue(t *testing.T) {
	components := map[string]string{"kpi": `<i>{{ .Hint }}</i>`}
	out := expandAndRun(t, `<.kpi hint="{{ .Done }}/{{ .Total }} tasks" />`, components, map[string]any{"Done": 3, "Total": 7})

	if !strings.Contains(out, "<i>3/7 tasks</i>") {
		t.Errorf("got %q", out)
	}
}

// The bare-action value keeps the raw expression, so non-string data (ints,
// bools) keeps working in the component's if/range (#72).
func TestExpandAll_bareActionAttrStaysExpression(t *testing.T) {
	components := map[string]string{"kpi": `{{ if .Value }}<b>{{ .Value }}</b>{{ end }}`}
	out := expandAndRun(t, `<.kpi value="{{ .Count }}" />`, components, map[string]any{"Count": 3})

	if !strings.Contains(out, "<b>3</b>") {
		t.Errorf("got %q", out)
	}
}

// An attribute value may reference a sibling attribute, like the body does.
func TestExpandAll_attrValueCanReferenceSiblingAttr(t *testing.T) {
	components := map[string]string{"badge": `<b>{{ .Label }}</b><i>{{ .Hint }}</i>`}
	out := expandAndRun(t, `<.badge label="{{ .Title }}!" hint="{{ .Label }} (hint)" />`, components, map[string]any{"Title": "Oi"})

	if !strings.Contains(out, "<b>Oi!</b>") || !strings.Contains(out, "<i>Oi! (hint)</i>") {
		t.Errorf("got %q", out)
	}
}

// A value the expander cannot interpolate must fail at boot instead of
// rendering `{{ ... }}` to the user (#72).
func TestExpandAll_controlActionInAttrFailsLoudly(t *testing.T) {
	components := map[string]string{"kpi": `<b>{{ .Value }}</b>`}
	_, err := ExpandAll(`<.kpi value="{{ if .X }}yes{{ end }}" />`, components)
	if err == nil {
		t.Fatal("expected an error for a control action inside an attribute value")
	}
	if !strings.Contains(err.Error(), "value") {
		t.Errorf("error should name the attribute: %v", err)
	}
}
