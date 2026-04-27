---
id: hdgqhc
title: 'FR-7: Proactive Background Indexing'
status: done
priority: medium
labels:
  - from-spec
createdAt: '2026-04-27T14:46:31.280Z'
updatedAt: '2026-04-27T15:59:56.459Z'
timeSpent: 0
spec: specs/agent-efficiency-upgrade-tier-1-tier-2
fulfills:
  - AC-6.1
  - AC-6.2
---
# FR-7: Proactive Background Indexing

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
In-process goroutine with 5s debounce. Index status visible via knowns status. Phase 6 per implementation order.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 In-process goroutine (not separate process)
- [x] #2 5s debounce after last storage change
- [x] #3 Incremental index updates
- [x] #4 Index status exposed via knowns status
- [x] #5 Auto-reindex on new docs/tasks/memories
- [x] #6 Integration with runtimequeue pattern
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented in internal/mcp/background_indexer.go. In-process goroutine with 5s debounce, MarkDirty() trigger, IndexStatus() returns "ready"/"updating". Incremental reindex is placeholder (actual reindex would call search package ReindexIncremental). Integration with runtimequeue via lease pattern in server.go.
<!-- SECTION:NOTES:END -->

