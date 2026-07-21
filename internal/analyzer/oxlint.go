package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/arleyS3/phanes-dna/internal/dna"
)

type OxlintDiagnostic struct {
	File     string `json:"filename"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	RuleID   string `json:"rule_id"`
	Line     int    `json:"line"`
}

// RunOxlint executes oxlint (Rust-based ultra-fast linter) for JS/TS projects
// and converts its output into Phanes DNA Violations.
func RunOxlint(targetDir string) ([]*dna.Violation, error) {
	oxPath, err := exec.LookPath("oxlint")
	if err != nil {
		// Fallback to npx oxlint
		oxPath, err = exec.LookPath("npx")
		if err != nil {
			return nil, fmt.Errorf("neither oxlint nor npx found in PATH")
		}
	}

	var cmd *exec.Cmd
	if filepathBase(oxPath) == "npx" {
		cmd = exec.Command(oxPath, "-y", "oxlint", "-f", "json", targetDir)
	} else {
		cmd = exec.Command(oxPath, "-f", "json", targetDir)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run() // Oxlint exits non-zero on lint errors

	if out.Len() == 0 {
		return nil, nil
	}

	var diagnostics []OxlintDiagnostic
	if err := json.Unmarshal(out.Bytes(), &diagnostics); err != nil {
		return nil, fmt.Errorf("unmarshal oxlint json: %w", err)
	}

	var violations []*dna.Violation
	for _, d := range diagnostics {
		violations = append(violations, &dna.Violation{
			Rule:     d.RuleID,
			Severity: d.Severity,
			Location: fmt.Sprintf("%s:%d", d.File, d.Line),
			Message:  fmt.Sprintf("Oxlint: %s", d.Message),
		})
	}

	return violations, nil
}

func filepathBase(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}
