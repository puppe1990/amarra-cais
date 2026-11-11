package cli

import "testing"

func TestToPlural(t *testing.T) {
	tests := map[string]string{
		"":         "",
		"y":        "ys",
		"dish":     "dishes",
		"recipe":   "recipes",
		"category": "categories",
		"city":     "cities",
		"boy":      "boys",
		"day":      "days",
		"key":      "keys",
		"toy":      "toys",
		"play":     "plays",
		"task":     "tasks",
		"class":    "classes",
		"box":      "boxes",
		"bookmark": "bookmarks",
		"recipes":  "recipes",
		"status":   "statuses",
	}
	for in, want := range tests {
		if got := toPlural(in); got != want {
			t.Errorf("toPlural(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDataForResource_PluralPascal(t *testing.T) {
	data := dataForResource("dish")
	if data.Plural != "dishes" {
		t.Errorf("Plural = %q, want dishes", data.Plural)
	}
	if data.PluralPascal != "Dishes" {
		t.Errorf("PluralPascal = %q, want Dishes", data.PluralPascal)
	}
}

func TestToSnake_pascalCase(t *testing.T) {
	tests := map[string]string{
		"Category":      "category",
		"BlogPost":      "blog_post",
		"category":      "category",
		"blog_post":     "blog_post",
		"user-settings": "user_settings",
		"ID":            "id",
	}
	for in, want := range tests {
		if got := toSnake(in); got != want {
			t.Errorf("toSnake(%q) = %q, want %q (#39)", in, got, want)
		}
	}
}

func TestNames(t *testing.T) {
	data := dataForHandler("user_settings")
	if data.Pascal != "UserSettings" {
		t.Errorf("Pascal = %q", data.Pascal)
	}
	if data.Snake != "user_settings" {
		t.Errorf("Snake = %q", data.Snake)
	}
	if data.Title != "User Settings" {
		t.Errorf("Title = %q", data.Title)
	}
}

func TestLowerFirst(t *testing.T) {
	tests := map[string]string{
		"":             "",
		"A":            "a",
		"UserSettings": "userSettings",
		"ID":           "iD",
	}
	for in, want := range tests {
		if got := lowerFirst(in); got != want {
			t.Errorf("lowerFirst(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestToPascalAndToTitle(t *testing.T) {
	tests := []struct{ in, pascal, title string }{
		{"user_settings", "UserSettings", "User Settings"},
		{"blog-post", "BlogPost", "Blog Post"},
		{"Category", "Category", "Category"},
	}
	for _, tt := range tests {
		if got := toPascal(tt.in); got != tt.pascal {
			t.Errorf("toPascal(%q) = %q, want %q", tt.in, got, tt.pascal)
		}
		if got := toTitle(tt.in); got != tt.title {
			t.Errorf("toTitle(%q) = %q, want %q", tt.in, got, tt.title)
		}
	}
}
