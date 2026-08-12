package tui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/jillesme/tourminal/internal/tour"
	"github.com/jillesme/tourminal/internal/workspace"
)

func (m *Model) updatePlayerKey(key string) tea.Cmd {
	switch key {
	case "q":
		return tea.Quit
	case "n", "]", "space":
		m.nextStep()
	case "p", "[":
		m.previousStep()
	case "g":
		m.stepCursor = m.stepIndex
		m.stepPicker = true
	case "r":
		m.prepareStep()
	}
	return nil
}

func (m *Model) updateTourPicker(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.tourCursor > 0 {
			m.tourCursor--
		}
	case "down", "j":
		if m.tourCursor < len(m.refs)-1 {
			m.tourCursor++
		}
	case "enter":
		if err := m.startTourAt(m.tourCursor, m.startStep); err != nil {
			m.stepError = err.Error()
		}
	}
	return m, nil
}

func (m *Model) updateStepPicker(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q":
		return m, tea.Quit
	case "esc", "g":
		m.stepPicker = false
	case "up", "k":
		if m.stepCursor > 0 {
			m.stepCursor--
		}
	case "down", "j":
		if m.stepCursor < len(m.tour.Steps)-1 {
			m.stepCursor++
		}
	case "enter":
		m.stepIndex = m.stepCursor
		m.stepPicker = false
		m.prepareStep()
	}
	return m, nil
}

func (m *Model) updateFinished(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q", "enter", "esc":
		return m, tea.Quit
	case "p", "[":
		m.screen = screenPlayer
		m.stepIndex = len(m.tour.Steps) - 1
		m.prepareStep()
	}
	return m, nil
}

func (m *Model) startTourAt(index, step int) error {
	loaded, err := tour.Load(m.refs[index].Path)
	if err != nil {
		return err
	}
	if len(loaded.Steps) == 0 {
		return fmt.Errorf("tour %q has no steps", loaded.Title)
	}
	if step < 1 {
		step = 1
	}
	if step > len(loaded.Steps) {
		return fmt.Errorf("step %d is outside tour %q (1-%d)", step, loaded.Title, len(loaded.Steps))
	}
	m.tour = loaded
	m.tourIndex = index
	m.tourWarning = strings.TrimSpace(m.refs[index].Warning + " " + workspace.GitRefWarning(m.root, loaded.Ref))
	m.stepIndex = step - 1
	m.screen = screenPlayer
	m.focus = focusNotes
	m.prepareStep()
	return nil
}

func (m *Model) nextStep() {
	if m.stepIndex+1 < len(m.tour.Steps) {
		m.stepIndex++
		m.prepareStep()
		return
	}
	if m.tour.NextTour != "" {
		for i, ref := range m.refs {
			if ref.Title == m.tour.NextTour {
				if err := m.startTourAt(i, 1); err != nil {
					m.showStepError(fmt.Errorf("open next tour %q: %w", ref.Title, err))
				}
				return
			}
		}
		m.showStepError(fmt.Errorf("next tour %q was not found", m.tour.NextTour))
		return
	}
	if next := nextNumberedTour(m.tour.Title, m.refs); next >= 0 {
		if err := m.startTourAt(next, 1); err != nil {
			m.showStepError(fmt.Errorf("open next tour %q: %w", m.refs[next].Title, err))
		}
		return
	}
	m.screen = screenFinished
}

func (m *Model) showStepError(err error) {
	m.stepError = err.Error()
	m.renderNotes()
	m.centerTarget()
}

func (m *Model) previousStep() {
	if m.stepIndex > 0 {
		m.stepIndex--
		m.prepareStep()
	}
}

var numberedTourPattern = regexp.MustCompile(`^#?(\d+)\s*[-:]`)

func stepMarkerPrefix(title, explicit string) string {
	if explicit != "" {
		return explicit
	}
	match := numberedTourPattern.FindStringSubmatch(title)
	if len(match) == 2 {
		return "CT" + match[1]
	}
	return ""
}

func nextNumberedTour(title string, refs []workspace.TourRef) int {
	match := numberedTourPattern.FindStringSubmatch(title)
	if len(match) != 2 {
		return -1
	}
	number, err := strconv.Atoi(match[1])
	if err != nil {
		return -1
	}
	wanted := strconv.Itoa(number + 1)
	for i, ref := range refs {
		candidate := numberedTourPattern.FindStringSubmatch(ref.Title)
		if len(candidate) == 2 && candidate[1] == wanted {
			return i
		}
	}
	return -1
}
