---
name: kn-debug
description: Use when debugging errors, test failures, build issues, or blocked tasks — structured triage to fix to learn
---

# Debugging

Systematic debugging: triage → reproduce → diagnose → fix → learn.

**Announce:** "Using kn-debug for [error/issue]."

**Core principle:** CLASSIFY FIRST → REPRODUCE → ROOT CAUSE → FIX → CAPTURE LEARNING.

## When to Use

- Build fails (compilation, type error, missing dependency)
- Test fails (assertion mismatch, timeout, flaky)
- Runtime crash or exception
- Integration failure (API mismatch, env config, auth)
- Task blocked with unclear cause
- User says "debug this", "fix this error", "why is this failing"

## Inputs

- Error message, stack trace, or failing command
- Optional: task ID (if debugging within a task context)

---

## Step 1: Triage — Classify the Issue

Classify before investigating. Misclassifying wastes time.

| Type | Signals |
|---|---|
| **Build failure** | Compilation error, type error, missing module, bundler failure |
| **Test failure** | Assertion mismatch, snapshot diff, timeout, flaky intermittent |
| **Runtime error** | Crash, uncaught exception, undefined behavior |
| **Integration failure** | HTTP 4xx/5xx, env variable missing, API schema mismatch |
| **Blocked task** | Circular dependency, conflicting changes, unclear requirement |

**Output:** One-line classification: `[TYPE] in [component]: [symptom]`

---

## Step 2: Check Known Patterns

Before deep investigation, search for known solutions (unified search includes docs, learnings, and memories):

```json
mcp_knowns_search({ "action": "search", "query": "<keywords from classification>", "type": "doc" })
```

Also check learnings docs:
```json
mcp_knowns_search({ "action": "search", "query": "<error pattern>", "type": "doc", "tag": "learning" })
```

**Query debug pattern memory specifically:**
```json
mcp_knowns_search({ "action": "search", "query": "<error pattern>", "type": "memory", "tag": "debug" })
```

If a known pattern matches → jump to Step 4 (Fix) using the documented resolution.

---

## Step 3: Reproduce & Diagnose

### 3a. Reproduce

Run the exact failing command verbatim:
```bash
# Whatever failed — run it exactly
<failing-command> 2>&1
```

Capture error output verbatim. Exact line numbers and messages matter.

Run twice — if intermittent, classify as flaky (check shared state, race conditions, test ordering).

### 3b. Read implicated files

Read exactly the files mentioned in the error output. Do not read the entire codebase.

### 3c. Check recent changes

```bash
git log --oneline -10
git diff HEAD~3 -- <failing-file>
```

If a recent commit introduced the failure → fix is likely adjusting that change.

### 3d. Check task context (if task ID provided)

```json
mcp_knowns_tasks({ "action": "get", "taskId": "<id>" })
```

Does the failure indicate the task was implemented against the wrong spec, or correctly but the spec was wrong?

### 3e. Narrow to root cause

Write a one-sentence root cause:

> Root cause: `<file>:<line>` — `<what is wrong and why>`

If you cannot write this sentence, you do not have the root cause yet. Do NOT proceed to Fix.

---

## Step 4: Fix

### Small fix (1–3 files, obvious change)

- Implement directly
- Run verification immediately:
```bash
# Re-run the originally failing command
<failing-command>
```

### Substantial fix (cross-cutting, logic redesign)

- If within a task, append notes about the issue:
```json
mcp_knowns_tasks({ "action": "update", "taskId": "<id>",
  "appendNotes": "🐛 Debug: <root cause summary>. Fix: <what was changed>"
})
```

- If standalone, consider creating a task:
```json
mcp_knowns_tasks({ "action": "create", "title": "Fix: <root cause summary>",
  "description": "Root cause: <detail>\nFix approach: <approach>",
  "priority": "high",
  "labels": ["bugfix"]
})
```

### Verify the fix

Run the exact command that originally failed. It must pass cleanly:
```bash
<original-failing-command>
```

Also run broader checks for regressions:
```bash
# Project-specific build/test/lint
go build ./...
go test ./...
```

If verification fails → return to Step 3 with new information. Do NOT report success.

---

## Step 5: Learn — Capture the Pattern

### Learning Capture Trigger (Inclusive)

Capture a learning if **ANY** of:
- Debug time ≥ 10 minutes (not trivial)
- Root cause is non-obvious (not immediately visible from error message)
- The fix could apply to future similar issues
- The error could recur in other parts of the codebase

**Rule**: Better to capture a small learning than to miss a significant one. If unsure, err on the side of capturing.

### New failure pattern worth remembering?

Ask: would this save ≥15 minutes if a future agent knew it?

**Quick pattern (< 5 min to describe):** save to memory with `category: debug` for fast recall in future `kn-debug` runs:
```json
mcp_knowns_memory({ "action": "add", "title": "<error pattern>",
  "content": "Root cause: <sentence>. Fix: <what resolves it>",
  "layer": "project",
  "category": "debug",
  "tags": ["debug", "<domain>"]
})
```

Note: The `debug` category enables future `kn-debug` sessions to query patterns by error type. This is distinct from `failure` which is more general.

**Detailed pattern (worth a full writeup):** create or update a learning doc:

```json
mcp_knowns_search({ "action": "search", "query": "<failure domain>", "type": "doc", "tag": "learning" })
```

**If existing learning doc found — update it:**
```json
mcp_knowns_docs({ "action": "update", "path": "<existing-path>",
  "appendContent": "\n\n## <Date> — <Classification>\n\n**Root cause:** <sentence>\n**Signal:** <how to recognize>\n**Fix:** <what resolves it>"
})
```

**If no existing doc — create new:**
```json
mcp_knowns_docs({ "action": "create", "title": "Learning: <domain> — <pattern>",
  "description": "Debugging pattern for <issue type>",
  "folder": "learnings",
  "tags": ["learning", "<domain>"],
  "content": "## Problem\n\n<what goes wrong>\n\n## Root Cause\n\n<why it happens>\n\n## Signal\n\n<how to recognize this pattern>\n\n## Fix\n\n<what resolves it>\n\n## Source\n\n@task-<id> (if applicable)"
})
```

### Known pattern that didn't work?

If the documented resolution failed or is outdated:
```json
mcp_knowns_docs({ "action": "update", "path": "<learning-path>",
  "appendContent": "\n\n⚠️ **Update <date>:** Resolution no longer accurate — <what changed>"
})
```

---

## Step 6: Post-Fix Review (Optional but Recommended)

After fixing and capturing the learning:

> Review fix before committing? (`/kn-review`)

This is especially important for:
- Security-related fixes
- Concurrency/race condition fixes
- Changes to shared/ core modules

```
After fix captured:
→ Run `/kn-review` to review the fix before committing
→ Or `/kn-commit` to commit directly if fix is simple and isolated
```

---

## Shared Output Contract

Required order for the final user-facing response:

1. Goal/result — what was debugged and whether it's fixed.
2. Key details — root cause, fix applied, verification status, learning captured (and whether it was reviewed).
3. Next action — resume implementation, or escalate if unfixable.

For `kn-debug`, the key details should cover:

- classification and root cause
- what was changed to fix it
- verification result (pass/fail)
- whether a learning was captured (and its category: `debug` vs `failure`)
- whether post-fix review was suggested

---

## Quick Reference

| Situation | First action |
|---|---|
| Build fails | `git log --oneline -10` — check recent changes |
| Test fails | Run test verbatim, capture exact assertion output |
| Flaky test | Run 5× — if intermittent, check shared state/ordering |
| Runtime crash | Read stack trace top-to-bottom, find first line in your code |
| Integration error | Check env vars, then API response body (not just status code) |
| Recurring issue | Search debug memories (`category: debug`) first |

## Related Skills

- `/kn-implement <id>` — resume implementation after fix
- `/kn-extract` — extract pattern if fix reveals reusable knowledge
- `/kn-review` — review fix before committing
- `/kn-commit` — commit the fix

## Checklist

- [ ] Issue classified
- [ ] Known patterns checked (including debug memory category)
- [ ] Reproduced with exact command
- [ ] Root cause identified (one sentence)
- [ ] Fix applied and verified
- [ ] Learning captured (if ≥10 min OR non-obvious root cause)
- [ ] **Post-fix review suggested**

## Red Flags

- Fixing symptoms without root cause
- Skipping reproduction — diagnosing from error message alone
- Not checking known patterns first (especially debug memory category)
- Committing fix without running verification
- Not capturing a learning when the fix took ≥10 minutes to find
- Not suggesting post-fix review for security/concurrency fixes