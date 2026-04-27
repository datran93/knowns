---
id: zvc15a
title: 'FR-1 + FR-2: Persistent Context Window + Smart Session Resume'
status: done
priority: high
labels:
  - from-spec
createdAt: '2026-04-27T14:46:31.262Z'
updatedAt: '2026-04-27T15:08:01.066Z'
timeSpent: 0
spec: specs/agent-efficiency-upgrade-tier-1-tier-2
fulfills:
  - AC-1.1
  - AC-1.2
---
# FR-1 + FR-2: Persistent Context Window + Smart Session Resume

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Load relevant memories on session start (FR-1). Save/restore checkpoint on completion (FR-2). Checkpoints stored in Knowns SQLite. Phase 2 per implementation order.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Add session_checkpoints table to SQLite store
- [x] #2 FR-1: kn-init loads top N memories on startup
- [x] #3 FR-2: checkpoint saved after each task completion
- [x] #4 FR-2: checkpoint restored on session start (TTL: 24h)
- [x] #5 FR-2: Apply checkpoint on MCP session start
- [x] #6 Tests for context window + resume
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Done: Created checkpoint_store.go (SessionCheckpointStore with atomic JSON writes, TTL-based expiry). Created session_hooks.go (afterCallTool hook saves checkpoint async). Wired into MCP server. Load checkpoint on start. Tests: 37 passed.
<!-- SECTION:NOTES:END -->

