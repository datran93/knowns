---
name: kn-deepwiki
description: Generate comprehensive, interactive repository wikis and deep knowledge bases for the project
---

# Deepwiki Workflow (Automated Knowledge Base)

This workflow guides you to autonomously transform a codebase or a specific feature into a structured, highly navigable wiki using Knowns docs. It combines high-level architectural overviews with deep implementation details, flow diagrams, and verifiable code citations.

## Inputs

- A topic, feature area, or request to document (e.g., "authentication flow", "full system onboarding")
- Current project root opened in the agent session
- Optional: `--regenerate` to update an existing wiki instead of creating new

---

## 🚀 Execution Phase

### Phase 0: Check for Existing Wiki

Before creating a new wiki, check if one already exists for this topic:

```json
mcp_knowns_search({ "action": "search", "query": "<topic>", "type": "doc" })
```

**If existing wiki found and `--regenerate` passed:**
- Compare existing wiki structure with current code state
- Identify what changed since last generation
- Update only changed pages/sections
- Report: "Regenerated @doc/wiki/<slug> — updated X sections"

**If existing wiki found without `--regenerate`:**
- Report: "Wiki already exists: @doc/wiki/<slug>. Use `--regenerate` to update it."
- Do NOT overwrite without explicit user request

---

### Phase 1: Discovery & Analysis

- **Analyze**: Scan the repository/feature to detect project type, domain layers, tech stack, and key components.
- Use `mcp__knowns__search` and `mcp__knowns__code` to trace actual code paths.
- Identify the main entry points, primary modules, and data flows.
- **External research**: For topics involving external libraries or recent tech, use `mcp__knowns__research` to fetch real-time docs:
  ```json
  mcp_knowns_research({ "action": "search_latest_syntax", "topic": "<tech-stack-topic>" })
  ```

---

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

---

### Phase 3: Validation

- Verify all file citations are accurate and files exist
- Ensure Mermaid diagrams have valid syntax
- Check the doc renders correctly

---

### Phase 4: Completeness Score

After validation, calculate a completeness score:

1. Count total symbols (functions, types, interfaces) discovered in the domain
2. Count how many have documentation in the wiki
3. Report:

```
Wiki Completeness: X/Y symbols documented (Z%)
Gaps:
- @doc/wiki/<slug>#section — <symbol> not documented
- Next: add documentation for remaining symbols
```

If completeness < 60%, flag as "incomplete wiki":
```
⚠️ Wiki is incomplete (Z% documented). Consider:
1. Add more sections for undocumented symbols
2. Break into multiple focused wikis
3. Accept partial documentation for initial version
```

---

### Phase 5: Delivery

- Present the generated doc path to the user
- Include completeness score
- Offer to deep-dive into specific components

---

## Regenerate Mode

When `--regenerate` is passed:

**Step 1:** Compare current code state with existing wiki:
```bash
# Check which files changed since wiki was created
git log --oneline --since="<wiki-creation-date>" -- <domain-path>
```

**Step 2:** Identify changed symbols:
```json
mcp_knowns_code({ "action": "symbols", "path": "<changed-file>" })
```

**Step 3:** Update only affected sections in the wiki. Keep unchanged sections intact.

**Step 4:** Report what changed:
```
Regenerated @doc/wiki/<slug>:
- Updated: Architecture diagram (module X renamed)
- Updated: Data Flow (new endpoint added to Y)
- Unchanged: 4 sections (no code changes detected)
Completeness: X/Y symbols (Z%)
```

---

## 🔴 Critical Constraints

1. **Single File**: Create ONLY ONE `wiki/{slug}.md` file per invocation. No subdirectories.
2. **Verifiable Depth**: Every technical claim MUST have a source. No hand-waving.
3. **First Principles**: Always explain _WHY_ before _WHAT_.
4. **Read-Only**: Do NOT modify source code; only create wiki docs.
5. **Use Knowns MCP tools**: `mcp__knowns__docs` for create/update, `mcp__knowns__search`/`mcp__knowns__code` for discovery.

---

## Tool Usage

```json
// Create single wiki doc
mcp_knowns_docs({ "action": "create", "path": "wiki/{slug}", "title": "...", "content": "..." })

// Update existing wiki (regenerate mode)
mcp_knowns_docs({ "action": "update", "path": "wiki/{slug}", "section": "3", "content": "..." })

// Search code for tracing
mcp_knowns_search({ "action": "search", "query": "..." })
mcp_knowns_code({ "action": "search", "query": "...", "neighbors": 5 })
mcp_knowns_code({ "action": "deps", "path": "..." })
```

---

## 📌 Usage Examples

`/kn-deepwiki "Generate a complete onboarding wiki for this repository"`
`/kn-deepwiki "Create a deep technical reference for the internal messaging system"`
`/kn-deepwiki "Document the authentication flow and user session management"`
`/kn-deepwiki --regenerate "Update the authentication wiki with recent changes"`