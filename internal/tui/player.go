package tui

import (
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/jillesme/tourminal/internal/render"
	"github.com/jillesme/tourminal/internal/resolver"
)

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
		options := render.Options{Dark: m.dark, NoColor: m.noColor}
		if highlighted, highlightErr := render.Source(resolved.Path, content, options); highlightErr == nil {
			content = highlighted
		} else if m.stepError == "" {
			m.stepError = "syntax highlighting: " + highlightErr.Error()
		}
	}
	if m.stepError != "" && content == "" {
		content = m.styles.error.Render("Unable to display this step\n\n" + render.TerminalText(m.stepError))
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
			if sourceLine == 0 || info.Soft {
				return strings.Repeat(" ", digits+3)
			}
			marker := " "
			style := m.styles.gutter
			if sourceLine == target {
				marker = "▶"
				style = m.styles.activeGutter
			}
			return style.Render(fmt.Sprintf("%s %*d ", marker, digits, sourceLine))
		}
		m.code.StyleLineFunc = func(index int) lipgloss.Style {
			line := m.sourceLineAt(index)
			if selectionStart > 0 && line >= selectionStart && line <= selectionEnd {
				return m.styles.selectedLine
			}
			if line == target {
				return m.styles.targetLine
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
	notes := m.notesRaw
	if m.stepError != "" {
		notes += "\n\n> **Step warning:** " + m.stepError
	}
	options := render.Options{Dark: m.dark, NoColor: m.noColor}
	markdown, err := render.Markdown(notes, max(20, width), options)
	if err != nil {
		markdown = m.notesRaw + "\n\n" + m.styles.error.Render(err.Error())
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
	separator := m.styles.divider.Render(strings.Repeat("─", width))
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
