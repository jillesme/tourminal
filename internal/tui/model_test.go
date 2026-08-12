package tui

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/jillesme/tourminal/internal/workspace"
)

func TestThemeModes(t *testing.T) {
	for input, want := range map[string]ThemeMode{
		"auto":    ThemeAuto,
		" LIGHT ": ThemeLight,
		"Dark":    ThemeDark,
	} {
		got, err := ParseThemeMode(input)
		if err != nil {
			t.Fatalf("ParseThemeMode(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseThemeMode(%q) = %q, want %q", input, got, want)
		}
	}
	if _, err := ParseThemeMode("sepia"); err == nil {
		t.Fatal("expected an invalid theme to fail")
	}

	root := t.TempDir()
	tourPath := filepath.Join(root, "theme.tour")
	contents := `{"title":"Theme","steps":[{"description":"# Theme\n\nReadable text"}]}`
	if err := os.WriteFile(tourPath, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	refs := []workspace.TourRef{{Path: tourPath, Title: "Theme"}}

	light, err := New(root, refs, 1, ThemeLight)
	if err != nil {
		t.Fatal(err)
	}
	if light.dark || light.Init() != nil {
		t.Fatal("explicit light mode should use the light palette without querying the terminal")
	}

	dark, err := New(root, refs, 1, ThemeDark)
	if err != nil {
		t.Fatal(err)
	}
	if !dark.dark || dark.Init() != nil {
		t.Fatal("explicit dark mode should use the dark palette without querying the terminal")
	}
	if light.notesRender == dark.notesRender {
		t.Fatal("light and dark modes rendered identical Markdown")
	}

	auto, err := New(root, refs, 1, ThemeAuto)
	if err != nil {
		t.Fatal(err)
	}
	if auto.Init() == nil {
		t.Fatal("auto mode should query the terminal background color")
	}
	auto.Update(tea.BackgroundColorMsg{Color: color.White})
	if auto.dark {
		t.Fatal("auto mode did not switch to the light palette")
	}
}

func TestPlayerFollowsContentAndFileSteps(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	tourPath := filepath.Join(root, "main.tour")
	contents := `{"title":"Walkthrough","steps":[{"description":"# Intro\n\nWelcome"},{"description":"# Code\n\nExplanation that must remain visible.","file":"main.go","line":3}]}`
	if err := os.WriteFile(tourPath, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	m, err := New(root, []workspace.TourRef{{Path: tourPath, Title: "Walkthrough"}}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if m.stepIndex != 0 || m.hasCode {
		t.Fatalf("unexpected initial state: step=%d hasCode=%v", m.stepIndex, m.hasCode)
	}
	m.nextStep()
	if m.stepIndex != 1 || !m.hasCode || m.resolved.TargetLine != 3 {
		t.Fatalf("unexpected file state: step=%d hasCode=%v line=%d", m.stepIndex, m.hasCode, m.resolved.TargetLine)
	}

	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if !m.compact {
		t.Fatal("expected compact layout")
	}
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "Step 2 of 2") || !strings.Contains(view, "main.go") ||
		!strings.Contains(view, "Explanation that must remain visible") {
		t.Fatalf("unexpected view: %q", view)
	}
	assertInlineDescription(t, view)
	if got := m.sourceLineAt(m.targetRow); got != 3 {
		t.Fatalf("target display row maps to source line %d, want 3", got)
	}
	if got := m.sourceLineAt(m.targetRow - 1); got != 0 {
		t.Fatalf("row before target maps to source line %d, want inserted description", got)
	}
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if m.compact {
		t.Fatal("expected split-pane layout")
	}
	view = ansi.Strip(m.View().Content)
	if !strings.Contains(view, "main.go") || !strings.Contains(view, "Explanation that must remain visible") {
		t.Fatalf("inline view omitted tour content: %q", view)
	}
	if strings.Contains(view, "CodeTour description") {
		t.Fatalf("inline description has a redundant label: %q", view)
	}
	assertInlineDescription(t, view)
}

func assertInlineDescription(t *testing.T, view string) {
	t.Helper()
	before := strings.Index(view, "package main")
	description := strings.Index(view, "Explanation that must remain visible")
	target := strings.Index(view, "func main()")
	if before < 0 || description < 0 || target < 0 || !(before < description && description < target) {
		t.Fatalf("description is not inline before the target: %q", view)
	}
}

func TestStepMarkerAndNumberedTourHelpers(t *testing.T) {
	if got := stepMarkerPrefix("1 - Intro", ""); got != "CT1" {
		t.Fatalf("marker = %q", got)
	}
	if got := stepMarkerPrefix("Anything", "custom"); got != "custom" {
		t.Fatalf("explicit marker = %q", got)
	}
	refs := []workspace.TourRef{{Title: "1 - Intro"}, {Title: "2: Details"}}
	if got := nextNumberedTour(refs[0].Title, refs); got != 1 {
		t.Fatalf("next numbered tour = %d", got)
	}
}
