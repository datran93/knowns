// Package mcp provides the Model Context Protocol server for the Knowns CLI.
package mcp

import (
	"context"
	"sync"
	"time"

	"github.com/datran93/knowns/internal/models"
	"github.com/datran93/knowns/internal/storage"
	"github.com/datran93/knowns/internal/validate"
	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// autoValidationHooks adds non-blocking validation to write operations.
// It runs asynchronously with a 5-minute TTL cache and never blocks the
// primary tool operation.
type autoValidationHooks struct {
	getStore func() *storage.Store
	getCfg   func() *models.AgentEfficiencySettings

	mu    sync.RWMutex
	cache map[cacheKey]*cachedValidation
}

type cacheKey struct {
	tool   string
	action string
	entity string // task ID, doc path, or memory ID
}

type cachedValidation struct {
	result  *validate.Result
	expires time.Time
}

// newAutoValidationHooks creates auto-validation hooks if the feature is enabled.
// Returns nil if autoValidation is disabled in config.
func newAutoValidationHooks(getStore func() *storage.Store, getCfg func() *models.AgentEfficiencySettings) *server.Hooks {
	hooks := &server.Hooks{}
	av := &autoValidationHooks{
		getStore: getStore,
		getCfg:   getCfg,
		cache:    make(map[cacheKey]*cachedValidation),
	}
	hooks.AddBeforeCallTool(av.beforeCallTool)
	return hooks
}

// writeOperationTags maps tool+action to the validation scope needed.
// Only write operations that create or modify entities need validation.
var writeOperationTags = map[string]string{
	// Tasks
	"tasks.create":  "tasks",
	"tasks.update":  "tasks",
	"tasks.delete":  "tasks",
	// Docs
	"docs.create":   "docs",
	"docs.update":   "docs",
	"docs.delete":   "docs",
	// Memory
	"memory.create": "memory",
	"memory.update": "memory",
	"memory.delete": "memory",
	// Templates
	"templates.create": "templates",
	"templates.update": "templates",
	"templates.delete": "templates",
}

// beforeCallTool runs non-blocking validation for write operations.
// It checks the cache first, then fires async validation if not cached.
func (av *autoValidationHooks) beforeCallTool(_ context.Context, id any, req *gomcp.CallToolRequest) {
	cfg := av.getCfg()
	if cfg == nil || !cfg.IsEnabled("autoValidation") {
		return
	}

	toolName := req.Params.Name
	args := req.GetArguments()
	action, _ := args["action"].(string)
	tag := toolName + "." + action

	scope, isWrite := writeOperationTags[tag]
	if !isWrite {
		return
	}

	// Extract entity identifier for cache key.
	var entity string
	switch toolName {
	case "tasks":
		if action == "create" {
			entity = "pending:" + tag // will be updated post-creation
		} else {
			entity, _ = args["taskId"].(string)
		}
	case "docs":
		entity, _ = args["path"].(string)
	case "memory":
		entity, _ = args["id"].(string)
		if entity == "" && action == "create" {
			entity = "pending:" + tag
		}
	case "templates":
		entity, _ = args["name"].(string)
	}

	key := cacheKey{tool: toolName, action: action, entity: entity}

	// Check cache.
	av.mu.RLock()
	cached, ok := av.cache[key]
	av.mu.RUnlock()

	if ok && time.Now().Before(cached.expires) {
		// Cached — log but don't block.
		av.logWarnings(cached.result, toolName, entity)
		return
	}

	// Fire async validation (non-blocking).
	go av.runValidation(tag, scope, entity, key)
}

// runValidation performs the validation asynchronously.
func (av *autoValidationHooks) runValidation(tag, scope, entity string, key cacheKey) {
	store := av.getStore()
	if store == nil {
		return
	}

	result := validate.Run(store, validate.Options{Scope: scope, Entity: entity})
	if len(result.Issues) == 0 {
		return
	}

	// Cache the result (5 min TTL).
	cfg := av.getCfg()
	ttlSeconds := 300
	if cfg != nil && cfg.AutoValidation != nil && cfg.AutoValidation.CacheTTLSeconds > 0 {
		ttlSeconds = cfg.AutoValidation.CacheTTLSeconds
	}

	av.mu.Lock()
	av.cache[key] = &cachedValidation{
		result:  result,
		expires: time.Now().Add(time.Duration(ttlSeconds) * time.Second),
	}
	av.mu.Unlock()

	// Prune expired entries.
	av.mu.Lock()
	now := time.Now()
	for k, v := range av.cache {
		if now.After(v.expires) {
			delete(av.cache, k)
		}
	}
	av.mu.Unlock()

	// Log warnings.
	av.logWarnings(result, tag, entity)
}

// logWarnings emits validation warnings as an info log.
// This is non-blocking and doesn't affect the tool response.
func (av *autoValidationHooks) logWarnings(result *validate.Result, toolName, entity string) {
	if result.WarningCount == 0 {
		return
	}
	// Use mcpLog (the server-level logger) for warnings.
	// The warnings are advisory only — they don't block operations.
	for _, iss := range result.Issues {
		if iss.Level == "warning" {
			mcpLog.Printf("[auto-validate] warning: %s (%s) — %s: %s",
				toolName, entity, iss.Code, iss.Message)
		}
	}
}