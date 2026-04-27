---
name: kn-extract
description: Use when extracting reusable patterns, decisions, failures, or knowledge into documentation
---

# Extracting Knowledge (with Compounding)

**Announce:** "Using kn-extract to extract knowledge."

**Core principle:** EXTRACT PATTERNS + DECISIONS + FAILURES → COMPOUND LEARNINGS.

## Inputs

- Usually a completed task ID
- Sometimes a code change, repeated pattern, or recurring support issue
- Optional: `--compound` flag for full 3-category analysis
- Optional: `--consolidate` flag to review and consolidate all existing learnings
- Optional: `--mid-task` flag to trigger extraction during implementation (not just after completion)

## Mode Detection

Check `$ARGUMENTS`:
- Contains `--consolidate` → Go to "Consolidation Mode" section
- Contains `--mid-task` → Go to "Mid-Task Extract Trigger" section
- Contains `--compound` → Run full 3-category analysis
- Otherwise → Continue with normal extraction flow

## Extraction Rules

- Extract patterns, decisions, AND failures — not just code patterns
- Prefer updating an existing doc over creating a duplicate
- Link the extracted knowledge back to the source task or source doc
- Only create a template if the pattern is genuinely reusable for generation
- Promote critical learnings to `learnings/critical-patterns` for future `kn-init` sessions

## Step 1: Identify Source

```json
mcp_knowns_tasks({ "action": "get", "taskId": "$ARGUMENTS" })
```

Look for three categories:

| Category | What to look for |
|----------|-----------------|
| **Patterns** | Reusable code patterns, architecture approaches, integration techniques |
| **Decisions** | Good calls, bad calls, trade-offs, surprises |
| **Failures** | Bugs, wrong assumptions, wasted effort, missing prerequisites |

## Step 2: Confidence Verification

**CRITICAL**: Before extracting, verify the source is genuine and not fabricated.

Verify the task exists and has real implementation:
```json
mcp_knowns_tasks({ "action": "get", "taskId": "<source-task-id>" })
```

For pattern extraction, verify the code actually exists:
```bash
# Check the referenced file and symbol actually exist
grep -n "<pattern-symbol>" <referenced-file>
```

If the task doesn't exist, the pattern isn't in code, or the finding feels "manufactured":
- Do NOT fabricate a finding
- If the task ran smoothly with no surprises, say that explicitly
- Document only what is verifiable

**Rule**: A short learning with 2 genuine entries is better than a long doc with invented ones.

## Step 3: Search for Existing Docs

```json
mcp_knowns_search({ "action": "search", "query": "<pattern/topic>", "type": "doc" })
```

**Don't duplicate.** Update existing docs when possible.

## Step 4: Three-Category Analysis

### 4a. Patterns

Identify reusable patterns:
- Code patterns: new utilities, abstractions worth standardizing
- Architecture patterns: structural decisions that worked
- Process patterns: workflow approaches that saved time

### 4b. Decisions

Identify significant decisions:
- **GOOD_CALL**: decisions that proved correct or saved time
- **BAD_CALL**: decisions that required rework
- **SURPRISE**: things that turned out differently than expected
- **TRADEOFF**: conscious choices where alternatives were considered

### 4c. Failures

Identify failures and wasted effort:
- Bugs introduced and root causes
- Wrong assumptions that required backtracking
- Missing prerequisites discovered mid-execution
- Test gaps that allowed regressions

## Step 5: Create/Update Documentation

### For patterns → pattern doc (same as before)

```json
mcp_knowns_docs({ "action": "create", "title": "Pattern: <Name>",
  "description": "Reusable pattern for <purpose>",
  "tags": ["pattern", "<domain>"],
  "folder": "patterns"
})
```

### For decisions + failures → learning doc

```json
mcp_knowns_docs({ "action": "create", "title": "Learning: <feature/domain>",
  "description": "Learnings from <task/feature>",
  "tags": ["learning", "<domain>"],
  "folder": "learnings",
  "content": "<see template below>"
})
```

**Learning doc template:**

```markdown
## Patterns

### <Pattern Name>
- **What:** <description>
- **When to use:** <applicable conditions>
- **Source:** @task-<id>

## Decisions

### <Decision>
- **Chose:** <what was chosen>
- **Over:** <what was rejected>
- **Tag:** GOOD_CALL / BAD_CALL / SURPRISE / TRADEOFF
- **Outcome:** <how it played out>
- **Recommendation:** <for future work>

## Failures

### <Failure>
- **What went wrong:** <description>
- **Root cause:** <not just symptom>
- **Time lost:** <estimate>
- **Prevention:** <what to do differently>
```

## Step 6: Save to Memory

For each extracted pattern or decision worth quick recall, save a concise memory entry alongside the doc:

```json
mcp_knowns_memory({ "action": "add", "title": "<pattern/decision name>",
  "content": "<2-3 sentence summary>. Full reference: @doc/<path>",
  "layer": "project",
  "category": "<pattern|decision|convention|failure>",
  "tags": ["<domain>"]
})
```

For debug patterns specifically, use `category: "debug"` so future `kn-debug` runs can query by error type:
```json
mcp_knowns_memory({ "action": "add", "title": "<error pattern>",
  "content": "Root cause: <sentence>. Fix: <what resolves it>",
  "layer": "project",
  "category": "debug",
  "tags": ["debug", "<domain>"]
})
```

Memory = fast agent recall in future sessions. Doc = full structured reference.
Do NOT duplicate the entire doc content — store a summary and link to the doc.
Skip this step if the extraction produced nothing generalizable.

## Step 7: Create Template (if code-generatable)

```json
mcp_knowns_templates({ "action": "create", "name": "<pattern-name>",
  "description": "Generate <what>",
  "doc": "patterns/<pattern-name>"
})
```

## Step 8: Promote Critical Learnings

For any finding that meets ALL criteria:
- Affects more than one future feature
- Would cause ≥30 minutes wasted effort if unknown
- Is generalizable, not implementation-specific

Check if critical-patterns doc exists:
```json
mcp_knowns_search({ "action": "search", "query": "critical patterns", "type": "doc", "tag": "learning" })
```

**If exists — append:**
```json
mcp_knowns_docs({ "action": "update", "path": "learnings/critical-patterns",
  "appendContent": "\n\n## [Date] <Learning Title>\n**Category:** pattern / decision / failure\n**Source:** @task-<id>\n**Tags:** [tag1, tag2]\n\n<2-4 sentence summary and what to do differently>\n\n**Full entry:** @doc/learnings/<slug>"
})
```

**If not exists — create:**
```json
mcp_knowns_docs({ "action": "create", "title": "Critical Patterns",
  "description": "Promoted learnings that save the most time. Read at session start.",
  "folder": "learnings",
  "tags": ["learning", "critical"],
  "content": "# Critical Patterns\n\nPromoted learnings from completed work. Read this at the start of every session via `kn-init`. These are lessons that cost the most to learn and save the most by knowing.\n\n---"
})
```

**Calibration:** Do NOT promote everything. If critical-patterns grows past 20-30 entries it becomes noise. Only promote learnings that would have saved ≥30 minutes if known in advance.

## Step 9: Validate

**CRITICAL:** After creating doc/template, validate to catch broken refs:

```json
mcp_knowns_validate({ "entity": "<doc-path>" })
```

If errors found, fix before continuing.

## Step 10: Link Back to Task

```json
mcp_knowns_tasks({ "action": "update", "taskId": "$ARGUMENTS",
  "appendNotes": "📚 Extracted to @doc/<path>"
})
```

---

# Mid-Task Extract Trigger

When `$ARGUMENTS` contains `--mid-task`, trigger extraction during implementation when a significant architectural decision is made.

## When to Trigger

During `kn-implement`, if ANY of these occur:
- A non-obvious design choice was made to solve a problem
- A tradeoff was explicitly weighed and one option was chosen over another
- A bug was discovered and the root cause + fix was non-trivial
- A pattern was identified that could be reused

## How to Execute

1. Pause implementation briefly
2. Run the extract workflow for the finding (Steps 1-10 condensed):
   - Identify the specific pattern/decision/failure (1 finding, not full 3-category)
   - Search for existing doc to avoid duplication
   - Create/update doc
   - Save to memory with `category: "debug"` if it's a bug pattern
   - Continue implementation

## Output Format (condensed)

```
📚 Mid-task extract: <finding summary>
   → @doc/<path> (or updated existing doc)
   Memory saved: <title>
   Continue: /kn-implement <task-id>
```

---

# Consolidation Mode (Dream Lite)

When `$ARGUMENTS` contains `--consolidate`:

**Announce:** "Using kn-extract --consolidate to review and consolidate learnings."

Scan all existing learnings docs, merge duplicates, flag outdated entries, and promote new critical patterns. Run on-demand when the learnings folder feels messy or after a batch of completed work.

## C-Step 1: Priority Queue for Consolidation

Don't read ALL docs blindly. Prioritize by:
- **Staleness**: Docs not updated in >6 months
- **Ref count**: Docs referenced by many other docs/tasks (high impact if wrong)
- **Broken refs**: Docs with existing validation errors

```json
mcp_knowns_docs({ "action": "list", "tag": "learning" })
```

Sort by priority before reading all of them.

## C-Step 2: Scan and Classify

Read docs in priority order. For each, classify as:

### Duplicates
- Two docs covering the same root cause or pattern
- Same advice phrased differently across docs
- → Merge into one, delete the other

### Outdated
- Fix/pattern references code that no longer exists
- Advice contradicts current architecture or conventions
- → Update or mark as outdated with date

### Missing Promotions
- Learning that meets critical criteria (affects multiple features, saves ≥30 min) but isn't in critical-patterns
- → Propose promotion

### Orphaned
- Learning that references a task or doc that no longer exists
- → Fix ref or note the context is lost

## C-Step 3: Apply Changes

For each issue found, present to user:

```
Consolidation findings (priority order):

1. MERGE: "Learning: auth token" + "Learning: JWT refresh" → same root cause
   → Merge into "Learning: auth token handling"?

2. OUTDATED: "Learning: webpack config" — references webpack.config.js which was removed
   → Mark outdated or delete?

3. PROMOTE: "Learning: Go test race conditions" — saved 2h on 3 separate tasks
   → Promote to critical-patterns?

4. ORPHAN: "Learning: SSE reconnect" — references @task-abc123 which doesn't exist
   → Keep content, remove broken ref?

Apply all? (yes / review each / skip)
```

**If "review each":** present one at a time, apply user's choice.
**If "yes":** apply all suggested changes.

## C-Step 4: Report

```
Consolidation complete:
- Merged: X docs
- Updated: X docs
- Promoted: X to critical-patterns
- Orphans fixed: X
- Total learnings: X docs (skipped low-priority: Y)
```

---

## Shared Output Contract

Required order for the final user-facing response:

1. Goal/result — what knowledge was extracted, updated, or intentionally not extracted.
2. Key details — docs created/updated, critical promotions, template status, validation.
3. Next action — recommend a concrete follow-up command only when a natural handoff exists.

For `kn-extract`, the key details should cover:

- what knowledge was extracted (patterns, decisions, failures)
- whether docs were created or updated
- whether critical learnings were promoted
- whether a template was created
- where the canonical knowledge now lives

When the extraction leads to a clear follow-up, include the best next command. If the correct outcome is a no-op or a completed doc update with no obvious continuation, stop after the result and key details.

## No-Op Case

If the work is too specific to generalize, say so explicitly and do not force a new doc.

**Do NOT fabricate findings.** If the task ran smoothly with no surprises, write that. A short learning with 2 genuine entries is better than a long doc with invented ones.

## What to Extract

| Source | Extract As | Template? |
|--------|------------|-----------|
| Code pattern | Pattern doc | ✅ Yes |
| API pattern | Integration guide | ✅ Yes |
| Decision (good/bad) | Learning doc | ❌ No |
| Failure / debugging | Learning doc | ❌ No |
| Error solution | Troubleshooting | ❌ No |
| Security approach | Guidelines | ❌ No |

## Checklist

- [ ] Three categories analyzed (patterns, decisions, failures)
- [ ] Knowledge is generalizable
- [ ] Includes working example (for patterns)
- [ ] Links back to source
- [ ] Critical learnings promoted (if applicable)
- [ ] Template created (if applicable)
- [ ] **Validated (no broken refs)**

## Red Flags

- Only extracting code patterns, ignoring decisions and failures
- Promoting everything as critical (noise kills the learning loop)
- Writing generic learnings ("test more carefully" is worthless)
- Fabricating findings when the task was straightforward
- Not checking existing docs before creating duplicates
- Extracting without verifying the source task/pattern actually exists