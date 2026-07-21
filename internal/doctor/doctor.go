package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/arley/phanes-dna/internal/ai"
	"github.com/arley/phanes-dna/internal/store"
)

// RunDoctorChecks performs read-only system health checks.
func RunDoctorChecks() {
	fmt.Println("🏥 Running Phanes DNA Ecosystem Diagnostics...")
	fmt.Println("--------------------------------------------------")

	// 1. Go Toolchain Check
	goPath, err := exec.LookPath("go")
	if err == nil {
		out, _ := exec.Command(goPath, "version").Output()
		fmt.Printf("  ✅ Go Toolchain: %s", string(out))
	} else {
		fmt.Println("  ⚠️ Go Toolchain: Not found in PATH (binary distribution mode active)")
	}

	// 2. Database Capability Check
	tmpDb := filepath.Join(os.TempDir(), "phanes-doctor-test.db")
	st, err := store.NewStore(tmpDb)
	_ = os.Remove(tmpDb)
	if err == nil {
		st.Close()
		fmt.Println("  ✅ SQLite Engine: Pure Go SQLite driver working properly")
	} else {
		fmt.Printf("  ❌ SQLite Engine Error: %v\n", err)
	}

	// 3. AI Provider Auto-Detection Check
	_, provName, err := ai.AutoDetectProvider(ai.Config{})
	if err == nil {
		fmt.Printf("  ✅ AI Provider: Active (%s)\n", provName)
	} else {
		fmt.Printf("  ⚠️ AI Provider Warning: %v\n", err)
	}

	// 4. Git Repository & Hooks Check
	cwd, _ := os.Getwd()
	gitDir := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		preCommit := filepath.Join(gitDir, "hooks", "pre-commit")
		if _, err := os.Stat(preCommit); err == nil {
			fmt.Println("  ✅ Git Hooks: Pre-commit hook installed and active")
		} else {
			fmt.Println("  ℹ️ Git Hooks: Repository detected, but pre-commit hook not installed yet (run 'phanes-dna hooks install')")
		}
	} else {
		fmt.Println("  ℹ️ Git Status: Current working directory is not a Git repository root")
	}

	// 5. Engram Reachability Check
	engramPath, err := exec.LookPath("engram")
	if err == nil {
		fmt.Printf("  ✅ Engram Memory: Installed at %s\n", engramPath)
	} else {
		fmt.Println("  ℹ️ Engram Memory: Not installed in PATH (optional persistent memory)")
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("🎉 Diagnostics complete.")
}
