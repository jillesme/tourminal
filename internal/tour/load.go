package tour

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const maxTourSize = 4 << 20

func Load(path string) (*Tour, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open tour: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxTourSize+1))
	if err != nil {
		return nil, fmt.Errorf("read tour: %w", err)
	}
	if len(data) > maxTourSize {
		return nil, fmt.Errorf("tour is larger than %d MiB", maxTourSize>>20)
	}

	var result Tour
	decoder := json.NewDecoder(newBytesReader(data))
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("parse tour JSON: %w", err)
	}
	if err := ensureSingleJSONValue(decoder); err != nil {
		return nil, err
	}
	result.Path = path
	if diagnostics := Validate(&result); len(diagnostics) > 0 {
		return nil, errors.New(diagnostics[0])
	}
	return &result, nil
}

func Validate(t *Tour) []string {
	var diagnostics []string
	if strings.TrimSpace(t.Title) == "" {
		diagnostics = append(diagnostics, "tour title is required")
	}
	if t.Steps == nil {
		diagnostics = append(diagnostics, "tour steps are required")
	}
	for i, step := range t.Steps {
		prefix := fmt.Sprintf("step %d", i+1)
		if strings.TrimSpace(step.Description) == "" {
			diagnostics = append(diagnostics, prefix+": description is required")
		}
		if step.Line < 0 {
			diagnostics = append(diagnostics, prefix+": line must be 1-based")
		}
		if step.Selection != nil {
			if step.Selection.Start.Line < 1 || step.Selection.End.Line < 1 ||
				step.Selection.Start.Character < 1 || step.Selection.End.Character < 1 {
				diagnostics = append(diagnostics, prefix+": selection positions must be 1-based")
			}
			if step.Selection.Start.Line > step.Selection.End.Line ||
				(step.Selection.Start.Line == step.Selection.End.Line &&
					step.Selection.Start.Character > step.Selection.End.Character) {
				diagnostics = append(diagnostics, prefix+": selection start must not follow its end")
			}
		}
	}
	return diagnostics
}

// Small local reader avoids exposing []byte ownership to json.Decoder callers.
type bytesReader struct {
	b []byte
	i int
}

func newBytesReader(b []byte) *bytesReader { return &bytesReader{b: b} }

func (r *bytesReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}

func ensureSingleJSONValue(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err == io.EOF {
		return nil
	} else if err != nil {
		return fmt.Errorf("parse trailing JSON: %w", err)
	}
	return errors.New("tour file contains more than one JSON value")
}
