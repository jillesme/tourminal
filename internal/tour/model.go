package tour

import "strings"

// Position is a 1-based source position, matching the CodeTour schema.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Selection struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

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

	Path string `json:"-"`
}

func (s Step) Label(number int) string {
	if s.Title != "" {
		return s.Title
	}
	for _, line := range strings.Split(strings.TrimSpace(s.Description), "\n") {
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
	return "Step " + itoa(number)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
