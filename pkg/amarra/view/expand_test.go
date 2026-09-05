package view

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func TestExpandAll_staticButton(t *testing.T) {
	components := map[string]string{
		"button": `<button type="{{ .Type }}">{{ .Inner }}</button>`,
	}
	got, err := ExpandAll(`<.button type="submit">Save</.button>`, components)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`{{ $type := "submit" }}`,
		`<button type="{{ $type }}">`,
		`Save`,
		`</button>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
	if strings.Contains(got, "<.button") {
		t.Errorf("left component tag in %q", got)
	}
}

func TestExpandAll_selfClosing(t *testing.T) {
	components := map[string]string{
		"flash": `<div class="flash">{{ .Inner }}</div>`,
	}
	got, err := ExpandAll(`<.flash />`, components)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "<.flash") {
		t.Fatalf("not expanded: %q", got)
	}
	if !strings.Contains(got, `<div class="flash">`) {
		t.Errorf("got %q", got)
	}
}

func TestExpandAll_nested(t *testing.T) {
	components := map[string]string{
		"form":   `<form>{{ .Inner }}</form>`,
		"button": `<button>{{ .Inner }}</button>`,
	}
	got, err := ExpandAll(`<.form><.button>Go</.button></.form>`, components)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "<.form") || strings.Contains(got, "<.button") {
		t.Fatalf("not fully expanded: %q", got)
	}
	if !strings.Contains(got, `<form>`) || !strings.Contains(got, `<button>`) || !strings.Contains(got, `Go`) {
		t.Errorf("got %q", got)
	}
}

func TestExpandAll_unknownComponent(t *testing.T) {
	_, err := ExpandAll(`<.nope />`, map[string]string{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExpandAll_dynamicAttr(t *testing.T) {
	components := map[string]string{
		"input": `<input value="{{ .Value }}">`,
	}
	got, err := ExpandAll(`<.input value="{{ .Item.Title }}" />`, components)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `{{ $value := .Item.Title }}`) {
		t.Errorf("got %q", got)
	}
}

func TestExpandAll_nestedSameAttrDoesNotLeak(t *testing.T) {
	components := map[string]string{
		"card":   `<section>{{ .Inner }}<i>{{ .Title }}</i></section>`,
		"button": `<button>{{ .Title }}</button>`,
	}
	src := `<.card title="outer"><.button title="inner">x</.button></.card>`
	got, err := ExpandAll(src, components)
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := template.New("t").Parse(got)
	if err != nil {
		t.Fatalf("parse expanded: %v\n%s", err, got)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `<button>inner</button>`) {
		t.Errorf("child title missing: %q", out)
	}
	if !strings.Contains(out, `<i>outer</i>`) {
		t.Errorf("parent title leaked/overwritten: %q", out)
	}
}
