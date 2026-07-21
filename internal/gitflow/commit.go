package gitflow

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/arleyS3/phanes-dna/internal/ai"
)

// GenerateConventionalCommit inspects the current Git branch and staged diff,
// then uses the active AI Provider to generate a Conventional Commit in the chosen language.
func GenerateConventionalCommit(ctx context.Context, prov ai.Provider, lang string) (string, error) {
	if lang == "" || lang == "auto" {
		lang = os.Getenv("PHANES_LANGUAGE")
	}
	if lang == "" {
		lang = "en"
	}

	langDesc := "English"
	switch strings.ToLower(lang) {
	case "es", "spanish", "español":
		langDesc = "Spanish"
	case "pt", "portuguese":
		langDesc = "Portuguese"
	}

	// 1. Get current branch name
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	branch := strings.TrimSpace(string(branchOut))
	if err != nil || branch == "" {
		branch = "main"
	}

	// 2. Get staged diff
	diffCmd := exec.Command("git", "diff", "--cached")
	var diffOut bytes.Buffer
	diffCmd.Stdout = &diffOut
	if err := diffCmd.Run(); err != nil || diffOut.Len() == 0 {
		return "", fmt.Errorf("no staged changes found (run 'git add' first)")
	}

	diffText := diffOut.String()
	if len(diffText) > 2000 {
		diffText = diffText[:2000] + "\n...[diff truncated]"
	}

	// Determine scope/prefix recommendation from GitFlow branch
	prefixHint := "feat"
	if strings.HasPrefix(branch, "bugfix/") || strings.HasPrefix(branch, "fix/") {
		prefixHint = "fix"
	} else if strings.HasPrefix(branch, "hotfix/") {
		prefixHint = "hotfix"
	} else if strings.HasPrefix(branch, "docs/") {
		prefixHint = "docs"
	} else if strings.HasPrefix(branch, "refactor/") {
		prefixHint = "refactor"
	}

	if prov == nil {
		// Fallback commit message if no AI provider active
		return fmt.Sprintf("%s: update codebase on %s", prefixHint, branch), nil
	}

	prompt := fmt.Sprintf(`Given the active GitFlow branch '%s' and the following staged git diff, generate a single line Conventional Commit message in %s (format: <type>(<scope>): <short description>).
Do not include markdown or explanations. Return ONLY the single line commit string.

Git Diff:
%s`, branch, langDesc, diffText)

	resp, err := prov.Analyze(ctx, ai.AnalyzeRequest{
		Prompt:      prompt,
		Temperature: 0.2,
	})
	if err != nil {
		return fmt.Sprintf("%s: update codebase on %s", prefixHint, branch), nil
	}

	commitMsg := strings.TrimSpace(resp.Content)
	commitMsg = strings.TrimPrefix(commitMsg, "`")
	commitMsg = strings.TrimSuffix(commitMsg, "`")

	return commitMsg, nil
}
