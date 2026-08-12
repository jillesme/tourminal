package render

import (
	"bytes"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// Source applies syntax highlighting appropriate for filename.
func Source(filename, source string, options Options) (string, error) {
	source = TerminalText(source)
	if options.NoColor {
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
	styleName := "github"
	if options.Dark {
		styleName = "github-dark"
	}
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}
	var output bytes.Buffer
	if err := formatters.TTY16m.Format(&output, style, iterator); err != nil {
		return source, err
	}
	return output.String(), nil
}
