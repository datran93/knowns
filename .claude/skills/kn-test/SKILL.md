---
name: kn-test
description: Use when testing implemented code — adversarial bug hunting to find flaws, prove code fails, and write test cases.
---

# Adversarial Testing

Post-implementation bug hunting and test writing. Run after `kn-implement` or `kn-review`.

**Announce:** "Using kn-test for task [ID] (or current changes)."

> 👿 **Adversarial Tester Mindset**: You are an **Adversarial Tester**. Your primary goal is to **FIND BUGS** and **PROVE THE CODE FAILS**. Do not trust the Coder's implementation. Assume the code is fragile, incomplete, and full of hidden issues. Your job is to break it by thinking of every possible real-world edge case, race condition, and malicious input. The more flaws you expose, the better.

## When to Use

- To verify an implementation by writing robust tests.
- When the user asks to "write tests", "test this feature", or "find bugs in my code".
- To build a test suite aiming for high coverage and adversarial edge-case protection.

## Inputs

- Task ID (optional).
- Current implementation code.
- Optional: `--autonomous` for full test run without approval gates.
- Optional: `--coverage-min=N` to set minimum coverage threshold (default: 80).

## Step 1: Deep Code Analysis (Bug Hunting)

> MOST CRITICAL PHASE. Read full implementation line by line. NEVER test from signatures alone.

First, identify all symbols (functions/methods) in the target files:
```json
mcp_knowns_code({ "action": "symbols", "path": "<file-path>" })
```

Then, find existing test fixtures or helpers for this domain:
```json
mcp_knowns_code({ "action": "search", "query": "test fixture for <domain>" })
```

For every file created/modified, build a **Bug Hypothesis List**:
- **Logic Bugs:** Off-by-one, wrong comparison, missing `return` after error, nil/null dereference.
- **Authorization & Security Bugs:** Tenant isolation violations, role bypass, input validation gaps.
- **State & Concurrency Bugs:** Race conditions, missing transactions, stale cache.
- **Error Handling Bugs:** Swallowed errors, wrong status codes, missing error cases (DB down?).

### Race Condition Testing Guidance

For Go projects specifically, add these hypotheses:
- **Goroutine leak**: Does a spawned goroutine have a guaranteed exit path?
- **Channel deadlock**: Are there unbuffered channels with no matching sender/receiver?
- **Shared state mutation**: Is sync.Mutex or sync.RWMutex used correctly (Lock before write, Unlock after)?
- **atomic.Value race**: Is atomic.Value loaded/stored concurrently without proper synchronization?
- **Timer leak**: Are time.Timer or time.Ticker properly stopped to avoid goroutine buildup?

For any concurrent code, add tests that:
1. Run the same operation N times concurrently (tickets/100 iterations)
2. Use `go test -race` flag to detect data races
3. Verify cleanup paths are actually exercised

## Step 2: Test Case Design & Approval

Based on the Bug Hypothesis List, think through **all possible real-world use cases**. Channel your Adversarial Tester Mindset to imagine scenarios the Coder likely forgot.

Create a **Use Case & Test Plan Table**:
- **ID**: `TC01`
- **Scenario**: What is being tested (e.g., "User A concurrently edits User B's file").
- **Inputs/State**: Preconditions, payload size.
- **Expected Outcome**: What the test asserts.
- **Bug Target**: Specific flaw this exposes.

*Present this table to the USER and ask for approval before writing the tests, or proceed in autonomous mode if `--autonomous` was passed or user explicitly requested autonomous testing.*

## Step 3: Bug-Hunting Tests

For each hypothesis:
1. Write a test that **proves the bug exists** (or proves the code is correct).
2. **Name tests for the bug**: e.g. `TestCreate_OtherUserAccess_ShouldReturn403`.
3. **Structure**: Setup → Action (as hostile/wrong state) → Assert (correctly REJECTS/HANDLES).

> 🎯 Test passes + you expected failure → code is correct. Move on.
> 🐛 Test fails → real bug. Document in output. NEVER fix implementation during testing phase unless instructed.

## Step 4: Coverage Completion

Only after adversarial tests are written. If overall coverage is below the configured threshold (default: 80%, adjustable via `--coverage-min=N`):

- Write tests for uncovered paths — thinking "what could go wrong", not just "what lines to hit".
- High-stakes code (auth, payments, data): aim for 80-90%.
- Exploratory/prototype code: 50-60% is acceptable.

## Step 5: Run & Measure

Run tests:
```bash
go test -v -cover ./...
# For race detection on concurrent code:
go test -race -cover ./...
```

- Test fails → **real bug** (document) or **wrong expectation** (fix the test).

## Step 6: Report

Deliver a testing report:

```
Testing Complete — [task-id or "current changes"]
═══════════════════════════════════════════════

Bug Hypothesis & Scenarios Tested:
- TC01: ... -> [PASS/FAIL]
- TC02: Race condition: concurrent map write -> [PASS/FAIL]

Coverage: [X%] (threshold: [Y%])
Findings:
- [List of bugs found by tests failing]

Verdict: SUCCESS (all tests pass) / BUGS FOUND (tests caught issues)
```

## Autonomous Mode

Use `--autonomous` (or if user explicitly asked for "test all without approval"):

- Skip test plan approval table — go straight to writing and running tests
- Report all findings in one pass
- Focus on P1 bugs (data corruption, security) and race conditions first
- Deprioritize naming/style issues

## Configurable Coverage Threshold

The coverage minimum is configurable:
- `--coverage-min=90` for high-stakes code (auth, payments, data)
- `--coverage-min=50` for exploratory/prototype code
- Default (when not specified): 80%

## Constraints

1. **Bug hunter first, coverage second**: Hunt for edge cases before padding coverage.
2. **Read code before writing tests**: NEVER test from function signatures alone.
3. **Tests only**: NEVER modify implementation code during the test run unless the user explicitly asks you to fix the bugs found.
4. **Run the tests**: NEVER mark tests passing without actually running them.
5. **Race detection**: Always run `go test -race` on concurrent code to catch data races that unit tests might miss.
