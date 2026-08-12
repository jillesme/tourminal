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

	light, err := New(root, refs, Config{StartStep: 1, Theme: ThemeLight})
	if err != nil {
		t.Fatal(err)
	}
	if light.dark || light.Init() != nil {
		t.Fatal("explicit light mode should use the light palette without querying the terminal")
	}

	dark, err := New(root, refs, Config{StartStep: 1, Theme: ThemeDark})
	if err != nil {
		t.Fatal(err)
	}
	if !dark.dark || dark.Init() != nil {
		t.Fatal("explicit dark mode should use the dark palette without querying the terminal")
	}
	if light.notesRender == dark.notesRender {
		t.Fatal("light and dark modes rendered identical Markdown")
	}

	auto, err := New(root, refs, Config{StartStep: 1, Theme: ThemeAuto})
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

	plain, err := New(root, refs, Config{NoColor: true})
	if err != nil {
		t.Fatal(err)
	}
	if plain.themeMode != ThemeAuto || plain.Init() != nil {
		t.Fatal("zero-value theme should be automatic without querying when color is disabled")
	}
	if strings.Contains(plain.View().Content, "\x1b[") {
		t.Fatalf("no-color view contains ANSI escapes: %q", plain.View().Content)
	}

	if _, err := New(root, refs, Config{Theme: "sepia"}); err == nil {
		t.Fatal("expected an invalid configured theme to fail")
	}
}

func TestKeyboardStateTransitions(t *testing.T) {
	root := t.TempDir()
	tourPath := filepath.Join(root, "keys.tour")
	contents := `{"title":"Keys","steps":[{"description":"First"},{"description":"Second"}]}`
	if err := os.WriteFile(tourPath, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	m, err := New(root, []workspace.TourRef{{Path: tourPath, Title: "Keys"}}, Config{})
	if err != nil {
		t.Fatal(err)
	}

	press(t, m, "n")
	if m.stepIndex != 1 {
		t.Fatalf("next step index = %d", m.stepIndex)
	}
	press(t, m, "p")
	if m.stepIndex != 0 {
		t.Fatalf("previous step index = %d", m.stepIndex)
	}
	press(t, m, "g")
	if !m.stepPicker {
		t.Fatal("step picker did not open")
	}
	press(t, m, "j")
	press(t, m, "enter")
	if m.stepPicker || m.stepIndex != 1 {
		t.Fatalf("picker state=%v step=%d", m.stepPicker, m.stepIndex)
	}
	press(t, m, "n")
	if m.screen != screenFinished || !strings.Contains(ansi.Strip(m.View().Content), "Tour complete") {
		t.Fatal("tour did not reach the finished screen")
	}
	press(t, m, "p")
	if m.screen != screenPlayer || m.stepIndex != 1 {
		t.Fatal("previous did not return from the finished screen")
	}
	press(t, m, "?")
	if !m.showHelp || !strings.Contains(ansi.Strip(m.View().Content), "Tourminal help") {
		t.Fatal("help did not open")
	}
	press(t, m, "esc")
	if m.showHelp {
		t.Fatal("help did not close")
	}
}

func TestTourPickerAndLinkedTourErrors(t *testing.T) {
	root := t.TempDir()
	firstPath := filepath.Join(root, "first.tour")
	secondPath := filepath.Join(root, "second.tour")
	writeTestTour(t, firstPath, `{"title":"First","nextTour":"Second","steps":[{"description":"First step"}]}`)
	writeTestTour(t, secondPath, `{"title":"Second","steps":[{"description":"Second step"}]}`)
	refs := []workspace.TourRef{{Path: firstPath, Title: "First"}, {Path: secondPath, Title: "Second"}}

	m, err := New(root, refs, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if m.screen != screenTours || !strings.Contains(ansi.Strip(m.View().Content), "Choose a CodeTour") {
		t.Fatal("multiple tours did not open the tour picker")
	}
	press(t, m, "j")
	press(t, m, "enter")
	if m.tour.Title != "Second" {
		t.Fatalf("selected tour = %q", m.tour.Title)
	}

	if err := m.startTourAt(0, 1); err != nil {
		t.Fatal(err)
	}
	m.nextStep()
	if m.tour.Title != "Second" {
		t.Fatalf("linked tour = %q", m.tour.Title)
	}

	if err := m.startTourAt(0, 1); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(secondPath); err != nil {
		t.Fatal(err)
	}
	m.nextStep()
	if m.tour.Title != "First" || !strings.Contains(m.stepError, "open next tour") ||
		!strings.Contains(ansi.Strip(m.View().Content), "Step warning") {
		t.Fatalf("linked-tour failure was not surfaced: tour=%q error=%q", m.tour.Title, m.stepError)
	}

	writeTestTour(t, firstPath, `{"title":"First","nextTour":"Missing","steps":[{"description":"First step"}]}`)
	refs = []workspace.TourRef{{Path: firstPath, Title: "First"}}
	m, err = New(root, refs, Config{})
	if err != nil {
		t.Fatal(err)
	}
	m.nextStep()
	if m.screen != screenPlayer || !strings.Contains(m.stepError, `next tour "Missing" was not found`) {
		t.Fatalf("missing link error = %q", m.stepError)
	}
}

func press(t *testing.T, m *Model, key string) {
	t.Helper()
	message := tea.KeyPressMsg(tea.Key{Text: key, Code: []rune(key)[0]})
	if key == "enter" {
		message = tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	} else if key == "esc" {
		message = tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
	}
	m.Update(message)
}

func writeTestTour(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
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

	m, err := New(root, []workspace.TourRef{{Path: tourPath, Title: "Walkthrough"}}, Config{StartStep: 1})
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
