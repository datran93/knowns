package mcp

import (
	"context"
	"sync/atomic"

	"github.com/datran93/knowns/internal/models"
	"github.com/datran93/knowns/internal/storage"
	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// sessionState holds the current session checkpoint data.
// Access is synchronized via atomic operations on the containing struct.
type sessionState struct {
	lastTaskID       string
	recentTasks      []string // last 10 task IDs
	recentDocs       []string // last 10 doc paths
	pendingDecisions []string
	codeGraphSummary string
}

// sessionHooks handles session checkpoint save/restore.
// It runs after task operations to capture state.
type sessionHooks struct {
	getStore func() *storage.Store
	getCfg   func() *models.AgentEfficiencySettings

	state atomic.Value // holds *sessionState
}

// newSessionHooks creates session hooks if sessionResume is enabled.
// Returns the server.Hooks and the internal sessionHooks (for LoadCheckpoint).
// Returns (nil, nil) if the feature is disabled.
func newSessionHooks(getStore func() *storage.Store, getCfg func() *models.AgentEfficiencySettings) (*server.Hooks, *sessionHooks) {
	cfg := getCfg()
	if cfg == nil || !cfg.IsEnabled("sessionResume") {
		return nil, nil
	}

	hooks := &server.Hooks{}
	sh := &sessionHooks{
		getStore: getStore,
		getCfg:   getCfg,
	}
	sh.state.Store(&sessionState{})

	hooks.AddAfterCallTool(sh.afterCallTool)
	return hooks, sh
}

// afterCallTool captures task operation results and saves checkpoints.
// It is non-blocking — checkpoint saves run asynchronously.
func (sh *sessionHooks) afterCallTool(_ context.Context, id any, req *gomcp.CallToolRequest, result any) {
	toolName := req.Params.Name
	args := req.GetArguments()
	action, _ := args["action"].(string)

	// Only track task operations for now.
	if toolName != "tasks" {
		return
	}

	// Capture current state for async save.
	state := sh.getState()
	updated := false

	switch action {
	case "create", "update", "delete":
		// Extract task ID.
		var taskID string
		switch action {
		case "create":
			// For create, the task ID is in the result (JSON with ID field).
			taskID = sh.extractTaskIDFromResult(result)
		case "update", "delete":
			taskID, _ = args["taskId"].(string)
		}

		if taskID != "" {
			state.lastTaskID = taskID
			state.recentTasks = sh.appendRecent(state.recentTasks, taskID, 10)
			updated = true
		}
	}

	if !updated {
		return
	}

	// Save asynchronously.
	go sh.saveCheckpoint(state)
}

// extractTaskIDFromResult extracts the task ID from a create result.
func (sh *sessionHooks) extractTaskIDFromResult(result any) string {
	// The result is a *gomcp.CallToolResult which contains text content.
	if r, ok := result.(*gomcp.CallToolResult); ok {
		for _, c := range r.Content {
			if tc, ok := c.(gomcp.TextContent); ok {
				// Try to parse as JSON and extract "id" field.
				if len(tc.Text) > 0 && tc.Text[0] == '{' {
					// Simple extraction without full JSON parse.
					// Format: "id": "abc123" or "id":'abc123'
					text := tc.Text
					for i := 0; i < len(text)-5; i++ {
						if text[i:i+4] == `"id"` {
							j := i + 4
							// Skip whitespace and colon.
							for j < len(text) && (text[j] == ' ' || text[j] == ':' || text[j] == '"') {
								j++
							}
							start := j
							for j < len(text) && text[j] != '"' && text[j] != '\'' {
								j++
							}
							if j > start {
								return text[start:j]
							}
						}
					}
				}
			}
		}
	}
	return ""
}

// appendRecent appends item to slice, maintaining maxLen items.
func (sh *sessionHooks) appendRecent(slice []string, item string, maxLen int) []string {
	// Remove if already present.
	result := make([]string, 0, maxLen)
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	// Prepend new item.
	result = append([]string{item}, result...)
	// Truncate.
	if len(result) > maxLen {
		result = result[:maxLen]
	}
	return result
}

// getState returns the current session state.
func (sh *sessionHooks) getState() *sessionState {
	return sh.state.Load().(*sessionState)
}

// saveCheckpoint writes the current session state to disk asynchronously.
func (sh *sessionHooks) saveCheckpoint(state *sessionState) {
	store := sh.getStore()
	if store == nil || store.Checkpoints == nil {
		return
	}

	// Get TTL from config.
	ttlHours := 24
	cfg := sh.getCfg()
	if cfg != nil && cfg.SessionResume != nil && cfg.SessionResume.CheckpointTTLHours > 0 {
		ttlHours = cfg.SessionResume.CheckpointTTLHours
	}
	store.Checkpoints.UpdateTTL(ttlHours)

	cp := &storage.SessionCheckpoint{
		ID:              "default",
		LastTaskID:      state.lastTaskID,
		RecentTasks:     state.recentTasks,
		RecentDocs:      state.recentDocs,
		PendingDecisions: state.pendingDecisions,
		CodeGraphSummary: state.codeGraphSummary,
	}

	if err := store.Checkpoints.Save(cp); err != nil {
		mcpLog.Printf("[session] checkpoint save failed: %v", err)
	}
}

// LoadCheckpoint loads the latest checkpoint and restores session state.
// Called on server startup.
func (sh *sessionHooks) LoadCheckpoint() {
	store := sh.getStore()
	if store == nil || store.Checkpoints == nil {
		return
	}

	cp, err := store.Checkpoints.LoadLatest()
	if err != nil || cp == nil {
		return
	}

	state := &sessionState{
		lastTaskID:       cp.LastTaskID,
		recentTasks:      cp.RecentTasks,
		recentDocs:       cp.RecentDocs,
		pendingDecisions: cp.PendingDecisions,
		codeGraphSummary: cp.CodeGraphSummary,
	}
	sh.state.Store(state)

	mcpLog.Printf("[session] checkpoint loaded: lastTask=%s, recentTasks=%d",
		cp.LastTaskID, len(cp.RecentTasks))
}