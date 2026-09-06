package manifest

import (
	"encoding/json"
	"github.com/jillesme/tourminal/internal/tour"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jillesme/tourminal/internal/workspace"
)

func TestBuildResolvesStepsAndLinks(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	tours := filepath.Join(root, ".tours")
	if err := os.MkdirAll(tours, 0o755); err != nil {
		t.Fatal(err)
	}
	firstPath := filepath.Join(tours, "first.tour")
	secondPath := filepath.Join(tours, "second.tour")
	first := `{"title":"First","nextTour":"Second","steps":[{"description":"Code","file":"main.go","pattern":"^func main","commands":["unsafe"]},{"description":"Embedded","file":"sample.go","contents":"one\ntwo"}]}`
	second := `{"title":"Second","steps":[{"description":"Done"}]}`
	if err := os.WriteFile(firstPath, []byte(first), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondPath, []byte(second), 0o600); err != nil {
		t.Fatal(err)
	}

	refs, diagnostics := workspace.Discover(root)
	result := Build(root, refs, diagnostics)
	if result.APIVersion != 1 || len(result.Diagnostics) != 0 || len(result.Tours) != 2 {
		t.Fatalf("unexpected manifest: %#v", result)
	}
	var firstEntry TourEntry
	for _, entry := range result.Tours {
		if entry.Title == "First" {
			firstEntry = entry
		}
	}
	if firstEntry.NextTourPath != secondPath {
		t.Fatalf("next path = %q, want %q", firstEntry.NextTourPath, secondPath)
	}
	if got := firstEntry.Steps[0]; got.Resolved.Kind != "file" || got.Resolved.TargetLine != 3 || len(got.Commands) != 1 {
		t.Fatalf("unexpected file step: %#v", got)
	}
	if got := firstEntry.Steps[1]; got.Resolved.Kind != "embedded" || got.Resolved.Source != "one\ntwo" {
		t.Fatalf("unexpected embedded step: %#v", got)
	}
}

func TestBuildRejectsFilesAndResolvesExactMarkers(t *testing.T) {
	root := t.TempDir()
	for name, contents := range map[string]string{"binary.dat": "hello\x00world", "large.txt": strings.Repeat("x", (2<<20)+1), "main.go": "// CT1.10\n// CT1.1\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name, file, pattern string
		line                int
		rejected            bool
	}{
		{"binary", "binary.dat", "", 0, true},
		{"oversize", "large.txt", "", 0, true},
		{"missing", "missing.go", "", 0, true},
		{"invalid line", "main.go", "", 999, true},
		{"missing pattern", "main.go", "MISSING", 0, true},
		{"ambiguous pattern", "main.go", "CT1", 0, true},
		{"exact marker", "main.go", "", 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tour.Tour{Title: "1 - Intro", Steps: []tour.Step{{Description: "x", File: tc.file, Pattern: tc.pattern, Line: tc.line}}})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, "main.tour")
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			result := Build(root, []workspace.TourRef{{Path: path}}, nil)
			step := result.Tours[0].Steps[0]
			if tc.rejected {
				if step.Error == "" || step.Resolved.Kind != "content" || step.Resolved.Path != "" {
					t.Fatalf("unsafe resolution: %#v", step)
				}
			} else if step.Error != "" || step.Resolved.TargetLine != 2 {
				t.Fatalf("incorrect marker: %#v", step)
			}
		})
	}
}
