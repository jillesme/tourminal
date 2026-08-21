package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/jillesme/tourminal/internal/render"
)

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

// View renders the current Tourminal screen.
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
	header := m.styles.header.Width(m.width).Render(ansi.Truncate(headerText, m.width, "…"))
	var body string
	if !m.hasCode {
		body = m.pane("Tour description", m.notes.View(), m.focus == focusNotes, m.width)
	} else {
		body = m.pane(m.codeTitle, m.code.View(), true, m.width)
	}
	footerText := " n next  p previous  g steps  ↑/↓ scroll  ? help  q quit "
	footer := m.styles.footer.Width(m.width).Render(ansi.Truncate(footerText, m.width, "…"))
	return header + "\n" + body + "\n" + footer
}

func (m *Model) pane(title, content string, active bool, width int) string {
	style := m.styles.paneTitle
	if active {
		style = m.styles.activePane
	}
	title = render.TerminalLine(title)
	title = style.Width(width).Render(ansi.Truncate(" "+title+" ", width, "…"))
	return title + "\n" + content
}

func (m *Model) tourPickerView() string {
	lines := []string{m.styles.title.Render("Choose a CodeTour"), ""}
	start, end := visibleWindow(m.tourCursor, len(m.refs), max(1, m.height-6))
	for i := start; i < end; i++ {
		prefix := "  "
		style := m.styles.normalItem
		if i == m.tourCursor {
			prefix = "▶ "
			style = m.styles.selectedItem
		}
		primary := ""
		if m.refs[i].Primary {
			primary = "  ★ primary"
		}
		lines = append(lines, style.Render(prefix+render.TerminalLine(m.refs[i].Title)+primary))
		if i == m.tourCursor && m.refs[i].Description != "" {
			lines = append(lines, m.styles.muted.Render("    "+render.TerminalText(m.refs[i].Description)))
		}
	}
	if m.stepError != "" {
		lines = append(lines, "", m.styles.error.Render(render.TerminalText(m.stepError)))
	}
	lines = append(lines, "", m.styles.footer.Render(" j/k move  Enter start  ? help  q quit "))
	return frame(m.width, m.height, strings.Join(lines, "\n"))
}

func (m *Model) stepPickerView() string {
	lines := []string{m.styles.title.Render(render.TerminalLine(m.tour.Title) + " — Steps"), ""}
	start, end := visibleWindow(m.stepCursor, len(m.tour.Steps), max(1, m.height-5))
	for i := start; i < end; i++ {
		prefix := "  "
		style := m.styles.normalItem
		if i == m.stepCursor {
			prefix = "▶ "
			style = m.styles.selectedItem
		}
		current := ""
		if i == m.stepIndex {
			current = "  • current"
		}
		lines = append(lines, style.Render(fmt.Sprintf("%s%d. %s%s", prefix, i+1, render.TerminalLine(m.tour.Steps[i].Label(i+1)), current)))
	}
	lines = append(lines, "", m.styles.footer.Render(" j/k move  Enter open  Esc return  q quit "))
	return frame(m.width, m.height, strings.Join(lines, "\n"))
}

func (m *Model) helpView() string {
	text := m.styles.title.Render("Tourminal help") + `

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
	text := m.styles.title.Render("Tour complete") + "\n\n" +
		fmt.Sprintf("You finished %s.\n\n", render.TerminalLine(m.tour.Title)) +
		m.styles.muted.Render("p: return to the last step   Enter/q: exit")
	return frame(m.width, m.height, text)
}

func visibleWindow(cursor, total, capacity int) (int, int) {
	start := max(cursor-capacity/2, 0)
	end := min(total, start+capacity)
	if end-start < capacity {
		start = max(0, end-capacity)
	}
	return start, end
}

func frame(width, height int, content string) string {
	return lipgloss.NewStyle().Width(max(1, width-4)).Height(max(1, height-2)).Padding(1, 2).Render(content)
}
