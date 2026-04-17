package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/howznguyen/knowns/internal/models"
)

func TestStoreResolveRawReference(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := filepath.Join(t.TempDir(), ".knowns")
	store := NewStore(root)
	if err := store.Init("resolve-test"); err != nil {
		t.Fatalf("init store: %v", err)
	}

	now := time.Now().UTC()
	if err := store.Docs.Create(&models.Doc{
		Path:      "guides/setup",
		Title:     "Setup Guide",
		Tags:      []string{"guide", "semantic"},
		Content:   "# Overview\n\nHello.",
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create doc: %v", err)
	}
	if err := store.Tasks.Create(&models.Task{
		ID:        "rag001",
		Title:     "Implement runtime",
		Status:    "in-progress",
		Priority:  "high",
		Labels:    []string{"semantic", "cli"},
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := store.Memory.Create(&models.MemoryEntry{
		ID:        "mem001",
		Title:     "Semantic note",
		Layer:     models.MemoryLayerProject,
		Category:  "pattern",
		Tags:      []string{"semantic"},
		Content:   "Remember this.",
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create memory: %v", err)
	}

	resolution, err := store.ResolveRawReference("@doc/guides/setup#overview{implements}")
	if err != nil {
		t.Fatalf("resolve doc ref: %v", err)
	}
	if !resolution.Found {
		t.Fatal("expected doc ref to resolve")
	}
	if resolution.Reference.Relation != "implements" {
		t.Fatalf("relation = %q, want implements", resolution.Reference.Relation)
	}
	if resolution.Reference.Fragment == nil || resolution.Reference.Fragment.Heading != "overview" {
		t.Fatalf("expected heading fragment, got %+v", resolution.Reference.Fragment)
	}
	if resolution.Entity == nil || resolution.Entity.Type != "doc" || resolution.Entity.Path != "guides/setup" {
		t.Fatalf("unexpected doc entity: %+v", resolution.Entity)
	}

	taskResolution, err := store.ResolveRawReference("@task-rag001{blocked-by}")
	if err != nil {
		t.Fatalf("resolve task ref: %v", err)
	}
	if !taskResolution.Found || taskResolution.Entity == nil {
		t.Fatal("expected task ref to resolve")
	}
	if taskResolution.Entity.Status != "in-progress" || taskResolution.Entity.Priority != "high" {
		t.Fatalf("unexpected task entity: %+v", taskResolution.Entity)
	}

	memoryResolution, err := store.ResolveRawReference("@memory-mem001")
	if err != nil {
		t.Fatalf("resolve memory ref: %v", err)
	}
	if !memoryResolution.Found || memoryResolution.Entity == nil {
		t.Fatal("expected memory ref to resolve")
	}
	if memoryResolution.Reference.Relation != models.SemanticReferenceRelationReferences {
		t.Fatalf("default relation = %q", memoryResolution.Reference.Relation)
	}
	if memoryResolution.Entity.MemoryLayer != models.MemoryLayerProject {
		t.Fatalf("memory layer = %q", memoryResolution.Entity.MemoryLayer)
	}

	memoryTitleResolution, err := store.ResolveRawReference("@memory-semantic-note{follows}")
	if err != nil {
		t.Fatalf("resolve memory title slug ref: %v", err)
	}
	if !memoryTitleResolution.Found || memoryTitleResolution.Entity == nil {
		t.Fatal("expected memory title slug ref to resolve")
	}
	if memoryTitleResolution.Entity.ID != "mem001" {
		t.Fatalf("memory title slug resolved id = %q, want mem001", memoryTitleResolution.Entity.ID)
	}
	if memoryTitleResolution.Reference.Relation != "follows" {
		t.Fatalf("memory title slug relation = %q, want follows", memoryTitleResolution.Reference.Relation)
	}
}

func TestStoreResolveRawReferenceInvalid(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), ".knowns"))
	if _, err := store.ResolveRawReference("not-a-ref"); err == nil {
		t.Fatal("expected invalid ref error")
	}
}
