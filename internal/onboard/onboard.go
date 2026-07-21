package onboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/arleyS3/phanes-dna/internal/ai"
	"github.com/arleyS3/phanes-dna/internal/dna"
	"github.com/arleyS3/phanes-dna/internal/store"
)

// GenerateOnboardingGuide acts as a Dev Onboarding Mentor, analyzing codebase structure,
// layer rules, and relevant code chunks to answer how to implement features or follow project conventions.
func GenerateOnboardingGuide(ctx context.Context, st *store.Store, prov ai.Provider, projectID int64, topic string) (string, error) {
	if topic == "" {
		topic = "How is this project structured and what are the main coding conventions?"
	}

	// 1. Fetch project & layer rules
	proj, _ := st.GetProject(ctx, projectID)
	projName := "Current Project"
	if proj != nil {
		projName = proj.Name
	}

	rules, _ := st.GetLayerRules(ctx)

	// 2. Perform vector search for relevant code examples
	var queryEmb []float32
	if prov != nil {
		embs, err := prov.Embed(ctx, []string{topic})
		if err == nil && len(embs) > 0 {
			queryEmb = embs[0]
		}
	}

	chunks, _ := st.SearchSimilar(ctx, queryEmb, 3)

	var codeExamples string
	for _, c := range chunks {
		comp := dna.Compress(c.Content, dna.Normal)
		codeExamples += fmt.Sprintf("\n--- Example File: %s (Lines %d-%d) ---\n%s\n", c.FilePath, c.StartLine, c.EndLine, comp.Text)
	}

	var layerRulesSummary string
	for _, r := range rules {
		layerRulesSummary += fmt.Sprintf("- Layer '%s' -> '%s' (Allowed: %v, Severity: %s)\n",
			r.Layer, r.TargetLayer, r.Allowed, r.Severity)
	}
	if layerRulesSummary == "" {
		layerRulesSummary = "Standard layered architecture (Controller -> Service -> Repository)."
	}

	if prov == nil {
		// Offline fallback response
		return fmt.Sprintf("📖 Onboarding Guide for '%s':\nProject: %s\n\nArchitecture Rules:\n%s\nRelevant Examples:\n%s",
			topic, projName, layerRulesSummary, codeExamples), nil
	}

	systemPrompt := `You are the Senior Resident Software Architect and Dev Onboarding Mentor.
Your goal is to onboard a new software developer to the codebase.
Explain step-by-step how to follow project conventions, create new components, or handle the requested topic.
Use concrete patterns and real examples derived from the provided codebase context.
Be direct, clear, and educational.`

	userPrompt := fmt.Sprintf(`Project Name: %s

Topic/Question: %s

Active Architecture Rules:
%s

Codebase Examples:
%s

Provide a structured, step-by-step Developer Onboarding guide explaining how to accomplish this in this project.`, projName, topic, layerRulesSummary, codeExamples)

	resp, err := prov.Analyze(ctx, ai.AnalyzeRequest{
		SystemPrompt: systemPrompt,
		Prompt:       userPrompt,
		Temperature:  0.3,
	})
	if err != nil || resp == nil || resp.Content == "" {
		return fmt.Sprintf("📖 Onboarding Guide for '%s':\nProject: %s\n\nArchitecture Rules:\n%s\nRelevant Examples:\n%s",
			topic, projName, layerRulesSummary, codeExamples), nil
	}

	return strings.TrimSpace(resp.Content), nil
}
