package resolver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jillesme/tourminal/internal/tour"
)

func TestResolveFileAnchors(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	if err := os.WriteFile(path, []byte("package main\n\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		step tour.Step
		line int
	}{
		{"line", tour.Step{File: "main.go", Line: 2}, 2},
		{"pattern", tour.Step{File: "main.go", Pattern: `^func main`}, 3},
		{"selection", tour.Step{File: "main.go", Selection: &tour.Selection{Start: tour.Position{Line: 1, Character: 1}, End: tour.Position{Line: 3, Character: 4}}}, 3},
		{"eof", tour.Step{File: "main.go"}, 4},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolved, err := Resolve(root, test.step)
			if err != nil {
				t.Fatal(err)
			}
			if resolved.TargetLine != test.line {
				t.Fatalf("line = %d, want %d", resolved.TargetLine, test.line)
			}
		})
	}
}

func TestResolveRejectsUnsafeAndInvalidSources(t *testing.T) {
	root := t.TempDir()
	_, err := Resolve(root, tour.Step{File: "../secret", Line: 1})
	if err == nil || !strings.Contains(err.Error(), "escapes workspace") {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "short.go"), []byte("one\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = Resolve(root, tour.Step{File: "short.go", Line: 99})
	if err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("unexpected error: %v", err)
	}

	external := t.TempDir()
	if err := os.WriteFile(filepath.Join(external, "secret.go"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(root, "linked")); err == nil {
		_, err = Resolve(root, tour.Step{File: "linked/secret.go", Line: 1})
		if err == nil || !strings.Contains(err.Error(), "symlink escapes") {
			t.Fatalf("unexpected symlink error: %v", err)
		}
	}
}

func TestResolveDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	resolved, err := Resolve(root, tour.Step{Directory: "src"})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Kind != Directory || !strings.Contains(resolved.Source, "main.go") {
		t.Fatalf("unexpected resolution: %#v", resolved)
	}
}
