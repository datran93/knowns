package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/howznguyen/knowns/internal/models"
	"github.com/howznguyen/knowns/internal/storage"
	"github.com/howznguyen/knowns/internal/workingmemory"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterWorkingMemoryTools registers session-scoped working memory tools.
func RegisterWorkingMemoryTools(s *server.MCPServer, getStore func() *storage.Store, getWM func() *workingmemory.Store) {

	// ── add_working_memory ──────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("add_working_memory",
			mcp.WithDescription("Add a working memory entry (session-scoped, cleared when the MCP session ends)."),
			mcp.WithString("content",
				mcp.Required(),
				mcp.Description("Memory content"),
			),
			mcp.WithString("title",
				mcp.Description("Memory title"),
			),
			mcp.WithString("category",
				mcp.Description("Category"),
			),
			mcp.WithArray("tags",
				mcp.Description("Tags"),
				mcp.WithStringItems(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			store := getStore()
			if store == nil {
				return noProjectError()
			}
			wm := getWM()

			content, err := req.RequireString("content")
			if err != nil {
				return errResult("content is required")
			}

			args := req.GetArguments()
			title, _ := stringArg(args, "title")
			category, _ := stringArg(args, "category")
			tags, _ := stringSliceArg(args, "tags")

			entry := wm.Add(&models.MemoryEntry{
				Title:    title,
				Layer:    models.MemoryLayerWorking,
				Category: category,
				Content:  content,
				Tags:     tags,
			})

			out, _ := json.MarshalIndent(entry, "", "  ")
			return mcp.NewToolResultText(string(out)), nil
		},
	)

	// ── get_working_memory ──────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("get_working_memory",
			mcp.WithDescription("Get a session-scoped working memory entry by ID."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Memory entry ID"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			store := getStore()
			if store == nil {
				return noProjectError()
			}
			wm := getWM()

			id, err := req.RequireString("id")
			if err != nil {
				return errResult("id is required")
			}

			entry, ok := wm.Get(id)
			if !ok {
				return errResult("working memory entry not found: " + id)
			}

			out, _ := json.MarshalIndent(entry, "", "  ")
			return mcp.NewToolResultText(string(out)), nil
		},
	)

	// ── list_working_memories ───────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("list_working_memories",
			mcp.WithDescription("List all session-scoped working memory entries."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			store := getStore()
			if store == nil {
				return noProjectError()
			}
			wm := getWM()

			entries := wm.List()
			if entries == nil {
				entries = []*models.MemoryEntry{}
			}

			out, _ := json.MarshalIndent(entries, "", "  ")
			return mcp.NewToolResultText(string(out)), nil
		},
	)

	// ── delete_working_memory ───────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("delete_working_memory",
			mcp.WithDescription("Delete a session-scoped working memory entry."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Memory entry ID"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			store := getStore()
			if store == nil {
				return noProjectError()
			}
			wm := getWM()

			id, err := req.RequireString("id")
			if err != nil {
				return errResult("id is required")
			}

			if !wm.Delete(id) {
				return errResult("working memory entry not found: " + id)
			}

			result := map[string]any{
				"deleted": true,
				"id":      id,
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return mcp.NewToolResultText(string(out)), nil
		},
	)

	// ── clear_working_memory ────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("clear_working_memory",
			mcp.WithDescription("Clear all session-scoped working memory entries."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			store := getStore()
			if store == nil {
				return noProjectError()
			}
			wm := getWM()

			count := wm.Clear()
			result := map[string]any{
				"cleared": count,
				"message": fmt.Sprintf("Cleared %d working memory entries.", count),
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return mcp.NewToolResultText(string(out)), nil
		},
	)
}
