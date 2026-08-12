package workspace

import (
	"fmt"
	"runtime"
	"strings"
	"unicode"
)

// EvaluateWhen supports CodeTour's documented platform variables without
// evaluating arbitrary JavaScript from a repository.
func EvaluateWhen(expression string) (bool, error) {
	if strings.TrimSpace(expression) == "" {
		return true, nil
	}
	p := whenParser{
		input: expression,
		values: map[string]bool{
			"isLinux":   runtime.GOOS == "linux",
			"isMac":     runtime.GOOS == "darwin",
			"isWindows": runtime.GOOS == "windows",
			"isWeb":     false,
			"true":      true,
			"false":     false,
		},
	}
	result, err := p.parseOr()
	if err != nil {
		return false, err
	}
	p.skipSpace()
	if p.pos != len(p.input) {
		return false, fmt.Errorf("unsupported token near %q", p.input[p.pos:])
	}
	return result, nil
}

type whenParser struct {
	input  string
	pos    int
	values map[string]bool
}

func (p *whenParser) parseOr() (bool, error) {
	left, err := p.parseAnd()
	if err != nil {
		return false, err
	}
	for {
		if !p.match("||") {
			return left, nil
		}
		right, err := p.parseAnd()
		if err != nil {
			return false, err
		}
		left = left || right
	}
}

func (p *whenParser) parseAnd() (bool, error) {
	left, err := p.parseUnary()
	if err != nil {
		return false, err
	}
	for {
		if !p.match("&&") {
			return left, nil
		}
		right, err := p.parseUnary()
		if err != nil {
			return false, err
		}
		left = left && right
	}
}

func (p *whenParser) parseUnary() (bool, error) {
	if p.match("!") {
		value, err := p.parseUnary()
		return !value, err
	}
	return p.parsePrimary()
}

func (p *whenParser) parsePrimary() (bool, error) {
	if p.match("(") {
		value, err := p.parseOr()
		if err != nil {
			return false, err
		}
		if !p.match(")") {
			return false, fmt.Errorf("missing closing parenthesis")
		}
		return value, nil
	}
	p.skipSpace()
	start := p.pos
	for p.pos < len(p.input) {
		r := rune(p.input[p.pos])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		p.pos++
	}
	if start == p.pos {
		return false, fmt.Errorf("expected a platform variable near %q", p.input[p.pos:])
	}
	name := p.input[start:p.pos]
	value, ok := p.values[name]
	if !ok {
		return false, fmt.Errorf("unsupported variable %q", name)
	}
	return value, nil
}

func (p *whenParser) match(token string) bool {
	p.skipSpace()
	if strings.HasPrefix(p.input[p.pos:], token) {
		p.pos += len(token)
		return true
	}
	return false
}

func (p *whenParser) skipSpace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}
