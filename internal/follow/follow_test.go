package follow

import (
	"testing"

	"github.com/jillesme/tourminal/internal/tour"
)

func TestEffectiveStepUsesMarkerFallback(t *testing.T) {
	item := &tour.Tour{
		Title: "1 - Intro",
		Steps: []tour.Step{{Description: "x", File: "main.go"}},
	}
	if got := EffectiveStep(item, 0).Pattern; got != `CT1\.1\b` {
		t.Fatalf("pattern = %q, want %q", got, `CT1\.1\b`)
	}

	item.StepMarker = "custom"
	if got := EffectiveStep(item, 0).Pattern; got != `custom\.1\b` {
		t.Fatalf("pattern = %q, want %q", got, `custom\.1\b`)
	}
}

func TestNextNumberedTour(t *testing.T) {
	titles := []string{"1 - Intro", "Other", "2: Details"}
	if got := NextNumberedTour(titles[0], titles); got != 2 {
		t.Fatalf("index = %d, want 2", got)
	}
	if got := NextNumberedTour("Other", titles); got != -1 {
		t.Fatalf("unmatched index = %d, want -1", got)
	}
}

func TestEffectiveStepPreservesSelection(t *testing.T) {
	item := &tour.Tour{Title: "1 - Intro", Steps: []tour.Step{{File: "main.go", Selection: &tour.Selection{Start: tour.Position{Line: 1, Character: 1}, End: tour.Position{Line: 1, Character: 2}}}}}
	if got := EffectiveStep(item, 0).Pattern; got != "" {
		t.Fatalf("selection acquired pattern %q", got)
	}
}
