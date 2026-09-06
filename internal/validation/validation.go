package validation

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/jillesme/tourminal/internal/follow"
	"github.com/jillesme/tourminal/internal/resolver"
	"github.com/jillesme/tourminal/internal/tour"
	"github.com/jillesme/tourminal/internal/workspace"
)

// Result contains actionable diagnostics from resolving a tour against a
// workspace. Errors make a tour invalid; warnings describe safe degradation.
type Result struct {
	Errors   []string
	Warnings []string
}

func (r *Result) addError(format string, args ...any) {
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
}

func (r *Result) addWarning(format string, args ...any) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, args...))
}

// Check resolves every location and anchor in t against root.
func Check(root string, t *tour.Tour) Result {
	var result Result

	if _, err := workspace.EvaluateWhen(t.When); err != nil {
		result.addError("when expression: %v", err)
	}
	if err := workspace.ValidateGitRef(root, t.Ref); err != nil {
		result.addError("ref: %v", err)
	} else if warning := workspace.GitRefWarning(root, t.Ref); warning != "" {
		result.addWarning("ref: %s", warning)
	}

	for index := range t.Steps {
		validateStep(root, index+1, follow.EffectiveStep(t, index), &result)
	}
	return result
}

func validateStep(root string, number int, step tour.Step, result *Result) {
	prefix := fmt.Sprintf("step %d", number)
	if step.Line > 0 && step.Pattern != "" {
		result.addError("%s: line and pattern cannot both be set", prefix)
	}

	locationKinds := 0
	for _, present := range []bool{step.File != "", step.Directory != "", step.URI != ""} {
		if present {
			locationKinds++
		}
	}
	if locationKinds > 1 {
		result.addError("%s: use only one of file, directory, or uri", prefix)
	}
	if step.Contents != "" && (step.Directory != "" || step.URI != "") {
		result.addError("%s: contents can only be combined with file", prefix)
	}
	if (step.Line > 0 || step.Pattern != "" || step.Selection != nil) && step.File == "" && step.Contents == "" {
		result.addError("%s: line, pattern, and selection require file or contents", prefix)
	}
	if step.URI != "" {
		parsed, err := url.ParseRequestURI(step.URI)
		if err != nil || !parsed.IsAbs() {
			result.addError("%s: uri must be absolute: %q", prefix, step.URI)
		}
	}
	if len(step.Commands) > 0 {
		result.addWarning("%s: contains %d command(s); Tourminal will not execute them", prefix, len(step.Commands))
	}

	if locationKinds > 1 || (step.Contents != "" && (step.Directory != "" || step.URI != "")) {
		return
	}

	if _, err := resolver.Resolve(root, step); err != nil {
		result.addError("%s: %v", prefix, err)
	}
}

// MissingNextTours checks cross-tour links after all tours have loaded.
func MissingNextTours(tours []*tour.Tour) []string {
	titles := make(map[string]int, len(tours))
	var errors []string
	for _, item := range tours {
		titles[item.Title]++
	}
	orderedTitles := make([]string, 0, len(titles))
	for title := range titles {
		orderedTitles = append(orderedTitles, title)
	}
	sort.Strings(orderedTitles)
	for _, title := range orderedTitles {
		count := titles[title]
		if count > 1 {
			errors = append(errors, fmt.Sprintf("tour title %q is used %d times; titles must be unique for nextTour links", title, count))
		}
	}
	for _, item := range tours {
		if item.NextTour == "" {
			continue
		}
		if titles[item.NextTour] == 0 {
			errors = append(errors, fmt.Sprintf("tour %q: nextTour %q was not found", strings.TrimSpace(item.Title), item.NextTour))
		}
	}
	return errors
}
