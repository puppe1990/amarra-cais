package cli

import (
	"fmt"
	"regexp"
	"strings"
)

type FieldDef struct {
	Name      string
	Pascal    string
	SQLType   string
	GoType    string
	HTMLType  string
	Widget    string
	Required  bool
	RefTable  string
	RefPascal string
}

func parseFields(spec string) ([]FieldDef, error) {
	if strings.TrimSpace(spec) == "" {
		return defaultFields()
	}

	var fields []FieldDef
	seen := make(map[string]bool)
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		part = normalizeFieldPart(part)
		name, typ, req, err := parseFieldPart(part)
		if err != nil {
			return nil, err
		}
		f, err := fieldFromNameType(name, typ, req)
		if err != nil {
			return nil, err
		}
		if err := validateFieldName(f.Name); err != nil {
			return nil, err
		}
		if seen[f.Name] {
			return nil, fmt.Errorf("duplicate field %q: each field may appear once", f.Name)
		}
		seen[f.Name] = true
		fields = append(fields, f)
	}
	if len(fields) == 0 {
		return defaultFields()
	}
	return fields, nil
}

var fieldNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// reservedFieldNames are the columns the migration already creates plus the
// SQLite keywords that break unquoted DDL/DML; a field named "order" used to
// generate a syntax error after store.go was already patched (#128).
var reservedFieldNames = map[string]bool{
	"id": true, "created_at": true, "updated_at": true,
	"abort": true, "action": true, "add": true, "after": true, "all": true,
	"alter": true, "always": true, "analyze": true, "and": true, "as": true,
	"asc": true, "attach": true, "autoincrement": true, "before": true,
	"begin": true, "between": true, "by": true, "cascade": true, "case": true,
	"cast": true, "check": true, "collate": true, "column": true, "commit": true,
	"conflict": true, "constraint": true, "create": true, "cross": true,
	"current": true, "database": true, "default": true, "deferrable": true,
	"deferred": true, "delete": true, "desc": true, "detach": true,
	"distinct": true, "do": true, "drop": true, "each": true, "else": true,
	"end": true, "escape": true, "except": true, "exclusive": true,
	"exists": true, "explain": true, "fail": true, "filter": true,
	"first": true, "following": true, "for": true, "foreign": true,
	"from": true, "full": true, "generated": true, "glob": true, "group": true,
	"groups": true, "having": true, "if": true, "ignore": true,
	"immediate": true, "in": true, "index": true, "indexed": true,
	"initially": true, "inner": true, "insert": true, "instead": true,
	"intersect": true, "into": true, "is": true, "isnull": true, "join": true,
	"key": true, "last": true, "left": true, "like": true, "limit": true,
	"match": true, "materialized": true, "natural": true, "no": true,
	"not": true, "nothing": true, "notnull": true, "null": true, "nulls": true,
	"of": true, "offset": true, "on": true, "or": true, "order": true,
	"others": true, "outer": true, "over": true, "partition": true,
	"plan": true, "pragma": true, "preceding": true, "primary": true,
	"query": true, "raise": true, "range": true, "recursive": true,
	"references": true, "regexp": true, "reindex": true, "release": true,
	"rename": true, "replace": true, "restrict": true, "returning": true,
	"right": true, "rollback": true, "row": true, "rows": true,
	"savepoint": true, "select": true, "set": true, "table": true,
	"temp": true, "temporary": true, "then": true, "ties": true, "to": true,
	"transaction": true, "trigger": true, "unbounded": true, "union": true,
	"unique": true, "update": true, "using": true, "vacuum": true,
	"values": true, "view": true, "virtual": true, "when": true,
	"where": true, "window": true, "with": true, "without": true,
}

func validateFieldName(name string) error {
	if !fieldNamePattern.MatchString(name) {
		return fmt.Errorf("invalid field name %q: use lowercase letters, digits and underscores, starting with a letter", name)
	}
	if reservedFieldNames[name] {
		return fmt.Errorf("field name %q is reserved (SQLite keyword or generated column) — pick another name", name)
	}
	return nil
}

func parseFieldPart(part string) (name, typ string, required bool, err error) {
	required = true
	if idx := strings.Index(part, ":"); idx >= 0 {
		name = strings.TrimSpace(part[:idx])
		rest := strings.TrimSpace(part[idx+1:])
		if strings.HasSuffix(rest, "?") {
			required = false
			rest = strings.TrimSuffix(rest, "?")
		}
		typ = rest
	} else {
		name = part
		typ = "string"
	}
	if name == "" {
		return "", "", false, fmt.Errorf("invalid field %q", part)
	}
	return name, typ, required, nil
}

func fieldFromNameType(name, typ string, required bool) (FieldDef, error) {
	pascal := toPascal(name)
	if toSnake(name) == "url" {
		pascal = "URL"
	}
	switch typ {
	case "string":
		return stringField(name, pascal, required, "text", "input"), nil
	case "text":
		return stringField(name, pascal, required, "text", "textarea"), nil
	case "url":
		f := stringField(name, pascal, required, "url", "input")
		if required {
			f.SQLType = "TEXT NOT NULL"
		} else {
			f.SQLType = "TEXT"
		}
		return f, nil
	case "bool":
		return FieldDef{Name: toSnake(name), Pascal: pascal, SQLType: "INTEGER NOT NULL DEFAULT 0", GoType: "bool", HTMLType: "checkbox", Widget: "checkbox", Required: false}, nil
	case "int":
		return intField(name, pascal, required), nil
	case "float":
		return floatField(name, pascal, required), nil
	case "date":
		return stringField(name, pascal, required, "date", "input"), nil
	case "references":
		return refFieldFromName(toSnake(name), required)
	default:
		return FieldDef{}, fmt.Errorf("unknown field type %q (use string, text, url, bool, int, float, date, references)", typ)
	}
}

func normalizeFieldPart(part string) string {
	idx := strings.Index(part, ":")
	if idx < 0 {
		return part
	}
	name := strings.TrimSpace(part[:idx])
	typ := strings.TrimSpace(part[idx+1:])
	switch typ {
	case "belongs_to":
		return toSnake(name) + "_id:references"
	case "belongs_to?":
		return toSnake(name) + "_id:references?"
	default:
		return part
	}
}

func refFieldFromName(name string, required bool) (FieldDef, error) {
	if !strings.HasSuffix(name, "_id") {
		return FieldDef{}, fmt.Errorf("references field %q must end with _id (or use name:belongs_to)", name)
	}
	singular := strings.TrimSuffix(name, "_id")
	if singular == "" {
		return FieldDef{}, fmt.Errorf("invalid references field %q", name)
	}
	refTable := toPlural(singular)
	refPascal := toPascal(singular)
	pascal := toPascal(name)

	f := FieldDef{
		Name:      name,
		Pascal:    pascal,
		HTMLType:  "select",
		Widget:    "select",
		Required:  required,
		RefTable:  refTable,
		RefPascal: refPascal,
	}
	if required {
		f.SQLType = fmt.Sprintf("INTEGER NOT NULL REFERENCES %s(id)", refTable)
		f.GoType = "int64"
	} else {
		f.SQLType = fmt.Sprintf("INTEGER REFERENCES %s(id)", refTable)
		f.GoType = "*int64"
	}
	return f, nil
}

func hasReferenceFields(fields []FieldDef) bool {
	for _, f := range fields {
		if f.RefTable != "" {
			return true
		}
	}
	return false
}

func uniqueReferenceFields(fields []FieldDef) []FieldDef {
	seen := make(map[string]bool)
	var out []FieldDef
	for _, f := range fields {
		if f.RefTable == "" || seen[f.RefTable] {
			continue
		}
		seen[f.RefTable] = true
		out = append(out, f)
	}
	return out
}

func stringField(name, pascal string, required bool, htmlType, widget string) FieldDef {
	f := FieldDef{
		Name:     toSnake(name),
		Pascal:   pascal,
		HTMLType: htmlType,
		Widget:   widget,
		Required: required,
	}
	if required {
		f.SQLType = "TEXT NOT NULL"
		if widget == "textarea" {
			f.SQLType = "TEXT NOT NULL DEFAULT ''"
		}
		f.GoType = "string"
	} else {
		f.SQLType = "TEXT"
		f.GoType = "*string"
	}
	return f
}

func floatField(name, pascal string, required bool) FieldDef {
	f := FieldDef{
		Name:     toSnake(name),
		Pascal:   pascal,
		HTMLType: "float",
		Widget:   "input",
		Required: required,
	}
	if required {
		f.SQLType = "REAL NOT NULL DEFAULT 0"
		f.GoType = "float64"
	} else {
		f.SQLType = "REAL"
		f.GoType = "*float64"
	}
	return f
}

func intField(name, pascal string, required bool) FieldDef {
	f := FieldDef{
		Name:     toSnake(name),
		Pascal:   pascal,
		HTMLType: "number",
		Widget:   "input",
		Required: required,
	}
	if required {
		f.SQLType = "INTEGER NOT NULL DEFAULT 0"
		f.GoType = "int64"
	} else {
		f.SQLType = "INTEGER"
		f.GoType = "*int64"
	}
	return f
}

func defaultFields() ([]FieldDef, error) {
	f, err := fieldFromNameType("name", "string", true)
	if err != nil {
		return nil, err
	}
	return []FieldDef{f}, nil
}

func fieldNeedsStrPtr(fields []FieldDef) bool {
	for _, f := range fields {
		if f.GoType == "*string" {
			return true
		}
	}
	return false
}

func fieldNeedsInt64Ptr(fields []FieldDef) bool {
	for _, f := range fields {
		if f.GoType == "*int64" {
			return true
		}
	}
	return false
}

func fieldNeedsFloat64Ptr(fields []FieldDef) bool {
	for _, f := range fields {
		if f.GoType == "*float64" {
			return true
		}
	}
	return false
}
