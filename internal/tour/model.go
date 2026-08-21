package tour

import (
	"strconv"
	"strings"
)

// Position is a 1-based source position, matching the CodeTour schema.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Selection identifies an inclusive source range.
type Selection struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Step describes one stop in a CodeTour.
type Step struct {
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description"`
	Icon        string     `json:"icon,omitempty"`
	File        string     `json:"file,omitempty"`
	Directory   string     `json:"directory,omitempty"`
	Contents    string     `json:"contents,omitempty"`
	URI         string     `json:"uri,omitempty"`
	View        string     `json:"view,omitempty"`
	Line        int        `json:"line,omitempty"`
	Pattern     string     `json:"pattern,omitempty"`
	Selection   *Selection `json:"selection,omitempty"`
	Commands    []string   `json:"commands,omitempty"`
	MarkerTitle string     `json:"markerTitle,omitempty"`
}

// Tour is the CodeTour document model supported by Tourminal.
type Tour struct {
	Schema      string `json:"$schema,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Ref         string `json:"ref,omitempty"`
	IsPrimary   bool   `json:"isPrimary,omitempty"`
	NextTour    string `json:"nextTour,omitempty"`
	StepMarker  string `json:"stepMarker,omitempty"`
	When        string `json:"when,omitempty"`
	Steps       []Step `json:"steps"`
}

// Label returns the best human-readable label available for the step.
func (s Step) Label(number int) string {
	if s.Title != "" {
		return s.Title
	}
	for line := range strings.SplitSeq(strings.TrimSpace(s.Description), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			label := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			if label != "" {
				return label
			}
			break
		}
	}
	if s.MarkerTitle != "" {
		return s.MarkerTitle
	}
	if s.Directory != "" {
		return s.Directory
	}
	if s.File != "" {
		return s.File
	}
	if s.URI != "" {
		return s.URI
	}
	return "Step " + strconv.Itoa(number)
}
