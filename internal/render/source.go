package render

import (
	"bytes"
	"os"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func Source(filename, source string) (string, error) {
	source = TerminalText(source)
	if os.Getenv("NO_COLOR") != "" {
		return source, nil
	}
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)
	iterator, err := lexer.Tokenise(nil, source)
	if err != nil {
		return source, err
	}
	style := styles.Get("github-dark")
	if style == nil {
		style = styles.Fallback
	}
	var output bytes.Buffer
	if err := formatters.TTY16m.Format(&output, style, iterator); err != nil {
		return source, err
	}
	return output.String(), nil
}
