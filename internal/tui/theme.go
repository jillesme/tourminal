package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

type ThemeMode string

const (
	ThemeAuto  ThemeMode = "auto"
	ThemeLight ThemeMode = "light"
	ThemeDark  ThemeMode = "dark"
)

func ParseThemeMode(value string) (ThemeMode, error) {
	mode := ThemeMode(strings.ToLower(strings.TrimSpace(value)))
	switch mode {
	case ThemeAuto, ThemeLight, ThemeDark:
		return mode, nil
	default:
		return "", fmt.Errorf("theme must be auto, light, or dark")
	}
}

type themeStyles struct {
	header       lipgloss.Style
	footer       lipgloss.Style
	title        lipgloss.Style
	paneTitle    lipgloss.Style
	activePane   lipgloss.Style
	divider      lipgloss.Style
	muted        lipgloss.Style
	error        lipgloss.Style
	normalItem   lipgloss.Style
	selectedItem lipgloss.Style
	gutter       lipgloss.Style
	activeGutter lipgloss.Style
	targetLine   lipgloss.Style
	selectedLine lipgloss.Style
}

func newThemeStyles(dark bool) themeStyles {
	pick := lipgloss.LightDark(dark)
	accent := pick(lipgloss.Color("#006D87"), lipgloss.Color("#5CCFE6"))
	muted := pick(lipgloss.Color("#5B6472"), lipgloss.Color("#8892A6"))

	return themeStyles{
		header: lipgloss.NewStyle().Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(pick(lipgloss.Color("#5B3FD6"), lipgloss.Color("#7C5CFC"))),
		footer: lipgloss.NewStyle().
			Foreground(pick(lipgloss.Color("#20242B"), lipgloss.Color("#D7DCE2"))).
			Background(pick(lipgloss.Color("#E4E7EC"), lipgloss.Color("#252A34"))),
		title: lipgloss.NewStyle().Bold(true).Foreground(accent),
		paneTitle: lipgloss.NewStyle().Bold(true).Foreground(muted).
			Background(pick(lipgloss.Color("#EEF1F5"), lipgloss.Color("#1B1E26"))),
		activePane: lipgloss.NewStyle().Bold(true).
			Foreground(pick(lipgloss.Color("#101828"), lipgloss.Color("#FFFFFF"))).
			Background(pick(lipgloss.Color("#D9DEE7"), lipgloss.Color("#343B4A"))),
		divider:      lipgloss.NewStyle().Foreground(pick(lipgloss.Color("#C5CBD3"), lipgloss.Color("#3A4150"))),
		muted:        lipgloss.NewStyle().Foreground(muted),
		error:        lipgloss.NewStyle().Foreground(pick(lipgloss.Color("#B42318"), lipgloss.Color("#FF6B6B"))).Bold(true),
		normalItem:   lipgloss.NewStyle().Foreground(pick(lipgloss.Color("#20242B"), lipgloss.Color("#D7DCE2"))),
		selectedItem: lipgloss.NewStyle().Foreground(accent).Bold(true),
		gutter:       lipgloss.NewStyle().Foreground(muted),
		activeGutter: lipgloss.NewStyle().Foreground(accent).Bold(true),
		targetLine:   lipgloss.NewStyle().Background(pick(lipgloss.Color("#E8F1FF"), lipgloss.Color("#202A3C"))),
		selectedLine: lipgloss.NewStyle().Background(pick(lipgloss.Color("#D8E8FF"), lipgloss.Color("#253856"))),
	}
}
