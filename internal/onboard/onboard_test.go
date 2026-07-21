package onboard

import (
	"context"
	"testing"

	"github.com/arley/phanes-dna/internal/dna"
	"github.com/arley/phanes-dna/internal/store"
)

func TestGenerateOnboardingGuide_OfflineFallback(t *testing.T) {
	st, err := store.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory store: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	projID, _ := st.InsertProject(ctx, &dna.Project{Name: "SocialPlatform", RootPath: "/tmp/social"})

	guide, err := GenerateOnboardingGuide(ctx, st, nil, projID, "How to add new API endpoint?")
	if err != nil {
		t.Fatalf("GenerateOnboardingGuide failed: %v", err)
	}

	if guide == "" {
		t.Errorf("expected non-empty onboarding guide string")
	}
}
