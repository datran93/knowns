package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/datran93/knowns/internal/models"
	"github.com/datran93/knowns/internal/search"
	"github.com/datran93/knowns/internal/server/routes"
	"github.com/datran93/knowns/internal/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterCodeTool registers the consolidated code intelligence MCP tool.
func RegisterCodeTool(s *server.MCPServer, getStore func() *storage.Store) {
	s.AddTool(
		mcp.NewTool("code",
			mcp.WithDescription("Code intelligence operations. Use 'action' to specify: search, symbols, deps, graph, trace, impact, callers."),
			mcp.WithString("action",
				mcp.Required(),
				mcp.Description("Action to perform"),
				mcp.Enum("search", "symbols", "deps", "graph", "trace", "impact", "callers"),
			),
			mcp.WithString("symbol",
				mcp.Description("Symbol name (required for trace, impact, callers)"),
			),
			mcp.WithString("query",
				mcp.Description("Search query (required for search)"),
			),
			mcp.WithString("mode",
				mcp.Description("Search mode: hybrid, semantic, or keyword (default: hybrid)"),
				mcp.Enum("hybrid", "semantic", "keyword"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Limit results (default: 10 for search, 100 for symbols, 200 for deps)"),
			),
			mcp.WithNumber("neighbors",
				mcp.Description("Max neighbors per match (default: 5) (search)"),
			),
			mcp.WithString("edgeTypes",
				mcp.Description("Optional comma-separated edge types to expand (search)"),
			),
			mcp.WithString("path",
				mcp.Description("Optional file path filter (symbols, trace, impact)"),
			),
			mcp.WithString("kind",
				mcp.Description("Optional symbol kind filter (symbols)"),
			),
			mcp.WithString("type",
				mcp.Description("Optional edge type filter: calls, contains, has_method, imports, instantiates, implements, extends (deps)"),
			),
			mcp.WithNumber("depth",
				mcp.Description("Max trace depth (default: 10) (trace)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			action, err := req.RequireString("action")
			if err != nil {
				return errResult("action is required")
			}
			switch action {
			case "search":
				return handleCodeSearch(getStore, req)
			case "symbols":
				return handleCodeSymbols(getStore, req)
			case "deps":
				return handleCodeDeps(getStore, req)
			case "graph":
				return handleCodeGraph(getStore, req)
			case "trace":
				return handleCodeTrace(getStore, req)
			case "impact":
				return handleCodeImpact(getStore, req)
			case "callers":
				return handleCodeCallers(getStore, req)
			default:
				return errResultf("unknown code action: %s", action)
			}
		},
	)
}

func handleCodeGraph(getStore func() *storage.Store, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	store := getStore()
	if store == nil {
		return noProjectError()
	}
	nodes, edges := routes.BuildCodeGraph(store)
	if nodes == nil {
		nodes = []routes.GraphNode{}
	}
	if edges == nil {
		edges = []routes.GraphEdge{}
	}
	out, _ := json.MarshalIndent(map[string]any{"nodes": nodes, "edges": edges}, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func handleCodeSymbols(getStore func() *storage.Store, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	store := getStore()
	if store == nil {
		return noProjectError()
	}
	db := store.SemanticDB()
	if db == nil {
		return mcp.NewToolResultText("[]"), nil
	}
	defer db.Close()

	args := req.GetArguments()
	pathFilter, _ := stringArg(args, "path")
	kindFilter, _ := stringArg(args, "kind")
	limit := 100
	if v, ok := intArg(args, "limit"); ok && v > 0 {
		limit = v
	}

	rows, err := db.Query(`SELECT id, doc_path, field, COALESCE(name, ''), COALESCE(signature, '') FROM chunks WHERE type = 'code' AND (? = '' OR doc_path = ?) AND (? = '' OR field = ?) ORDER BY doc_path, name, id LIMIT ?`, pathFilter, pathFilter, kindFilter, kindFilter, limit)
	if err != nil {
		return errFailed("list code symbols", err)
	}
	defer rows.Close()

	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, docPath, kind, name, signature string
		if err := rows.Scan(&id, &docPath, &kind, &name, &signature); err != nil {
			continue
		}
		items = append(items, map[string]any{
			"id":        id,
			"path":      docPath,
			"kind":      kind,
			"name":      name,
			"signature": signature,
		})
	}

	out, _ := json.MarshalIndent(items, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func handleCodeDeps(getStore func() *storage.Store, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	store := getStore()
	if store == nil {
		return noProjectError()
	}
	db := store.SemanticDB()
	if db == nil {
		return mcp.NewToolResultText("[]"), nil
	}
	defer db.Close()

	args := req.GetArguments()
	edgeType, _ := stringArg(args, "type")
	limit := 200
	if v, ok := intArg(args, "limit"); ok && v > 0 {
		limit = v
	}

	rows, err := db.Query(`SELECT from_id, to_id, edge_type, from_path, to_path, raw_target, resolution_status, resolution_confidence FROM code_edges WHERE (? = '' OR edge_type = ?) ORDER BY from_id, edge_type, to_id LIMIT ?`, edgeType, edgeType, limit)
	if err != nil {
		return errFailed("list code dependencies", err)
	}
	defer rows.Close()

	items := make([]map[string]any, 0)
	for rows.Next() {
		var fromID, toID, kind, fromPath, toPath, rawTarget, status, confidence string
		if err := rows.Scan(&fromID, &toID, &kind, &fromPath, &toPath, &rawTarget, &status, &confidence); err != nil {
			continue
		}
		items = append(items, map[string]any{
			"from":       fromID,
			"to":         toID,
			"type":       kind,
			"fromPath":   fromPath,
			"toPath":     toPath,
			"rawTarget":  rawTarget,
			"status":     status,
			"confidence": confidence,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return fmt.Sprint(items[i]["from"], items[i]["type"], items[i]["to"]) < fmt.Sprint(items[j]["from"], items[j]["type"], items[j]["to"])
	})

	out, _ := json.MarshalIndent(items, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func handleCodeSearch(getStore func() *storage.Store, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	store := getStore()
	if store == nil {
		return noProjectError()
	}

	query, err := req.RequireString("query")
	if err != nil {
		return errResult(err.Error())
	}
	args := req.GetArguments()
	mode := "hybrid"
	if v, ok := stringArg(args, "mode"); ok && v != "" {
		mode = v
	}
	limit := 10
	if v, ok := intArg(args, "limit"); ok && v > 0 {
		limit = v
	}
	neighbors := 5
	if v, ok := intArg(args, "neighbors"); ok && v >= 0 {
		neighbors = v
	}
	edgeTypesCSV, _ := stringArg(args, "edgeTypes")

	embedder, vecStore, _ := search.InitSemantic(store)
	if embedder != nil {
		defer embedder.Close()
	}
	if vecStore != nil {
		defer vecStore.Close()
	}

	graph, err := search.SearchCodeWithNeighbors(store, embedder, vecStore, models.RetrievalOptions{
		Query: query,
		Mode:  mode,
		Limit: limit,
	}, splitCSV(edgeTypesCSV), neighbors)
	if err != nil {
		return errFailed("search code", err)
	}

	out, _ := json.MarshalIndent(graph, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// codeTraceNode represents a node in the call graph trace.
type codeTraceNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
	Line int    `json:"line,omitempty"`
}

// codeTraceEdge represents an edge in the call graph trace.
type codeTraceEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

func handleCodeTrace(getStore func() *storage.Store, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	store := getStore()
	if store == nil {
		return noProjectError()
	}
	db := store.SemanticDB()
	if db == nil {
		return mcp.NewToolResultText(`{"nodes": [], "edges": [], "maxDepth": 0}`), nil
	}
	defer db.Close()

	args := req.GetArguments()
	symbol, err := req.RequireString("symbol")
	if err != nil {
		return errResult("symbol is required for trace")
	}
	depth := 10
	if v, ok := intArg(args, "depth"); ok && v > 0 {
		depth = v
	}
	pathFilter, _ := stringArg(args, "path")

	// Find all edges related to this symbol (both as caller and callee)
	pattern := "%" + symbol + "%"
	rows, err := db.Query(`SELECT from_id, to_id, edge_type, from_path, to_path FROM code_edges WHERE from_id LIKE ? OR to_id LIKE ? OR raw_target LIKE ?`, pattern, pattern, pattern)
	if err != nil {
		return errFailed("trace code dependencies", err)
	}
	defer rows.Close()

	// Build adjacency lists for BFS
	type edgeInfo struct {
		from, to   string
		edgeType   string
		fromPath   string
		toPath     string
	}
	var allEdges []edgeInfo
	nodeSet := make(map[string]bool)

	for rows.Next() {
		var fromID, toID, edgeType, fromPath, toPath string
		if err := rows.Scan(&fromID, &toID, &edgeType, &fromPath, &toPath); err != nil {
			continue
		}
		// Apply path filter if provided
		if pathFilter != "" && fromPath != pathFilter && toPath != pathFilter {
			continue
		}
		allEdges = append(allEdges, edgeInfo{from: fromID, to: toID, edgeType: edgeType, fromPath: fromPath, toPath: toPath})
		nodeSet[fromID] = true
		nodeSet[toID] = true
	}

	// BFS from any node containing the symbol to build the trace graph
	visited := make(map[string]bool)
	var traceNodes []codeTraceNode
	var traceEdges []codeTraceEdge
	queue := make([]string, 0)

	// Find starting nodes - any node that matches the symbol pattern
	for nodeID := range nodeSet {
		if strings.Contains(nodeID, symbol) {
			queue = append(queue, nodeID)
			visited[nodeID] = true
		}
	}

	currentDepth := 0
	for len(queue) > 0 && currentDepth < depth {
		nextQueue := make([]string, 0)
		for _, nodeID := range queue {
			// Add node
			traceNodes = append(traceNodes, codeTraceNode{
				ID:   nodeID,
				Name: extractSymbolName(nodeID),
				Type: "function",
				Path: extractDocPath(nodeID),
			})
			// Find edges
			for _, e := range allEdges {
				if e.from == nodeID && !visited[e.to] {
					visited[e.to] = true
					nextQueue = append(nextQueue, e.to)
					traceEdges = append(traceEdges, codeTraceEdge{
						From: e.from,
						To:   e.to,
						Type: e.edgeType,
					})
				}
				if e.to == nodeID && !visited[e.from] {
					// Reverse direction for incoming calls
					visited[e.from] = true
					nextQueue = append(nextQueue, e.from)
					traceEdges = append(traceEdges, codeTraceEdge{
						From: e.from,
						To:   e.to,
						Type: e.edgeType,
					})
				}
			}
		}
		queue = nextQueue
		currentDepth++
	}

	// Deduplicate nodes
	nodeMap := make(map[string]codeTraceNode)
	for _, n := range traceNodes {
		nodeMap[n.ID] = n
	}
	dedupedNodes := make([]codeTraceNode, 0, len(nodeMap))
	for _, n := range nodeMap {
		dedupedNodes = append(dedupedNodes, n)
	}

	out, _ := json.MarshalIndent(map[string]any{
		"nodes":    dedupedNodes,
		"edges":    traceEdges,
		"maxDepth": depth,
	}, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func handleCodeImpact(getStore func() *storage.Store, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	store := getStore()
	if store == nil {
		return noProjectError()
	}
	db := store.SemanticDB()
	if db == nil {
		return mcp.NewToolResultText(`{"symbol": "", "files": [], "count": 0}`), nil
	}
	defer db.Close()

	symbol, err := req.RequireString("symbol")
	if err != nil {
		return errResult("symbol is required for impact")
	}
	args := req.GetArguments()
	pathFilter, _ := stringArg(args, "path")

	// Find all edges where this symbol is the target (things that depend on it)
	pattern := "%" + symbol + "%"
	rows, err := db.Query(`SELECT from_id, to_id, edge_type, from_path, to_path FROM code_edges WHERE to_id LIKE ? OR raw_target LIKE ?`, pattern, pattern)
	if err != nil {
		return errFailed("analyze code impact", err)
	}
	defer rows.Close()

	// BFS to find all reachable nodes from the symbol
	visited := make(map[string]bool)
	fileSet := make(map[string]bool)
	var queue []string

	// Seed with direct dependencies
	for rows.Next() {
		var fromID, toID, edgeType, fromPath, toPath string
		if err := rows.Scan(&fromID, &toID, &edgeType, &fromPath, &toPath); err != nil {
			continue
		}
		if pathFilter != "" && fromPath != pathFilter && toPath != pathFilter {
			continue
		}
		if strings.Contains(toID, symbol) || strings.Contains(fromID, symbol) {
			if !visited[fromID] {
				visited[fromID] = true
				queue = append(queue, fromID)
				fileSet[fromPath] = true
			}
		}
	}

	// BFS for transitive dependencies
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		rows2, err := db.Query(`SELECT from_id, to_id, edge_type, from_path, to_path FROM code_edges WHERE from_id = ?`, current)
		if err != nil {
			continue
		}
		for rows2.Next() {
			var fromID, toID, edgeType, fromPath, toPath string
			if err := rows2.Scan(&fromID, &toID, &edgeType, &fromPath, &toPath); err != nil {
				continue
			}
			if pathFilter != "" && fromPath != pathFilter && toPath != pathFilter {
				continue
			}
			if !visited[toID] {
				visited[toID] = true
				queue = append(queue, toID)
				fileSet[toPath] = true
			}
		}
		rows2.Close()
	}

	files := make([]string, 0, len(fileSet))
	for f := range fileSet {
		files = append(files, f)
	}
	sort.Strings(files)

	out, _ := json.MarshalIndent(map[string]any{
		"symbol": symbol,
		"files":  files,
		"count":  len(files),
	}, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func handleCodeCallers(getStore func() *storage.Store, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	store := getStore()
	if store == nil {
		return noProjectError()
	}
	db := store.SemanticDB()
	if db == nil {
		return mcp.NewToolResultText(`{"symbol": "", "callers": []}`), nil
	}
	defer db.Close()

	symbol, err := req.RequireString("symbol")
	if err != nil {
		return errResult("symbol is required for callers")
	}

	// Find all edges where this symbol is the target (incoming calls)
	pattern := "%" + symbol + "%"
	rows, err := db.Query(`SELECT from_id, to_id, edge_type, from_path, to_path FROM code_edges WHERE to_id LIKE ? OR raw_target LIKE ?`, pattern, pattern)
	if err != nil {
		return errFailed("find code callers", err)
	}
	defer rows.Close()

	type caller struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Path string `json:"path"`
		Line int    `json:"line,omitempty"`
		Type string `json:"type"`
	}
	var callers []caller

	for rows.Next() {
		var fromID, toID, edgeType, fromPath, toPath string
		if err := rows.Scan(&fromID, &toID, &edgeType, &fromPath, &toPath); err != nil {
			continue
		}
		// Only include if to_id actually matches the symbol (not just raw_target)
		if !strings.Contains(toID, symbol) {
			continue
		}
		callers = append(callers, caller{
			ID:   fromID,
			Name: extractSymbolName(fromID),
			Path: fromPath,
			Type: edgeType,
		})
	}

	out, _ := json.MarshalIndent(map[string]any{
		"symbol":  symbol,
		"callers": callers,
	}, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

// extractSymbolName extracts symbol name from a code chunk ID like "code::path/file.go::funcName".
func extractSymbolName(id string) string {
	parts := strings.SplitN(id, "::", 3)
	if len(parts) != 3 {
		return id
	}
	if parts[2] == "__file__" {
		return filepath.Base(parts[1])
	}
	return parts[2]
}

// extractDocPath extracts the doc path from a code chunk ID.
func extractDocPath(id string) string {
	parts := strings.SplitN(id, "::", 3)
	if len(parts) >= 2 {
		return parts[1]
	}
	return id
}
