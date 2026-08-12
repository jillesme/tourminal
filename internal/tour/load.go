package tour

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const maxTourSize = 4 << 20

// Load parses and structurally validates the CodeTour at path.
func Load(path string) (*Tour, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open tour %q: %w", path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxTourSize+1))
	if err != nil {
		return nil, fmt.Errorf("read tour %q: %w", path, err)
	}
	if len(data) > maxTourSize {
		return nil, fmt.Errorf("tour %q is larger than %d MiB", path, maxTourSize>>20)
	}

	var result Tour
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("parse tour JSON %q: %w", path, err)
	}
	if err := ensureSingleJSONValue(decoder); err != nil {
		return nil, fmt.Errorf("parse tour JSON %q: %w", path, err)
	}
	if diagnostics := validate(&result); len(diagnostics) > 0 {
		return nil, fmt.Errorf("validate tour %q: %s", path, diagnostics[0])
	}
	return &result, nil
}

func validate(t *Tour) []string {
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

func ensureSingleJSONValue(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err == io.EOF {
		return nil
	} else if err != nil {
		return fmt.Errorf("parse trailing JSON: %w", err)
	}
	return errors.New("tour file contains more than one JSON value")
}
