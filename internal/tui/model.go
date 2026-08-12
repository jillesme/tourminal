package tui

import (
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/jillesme/tourminal/internal/render"
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

func New(root string, refs []workspace.TourRef, startStep int) (*Model, error) {
	if len(refs) == 0 {
		return nil, fmt.Errorf("no CodeTours found in %s", root)
	}
	m := &Model{
		root: root, refs: refs, startStep: startStep,
		width: 100, height: 30,
		code:  viewport.New(viewport.WithWidth(60), viewport.WithHeight(24)),
		notes: viewport.New(viewport.WithWidth(39), viewport.WithHeight(24)),
	}
	m.code.FillHeight = true
	m.code.SoftWrap = false
	m.code.SetHorizontalStep(8)
	m.notes.FillHeight = true
	m.notes.SoftWrap = true
	if len(refs) == 1 {
		if err := m.startTourAt(0, startStep); err != nil {
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

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
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
				_ = m.startTourAt(i, 1)
				return
			}
		}
	} else if next := nextNumberedTour(m.tour.Title, m.refs); next >= 0 {
		_ = m.startTourAt(next, 1)
		return
	}
	m.screen = screenFinished
}

func (m *Model) previousStep() {
	if m.stepIndex > 0 {
		m.stepIndex--
		m.prepareStep()
	}
}

func (m *Model) prepareStep() {
	step := m.tour.Steps[m.stepIndex]
	if step.File != "" && step.Line == 0 && step.Pattern == "" {
		if marker := stepMarkerPrefix(m.tour.Title, m.tour.StepMarker); marker != "" {
			step.Pattern = regexp.QuoteMeta(marker + "." + strconv.Itoa(m.stepIndex+1))
		}
	}
	resolved, err := resolver.Resolve(m.root, step)
	m.resolved = resolved
	m.stepError = ""
	if err != nil {
		m.stepError = err.Error()
	}

	m.hasCode = resolved.Kind == resolver.File || resolved.Kind == resolver.Directory
	m.codeTitle = resolved.DisplayPath
	if m.codeTitle == "" && m.hasCode {
		if resolved.Path != "" {
			m.codeTitle = filepath.Base(resolved.Path)
		} else {
			m.codeTitle = "Embedded source"
		}
	}

	content := render.TerminalText(resolved.Source)
	if resolved.Kind == resolver.File && content != "" {
		if highlighted, highlightErr := render.Source(resolved.Path, content); highlightErr == nil {
			content = highlighted
		} else if m.stepError == "" {
			m.stepError = "syntax highlighting: " + highlightErr.Error()
		}
	}
	if m.stepError != "" && content == "" {
		content = errorStyle.Render("Unable to display this step\n\n" + render.TerminalText(m.stepError))
	}
	m.codeSource = content

	m.notesRaw = step.Description
	if resolved.Kind == resolver.URI {
		m.notesRaw += "\n\n---\n\n**URI:** " + resolved.DisplayPath
	}
	if step.View != "" {
		m.notesRaw += "\n\n> **Terminal note:** VS Code view `" + step.View + "` is not available here."
	}
	if len(step.Commands) > 0 {
		m.notesRaw += "\n\n> **Safety:** This step contains commands. They were not executed."
	}
	if m.tourWarning != "" {
		m.notesRaw += "\n\n> **Tour warning:** " + m.tourWarning
	}
	if m.stepError != "" && content != "" {
		m.notesRaw += "\n\n> **Step warning:** " + m.stepError
	}
	m.renderNotes()
	m.configureCodeViewport()
	if m.hasCode {
		m.focus = focusCode
	} else {
		m.focus = focusNotes
	}
	if resolved.TargetLine > 0 {
		m.centerTarget()
	}
}

func (m *Model) configureCodeViewport() {
	target := m.resolved.TargetLine
	selectionStart := m.resolved.SelectionStart
	selectionEnd := m.resolved.SelectionEnd
	if m.resolved.Kind == resolver.File {
		digits := m.sourceLineDigits()
		m.code.LeftGutterFunc = func(info viewport.GutterContext) string {
			sourceLine := m.sourceLineAt(info.Index)
			if sourceLine == 0 {
				return strings.Repeat(" ", digits+3)
			}
			if info.Soft {
				return strings.Repeat(" ", digits+3)
			}
			marker := " "
			style := gutterStyle
			if sourceLine == target {
				marker = "▶"
				style = activeGutterStyle
			}
			return style.Render(fmt.Sprintf("%s %*d ", marker, digits, sourceLine))
		}
		m.code.StyleLineFunc = func(index int) lipgloss.Style {
			line := m.sourceLineAt(index)
			if selectionStart > 0 && line >= selectionStart && line <= selectionEnd {
				return selectedLineStyle
			}
			if line == target {
				return targetLineStyle
			}
			return lipgloss.NewStyle()
		}
	} else {
		m.code.LeftGutterFunc = nil
		m.code.StyleLineFunc = nil
	}
}

func (m *Model) centerTarget() {
	if m.targetRow < 0 {
		return
	}
	height := m.code.Height()
	targetPosition := height / 2
	if m.noteRows > 0 {
		// Keep the source line immediately before the inserted description in
		// view whenever the terminal is tall enough for the complete block.
		targetPosition = max(targetPosition, min(height-1, m.noteRows+1))
	}
	offset := m.targetRow - targetPosition
	if offset < 0 {
		offset = 0
	}
	m.code.SetYOffset(offset)
}

func (m *Model) renderNotes() {
	width := m.notes.Width() - 4
	if m.hasCode {
		width = m.code.Width() - m.sourceGutterWidth() - 4
	}
	markdown, err := render.Markdown(m.notesRaw, max(20, width))
	if err != nil {
		markdown = m.notesRaw + "\n\n" + errorStyle.Render(err.Error())
	}
	m.notesRender = markdown
	m.notes.SetContent(markdown)
	m.notes.GotoTop()
	m.composeCodeContent()
	m.layoutHeights()
}

func (m *Model) composeCodeContent() {
	if !m.hasCode {
		m.codeLines = nil
		m.targetRow = -1
		m.noteRows = 0
		return
	}

	sourceLines := strings.Split(m.codeSource, "\n")
	insertAt := 0
	if m.resolved.Kind == resolver.File && m.resolved.TargetLine > 0 {
		insertAt = min(len(sourceLines), m.resolved.TargetLine-1)
	}

	description := m.inlineDescription()
	displayLines := make([]string, 0, len(sourceLines)+len(description))
	m.codeLines = make([]int, 0, cap(displayLines))
	m.targetRow = -1
	m.noteRows = len(description)

	appendDescription := func() {
		displayLines = append(displayLines, description...)
		for range description {
			m.codeLines = append(m.codeLines, 0)
		}
	}

	for index, line := range sourceLines {
		if index == insertAt {
			appendDescription()
		}
		displayLines = append(displayLines, line)
		sourceLine := index + 1
		m.codeLines = append(m.codeLines, sourceLine)
		if sourceLine == m.resolved.TargetLine {
			m.targetRow = len(displayLines) - 1
		}
	}
	if insertAt == len(sourceLines) {
		appendDescription()
	}

	m.code.SetContentLines(displayLines)
}

func (m *Model) inlineDescription() []string {
	if strings.TrimSpace(m.notesRender) == "" {
		return nil
	}
	width := max(1, m.code.Width()-m.sourceGutterWidth())
	separator := dividerStyle.Render(strings.Repeat("─", width))
	lines := []string{separator}
	lines = append(lines, strings.Split(m.notesRender, "\n")...)
	lines = append(lines, separator)
	return lines
}

func (m *Model) sourceLineDigits() int {
	total := strings.Count(m.resolved.Source, "\n") + 1
	return int(math.Log10(float64(max(1, total)))) + 1
}

func (m *Model) sourceGutterWidth() int {
	if m.resolved.Kind != resolver.File {
		return 0
	}
	return m.sourceLineDigits() + 3
}

func (m *Model) sourceLineAt(displayRow int) int {
	if displayRow < 0 || displayRow >= len(m.codeLines) {
		return 0
	}
	return m.codeLines[displayRow]
}

func (m *Model) resize(width, height int) {
	m.width = max(30, width)
	m.height = max(8, height)
	m.compact = m.width < 100
	m.code.SetWidth(m.width)
	m.notes.SetWidth(m.width)
	if m.tour != nil {
		m.renderNotes()
		m.centerTarget()
	} else {
		m.layoutHeights()
	}
}

func (m *Model) layoutHeights() {
	bodyHeight := max(3, m.height-2)
	if !m.hasCode {
		m.notes.SetHeight(max(2, bodyHeight-1))
		m.code.SetHeight(max(2, bodyHeight-1))
		return
	}

	// File steps are a single source viewport. Their Markdown description is
	// inserted directly before the target source line.
	m.code.SetHeight(max(2, bodyHeight-1))
}

func (m *Model) View() tea.View {
	var content string
	switch {
	case m.showHelp:
		content = m.helpView()
	case m.stepPicker:
		content = m.stepPickerView()
	case m.screen == screenTours:
		content = m.tourPickerView()
	case m.screen == screenFinished:
		content = m.finishedView()
	default:
		content = m.playerView()
	}
	view := tea.NewView(content)
	view.AltScreen = true
	view.WindowTitle = "Tourminal"
	return view
}

func (m *Model) playerView() string {
	headerText := fmt.Sprintf(" Tourminal  %s  •  Step %d of %d ", render.TerminalLine(m.tour.Title), m.stepIndex+1, len(m.tour.Steps))
	header := headerStyle.Width(m.width).Render(ansi.Truncate(headerText, m.width, "…"))
	var body string
	if !m.hasCode {
		body = m.pane("Tour description", m.notes.View(), m.focus == focusNotes, m.width)
	} else {
		body = m.pane(m.codeTitle, m.code.View(), true, m.width)
	}
	footerText := " n next  p previous  g steps  ↑/↓ scroll  ? help  q quit "
	footer := footerStyle.Width(m.width).Render(ansi.Truncate(footerText, m.width, "…"))
	return header + "\n" + body + "\n" + footer
}

func (m *Model) pane(title, content string, active bool, width int) string {
	style := paneTitleStyle
	if active {
		style = activePaneTitleStyle
	}
	title = render.TerminalLine(title)
	title = style.Width(width).Render(ansi.Truncate(" "+title+" ", width, "…"))
	return title + "\n" + content
}

func (m *Model) tourPickerView() string {
	lines := []string{titleStyle.Render("Choose a CodeTour"), ""}
	start, end := visibleWindow(m.tourCursor, len(m.refs), max(1, m.height-6))
	for i := start; i < end; i++ {
		prefix := "  "
		style := normalItemStyle
		if i == m.tourCursor {
			prefix = "▶ "
			style = selectedItemStyle
		}
		primary := ""
		if m.refs[i].Primary {
			primary = "  ★ primary"
		}
		lines = append(lines, style.Render(prefix+render.TerminalLine(m.refs[i].Title)+primary))
		if i == m.tourCursor && m.refs[i].Description != "" {
			lines = append(lines, mutedStyle.Render("    "+render.TerminalText(m.refs[i].Description)))
		}
	}
	if m.stepError != "" {
		lines = append(lines, "", errorStyle.Render(render.TerminalText(m.stepError)))
	}
	lines = append(lines, "", footerStyle.Render(" j/k move  Enter start  ? help  q quit "))
	return frame(m.width, m.height, strings.Join(lines, "\n"))
}

func (m *Model) stepPickerView() string {
	lines := []string{titleStyle.Render(render.TerminalLine(m.tour.Title) + " — Steps"), ""}
	start, end := visibleWindow(m.stepCursor, len(m.tour.Steps), max(1, m.height-5))
	for i := start; i < end; i++ {
		prefix := "  "
		style := normalItemStyle
		if i == m.stepCursor {
			prefix = "▶ "
			style = selectedItemStyle
		}
		current := ""
		if i == m.stepIndex {
			current = "  • current"
		}
		lines = append(lines, style.Render(fmt.Sprintf("%s%d. %s%s", prefix, i+1, render.TerminalLine(m.tour.Steps[i].Label(i+1)), current)))
	}
	lines = append(lines, "", footerStyle.Render(" j/k move  Enter open  Esc return  q quit "))
	return frame(m.width, m.height, strings.Join(lines, "\n"))
}

func (m *Model) helpView() string {
	text := titleStyle.Render("Tourminal help") + `

Tour navigation
  n, ], Space      next step
  p, [             previous step
  g                 choose a step
  r                 reload current step

Step navigation
  ↑/↓, j/k          scroll focused pane
  ←/→, h/l          scroll code horizontally
  PgUp/PgDn         scroll a page

General
  ? or Esc          close help
  q or Ctrl-C       quit

Tour commands are never executed automatically.`
	return frame(m.width, m.height, text)
}

func (m *Model) finishedView() string {
	text := titleStyle.Render("Tour complete") + "\n\n" +
		fmt.Sprintf("You finished %s.\n\n", render.TerminalLine(m.tour.Title)) +
		mutedStyle.Render("p: return to the last step   Enter/q: exit")
	return frame(m.width, m.height, text)
}

func visibleWindow(cursor, total, capacity int) (int, int) {
	start := cursor - capacity/2
	if start < 0 {
		start = 0
	}
	end := min(total, start+capacity)
	if end-start < capacity {
		start = max(0, end-capacity)
	}
	return start, end
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

func frame(width, height int, content string) string {
	return lipgloss.NewStyle().Width(max(1, width-4)).Height(max(1, height-2)).Padding(1, 2).Render(content)
}

var (
	purple               = lipgloss.Color("#7C5CFC")
	cyan                 = lipgloss.Color("#5CCFE6")
	muted                = lipgloss.Color("#8892A6")
	headerStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(purple)
	footerStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#D7DCE2")).Background(lipgloss.Color("#252A34"))
	titleStyle           = lipgloss.NewStyle().Bold(true).Foreground(cyan)
	paneTitleStyle       = lipgloss.NewStyle().Bold(true).Foreground(muted).Background(lipgloss.Color("#1B1E26"))
	activePaneTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#343B4A"))
	dividerStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#3A4150"))
	mutedStyle           = lipgloss.NewStyle().Foreground(muted)
	errorStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
	normalItemStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#D7DCE2"))
	selectedItemStyle    = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	gutterStyle          = lipgloss.NewStyle().Foreground(muted)
	activeGutterStyle    = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	targetLineStyle      = lipgloss.NewStyle().Background(lipgloss.Color("#202A3C"))
	selectedLineStyle    = lipgloss.NewStyle().Background(lipgloss.Color("#253856"))
)
