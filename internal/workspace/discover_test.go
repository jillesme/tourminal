package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverOfficialLocationsAndPrimaryFirst(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".tours", "nested", "second.tour"), `{"title":"B","steps":[{"description":"b"}]}`)
	writeFile(t, filepath.Join(root, ".vscode", "main.tour"), `{"title":"A","isPrimary":true,"steps":[{"description":"a"}]}`)
	writeFile(t, filepath.Join(root, ".github", "tours", "ignored.txt"), `{"title":"ignored","steps":[]}`)

	refs, diagnostics := Discover(root)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics: %v", diagnostics)
	}
	if len(refs) != 2 {
		t.Fatalf("got %d tours, want 2", len(refs))
	}
	if refs[0].Title != "A" || !refs[0].Primary || refs[1].Title != "B" {
		t.Fatalf("unexpected order: %#v", refs)
	}

	resolved, err := ResolveRoot(filepath.Join(root, ".tours", "nested"))
	if err != nil {
		t.Fatal(err)
	}
	if resolved != root {
		t.Fatalf("root = %s, want %s", resolved, root)
	}
}

func TestDiscoverAllIncludesNonMatchingTours(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".tours", "never-here.tour"), `{"title":"Other platform","when":"false","steps":[{"description":"x"}]}`)

	refs, diagnostics := Discover(root)
	if len(diagnostics) != 0 || len(refs) != 0 {
		t.Fatalf("Discover refs=%v diagnostics=%v", refs, diagnostics)
	}
	refs, diagnostics = DiscoverAll(root)
	if len(diagnostics) != 0 || len(refs) != 1 {
		t.Fatalf("DiscoverAll refs=%v diagnostics=%v", refs, diagnostics)
	}
}

func TestRootForTour(t *testing.T) {
	root := t.TempDir()
	paths := []string{
		filepath.Join(root, ".tours", "nested", "x.tour"),
		filepath.Join(root, ".github", "tours", "x.tour"),
		filepath.Join(root, ".vscode", "main.tour"),
		filepath.Join(root, "main.tour"),
	}
	for _, path := range paths {
		if got := RootForTour(path); got != root {
			t.Errorf("RootForTour(%s) = %s, want %s", path, got, root)
		}
	}
}

func TestEvaluateWhen(t *testing.T) {
	value, err := EvaluateWhen("(isMac || isLinux || isWindows) && !isWeb")
	if err != nil || !value {
		t.Fatalf("value=%v error=%v", value, err)
	}
	if _, err := EvaluateWhen("process.exit()"); err == nil {
		t.Fatal("expected arbitrary JavaScript to be rejected")
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
