// Package skills exposes the agent skills bundled with the tourminal binary.
package skills

import (
	_ "embed"
	"strings"
)

//go:embed create-codetour/SKILL.md
var createCodeTourSkill string

//go:embed create-codetour/references/tour-schema.md
var createCodeTourSchema string

// CreateCodeTour returns a self-contained prompt containing the skill and all
// references needed to author a CodeTour without access to this repository.
func CreateCodeTour() string {
	return strings.TrimSpace(createCodeTourSkill) +
		"\n\n---\n\n# Bundled reference: references/tour-schema.md\n\n" +
		strings.TrimSpace(createCodeTourSchema) + "\n"
}
