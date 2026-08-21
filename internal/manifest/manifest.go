// Package manifest builds a versioned, editor-neutral representation of
// CodeTours for integrations such as the Tourminal Neovim plugin.
package manifest

import (
	"strings"

	"github.com/jillesme/tourminal/internal/follow"
	"github.com/jillesme/tourminal/internal/resolver"
	"github.com/jillesme/tourminal/internal/tour"
	"github.com/jillesme/tourminal/internal/workspace"
)

// APIVersion is incremented when the JSON contract changes incompatibly.
const APIVersion = 1

// Manifest is the JSON document consumed by editor integrations.
type Manifest struct {
	APIVersion  int         `json:"apiVersion"`
	Root        string      `json:"root"`
	Diagnostics []string    `json:"diagnostics"`
	Tours       []TourEntry `json:"tours"`
}

// TourEntry contains the metadata and eagerly resolved steps for one tour.
type TourEntry struct {
	Path         string      `json:"path"`
	Title        string      `json:"title"`
	Description  string      `json:"description,omitempty"`
	Ref          string      `json:"ref,omitempty"`
	Primary      bool        `json:"primary"`
	Warning      string      `json:"warning,omitempty"`
	NextTour     string      `json:"nextTour,omitempty"`
	NextTourPath string      `json:"nextTourPath,omitempty"`
	Steps        []StepEntry `json:"steps"`
}

// StepEntry contains presentation metadata plus its resolved location.
type StepEntry struct {
	Number      int             `json:"number"`
	Label       string          `json:"label"`
	Title       string          `json:"title,omitempty"`
	Description string          `json:"description"`
	Icon        string          `json:"icon,omitempty"`
	File        string          `json:"file,omitempty"`
	Directory   string          `json:"directory,omitempty"`
	URI         string          `json:"uri,omitempty"`
	View        string          `json:"view,omitempty"`
	Line        int             `json:"line,omitempty"`
	Pattern     string          `json:"pattern,omitempty"`
	Selection   *tour.Selection `json:"selection,omitempty"`
	Commands    []string        `json:"commands,omitempty"`
	Resolved    ResolvedEntry   `json:"resolved"`
	Error       string          `json:"error,omitempty"`
}

// ResolvedEntry is the portion of a resolved step needed by an editor. Source
// is included for embedded content and directory listings, but workspace files
// are opened directly by the editor.
type ResolvedEntry struct {
	Kind           string `json:"kind"`
	Path           string `json:"path,omitempty"`
	DisplayPath    string `json:"displayPath,omitempty"`
	Source         string `json:"source,omitempty"`
	TargetLine     int    `json:"targetLine,omitempty"`
	SelectionStart int    `json:"selectionStart,omitempty"`
	SelectionEnd   int    `json:"selectionEnd,omitempty"`
}

// Build creates a manifest for refs rooted at root. A tour that changes or
// becomes unreadable during the build is reported as a diagnostic and skipped.
func Build(root string, refs []workspace.TourRef, diagnostics []error) Manifest {
	result := Manifest{
		APIVersion:  APIVersion,
		Root:        root,
		Diagnostics: make([]string, 0, len(diagnostics)),
		Tours:       make([]TourEntry, 0, len(refs)),
	}
	for _, diagnostic := range diagnostics {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Error())
	}

	loadedTours := make([]*tour.Tour, 0, len(refs))
	loadedRefs := make([]workspace.TourRef, 0, len(refs))
	for _, ref := range refs {
		loaded, err := tour.Load(ref.Path)
		if err != nil {
			result.Diagnostics = append(result.Diagnostics, ref.Path+": "+err.Error())
			continue
		}
		loadedTours = append(loadedTours, loaded)
		loadedRefs = append(loadedRefs, ref)
	}

	titles := make([]string, len(loadedTours))
	for index, item := range loadedTours {
		titles[index] = item.Title
	}

	for index, item := range loadedTours {
		ref := loadedRefs[index]
		entry := TourEntry{
			Path: ref.Path, Title: item.Title, Description: item.Description,
			Ref: item.Ref, Primary: ref.Primary, NextTour: item.NextTour,
			Warning:      strings.TrimSpace(ref.Warning + " " + workspace.GitRefWarning(root, item.Ref)),
			NextTourPath: nextTourPath(item, loadedRefs, titles),
			Steps:        make([]StepEntry, 0, len(item.Steps)),
		}
		for stepIndex, rawStep := range item.Steps {
			effective := follow.EffectiveStep(item, stepIndex)
			resolved, resolveErr := resolver.Resolve(root, effective)
			stepEntry := StepEntry{
				Number: stepIndex + 1, Label: rawStep.Label(stepIndex + 1),
				Title: rawStep.Title, Description: rawStep.Description, Icon: rawStep.Icon,
				File: rawStep.File, Directory: rawStep.Directory, URI: rawStep.URI,
				View: rawStep.View, Line: rawStep.Line, Pattern: effective.Pattern,
				Selection: rawStep.Selection, Commands: rawStep.Commands,
				Resolved: resolvedEntry(resolved, rawStep),
			}
			if resolveErr != nil {
				stepEntry.Error = resolveErr.Error()
			}
			entry.Steps = append(entry.Steps, stepEntry)
		}
		result.Tours = append(result.Tours, entry)
	}
	return result
}

func nextTourPath(item *tour.Tour, refs []workspace.TourRef, titles []string) string {
	if item.NextTour != "" {
		for index, title := range titles {
			if title == item.NextTour {
				return refs[index].Path
			}
		}
		return ""
	}
	if index := follow.NextNumberedTour(item.Title, titles); index >= 0 {
		return refs[index].Path
	}
	return ""
}

func resolvedEntry(resolved resolver.ResolvedStep, step tour.Step) ResolvedEntry {
	entry := ResolvedEntry{
		Kind: resolvedKind(resolved.Kind, step), Path: resolved.Path,
		DisplayPath: resolved.DisplayPath, TargetLine: resolved.TargetLine,
		SelectionStart: resolved.SelectionStart, SelectionEnd: resolved.SelectionEnd,
	}
	if resolved.Kind == resolver.Directory || step.Contents != "" {
		entry.Source = resolved.Source
	}
	return entry
}

func resolvedKind(kind resolver.Kind, step tour.Step) string {
	switch kind {
	case resolver.File:
		if step.Contents != "" {
			return "embedded"
		}
		return "file"
	case resolver.Directory:
		return "directory"
	case resolver.URI:
		return "uri"
	default:
		return "content"
	}
}
