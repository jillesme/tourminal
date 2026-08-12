package render

import "strings"

// TerminalText removes control characters that could alter the terminal
// outside Tourminal's own rendering. Newlines and tabs remain available for
// source code and Markdown layout.
func TerminalText(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	return strings.Map(func(r rune) rune {
		switch {
		case r == '\n' || r == '\t':
			return r
		case r < 0x20 || (r >= 0x7f && r <= 0x9f):
			return '\uFFFD'
		case r >= 0x202a && r <= 0x202e:
			return '\uFFFD'
		case r >= 0x2066 && r <= 0x2069:
			return '\uFFFD'
		default:
			return r
		}
	}, value)
}

// TerminalLine applies TerminalText and also prevents untrusted values such as
// filenames and titles from creating additional terminal rows.
func TerminalLine(value string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return ' '
		}
		return r
	}, TerminalText(value))
}
