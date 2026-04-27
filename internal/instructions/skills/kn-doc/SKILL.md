---
name: kn-doc
description: Use when working with Knowns documentation - viewing, searching, creating, or updating docs
---

# Working with Documentation

**Announce:** "Using kn-doc to work with documentation."

**Core principle:** SEARCH BEFORE CREATING - avoid duplicates.

## Inputs

- Doc path, topic, folder, or task/spec reference
- Whether this is a create, update, or search request
- Optional: `--age` flag to check doc staleness (warn if >6 months old)

## Preflight

- Search before creating
- Prefer section edits for targeted changes
- Preserve doc structure and metadata unless the user asked for a restructure
- Validate refs after doc changes

## Quick Reference

```json
// List docs
mcp_knowns_docs({ "action": "list" })

// View doc (smart mode)
mcp_knowns_docs({ "action": "get", "path": "<path>", "smart": true })

// Search docs
mcp_knowns_search({ "action": "search", "query": "<query>", "type": "doc" })

// Create doc (MUST include description)
mcp_knowns_docs({ "action": "create", "title": "<title>",
  "description": "<brief description of what this doc covers>",
  "tags": ["tag1", "tag2"],
  "folder": "folder"
})

// Update content
mcp_knowns_docs({ "action": "update", "path": "<path>",
  "content": "content"
})

// Update metadata (title, description, tags)
mcp_knowns_docs({ "action": "update", "path": "<path>",
  "title": "New Title",
  "description": "Updated description",
  "tags": ["new", "tags"]
})

// Update section only
mcp_knowns_docs({ "action": "update", "path": "<path>",
  "section": "2",
  "content": "## 2. New Content\n\n..."
})
```

## Creating Documents

1. Search first (avoid duplicates)
2. Choose location:

| Type | Folder |
|------|--------|
| Core | (root) |
| Guide | `guides` |
| Pattern | `patterns` |
| API | `api` |

3. Create with **title + description + tags**
4. Add content
5. **Validate** after creating

**CRITICAL:** Always include `description` - validate will fail without it!

## Updating Documents

**Section edit is most efficient:**
```json
mcp_knowns_docs({ "action": "update", "path": "<path>",
  "section": "3",
  "content": "## 3. New Content\n\n..."
})
```

## Validate After Changes

**CRITICAL:** After creating/updating docs, validate:

```json
// Validate specific doc (saves tokens)
mcp_knowns_validate({ "entity": "<doc-path>" })

// Or validate all docs
mcp_knowns_validate({ "scope": "docs" })
```

If errors found, fix before continuing.

---

## Doc Age Check

When `--age` flag is passed, check for potentially stale docs:

For any doc being updated or referenced:
- Check `info.modified` or `CreatedAt` timestamp
- Docs older than **6 months** without updates → flag as potentially stale:

```
⚠️ Doc age warning: @doc/<path> was last updated YYYY-MM-DD (N months ago).
   Consider reviewing for accuracy.
```

This applies especially to:
- Architecture docs (tech stacks evolve fast)
- Setup/guides (CLI flags, config formats change)
- Pattern docs (best practices evolve)

---

## Diagram Recommendation Heuristic

When creating or updating docs, suggest Mermaid diagrams when:

| Condition | Recommendation |
|-----------|----------------|
| Flow has >3 decision points | Use `graph TD` flowchart |
| Sequence involves >2 actors/services | Use `sequenceDiagram` |
| State transitions exist | Use `stateDiagram-v2` |
| Hierarchical relationships (e.g., module tree) | Use `graph TD` with subgraphs |
| Text description would require >3 paragraphs to explain | Use diagram instead |

**Do NOT force diagrams** — if a simple paragraph explains it clearly, use text.

---

## Mermaid Diagrams

WebUI supports mermaid rendering. Use for:
- Architecture diagrams
- Flowcharts
- Sequence diagrams
- Entity relationships

````markdown
```mermaid
graph TD
    A[Start] --> B{Decision}
    B -->|Yes| C[Action]
    B -->|No| D[End]
```
````

Diagrams render automatically in WebUI preview.

---

## Shared Output Contract

All built-in skills in scope must end with the same user-facing information order: `kn-init`, `kn-spec`, `kn-plan`, `kn-research`, `kn-implement`, `kn-verify`, `kn-doc`, `kn-template`, `kn-extract`, and `kn-commit`.

Required order for the final user-facing response:

1. Goal/result - state what doc was created, updated, inspected, or ruled out.
2. Key details - include the canonical doc path, any important refs added or fixed, validation result, and doc age warnings if applicable.
3. Next action - recommend a concrete follow-up command only when a natural handoff exists.

Keep this concise for CLI use. Documentation-specific content may extend the key-details section, but must not replace or reorder the shared structure.

For `kn-doc`, the key details should cover:

- whether the doc was created, updated, or only inspected
- the canonical doc path
- any important refs added or fixed
- validation result
- doc age warnings if applicable

When doc work naturally leads to another action, include the best next command. If the request ends with inspection or a fully validated update, do not force a handoff.

## Checklist

- [ ] Searched for existing docs
- [ ] Created with **description** (required!)
- [ ] Used section editing for updates
- [ ] Used mermaid for complex flows (optional)
- [ ] **Validated after changes**
- [ ] Doc age checked (if `--age` passed)

## Red Flags

- Creating near-duplicate docs instead of updating an existing one
- Replacing a full doc when only one section needed a change
- Leaving broken refs after an edit
- Updating a stale doc without checking if content is still accurate