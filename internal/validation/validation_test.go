package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jillesme/tourminal/internal/tour"
)

func TestTourValidatesAnchors(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		step tour.Step
		want string
	}{
		{"missing file", tour.Step{Description: "x", File: "missing.go", Line: 1}, "no such file"},
		{"bad regex", tour.Step{Description: "x", File: "main.go", Pattern: "["}, "unsupported pattern"},
		{"no regex match", tour.Step{Description: "x", File: "main.go", Pattern: "missing"}, "does not match"},
		{"ambiguous regex", tour.Step{Description: "x", File: "main.go", Pattern: "a"}, "matches main.go 4 times"},
		{"line and regex", tour.Step{Description: "x", File: "main.go", Line: 1, Pattern: "package"}, "cannot both be set"},
		{"location conflict", tour.Step{Description: "x", File: "main.go", Directory: "."}, "use only one"},
		{"relative uri", tour.Step{Description: "x", URI: "docs/page"}, "uri must be absolute"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Check(root, &tour.Tour{Title: "x", Steps: []tour.Step{test.step}})
			if !contains(result.Errors, test.want) {
				t.Fatalf("errors=%v, want substring %q", result.Errors, test.want)
			}
		})
	}
}

func TestTourAcceptsUniquePatternAndWarnsAboutCommands(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	result := Check(root, &tour.Tour{Title: "x", Steps: []tour.Step{{
		Description: "x", File: "main.go", Pattern: "^func main", Commands: []string{"workbench.action.files.save"},
	}}})
	if len(result.Errors) != 0 || !contains(result.Warnings, "will not execute") {
		t.Fatalf("result=%#v", result)
	}
}

func TestMissingNextTours(t *testing.T) {
	tours := []*tour.Tour{{Title: "One", NextTour: "Missing"}, {Title: "Two"}}
	if got := MissingNextTours(tours); len(got) != 1 || !strings.Contains(got[0], "Missing") {
		t.Fatalf("got %v", got)
	}
}

func contains(values []string, substring string) bool {
	for _, value := range values {
		if strings.Contains(value, substring) {
			return true
		}
	}
	return false
}
