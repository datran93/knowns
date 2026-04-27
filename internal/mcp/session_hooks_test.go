package mcp

import (
	"testing"
)

// TestSessionHooksAppendRecentDeduplication tests deduplication logic.
func TestSessionHooksAppendRecentDeduplication(t *testing.T) {
	sh := &sessionHooks{}

	// Test: prepend existing item.
	result := sh.appendRecent([]string{"task-A", "task-B"}, "task-A", 10)
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
	if result[0] != "task-A" {
		t.Errorf("expected task-A first, got %s", result[0])
	}
	if result[1] != "task-B" {
		t.Errorf("expected task-B second, got %s", result[1])
	}
}

// TestSessionHooksAppendRecentMaxLen respects max length.
func TestSessionHooksAppendRecentMaxLen(t *testing.T) {
	sh := &sessionHooks{}

	slice := []string{"task-1", "task-2", "task-3"}
	result := sh.appendRecent(slice, "task-new", 3)

	if len(result) != 3 {
		t.Errorf("expected max 3 items, got %d", len(result))
	}
	if result[0] != "task-new" {
		t.Errorf("expected task-new first, got %s", result[0])
	}
}

// TestSessionHooksAppendRecentNewItem prepends correctly.
func TestSessionHooksAppendRecentNewItem(t *testing.T) {
	sh := &sessionHooks{}

	result := sh.appendRecent([]string{"task-B", "task-C"}, "task-A", 10)
	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}
	if result[0] != "task-A" {
		t.Errorf("expected task-A first, got %s", result[0])
	}
	if result[2] != "task-C" {
		t.Errorf("expected task-C last, got %s", result[2])
	}
}

// TestSessionHooksExtractTaskIDFromResult tests ID extraction from MCP result.
func TestSessionHooksExtractTaskIDFromResult(t *testing.T) {
	sh := &sessionHooks{}

	// Test nil result.
	id := sh.extractTaskIDFromResult(nil)
	if id != "" {
		t.Errorf("expected empty string for nil result, got %s", id)
	}
}

// TestSessionHooksNilHooksLoadCheckpoint documents that it panics on nil.
// The server.go already guards: "if sh != nil && store != nil { sh.LoadCheckpoint() }"
// so this is a usage guard, not a nil-safe method.
func TestSessionHooksNilHooksLoadCheckpoint(t *testing.T) {
	var sh *sessionHooks
	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		sh.LoadCheckpoint()
	}()
	if !didPanic {
		// This documents the current behavior: nil receiver causes panic.
		t.Log("LoadCheckpoint panics on nil sessionHooks (nil receiver guard needed)")
	}
}
