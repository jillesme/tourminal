package manifest

import (
	"os"
	"path/filepath"
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
