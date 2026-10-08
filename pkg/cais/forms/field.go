package forms

import (
	"fmt"
	"html/template"
	"strings"
)

// FieldData describes a single form field for templates.
type FieldData struct {
	Name     string
	Label    string
	Value    string
	Type     string
	Required bool
	Error    string
}

// MakeField builds FieldData with validation error from a map. value is `any`
// so an optional model field (e.g. a nil *string on the New page) does not
// panic the template: it coerces to "" instead of dereferencing nil.
func MakeField(name, label string, value any, htmlType string, required bool, errors map[string]string) FieldData {
	if htmlType == "" {
		htmlType = "text"
	}
	return FieldData{
		Name:     name,
		Label:    label,
		Value:    stringValue(value),
		Type:     htmlType,
		Required: required,
		Error:    FieldError(errors, name),
	}
}

func stringValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case *string:
		if v == nil {
			return ""
		}
		return *v
	default:
		return fmt.Sprint(v)
	}
}

// FieldInput renders a labeled input, textarea, or checkbox with optional error text.
// Type "password" always includes the show/hide eye toggle (same as fieldPassword).
func FieldInput(f FieldData) template.HTML {
	if f.Type == "password" {
		return FieldPassword(f)
	}
	var b strings.Builder
	switch f.Type {
	case "textarea":
		b.WriteString(`<div><label class="block text-sm font-medium text-foam/70 mb-1" for="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`">`)
		b.WriteString(template.HTMLEscapeString(f.Label))
		b.WriteString(`</label><textarea class="w-full border border-copper/40 bg-ink text-foam rounded-lg px-3 py-2 min-h-[80px] focus:ring-2 focus:ring-copper outline-none" id="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`" name="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`">`)
		b.WriteString(template.HTMLEscapeString(f.Value))
		b.WriteString(`</textarea>`)
	case "float":
		b.WriteString(`<div><label class="block text-sm font-medium text-foam/70 mb-1" for="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`">`)
		b.WriteString(template.HTMLEscapeString(f.Label))
		b.WriteString(`</label><input class="w-full border border-copper/40 bg-ink text-foam rounded-lg px-3 py-2 focus:ring-2 focus:ring-copper outline-none" type="number" step="any" id="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`" name="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`" value="`)
		b.WriteString(template.HTMLEscapeString(f.Value))
		b.WriteString(`"`)
		if f.Required {
			b.WriteString(` required`)
		}
		b.WriteString(` />`)
	case "checkbox":
		b.WriteString(`<label class="flex items-center gap-2 text-sm text-foam/80"><input type="checkbox" name="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`" class="rounded border-copper/40 bg-ink text-copper"`)
		if f.Value == "true" || f.Value == "1" || strings.EqualFold(f.Value, "on") {
			b.WriteString(` checked`)
		}
		b.WriteString(` />`)
		b.WriteString(template.HTMLEscapeString(f.Label))
		b.WriteString(`</label>`)
	default:
		b.WriteString(`<div><label class="block text-sm font-medium text-foam/70 mb-1" for="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`">`)
		b.WriteString(template.HTMLEscapeString(f.Label))
		b.WriteString(`</label><input class="w-full border border-copper/40 bg-ink text-foam rounded-none px-3 py-2 focus:ring-2 focus:ring-copper outline-none" type="`)
		b.WriteString(template.HTMLEscapeString(f.Type))
		b.WriteString(`" id="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`" name="`)
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(`" value="`)
		b.WriteString(template.HTMLEscapeString(f.Value))
		b.WriteString(`"`)
		if f.Required {
			b.WriteString(` required`)
		}
		b.WriteString(` />`)
	}
	if f.Error != "" && f.Type != "checkbox" {
		b.WriteString(`<p class="text-red-600 text-sm mt-1">`)
		b.WriteString(template.HTMLEscapeString(f.Error))
		b.WriteString(`</p>`)
	}
	if f.Type != "checkbox" {
		b.WriteString(`</div>`)
	}
	return template.HTML(b.String())
}
