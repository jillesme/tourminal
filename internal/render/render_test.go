package render

import (
	"strings"
	"testing"
)

func TestSourceHighlighting(t *testing.T) {
	results := make(map[bool]string)
	for _, dark := range []bool{false, true} {
		result, err := Source("main.go", "package main\n", Options{Dark: dark})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(result, "\x1b[") || !strings.Contains(result, "package") {
			t.Fatalf("expected highlighted source, got %q", result)
		}
		results[dark] = result
	}
	if results[false] == results[true] {
		t.Fatal("light and dark syntax highlighting are identical")
	}
}

func TestMarkdown(t *testing.T) {
	results := make(map[bool]string)
	for _, dark := range []bool{false, true} {
		result, err := Markdown("# Hello\n\n* one\n* two", 40, Options{Dark: dark})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(result, "Hello") || !strings.Contains(result, "one") {
			t.Fatalf("unexpected markdown: %q", result)
		}
		results[dark] = result
	}
	if results[false] == results[true] {
		t.Fatal("light and dark Markdown rendering are identical")
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
	result, err := Source("main.go", "package main\x1b]0;owned\a\n", Options{Dark: true, NoColor: true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "\x1b") || strings.Contains(result, "\a") {
		t.Fatalf("unsafe source output: %q", result)
	}
}

func TestNoColorRendering(t *testing.T) {
	source, err := Source("main.go", "package main\n", Options{NoColor: true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(source, "\x1b[") {
		t.Fatalf("source contains ANSI escapes: %q", source)
	}

	markdown, err := Markdown("# Hello", 40, Options{NoColor: true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(markdown, "\x1b[") {
		t.Fatalf("markdown contains ANSI escapes: %q", markdown)
	}
}
