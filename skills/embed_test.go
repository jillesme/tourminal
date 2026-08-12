package skills

import (
	"strings"
	"testing"
)

func TestCreateCodeTourIsSelfContained(t *testing.T) {
	content := CreateCodeTour()
	for _, expected := range []string{
		"name: create-codetour",
		"## Workflow",
		"# Bundled reference: references/tour-schema.md",
		`"$schema": "https://aka.ms/codetour-schema"`,
		"git rev-parse HEAD",
		"Pinned snapshot",
		"tour validate",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("bundled skill does not contain %q", expected)
		}
	}
	if strings.Contains(content, "TODO") {
		t.Fatal("bundled skill contains an unfinished TODO")
	}
	if strings.Contains(content, "tourminal validate") {
		t.Fatal("bundled skill uses the compatibility command")
	}
}
