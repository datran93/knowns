---
name: kn-commit
description: Use when committing code changes with proper conventional commit format and verification
---

# Committing Changes

**Announce:** "Using kn-commit to commit changes."

**Core principle:** VERIFY BEFORE COMMITTING - check staged changes, ask for confirmation.

## Inputs

- Current staged changes
- Relevant task IDs, scope, and reason for the change
- Optional: `--force` to skip unrelated change detection (use with caution)

## Preflight

- Confirm the correct files are staged
- Check whether the commit should reference a task or feature area
- **Refuse to commit** if the staged diff looks unrelated or mixed across multiple concerns

## Step 1: Working Tree Safety Check

Before staging, check what exists in the working tree:

```bash
git status --short
```

**Unrelated changes check**: If there are modified files that are NOT part of the current commit scope:

```
⚠️ Working tree has unrelated changes:
  M <file-1>    <- not part of this commit
  M <file-2>    <- not part of this commit

Options:
1. git stash    (stash unrelated, keep staged, then commit)
2. git add <specific-files>   (stage only the intended files)
3. Abort        (don't commit until unrelated changes are resolved)

Do NOT use `git add -A` when unrelated changes exist.
```

If unrelated changes exist and user doesn't use `--force`:
- Report the conflict clearly
- Do NOT proceed to `git add -A`
- Ask user to resolve before committing

## Step 2: Review Staged Changes

```bash
git status
git diff --staged
```

**Unrelated change detector**: Scan staged files against expected scope.
If staged files include paths unrelated to the current task/spec:
```
⚠️ Staged diff includes unrelated files:
  <unrelated-path-1>
  <unrelated-path-2>

These files don't match the current implementation scope.
Options:
1. Unstage them: git restore --staged <file>
2. Commit anyway (use --force to bypass)
```

**Abort** if unrelated staged files exist and user hasn't approved.

## Step 3: Generate Commit Message

**Format:**
```
<type>(<scope>): <message>

- Bullet point summarizing change
```

**Types:** feat, fix, docs, style, refactor, perf, test, chore

**Rules:**
- Title lowercase, no period, max 50 chars
- Body explains *why*, not just *what*

## Step 4: Ask for Confirmation

```
Ready to commit:

feat(auth): add JWT token refresh

- Added refresh token endpoint

Working tree: clean (no unrelated changes)
Staged files: 4 files, +127/-34 lines

Proceed? (yes/no/edit)
```

**Wait for user approval.**

## Step 5: Commit

```bash
git commit -m "feat(auth): add JWT token refresh

- Added refresh token endpoint"
```

## Post-Commit Actions (Optional)

After successful commit, offer post-commit actions:

```
✓ Committed: abc1234 — feat(auth): add JWT token refresh

Post-commit options:
1. Push: git push origin main
2. Tag: git tag v1.2.0
3. CI trigger: (if applicable)
4. Next workflow: /kn-verify
```

Only suggest these if relevant — don't force a handoff.

---

## Final Response Contract

All built-in skills in scope must end with the same user-facing information order: `kn-init`, `kn-spec`, `kn-plan`, `kn-research`, `kn-implement`, `kn-verify`, `kn-doc`, `kn-template`, `kn-extract`, and `kn-commit`.

Required order for the final user-facing response:

1. Goal/result - state whether a commit was proposed, blocked, or created.
2. Key details - include the proposed commit message, staged file summary, working tree status, and approval status.
3. Next action - recommend a concrete follow-up command only when a natural handoff exists.

Keep this concise for CLI use. Verification-specific content may extend the key-details section, but must not replace or reorder the shared structure.

For `kn-commit`, the key details should cover:

- the proposed commit title
- staged file count and diff size
- working tree safety status
- any concerns about the staged diff
- a clear approval prompt

## Guidelines

- Only commit staged files
- NO "Co-Authored-By" lines
- NO "Generated with Claude Code" ads
- Ask before committing
- **Detect unrelated changes** before staging

## Next Step Suggestion

When a follow-up is natural, recommend exactly one next command:

- after proposing a commit: no command, wait for approval
- after a successful commit tied to active work: `/kn-verify`
- after a successful standalone commit: `/kn-extract` or `/kn-review` if it's a meaningful change
- after commit with pushed code: `/kn-verify` to confirm repo health

## Checklist

- [ ] Reviewed staged changes
- [ ] **Unrelated changes detected and resolved**
- [ ] Message follows convention
- [ ] User approved
- [ ] Next action suggested when applicable

## Abort Conditions

- Nothing staged
- Staged diff includes unrelated work that should be split
- Working tree has unrelated changes (unless `--force` used)
- User has not explicitly approved the final message

## Red Flags

- Using `git add -A` when unrelated changes exist in working tree
- Committing without checking what is actually staged
- Ignoring pre-existing uncommitted changes that might get accidentally bundled