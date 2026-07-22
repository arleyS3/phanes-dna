package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arleyS3/phanes-dna/internal/ai"
	"github.com/arleyS3/phanes-dna/internal/analyzer"
	"github.com/arleyS3/phanes-dna/internal/dna"
	"github.com/arleyS3/phanes-dna/internal/doctor"
	"github.com/arleyS3/phanes-dna/internal/gitflow"
	"github.com/arleyS3/phanes-dna/internal/githooks"
	"github.com/arleyS3/phanes-dna/internal/mcp"
	"github.com/arleyS3/phanes-dna/internal/onboard"
	"github.com/arleyS3/phanes-dna/internal/generator"
	"github.com/arleyS3/phanes-dna/internal/setup"
	"github.com/arleyS3/phanes-dna/internal/store"
	"github.com/arleyS3/phanes-dna/internal/tui"
)

func main() {
	dbPath := os.Getenv("PHANES_DB_PATH")
	if dbPath == "" {
		dbPath = "phanes.db"
	}

	st, err := store.NewStore(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening store: %v\n", err)
		os.Exit(1)
	}

	var command string
	if len(os.Args) < 2 {
		command = tui.RunInteractiveMenu()
	} else {
		command = os.Args[1]
	}

	if command == "exit" {
		fmt.Println("👋 Goodbye!")
		os.Exit(0)
	}

	// Serve & doctor handle output themselves
	if command != "serve" && command != "doctor" {
		_, provName, err := ai.AutoDetectProvider(ai.Config{})
		if err != nil {
			fmt.Printf("⚠️ Provider auto-detection warning: %v (continuing with offline mode)\n", err)
		} else {
			fmt.Printf("🤖 Connected to AI Provider: %s\n", provName)
		}
	}

	switch command {
	case "analyze":
		runAnalyze(st)
	case "serve":
		prov, _, _ := ai.AutoDetectProvider(ai.Config{})
		runServe(st, prov)
	case "review":
		runReview(st)
	case "onboard":
		prov, _, _ := ai.AutoDetectProvider(ai.Config{})
		runOnboard(st, prov)
	case "generate":
		runGenerate()
	case "setup":
		runSetup()
	case "setup-rules":
		if err := setup.AskQuestionsAndGenerateRules(); err != nil {
			fmt.Fprintf(os.Stderr, "Rules setup failed: %v\n", err)
			os.Exit(1)
		}
	case "review-commit":
		if len(os.Args) < 3 {
			fmt.Println("Usage: phanes-dna review-commit <msg_file>")
			os.Exit(1)
		}
		rulesPath := filepath.Join(".", "PHANES_RULES.md")
		if err := githooks.ValidateCommitMessageFromFile(os.Args[2], rulesPath); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Phanes DNA: Commit validation error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	case "review-commit-sha":
		if len(os.Args) < 3 {
			fmt.Println("Usage: phanes-dna review-commit-sha <sha>")
			os.Exit(1)
		}
		rulesPath := filepath.Join(".", "PHANES_RULES.md")
		if err := githooks.ValidateCommitSha(os.Args[2], rulesPath); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Phanes DNA: Commit validation error on SHA %s: %v\n", os.Args[2], err)
			os.Exit(1)
		}
		os.Exit(0)
	case "hooks":
		runHooks()
	case "commit":
		prov, _, _ := ai.AutoDetectProvider(ai.Config{})
		runCommit(prov)
	case "doctor":
		doctor.RunDoctorChecks()
	case "tui":
		action := tui.RunInteractiveMenu()
		if action != "exit" {
			os.Args = []string{os.Args[0], action}
			main()
		}
	case "version":
		fmt.Println("phanes-dna v0.5.0 (Dev Onboarding Agent + Multi-Stack)")
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Phanes DNA — Resident Software Architect")
	fmt.Println("\nUsage:")
	fmt.Println("  phanes-dna                   Launch interactive terminal UI menu")
	fmt.Println("  phanes-dna analyze [path]    Analyze multi-stack project (Java, Go, Python, TS/JS)")
	fmt.Println("  phanes-dna serve             Start MCP stdio server")
	fmt.Println("  phanes-dna review [--strict] [--ci] Run architecture compliance review")
	fmt.Println("  phanes-dna onboard [topic]   Dev Onboarding Mentor: ask how to build features & conventions")
	fmt.Println("  phanes-dna commit [--lang]   Generate Conventional Commit from branch & staged diff")
	fmt.Println("  phanes-dna generate <name>   Generate feature scaffolding files according to PHANES_RULES.md")
	fmt.Println("  phanes-dna setup [agent|rules] Install MCP config for agent, or run rules setup questionnaire")
	fmt.Println("  phanes-dna hooks [install|uninstall] [type] Manage Git pre-commit & pre-push hooks")
	fmt.Println("  phanes-dna doctor            Execute environment health & ecosystem diagnostics")
	fmt.Println("  phanes-dna version           Show binary version")
}

func runAnalyze(st *store.Store) {
	targetDir := "."
	if len(os.Args) > 2 {
		targetDir = os.Args[2]
	}

	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid path: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	projName := filepath.Base(absPath)
	projID, err := st.InsertProject(ctx, &dna.Project{
		Name:       projName,
		RootPath:   absPath,
		DetectedAt: time.Now(),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to insert project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔍 Analyzing multi-stack project '%s' at %s...\n", projName, absPath)
	registry := analyzer.NewRegistry()
	filesCount := 0

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		anal := registry.GetAnalyzer(path)
		if anal == nil {
			return nil
		}

		rel, _ := filepath.Rel(absPath, path)
		sf, err := anal.AnalyzeFile(path)
		if err != nil {
			fmt.Printf("  ⚠️ Failed to parse %s: %v\n", rel, err)
			return nil
		}
		sf.Path = rel

		fileID, err := st.InsertSourceFile(ctx, projID, sf)
		if err != nil {
			return err
		}

		// Store AST nodes
		for _, node := range sf.AST {
			_, _ = st.InsertASTNode(ctx, fileID, nil, &node)
		}

		// Store chunk content
		content, _ := os.ReadFile(path)
		chunk := &dna.Chunk{
			FilePath:  rel,
			StartLine: 1,
			EndLine:   len(sf.AST),
			Content:   string(content),
		}

		_, _ = st.InsertChunk(ctx, fileID, chunk)
		filesCount++
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Analysis failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Analysis complete. Indexed %d multi-stack source files into project ID %d.\n", filesCount, projID)
}

func runServe(st *store.Store, prov ai.Provider) {
	ctx := context.Background()
	projects, err := st.ListProjects(ctx)
	var projID int64 = 1
	if err == nil && len(projects) > 0 {
		projID = projects[0].ID
	}

	if err := mcp.ServeStdio(st, prov, projID); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server stopped: %v\n", err)
		os.Exit(1)
	}
}

func runReview(st *store.Store) {
	fs := flag.NewFlagSet("review", flag.ExitOnError)
	strict := fs.Bool("strict", false, "Exit with non-zero code if architecture violations are found")
	ci := fs.Bool("ci", false, "Output GitHub Actions annotations format")
	if len(os.Args) > 2 {
		fs.Parse(os.Args[2:])
	}

	ctx := context.Background()
	violations, err := st.GetViolations(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch violations: %v\n", err)
		os.Exit(1)
	}

	if len(violations) == 0 {
		fmt.Println("✅ No architecture violations found.")
		return
	}

	if *ci {
		fmt.Printf("::group::Phanes DNA Review (%d violations)\n", len(violations))
		for _, v := range violations {
			compMsg := dna.Compress(v.Message, dna.Normal).Text
			fmt.Printf("::error file=%s,title=%s::%s\n", v.Location, v.Rule, compMsg)
		}
		fmt.Println("::endgroup::")
	} else {
		fmt.Printf("🚨 Found %d architecture violation(s):\n", len(violations))
		for i, v := range violations {
			compMsg := dna.Compress(v.Message, dna.Normal).Text
			fmt.Printf(" [%d] %s: %s (%s)\n     Location: %s\n     Caveman Summary: %s\n",
				i+1, v.Rule, v.Message, v.Severity, v.Location, compMsg)
		}
	}

	if *strict || *ci {
		os.Exit(1)
	}
}

func runOnboard(st *store.Store, prov ai.Provider) {
	ctx := context.Background()
	projects, err := st.ListProjects(ctx)
	var projID int64 = 1
	if err == nil && len(projects) > 0 {
		projID = projects[0].ID
	}

	topic := ""
	if len(os.Args) > 2 {
		topic = strings.Join(os.Args[2:], " ")
	} else {
		fmt.Print("❓ Enter your onboarding question or topic (e.g. 'How to add a new endpoint?'): ")
		reader := bufio.NewReader(os.Stdin)
		topic, _ = reader.ReadString('\n')
		topic = strings.TrimSpace(topic)
	}

	fmt.Println("🎓 Generating Dev Onboarding Guide...")
	guide, err := onboard.GenerateOnboardingGuide(ctx, st, prov, projID, topic)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Onboarding guide error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n" + guide)
}



func runSetup() {
	target := "all"
	if len(os.Args) > 2 {
		target = os.Args[2]
	}

	if target == "rules" {
		if err := setup.AskQuestionsAndGenerateRules(); err != nil {
			fmt.Fprintf(os.Stderr, "Rules setup failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("⚙️ Setting up Phanes DNA MCP stdio server for target '%s'...\n", target)
	if err := setup.InstallMCPConfig(target); err != nil {
		fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("🚀 MCP Setup complete. Restart your AI assistant to auto-discover Phanes DNA tools.")
}

func runHooks() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: phanes-dna hooks [install|uninstall] [type]")
		os.Exit(1)
	}

	sub := os.Args[2]
	gitDir, _ := os.Getwd()

	switch sub {
	case "install":
		hookType := "all"
		if len(os.Args) > 3 {
			hookType = os.Args[3]
		}
		if err := githooks.InstallHooks(gitDir, hookType); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to install hooks: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("🚀 Git hooks installed successfully.")
	case "uninstall":
		if err := githooks.UninstallHooks(gitDir); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to uninstall hooks: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Println("Unknown subcommand for hooks. Use install or uninstall.")
		os.Exit(1)
	}
}

func runCommit(prov ai.Provider) {
	fs := flag.NewFlagSet("commit", flag.ExitOnError)
	lang := fs.String("lang", "auto", "Target language for commit message (es, en, pt, auto)")
	if len(os.Args) > 2 {
		fs.Parse(os.Args[2:])
	}

	ctx := context.Background()
	commitMsg, err := gitflow.GenerateConventionalCommit(ctx, prov, *lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating commit message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("💡 Generated Conventional Commit:\n   \033[1;32m%s\033[0m\n", commitMsg)
}

func runGenerate() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: phanes-dna generate <feature_name>")
		os.Exit(1)
	}

	featureName := os.Args[2]
	if err := generator.GenerateScaffolding(featureName); err != nil {
		fmt.Fprintf(os.Stderr, "Scaffolding generation failed: %v\n", err)
		os.Exit(1)
	}
}
