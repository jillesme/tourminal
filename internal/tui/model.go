package tui

import (
	"fmt"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/jillesme/tourminal/internal/resolver"
	"github.com/jillesme/tourminal/internal/tour"
	"github.com/jillesme/tourminal/internal/workspace"
)

type screen int

const (
	screenTours screen = iota
	screenPlayer
	screenFinished
)

type focus int

const (
	focusCode focus = iota
	focusNotes
)

// Config controls a Tourminal player. Its zero value starts at the first step,
// detects the terminal theme automatically, and enables color.
type Config struct {
	StartStep int
	Theme     ThemeMode
	NoColor   bool
}

// Model implements the Bubble Tea model for the Tourminal player.
type Model struct {
	root       string
	refs       []workspace.TourRef
	tour       *tour.Tour
	tourIndex  int
	stepIndex  int
	startStep  int
	tourCursor int
	stepCursor int

	screen     screen
	focus      focus
	showHelp   bool
	stepPicker bool
	width      int
	height     int
	compact    bool
	hasCode    bool
	themeMode  ThemeMode
	dark       bool
	noColor    bool
	styles     themeStyles

	code        viewport.Model
	notes       viewport.Model
	resolved    resolver.ResolvedStep
	stepError   string
	codeTitle   string
	notesRaw    string
	notesRender string
	codeSource  string
	codeLines   []int
	targetRow   int
	noteRows    int
	tourWarning string
}

// New creates a Tourminal player for refs rooted at root.
func New(root string, refs []workspace.TourRef, config Config) (*Model, error) {
	if len(refs) == 0 {
		return nil, fmt.Errorf("no CodeTours found in %s", root)
	}
	themeMode := config.Theme
	if themeMode == "" {
		themeMode = ThemeAuto
	}
	if themeMode != ThemeAuto && themeMode != ThemeLight && themeMode != ThemeDark {
		return nil, fmt.Errorf("invalid theme %q", themeMode)
	}
	dark := themeMode != ThemeLight
	m := &Model{
		root: root, refs: refs, startStep: config.StartStep,
		width: 100, height: 30,
		code:      viewport.New(viewport.WithWidth(60), viewport.WithHeight(24)),
		notes:     viewport.New(viewport.WithWidth(39), viewport.WithHeight(24)),
		themeMode: themeMode, dark: dark, noColor: config.NoColor,
		styles: newThemeStyles(dark, config.NoColor),
	}
	m.code.FillHeight = true
	m.code.SoftWrap = false
	m.code.SetHorizontalStep(8)
	m.notes.FillHeight = true
	m.notes.SoftWrap = true
	if len(refs) == 1 {
		if err := m.startTourAt(0, config.StartStep); err != nil {
			return nil, err
		}
	} else {
		m.screen = screenTours
		for i, ref := range refs {
			if ref.Primary {
				m.tourCursor = i
				break
			}
		}
	}
	m.resize(m.width, m.height)
	return m, nil
}

// Init requests the terminal background when automatic theming is enabled.
func (m *Model) Init() tea.Cmd {
	if m.themeMode == ThemeAuto && !m.noColor {
		return tea.RequestBackgroundColor
	}
	return nil
}

// Update applies a Bubble Tea message to the player state.
func (m *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.BackgroundColorMsg:
		if m.themeMode == ThemeAuto {
			m.setDarkTheme(msg.IsDark())
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyPressMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if m.showHelp {
			if key == "q" {
				return m, tea.Quit
			}
			m.showHelp = false
			return m, nil
		}
		if key == "?" {
			m.showHelp = true
			return m, nil
		}
		if m.stepPicker {
			return m.updateStepPicker(key)
		}
		switch m.screen {
		case screenTours:
			return m.updateTourPicker(key)
		case screenFinished:
			return m.updateFinished(key)
		case screenPlayer:
			if command := m.updatePlayerKey(key); command != nil {
				return m, command
			}
		}
	}

	if m.screen == screenPlayer {
		var cmd tea.Cmd
		if m.hasCode {
			m.code, cmd = m.code.Update(message)
		} else {
			m.notes, cmd = m.notes.Update(message)
		}
		return m, cmd
	}
	return m, nil
}

func (m *Model) setDarkTheme(dark bool) {
	if m.dark == dark {
		return
	}
	m.dark = dark
	m.styles = newThemeStyles(dark, m.noColor)
	if m.tour != nil {
		m.prepareStep()
	}
}
