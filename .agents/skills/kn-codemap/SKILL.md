---
name: kn-codemap
description: Generate hierarchical codebase visualizations and execution flow maps with Mermaid diagrams
---

# Codemap Workflow (Codebase Visualization)

This workflow guides you to create a visual representation of the project's architecture, data flows, and module relationships using Knowns docs. It helps onboard new developers and refresh understanding of how the system is structured.

## Inputs

- A visualization request (e.g., "visualize the module dependency graph", "map HTTP request flow")
- Optional: specific entry point or user journey to trace

## 🚀 Execution Phase

### Phase 1: Structural Discovery

- Use `mcp__knowns__code` (`search`, `deps`, `symbols`) to map the high-level project structure and find exact symbol locations.
- Identify the main entry points (e.g., `main.go`, `cmd/`, `internal/`).
- List the primary modules, services, and directories.

### Phase 2: Execution Flow Mapping

- Pick a core user journey (e.g., "User Login" or "CLI Command Execution").
- Trace the call stack across files and services.
- Note the data transformations and external dependencies (DB, file system, third-party APIs).

### Phase 3: Visual Generation (Mermaid)

- Generate a `codemap/{slug}.md` file.
- **MANDATORY**: Include at least one **Architecture Diagram** (Mermaid `graph TD` or `graph LR`).
- **MANDATORY**: Include at least one **Sequence Diagram** (Mermaid `sequenceDiagram`) for a primary execution flow.
- Use subgraphs to group related items (e.g., UI layer, API layer, Service layer, Storage layer).

### Phase 4: Narrative Summary

- Beneath each diagram, provide a narrative explanation of "The Life of a Request".
- Link to actual files using `[filename:line_number](file_path)` format.

### Phase 5: Result Delivery

- Present the generated codemap doc to the user.
- Offer to deep-dive into specific components.

## 🔴 Critical Constraints

1. **Fact-Based**: Do not guess names. Verify every function call and import using code search.
2. **Visual Focus**: The goal is clarity. Use subgraphs in Mermaid to group related items.
3. **No Placeholders**: Every node in the diagram must correspond to a real code entity.
4. **Use Knowns MCP tools**: Use `mcp__knowns__code` for code intelligence, `mcp__knowns__docs` for creating the output doc.

## Tool Usage

```json
// Search code structure
mcp_knowns_code({ "action": "search", "query": "main entry", "neighbors": 5 })
mcp_knowns_code({ "action": "symbols", "path": "internal/cli" })
mcp_knowns_code({ "action": "deps", "path": "cmd/knowns/main.go" })

// Create codemap doc
mcp_knowns_docs({ "action": "create", "path": "codemap/architecture", "title": "Architecture Codemap", "content": "..." })
```

## 📌 Usage Examples

`/kn-codemap "Visualize the entire module dependency graph"`
`/kn-codemap "Map the flow of an incoming HTTP request from route to database"`
`/kn-codemap "Show how CLI commands are structured and routed"`
