---
name: kn-go
description: Use when implementing all tasks from an approved spec in one continuous run without manual review gates
---

# Go Mode — Full Pipeline Execution

Run the entire SDD pipeline from an approved spec: generate tasks → plan each → implement each → verify → commit. No manual review gates between steps.

**Announce:** "Using kn-go for spec [name]."

**Core principle:** SPEC APPROVED → GENERATE TASKS → PLAN → IMPLEMENT ALL → VERIFY → COMMIT.

## When to Use

- User has an approved spec and wants to execute everything in one shot
- User says "run all", "go mode", "execute everything", or similar
- The spec is already approved (tag: `spec`, `approved`)

## When NOT to Use

- Spec is still draft — redirect to `/kn-spec` first
- User wants to review each task individually — use `/kn-plan <id>` + `/kn-implement <id>`
- Spec has unresolved open questions — resolve them first

## Inputs

- Spec path: `specs/<name>` (from `$ARGUMENTS` or ask user)
- Optional: `--dry-run` to preview tasks without executing

## Process

Complete these phases in order. Do not skip phases.

---

### Phase 1: Validate Spec

```json
mcp_knowns_docs({ "action": "get", "path": "specs/<name>", "smart": true })
```

**Check:**
- Tags include `approved` — if not, STOP: "Spec not approved. Run `/kn-spec <name>` first."
- Has Acceptance Criteria — if empty, STOP: "Spec has no ACs."
- No unresolved open questions marked as blocking

```json
mcp_knowns_validate({ "entity": "specs/<name>" })
```

If validation errors → fix or report before continuing.

---

### Phase 2: Generate Tasks

Parse spec for requirements and generate tasks. Same logic as `kn-plan --from @doc/specs/<name>` but **skip the approval gate**.

```json
mcp_knowns_tasks({ "action": "create", "title": "<requirement title>",
  "description": "<from spec>",
  "spec": "specs/<name>",
  "fulfills": ["AC-1", "AC-2"],
  "priority": "medium",
  "labels": ["from-spec", "go-mode"]
})
```

Add implementation ACs per task:
```json
mcp_knowns_tasks({ "action": "update", "taskId": "<id>",
  "addAc": ["Step 1", "Step 2", "Tests"]
})
```

**Report:** "Created X tasks from spec. Starting implementation..."

---

### Phase 3: Plan + Implement Each Task

Loop through all generated tasks in dependency order (foundational first, dependent last).

For each task:

#### 3a. Take ownership + plan

```json
mcp_knowns_tasks({ "action": "update", "taskId": "<id>",
  "status": "in-progress"
})
mcp_knowns_time({ "action": "start", "taskId": "<id>" })
```

- Research context: follow refs, search related docs/memories, check templates
- Use `search` for discovery first. If a task/spec needs assembled execution context, use `mcp_knowns_search({ "action": "retrieve", "query": "<keywords>" })` before drafting or executing the plan. Fall back to CLI `knowns retrieve "<keywords>" --json` if MCP is unavailable.
- Draft and save plan directly (no approval gate)

```json
mcp_knowns_search({ "action": "search", "query": "<task keywords>", "type": "memory" })
```

```json
mcp_knowns_tasks({ "action": "update", "taskId": "<id>",
  "plan": "1. Step one\n2. Step two\n3. Tests"
})
```

#### 3b. Implement

- Work through plan steps
- Check ACs as completed
- Run tests/lint/build after each task

```json
mcp_knowns_tasks({ "action": "update", "taskId": "<id>",
  "checkAc": [1, 2, 3],
  "appendNotes": "Implemented: brief summary"
})
```

#### 3c. Complete task

```json
mcp_knowns_time({ "action": "stop", "taskId": "<id>" })
mcp_knowns_tasks({ "action": "update", "taskId": "<id>",
  "status": "done"
})
```

#### 3d. Quick validate

```json
mcp_knowns_validate({ "entity": "<id>" })
```

If errors → fix before moving to next task.

**Progress report between tasks:**
> "✓ Task X/Y done: [title]. Continuing..."

---

### Phase 4: Full Verification

After all tasks complete:

```json
mcp_knowns_validate({ "scope": "sdd" })
```

**Report SDD coverage:**
```
SDD Coverage Report
═══════════════════
Spec: specs/<name>
Tasks: X/X complete (100%)
ACs: Y/Z verified
```

If coverage < 100% → identify gaps and fix.

Also run project-level checks:
```bash
# Build/test/lint — adapt to project
go build ./...
go test ./...
```

---

### Phase 5: Pre-Flight Commit Safety Check

Before staging anything, verify the working tree state:

```bash
git status --short
```

**If unrelated changes exist** (files modified that are NOT in the spec implementation):

```
⚠️ Working tree has unrelated changes:
  M <unrelated-file-1>
  M <unrelated-file-2>

Cannot safely stage spec changes. Options:
1. Stash unrelated changes: git stash
2. Commit unrelated changes separately
3. Abort (stage manually after cleanup)
```

Do NOT proceed to commit until unrelated changes are resolved.

**If working tree is clean or only has spec-related changes:**

Stage and commit:
```bash
git add -A
git diff --staged --stat
```

---

### Phase 6: Rollback Guidance

If a mid-pipeline error leaves the codebase broken:

**Build/test fails during a task:**
1. Fix the error (compile error, test failure, etc.)
2. Re-run the failing command to confirm fix
3. Continue pipeline

**If unfixable:**
1. Mark task as `blocked`
2. Append notes: `Pipeline blocked: <error summary>`
3. Report at end of pipeline
4. Do NOT proceed to commit with broken code

**Context budget exceeded (context > 60% during implementation):**
1. Finish the current task
2. **Checkpoint**: commit completed work so far
3. Report progress: "Pipeline checkpoint at task X/Y. Remaining: Z tasks."
4. Suggest: "Run `/kn-go specs/<name>` again to continue remaining tasks."

**Rollback note:** If a task fails and you cannot fix it in the current pipeline run:
- Mark task as `blocked` with notes
- Continue to next task if possible
- At end of pipeline, report all blocked tasks
- User can fix blocked tasks manually and re-run `/kn-go specs/<name>`

---

### Phase 7: Commit

Generate commit message:
```
feat(<scope>): implement <spec-name>

- Task 1: <title>
- Task 2: <title>
- ...
- All ACs verified via SDD
```

**This is the ONE gate in go mode — ask user before committing:**

> Pipeline complete. X tasks done, SDD verified.
>
> Working tree status: [clean / has unrelated changes]
>
> Ready to commit:
> ```
> feat(<scope>): implement <spec-name>
> ```
> Proceed? (yes/no/edit)

---

## Context Budget

If context exceeds ~60% during implementation:

1. Finish the current task
2. **Checkpoint**: commit completed work so far
3. Report progress and remaining tasks
4. Suggest: "Run `/kn-go specs/<name>` again to continue remaining tasks."

The skill will detect already-done tasks and skip them on re-run.

---

## Re-run Behavior

When invoked on a spec that already has tasks:

1. List existing tasks linked to the spec
2. Filter to `todo` and `in-progress` only
3. Skip `done` tasks
4. Continue from where it left off

```json
mcp_knowns_tasks({ "action": "list", "spec": "specs/<name>" })
```

---

## Error Handling Summary

| Situation | Action |
|-----------|--------|
| Build/test fails | Fix → re-run → continue |
| Unfixable error | Mark blocked → append notes → continue |
| Spec not approved | HARD ABORT — stop immediately |
| Unrelated changes in working tree | Abort staging → ask user to resolve |
| Context budget exceeded | Checkpoint → commit → report → re-run |

---

## Shared Output Contract

Required order for the final user-facing response:

1. Goal/result — state what was completed across the full pipeline run.
2. Key details — tasks completed, tasks blocked, SDD coverage, build/test status, working tree status.
3. Next action — commit confirmation, or remaining work if interrupted.

For `kn-go`, the key details should cover:

- total tasks created and completed
- any blocked or skipped tasks
- SDD coverage percentage
- build/test/lint status
- working tree safety check result
- commit proposal

---

## Dry Run Mode

With `--dry-run`:
- Phase 1: validate spec ✓
- Phase 2: generate task preview (don't create) ✓
- Phase 3-5: skip

Show what would be created and ask user to confirm before running for real.

---

## Checklist

- [ ] Spec is approved (HARD ABORT if not)
- [ ] Spec validated (no broken refs)
- [ ] Tasks generated with fulfills mapping
- [ ] Each task: planned → implemented → ACs checked → validated → done
- [ ] SDD verification passed
- [ ] Build/test/lint passed
- [ ] **Pre-flight: working tree safety check**
- [ ] **Rollback guidance documented**
- [ ] User approved commit
- [ ] Commit created

## Red Flags

- Running on a draft spec
- Skipping task validation between tasks
- Not checking ACs before marking done
- Committing without working tree safety check
- Ignoring build/test failures
- Not reporting progress between tasks
- Continuing past context budget limit without checkpointing