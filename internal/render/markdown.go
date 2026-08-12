package render

import (
	"fmt"
	"strings"

	"charm.land/glamour/v2"
)

func Markdown(markdown string, width int) (string, error) {
	markdown = TerminalText(markdown)
	if width < 20 {
		width = 20
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
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
