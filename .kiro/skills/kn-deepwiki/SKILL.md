---
name: kn-deepwiki
description: Generate comprehensive, interactive repository wikis and deep knowledge bases for the project
---

# Deepwiki Workflow (Automated Knowledge Base)

This workflow guides you to autonomously transform a codebase or a specific feature into a structured, highly navigable wiki using Knowns docs. It combines high-level architectural overviews with deep implementation details, flow diagrams, and verifiable code citations.

## Inputs

- A topic, feature area, or request to document (e.g., "authentication flow", "full system onboarding")
- Current project root opened in the agent session

## 🚀 Execution Phase

### Phase 1: Discovery & Analysis

- **Analyze**: Scan the repository/feature to detect project type, domain layers, tech stack, and key components.
- Use `mcp__knowns__search` and `mcp__knowns__code` to trace actual code paths.
- Identify the main entry points, primary modules, and data flows.

### Phase 2: Content Generation

- Generate a single `wiki/{slug}.md` file containing all the knowledge.
- **Structure** (use these headings in order):
  1. **Overview** — System purpose, key insights, target users
  2. **Architecture** — High-level design with Mermaid diagram (graph TD or similar)
  3. **Data Flow** — Key execution paths with Mermaid sequence diagram
  4. **Key Components** — Module-by-module breakdown with file citations
  5. **Patterns & Conventions** — Important patterns discovered
- **Traceability**: Every technical claim must cite actual files using `[filename:line_number](path)` format.
- **Visuals**: Include at least 1 Mermaid diagram (architecture or sequence).
- **Tone**: Professional, technical, concise, structured.

### Phase 3: Validation

- Verify all file citations are accurate and files exist.
- Ensure Mermaid diagrams have valid syntax.
- Check the doc renders correctly.

### Phase 4: Delivery

- Present the generated doc path to the user.
- Offer to deep-dive into specific components.

## 🔴 Critical Constraints

1. **Single File**: Create ONLY ONE `wiki/{slug}.md` file per invocation. No subdirectories.
2. **Verifiable Depth**: Every technical claim MUST have a source. No hand-waving.
3. **First Principles**: Always explain _WHY_ before _WHAT_.
4. **Read-Only**: Do NOT modify source code; only create wiki docs.
5. **Use Knowns MCP tools**: `mcp__knowns__docs` for create, `mcp__knowns__search`/`mcp__knowns__code` for discovery.

## Tool Usage

```json
// Create single wiki doc
mcp_knowns_docs({ "action": "create", "path": "wiki/{slug}", "title": "...", "content": "..." })

// Search code for tracing
mcp_knowns_search({ "action": "search", "query": "..." })
mcp_knowns_code({ "action": "search", "query": "...", "neighbors": 5 })
mcp_knowns_code({ "action": "deps", "path": "..." })
```

## 📌 Usage Examples

`/kn-deepwiki "Generate a complete onboarding wiki for this repository"`
`/kn-deepwiki "Create a deep technical reference for the internal messaging system"`
`/kn-deepwiki "Document the authentication flow and user session management"`
