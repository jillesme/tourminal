package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestSkillCommandAndFlag(t *testing.T) {
	for _, args := range [][]string{{"skill"}, {"--skill"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := runWithIO(args, &stdout, &stderr); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(stdout.String(), "name: create-codetour") ||
				!strings.Contains(stdout.String(), "CodeTour Authoring Reference") {
				t.Fatalf("unexpected skill output: %q", stdout.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("unexpected stderr: %q", stderr.String())
			}
		})
	}
}

func TestSkillCommandRejectsArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runWithIO([]string{"skill", "extra"}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "usage: tourminal skill") {
		t.Fatalf("error = %v", err)
	}
}

func TestHelpAndVersionSucceed(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"validate", "--help"}, {"skill", "--help"}} {
		var stdout, stderr bytes.Buffer
		if err := runWithIO(args, &stdout, &stderr); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if !strings.Contains(stderr.String(), "Usage:") {
			t.Fatalf("%v: missing usage in %q", args, stderr.String())
		}
	}
	var stdout, stderr bytes.Buffer
	if err := runWithIO([]string{"version"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "tourminal dev") {
		t.Fatalf("version=%q", stdout.String())
	}
}

func TestThemeConfiguration(t *testing.T) {
	t.Setenv("TOURMINAL_THEME", "")
	if got := defaultTheme(); got != "auto" {
		t.Fatalf("default theme = %q, want auto", got)
	}

	t.Setenv("TOURMINAL_THEME", "light")
	if got := defaultTheme(); got != "light" {
		t.Fatalf("environment theme = %q, want light", got)
	}

	var stdout, stderr bytes.Buffer
	err := runWithIO([]string{"--theme", "sepia"}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "theme must be auto, light, or dark") {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateResolvesAnchors(t *testing.T) {
	root := t.TempDir()
	path := root + "/broken.tour"
	contents := `{"title":"Broken","steps":[{"description":"x","file":"missing.go","pattern":"["}]}`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	err := runWithIO([]string{"validate", path}, &stdout, &stderr)
	if err == nil || !strings.Contains(stderr.String(), "invalid:") {
		t.Fatalf("error=%v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
}
