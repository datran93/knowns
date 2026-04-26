---
name: kn-review
description: Use when reviewing implemented code before committing — multi-perspective review with adversarial mindset to find bugs and prove code breaks.
---

# Adversarial Code Review

Post-implementation quality review. Run after `kn-implement`, before `kn-test` or `kn-commit`.

**Announce:** "Using kn-review for task [ID] (or current changes)."

> 👿 **Adversarial Mindset**: You are an **Adversarial Reviewer**. Your goal is NOT to confirm that the code works; your goal is to **PROVE THAT IT BREAKS**. Assume the Coder has made mistakes, overlooked edge cases, or misunderstood the requirements. Hunt for race conditions, nil pointer dereferences, logic gaps, and security vulnerabilities. Be ruthless but objective.

## When to Use

- After implementing a task, before testing or committing.
- When the user asks to "review my code", "check this", or "review before commit".
- To find vulnerabilities, logical errors, or unhandled edge cases in an implementation.

## Inputs

- Task ID (optional — if provided, reviews against task ACs and spec).
- Current git diff (always).

## Step 1: Gather Review Context

```bash
git diff --stat
git diff
```

If task ID provided:
```json
mcp_knowns_tasks({ "action": "get", "taskId": "$ARGUMENTS" })
```

If task has spec:
```json
mcp_knowns_docs({ "action": "get", "path": "<spec-path>", "smart": true })
```

**Validate Integrity:** Check for broken references, missing ACs, and orphan docs.
```json
mcp_knowns_validate({ "scope": "sdd" })
```

**Check Impact (Breakage Analysis):** Find callers of modified symbols to confirm no unintended breakage.
```json
mcp_knowns_code({ "action": "deps", "type": "calls" })
```

Search for relevant conventions and past review patterns:
```json
mcp_knowns_search({ "action": "search", "query": "<feature area>", "type": "memory" })
```

## Step 2: Adversarial Semantic Audit

Review the diff ruthlessly from the following perspectives:

### 2a. Adversarial Edge-Case Hunting
- **Extreme Inputs**: What happens if the input is nil, empty, extremely large, malformed, or negative?
- **State Failures**: What happens if the DB drops, a transaction fails halfway, or a 3rd-party API times out?
- **Concurrency**: Are there race conditions? Is shared state mutated unsafely?
- **Bypass**: Can a malicious user skip validation or spoof an identity?

### 2b. Security Review
- Input validation — user input sanitized?
- Auth — proper authorization checks?
- Secrets — no hardcoded credentials or tokens?
- Data exposure — sensitive data in logs, responses, or error messages?
> **Rule:** Read the actual code. NEVER reduce this to "any obvious issues".

### 2c. Performance Review
- DB queries: indexed? bounded (LIMIT)? Could return unbounded rows?
- Loops: N+1 queries inside? Bounded?
- Async: context cancellation, timeouts, error propagation?
- Memory: unbounded allocations?

### 2d. Design Conformity & Completeness
- Traceability: Does each file implement what the design specified?
- Creep: Did the Coder ADD anything NOT in the design?
- Missing tests for new logic or unhandled edge cases.
- ACs from task not fully met (if task provided).

## Step 3: Triage Findings

Classify each finding strictly:

| Severity | Criteria | Action |
|----------|----------|--------|
| **P1** | Security vuln, data corruption, breaking change, bypass, stub shipped | **Blocks commit — must fix** |
| **P2** | Performance issue, concurrency risk, missing test for logic | Should fix before commit |
| **P3** | Minor cleanup, naming, style | Record for later |

**Calibration:** Label accurately. NEVER downplay blocking issues. Be ruthless.

## Step 4: Report Findings

Present findings grouped by severity:

```
Review Complete — [task-id or "current changes"]
═══════════════════════════════════════════════

P1 (blocks commit): X findings
- [file:line] Description — why it breaks the system

P2 (should fix): X findings
- [file:line] Description — impact

P3 (nice to have): X findings
- [file:line] Description

Verdict: PASS / BLOCKED (P1 exists)
```

## Step 5: Handle Results

### If P1 findings exist — HARD GATE
> ⛔ P1 findings block commit. Fix these first:
> 1. [Finding + suggested fix]
>
> After fixing, run `/kn-review` again.
Do NOT proceed to commit. Do NOT offer to skip P1.

### If only P2/P3
> ✓ No blocking issues. P2 findings recommended:
> 1. [Finding + suggested fix]
>
> Options:
> - Fix P2s now, then proceed
> - Commit as-is: `/kn-commit`
> - Create follow-up task for P2s

### If clean
> ✓ Review passed. No issues found.
>
> Ready: `/kn-test` or `/kn-commit`

## Constraints & Red Flags
1. **Report, NEVER fix**: Identify and document issues — NEVER modify code directly during review unless instructed.
2. **Adversarial distance**: Do not trust the implementation. Prove it is safe before approving.
3. **Severity honesty**: Label accurately. NEVER downplay blocking issues.
4. **Read the code**: Every finding MUST reference specific files and line numbers.
