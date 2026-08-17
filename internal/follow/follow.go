// Package follow contains editor-neutral CodeTour playback behavior.
package follow

import (
	"regexp"
	"strconv"

	"github.com/jillesme/tourminal/internal/tour"
)

var numberedTourPattern = regexp.MustCompile(`^#?(\d+)\s*[-:]`)

// EffectiveStep returns the step location that should be resolved during
// playback. Marker-based tours derive an implicit pattern for file steps that
// do not specify another anchor.
func EffectiveStep(item *tour.Tour, index int) tour.Step {
	step := item.Steps[index]
	if step.File == "" || step.Line != 0 || step.Pattern != "" {
		return step
	}
	if marker := StepMarkerPrefix(item.Title, item.StepMarker); marker != "" {
		step.Pattern = regexp.QuoteMeta(marker + "." + strconv.Itoa(index+1))
	}
	return step
}

// StepMarkerPrefix returns the explicit step marker or the conventional
// marker derived from a numbered tour title.
func StepMarkerPrefix(title, explicit string) string {
	if explicit != "" {
		return explicit
	}
	match := numberedTourPattern.FindStringSubmatch(title)
	if len(match) == 2 {
		return "CT" + match[1]
	}
	return ""
}

// NextNumberedTour returns the index of the next numbered tour, or -1 when
// title is not numbered or no matching successor exists.
func NextNumberedTour(title string, candidates []string) int {
	match := numberedTourPattern.FindStringSubmatch(title)
	if len(match) != 2 {
		return -1
	}
	number, err := strconv.Atoi(match[1])
	if err != nil {
		return -1
	}
	wanted := strconv.Itoa(number + 1)
	for index, candidate := range candidates {
		candidateMatch := numberedTourPattern.FindStringSubmatch(candidate)
		if len(candidateMatch) == 2 && candidateMatch[1] == wanted {
			return index
		}
	}
	return -1
}
