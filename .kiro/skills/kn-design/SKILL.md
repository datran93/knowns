---
name: kn-design
description: Use when creating a design document from an approved spec before planning — produces architecture decisions, component breakdown, and data flow for a feature
---

# Design from Spec

Systematic design: read spec → gather context → draft design → interactive review → save.

**Announce:** "Using kn-design for spec [name]."

**Core principle:** SPEC APPROVED → GATHER CONTEXT → DRAFT DESIGN → INTERACTIVE REVIEW → SAVE DESIGN DOC.

## When to Use

- User says "design this spec", "create design from spec", or `/kn-design specs/<name>`
- `kn-go` auto-invokes this skill before planning when design doesn't exist
- After a spec is approved and before `/kn-plan` or `/kn-go` execution

## Inputs

- Spec path: `specs/<name>` (from arguments or auto-invoked by kn-go)
- No additional inputs required

---

## Phase 1: Validate Spec

```json
mcp_knowns_docs({ "action": "get", "path": "specs/<name>", "smart": true })
```

**Check:**
- Tags include `approved` — if not, STOP: "Spec not approved. Run `/kn-spec <name>` first."
- Has Acceptance Criteria — if empty, STOP: "Spec has no ACs."

```json
mcp_knowns_validate({ "entity": "specs/<name>" })
```

If validation errors → fix or report before continuing.

---

## Phase 2: Gather Codebase Context

Before drafting, analyze the codebase to ensure design aligns with existing patterns.

### 2.1 Scan Related Code Patterns

Search for existing patterns and architecture docs relevant to the spec:

```json
mcp_knowns_search({ "action": "search", "query": "<spec-keywords>", "type": "doc" })
mcp_knowns_search({ "action": "search", "query": "<spec-keywords>", "type": "memory" })
```

### 2.2 Check Architecture Docs

If the spec references architecture docs:
```json
mcp_knowns_docs({ "action": "get", "path": "<referenced-doc>", "smart": true })
```

### 2.3 Check Existing Skills/Templates

Look at existing skill structure for consistency:
```json
mcp_knowns_templates({ "action": "list" })
```

---

## Phase 3: Draft Design Document

Based on spec requirements and codebase context, draft a design document with three main sections:

### 3.1 Architecture Decisions

Key choices that constrain implementation (the "why" decisions):
- Technology choices and rationale
- Data model decisions
- API design decisions
- Integration patterns

### 3.2 Component Breakdown

What each component does, its responsibilities:
- CLI components (if applicable)
- MCP/tool components
- Storage components
- UI components (if applicable)

### 3.3 Data Flow

How data moves through the system:
- Inputs → Processing → Outputs
- Key workflows
- Error handling paths

---

## Phase 4: Interactive Review

Show the draft design to the user and iterate based on feedback.

### 4.1 Present Draft

Present the draft design with clear sections:

```
Design Draft for specs/<name>
═══════════════════════════════════════

## Architecture Decisions

[Draft content]

## Component Breakdown

[Draft content]

## Data Flow

[Draft content]

─────────────────────────────────────────
Please review. Approve to save, or request changes.
```

### 4.2 Handle Feedback

**If user approves:**
→ Proceed to Phase 5 (Save)

**If user requests changes:**
→ Update draft based on feedback
→ Re-present for review
→ Repeat until approved or cancelled

**If user cancels:**
→ Report: "Design draft cancelled. No design doc saved."
→ STOP

---

## Phase 5: Save Design Document

### 5.1 Create designs/ directory if needed

Designs are stored in `designs/<spec-name>.md` (git-tracked, same as spec docs).

### 5.2 Write Design Document

```markdown
---
title: Design: <spec-name>
description: Design document generated from spec <spec-name>
spec: specs/<spec-name>
createdAt: <timestamp>
---

# Design: <spec-name>

## Architecture Decisions

[Content]

## Component Breakdown

[Content]

## Data Flow

[Content]
```

### 5.3 Verify Saved

```json
mcp_knowns_docs({ "action": "get", "path": "designs/<spec-name>" })
```

Confirm the file exists and is readable.

---

## Phase 6: Report Completion

```
✓ Design approved: designs/<spec-name>.md

Next step — generate tasks from design:

Run: /kn-plan --from @doc/designs/<name>
```

---

## Shared Output Contract

Required order for the final user-facing response:

1. Goal/result — state what design was produced and whether it was saved.
2. Key details — include architecture decisions, component count, data flow summary.
3. Next action — recommend a concrete follow-up command only when a natural handoff exists.

For `kn-design`, the key details should cover:
- spec analyzed
- design decisions extracted
- interactive review outcome (approved/revised/cancelled)
- file saved location

---

## Integration with kn-go

When `kn-go` runs on a spec without an existing design:

1. **Phase 0.5 (Pre-Planning Design Check):**
   ```
   Checking for existing design at designs/<spec-name>.md...
   ```

2. **If design doesn't exist:**
   ```
   No design found. Invoking /kn-design specs/<name>...
   ```
   → Auto-invoke this skill before proceeding to Phase 2 of kn-go

3. **If design exists:**
   ```
   Design found at designs/<spec-name>.md. Skipping design generation.
   ```
   → Continue directly to kn-go planning phase

---

## Related Skills

- `/kn-spec <name>` — Create spec for complex features
- `/kn-plan --from @doc/designs/<name>` — Generate tasks from design

---

## Checklist

- [ ] Spec validated (approved, has ACs)
- [ ] Codebase context gathered (patterns, architecture docs)
- [ ] Draft design produced (Architecture Decisions, Component Breakdown, Data Flow)
- [ ] Interactive review completed (approved by user)
- [ ] Design saved to designs/<spec-name>.md
- [ ] **Next step suggested** (/kn-plan --from @doc/designs/<name>)

## Red Flags

- Creating design from unapproved spec
- Skipping codebase context analysis
- Saving design without user review/approval
- Not suggesting next step after save
- Design doesn't align with existing codebase patterns