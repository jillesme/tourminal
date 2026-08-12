package tour

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAndLabels(t *testing.T) {
	path := writeTour(t, `{
  "title": "Example",
  "steps": [
    {"description": "### Introduction\n\nHello"},
    {"description": "No heading", "file": "main.go"}
  ]
}`)
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Title != "Example" || len(loaded.Steps) != 2 {
		t.Fatalf("unexpected tour: %#v", loaded)
	}
	if got := loaded.Steps[0].Label(1); got != "Introduction" {
		t.Fatalf("heading label = %q", got)
	}
	if got := loaded.Steps[1].Label(2); got != "main.go" {
		t.Fatalf("file label = %q", got)
	}
}

func TestLoadRejectsInvalidTour(t *testing.T) {
	tests := []struct {
		name, contents, want string
	}{
		{"missing title", `{"steps":[]}`, "title is required"},
		{"missing description", `{"title":"x","steps":[{}]}`, "description is required"},
		{"invalid selection", `{"title":"x","steps":[{"description":"x","selection":{"start":{"line":2,"character":1},"end":{"line":1,"character":1}}}]}`, "selection start"},
		{"trailing value", `{"title":"x","steps":[]} {}`, "more than one JSON value"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Load(writeTour(t, test.contents))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func writeTour(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.tour")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
