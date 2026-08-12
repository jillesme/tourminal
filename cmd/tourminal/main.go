package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/jillesme/tourminal/internal/render"
	"github.com/jillesme/tourminal/internal/tour"
	"github.com/jillesme/tourminal/internal/tui"
	"github.com/jillesme/tourminal/internal/validation"
	"github.com/jillesme/tourminal/internal/workspace"
	"github.com/jillesme/tourminal/skills"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "tourminal:", render.TerminalLine(err.Error()))
		os.Exit(1)
	}
}

func run(args []string) error {
	return runWithIO(args, os.Stdout, os.Stderr)
}

func runWithIO(args []string, stdout, stderr io.Writer) error {
	if len(args) > 0 && args[0] == "validate" {
		return validateCommand(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "skill" {
		return skillCommand(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "version" {
		if len(args) != 1 {
			return fmt.Errorf("usage: tourminal version")
		}
		fmt.Fprintln(stdout, versionString())
		return nil
	}

	flags := flag.NewFlagSet("tourminal", flag.ContinueOnError)
	flags.SetOutput(stderr)
	tourPath := flags.String("tour", "", "open a specific .tour file")
	step := flags.Int("step", 1, "start at a 1-based step")
	noColor := flags.Bool("no-color", false, "disable color output")
	themeName := flags.String("theme", defaultTheme(), "color theme: auto, light, or dark")
	showSkill := flags.Bool("skill", false, "print the bundled tour-creation skill and exit")
	showVersion := flags.Bool("version", false, "print version and exit")
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: tourminal [flags] [workspace]")
		fmt.Fprintln(flags.Output(), "       tourminal validate [workspace-or-tour]")
		fmt.Fprintln(flags.Output(), "       tourminal skill")
		fmt.Fprintln(flags.Output(), "       tourminal version")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *showVersion {
		fmt.Fprintln(stdout, versionString())
		return nil
	}
	if *showSkill {
		return skillCommand(nil, stdout, stderr)
	}
	if *noColor {
		_ = os.Setenv("NO_COLOR", "1")
	}
	themeMode, err := tui.ParseThemeMode(*themeName)
	if err != nil {
		return err
	}

	var root string
	var refs []workspace.TourRef
	if *tourPath != "" {
		abs, err := filepath.Abs(*tourPath)
		if err != nil {
			return err
		}
		loaded, err := tour.Load(abs)
		if err != nil {
			return err
		}
		root = workspace.RootForTour(abs)
		refs = []workspace.TourRef{{Path: abs, Title: loaded.Title, Description: loaded.Description, Primary: loaded.IsPrimary}}
	} else {
		if flags.NArg() > 1 {
			return fmt.Errorf("expected at most one workspace path")
		}
		start := ""
		if flags.NArg() == 1 {
			start = flags.Arg(0)
		}
		var err error
		root, err = workspace.ResolveRoot(start)
		if err != nil {
			return err
		}
		var diagnostics []error
		refs, diagnostics = workspace.Discover(root)
		for _, diagnostic := range diagnostics {
			fmt.Fprintln(stderr, "warning:", render.TerminalLine(diagnostic.Error()))
		}
	}

	model, err := tui.New(root, refs, *step, themeMode)
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(model).Run()
	return err
}

func defaultTheme() string {
	if value := strings.TrimSpace(os.Getenv("TOURMINAL_THEME")); value != "" {
		return value
	}
	return string(tui.ThemeAuto)
}

func versionString() string {
	result := "tourminal " + version
	var details []string
	if commit != "" && commit != "unknown" {
		details = append(details, "commit "+commit)
	}
	if date != "" && date != "unknown" {
		details = append(details, "built "+date)
	}
	if len(details) > 0 {
		result += " (" + strings.Join(details, ", ") + ")"
	}
	return result
}

func skillCommand(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("skill", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { fmt.Fprintln(flags.Output(), "Usage: tourminal skill") }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("usage: tourminal skill")
	}
	_, err := fmt.Fprint(stdout, skills.CreateCodeTour())
	return err
}

func validateCommand(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("validate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { fmt.Fprintln(flags.Output(), "Usage: tourminal validate [workspace-or-tour]") }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 1 {
		return fmt.Errorf("usage: tourminal validate [workspace-or-tour]")
	}
	path := ""
	if flags.NArg() == 1 {
		path = flags.Arg(0)
	}

	if path != "" {
		info, statErr := os.Stat(path)
		if statErr != nil && strings.EqualFold(filepath.Ext(path), ".tour") {
			return statErr
		}
		if statErr == nil && !info.IsDir() {
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			loaded, err := tour.Load(abs)
			if err != nil {
				return err
			}
			return reportTourValidation(workspace.RootForTour(abs), abs, loaded, stdout, stderr)
		}
	}

	root, err := workspace.ResolveRoot(path)
	if err != nil {
		return err
	}
	refs, diagnostics := workspace.DiscoverAll(root)
	errorCount := len(diagnostics)
	for _, diagnostic := range diagnostics {
		fmt.Fprintln(stderr, "invalid:", render.TerminalLine(diagnostic.Error()))
	}
	if len(refs) == 0 {
		if errorCount > 0 {
			return fmt.Errorf("%d invalid tour(s)", errorCount)
		}
		return fmt.Errorf("no CodeTours found in %s", root)
	}

	loadedTours := make([]*tour.Tour, 0, len(refs))
	for _, ref := range refs {
		loaded, loadErr := tour.Load(ref.Path)
		if loadErr != nil {
			// DiscoverAll already emitted this diagnostic, so avoid counting it
			// twice if the file changed between discovery and validation.
			continue
		}
		loadedTours = append(loadedTours, loaded)
		result := validation.Tour(root, loaded)
		for _, warning := range result.Warnings {
			fmt.Fprintf(stderr, "warning: %s: %s\n", render.TerminalLine(ref.Path), render.TerminalLine(warning))
		}
		for _, validationError := range result.Errors {
			fmt.Fprintf(stderr, "invalid: %s: %s\n", render.TerminalLine(ref.Path), render.TerminalLine(validationError))
		}
		errorCount += len(result.Errors)
		if len(result.Errors) == 0 {
			fmt.Fprintf(stdout, "valid: %s (%d steps)\n", render.TerminalLine(loaded.Title), len(loaded.Steps))
		}
	}
	for _, linkError := range validation.MissingNextTours(loadedTours) {
		fmt.Fprintln(stderr, "invalid:", render.TerminalLine(linkError))
		errorCount++
	}
	if errorCount > 0 {
		return fmt.Errorf("%d validation error(s)", errorCount)
	}
	return nil
}

func reportTourValidation(root, path string, loaded *tour.Tour, stdout, stderr io.Writer) error {
	result := validation.Tour(root, loaded)
	for _, warning := range result.Warnings {
		fmt.Fprintf(stderr, "warning: %s: %s\n", render.TerminalLine(path), render.TerminalLine(warning))
	}
	for _, validationError := range result.Errors {
		fmt.Fprintf(stderr, "invalid: %s: %s\n", render.TerminalLine(path), render.TerminalLine(validationError))
	}
	if len(result.Errors) > 0 {
		return fmt.Errorf("%d validation error(s)", len(result.Errors))
	}
	fmt.Fprintf(stdout, "valid: %s (%d steps)\n", render.TerminalLine(loaded.Title), len(loaded.Steps))
	return nil
}
