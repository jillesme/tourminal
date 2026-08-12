package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitRefs(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	root := t.TempDir()
	runGit(t, root, "init", "--quiet")
	runGit(t, root, "config", "user.name", "Tourminal Test")
	runGit(t, root, "config", "user.email", "tourminal@example.com")
	path := filepath.Join(root, "main.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "main.go")
	runGit(t, root, "commit", "--quiet", "-m", "first")
	first, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	for _, ref := range []string{"", "HEAD", first} {
		if err := ValidateGitRef(root, ref); err != nil {
			t.Fatalf("ValidateGitRef(%q): %v", ref, err)
		}
	}
	if err := ValidateGitRef(root, "does-not-exist"); err == nil {
		t.Fatal("expected an unknown ref to fail")
	}
	if warning := GitRefWarning(root, first); warning != "" {
		t.Fatalf("warning at matching ref = %q", warning)
	}

	if err := os.WriteFile(path, []byte("package main\n\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "main.go")
	runGit(t, root, "commit", "--quiet", "-m", "second")
	warning := GitRefWarning(root, first)
	if !strings.Contains(warning, first) || !strings.Contains(warning, "Source lines may have drifted") {
		t.Fatalf("unexpected drift warning: %q", warning)
	}
	if got := shortHash(first); got != first[:8] {
		t.Fatalf("shortHash = %q", got)
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	commandArgs := append([]string{"-C", root}, args...)
	if output, err := exec.Command("git", commandArgs...).CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, output)
	}
}
