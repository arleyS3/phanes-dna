package store

import (
	"context"
	"testing"
	"time"

	"github.com/arleyS3/phanes-dna/internal/dna"
)

func TestProjectRoundTrip(t *testing.T) {
	s, err := NewStore(":memory:")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	now := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)

	p := &dna.Project{
		Name:       "test-project",
		RootPath:   "/tmp/test",
		DetectedAt: now,
	}

	id, err := s.InsertProject(ctx, p)
	if err != nil {
		t.Fatalf("InsertProject: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	got, err := s.GetProject(ctx, id)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}

	if got.Name != p.Name {
		t.Errorf("Name: want %q, got %q", p.Name, got.Name)
	}
	if got.RootPath != p.RootPath {
		t.Errorf("RootPath: want %q, got %q", p.RootPath, got.RootPath)
	}
	if !got.DetectedAt.Equal(now) {
		t.Errorf("DetectedAt: want %v, got %v", now, got.DetectedAt)
	}
}

func TestListProjects(t *testing.T) {
	s, err := NewStore(":memory:")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	for _, name := range []string{"alpha", "beta"} {
		_, err := s.InsertProject(ctx, &dna.Project{
			Name:       name,
			RootPath:   "/tmp/" + name,
			DetectedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("InsertProject(%s): %v", name, err)
		}
	}

	projects, err := s.ListProjects(ctx)
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
}

func TestSourceFileRoundTrip(t *testing.T) {
	s, err := NewStore(":memory:")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	projID, err := s.InsertProject(ctx, &dna.Project{
		Name:       "p",
		RootPath:   "/p",
		DetectedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("InsertProject: %v", err)
	}

	sf := &dna.SourceFile{
		Path:   "src/main/java/com/example/UserController.java",
		Layers: []string{"controller"},
	}
	fileID, err := s.InsertSourceFile(ctx, projID, sf)
	if err != nil {
		t.Fatalf("InsertSourceFile: %v", err)
	}
	if fileID <= 0 {
		t.Fatalf("expected positive fileID, got %d", fileID)
	}

	files, err := s.GetSourceFilesByProject(ctx, projID)
	if err != nil {
		t.Fatalf("GetSourceFilesByProject: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Path != sf.Path {
		t.Errorf("Path: want %q, got %q", sf.Path, files[0].Path)
	}
}
