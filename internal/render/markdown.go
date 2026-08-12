package render

import (
	"fmt"
	"strings"

	"charm.land/glamour/v2"
)

// Markdown renders a Markdown description for display in the terminal.
func Markdown(markdown string, width int, options Options) (string, error) {
	markdown = TerminalText(markdown)
	if width < 20 {
		width = 20
	}
	style := "light"
	if options.NoColor {
		style = "notty"
	} else if options.Dark {
		style = "dark"
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
		glamour.WithEmoji(),
	)
	if err != nil {
		return "", fmt.Errorf("create markdown renderer: %w", err)
	}
	result, err := renderer.Render(markdown)
	if err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return strings.TrimSpace(result), nil
}
