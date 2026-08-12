package workspace

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jillesme/tourminal/internal/tour"
)

// TourRef contains the metadata needed to list and open a discovered tour.
type TourRef struct {
	Path        string
	Title       string
	Description string
	Primary     bool
	Warning     string
}

var mainTourFiles = []string{".tour", filepath.Join(".vscode", "main.tour"), "main.tour"}
var tourDirectories = []string{filepath.Join(".vscode", "tours"), filepath.Join(".github", "tours"), ".tours"}

// ResolveRoot finds the nearest ancestor containing a standard CodeTour
// location, starting from start.
func ResolveRoot(start string) (string, error) {
	if start == "" {
		var err error
		start, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return RootForTour(abs)
	}

	for dir := abs; ; dir = filepath.Dir(dir) {
		if hasTourLocation(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return abs, nil
}

// RootForTour returns the workspace root implied by a tour path.
func RootForTour(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve tour path: %w", err)
	}
	dir := filepath.Dir(abs)
	base := filepath.Base(abs)
	if base == ".tour" || base == "main.tour" {
		if filepath.Base(dir) == ".vscode" {
			return filepath.Dir(dir), nil
		}
		return dir, nil
	}

	for current := dir; ; current = filepath.Dir(current) {
		if filepath.Base(current) == ".tours" {
			return filepath.Dir(current), nil
		}
		if filepath.Base(current) == "tours" {
			parent := filepath.Dir(current)
			if base := filepath.Base(parent); base == ".vscode" || base == ".github" {
				return filepath.Dir(parent), nil
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	return dir, nil
}

// Discover finds tours that apply to the current platform.
func Discover(root string) ([]TourRef, []error) {
	return discover(root, true)
}

// DiscoverAll loads every tour regardless of whether its platform-specific
// when expression matches the current machine. Validation uses this so CI on
// one operating system still checks tours intended for another.
func DiscoverAll(root string) ([]TourRef, []error) {
	return discover(root, false)
}

func discover(root string, filterWhen bool) ([]TourRef, []error) {
	seen := make(map[string]bool)
	var refs []TourRef
	var diagnostics []error

	add := func(path string) {
		path = filepath.Clean(path)
		if seen[path] {
			return
		}
		seen[path] = true
		loaded, err := tour.Load(path)
		if err != nil {
			diagnostics = append(diagnostics, fmt.Errorf("%s: %w", path, err))
			return
		}
		whenMatches, whenErr := EvaluateWhen(loaded.When)
		if filterWhen && whenErr == nil && !whenMatches {
			return
		}
		warning := ""
		if whenErr != nil {
			warning = "The tour's 'when' expression could not be evaluated safely: " + whenErr.Error()
		}
		refs = append(refs, TourRef{
			Path: path, Title: loaded.Title, Description: loaded.Description,
			Primary: loaded.IsPrimary || isImplicitPrimary(loaded.Title),
			Warning: warning,
		})
	}

	for _, rel := range mainTourFiles {
		path := filepath.Join(root, rel)
		if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
			add(path)
		}
	}
	for _, rel := range tourDirectories {
		dir := filepath.Join(root, rel)
		walkErr := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				diagnostics = append(diagnostics, err)
				return nil
			}
			if entry.IsDir() {
				return nil
			}
			if strings.EqualFold(filepath.Ext(entry.Name()), ".tour") {
				add(path)
			}
			return nil
		})
		if walkErr != nil && !os.IsNotExist(walkErr) {
			diagnostics = append(diagnostics, fmt.Errorf("walk tour directory %s: %w", dir, walkErr))
		}
	}

	sort.SliceStable(refs, func(i, j int) bool {
		if refs[i].Primary != refs[j].Primary {
			return refs[i].Primary
		}
		return strings.ToLower(refs[i].Title) < strings.ToLower(refs[j].Title)
	})
	return refs, diagnostics
}

func isImplicitPrimary(title string) bool {
	title = strings.TrimPrefix(title, "#")
	return strings.HasPrefix(title, "1 - ") || strings.HasPrefix(title, "1: ")
}

func hasTourLocation(root string) bool {
	for _, rel := range mainTourFiles {
		if info, err := os.Stat(filepath.Join(root, rel)); err == nil && info.Mode().IsRegular() {
			return true
		}
	}
	for _, rel := range tourDirectories {
		if info, err := os.Stat(filepath.Join(root, rel)); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}
