package render

import (
	"strings"
	"testing"
)

func TestSourceHighlighting(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	result, err := Source("main.go", "package main\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "\x1b[") || !strings.Contains(result, "package") {
		t.Fatalf("expected highlighted source, got %q", result)
	}
}

func TestMarkdown(t *testing.T) {
	result, err := Markdown("# Hello\n\n* one\n* two", 40)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "Hello") || !strings.Contains(result, "one") {
		t.Fatalf("unexpected markdown: %q", result)
	}
}

func TestTerminalTextRemovesControlSequences(t *testing.T) {
	input := "safe\x1b]52;c;secret\a\rnext\x9b31m\u202e"
	result := TerminalText(input)
	for _, forbidden := range []string{"\x1b", "\a", "\r", "\x9b", "\u202e"} {
		if strings.Contains(result, forbidden) {
			t.Fatalf("result still contains %q: %q", forbidden, result)
		}
	}
	if got := TerminalText("one\r\ntwo\tthree"); got != "one\ntwo\tthree" {
		t.Fatalf("line and tab handling = %q", got)
	}
	if got := TerminalLine("one\ntwo\tthree"); got != "one two three" {
		t.Fatalf("single-line handling = %q", got)
	}
}

func TestSourceSanitizesBeforeHighlighting(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	result, err := Source("main.go", "package main\x1b]0;owned\a\n")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "\x1b") || strings.Contains(result, "\a") {
		t.Fatalf("unsafe source output: %q", result)
	}
}
