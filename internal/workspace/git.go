package workspace

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ValidateGitRef verifies that a tour's requested ref exists and identifies a
// commit in the workspace repository.
func ValidateGitRef(root, ref string) error {
	if ref == "" || ref == "HEAD" {
		return nil
	}
	if _, err := gitOutput(root, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("git executable not found")
		}
		return fmt.Errorf("git ref %q does not resolve to a commit", ref)
	}
	return nil
}

func GitRefWarning(root, wanted string) string {
	if wanted == "" || wanted == "HEAD" {
		return ""
	}
	branch, branchErr := gitOutput(root, "rev-parse", "--abbrev-ref", "HEAD")
	head, headErr := gitOutput(root, "rev-parse", "HEAD")
	if branchErr != nil || headErr != nil {
		return fmt.Sprintf("This tour targets git ref %q; the current ref could not be verified.", wanted)
	}
	if wanted == branch || wanted == head {
		return ""
	}
	wantedCommit, err := gitOutput(root, "rev-parse", "--verify", "--end-of-options", wanted+"^{commit}")
	if err == nil && wantedCommit == head {
		return ""
	}
	return fmt.Sprintf("This tour targets git ref %q, but the workspace is on %q (%s). Source lines may have drifted.", wanted, branch, shortHash(head))
}

func gitOutput(root string, args ...string) (string, error) {
	commandArgs := append([]string{"-C", root}, args...)
	output, err := exec.Command("git", commandArgs...).Output()
	return strings.TrimSpace(string(output)), err
}

func shortHash(hash string) string {
	if len(hash) > 8 {
		return hash[:8]
	}
	return hash
}
