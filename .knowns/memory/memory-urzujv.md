---
id: urzujv
title: 'Bug fix: stale test binary hooks in settings.json'
layer: project
category: debug
tags:
  - runtimeinstall
  - debug
  - hooks
  - claude
createdAt: '2026-04-28T15:54:51.512Z'
updatedAt: '2026-04-28T16:31:03.773Z'
---

**Bug**: `make test` corrupt `~/.claude/settings.json` with stale SessionStart hooks pointing to test binary path `/var/folders/.../cli.test`.

**Root cause**: `runtimeinstall.Install()` calls `os.Executable()` which returns the Go test binary path during `go test`. Hook path stored in settings.json was the ephemeral test binary.

**Fix**: `hookCommandPath()` and `legacyPromptCommandPath()` now detect test binary paths via `isTestBinaryPath()` (checks for `/var/folders` + `cli.test`) and replace with `./bin/knowns` (resolved from working directory at test time).

**Files**: `internal/runtimeinstall/runtimeinstall.go`

**Note**: `statusMessage` in Claude settings.json hook groups can be at group-level OR hook-entry level. All helper functions (`isStaleKnownsHookGroup`, `isSameManagedHook`, etc.) must check both locations.
