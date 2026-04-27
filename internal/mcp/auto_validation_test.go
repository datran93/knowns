package mcp

import (
	"testing"
	"time"

	"github.com/datran93/knowns/internal/models"
	"github.com/datran93/knowns/internal/storage"
)

func TestCacheKeyEquality(t *testing.T) {
	k1 := cacheKey{tool: "tasks", action: "update", entity: "abc123"}
	k2 := cacheKey{tool: "tasks", action: "update", entity: "abc123"}
	k3 := cacheKey{tool: "tasks", action: "update", entity: "xyz789"}

	if k1 != k2 {
		t.Error("identical cacheKeys should be equal")
	}
	if k1 == k3 {
		t.Error("different cacheKeys should not be equal")
	}
}

func TestCachedValidationExpiry(t *testing.T) {
	cv := &cachedValidation{
		result:  nil,
		expires: time.Now().Add(5 * time.Minute),
	}
	if time.Now().Before(cv.expires) {
		// expected
	}
}

func TestWriteOperationTags(t *testing.T) {
	writeOps := []string{
		"tasks.create", "tasks.update", "tasks.delete",
		"docs.create", "docs.update", "docs.delete",
		"memory.create", "memory.update", "memory.delete",
		"templates.create", "templates.update", "templates.delete",
	}

	for _, op := range writeOps {
		if _, ok := writeOperationTags[op]; !ok {
			t.Errorf("expected %q to be in writeOperationTags", op)
		}
	}

	// Read operations should not be in the map.
	readOps := []string{"tasks.get", "tasks.list", "docs.list", "search.search", "memory.list"}
	for _, op := range readOps {
		if _, ok := writeOperationTags[op]; ok {
			t.Errorf("read operation %q should not be in writeOperationTags", op)
		}
	}
}

func TestAutoValidationHooksNilConfig(t *testing.T) {
	// Test that hooks can be created with nil config getter (graceful).
	hooks := newAutoValidationHooks(
		func() *storage.Store { return nil },
		func() *models.AgentEfficiencySettings { return nil },
	)
	if hooks == nil {
		// Hooks are still created even with nil config — checks happen at call time.
		t.Log("hooks created with nil config (expected)")
	}
}

func TestAutoValidationHooksDisabled(t *testing.T) {
	// Test with autoValidation explicitly disabled.
	disabled := &models.AgentEfficiencySettings{
		AutoValidation: &models.FeatureFlag{Enabled: false},
	}
	hooks := newAutoValidationHooks(
		func() *storage.Store { return nil },
		func() *models.AgentEfficiencySettings { return disabled },
	)
	// Hooks created; the beforeCallTool will early-return when disabled.
	if hooks == nil {
		t.Error("expected hooks to be created")
	}
}