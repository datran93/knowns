---
id: lxm0j7
title: 'FR-3: Semantic Code Navigation'
status: done
priority: high
labels:
  - from-spec
createdAt: '2026-04-27T14:46:31.266Z'
updatedAt: '2026-04-27T15:11:50.844Z'
timeSpent: 0
spec: specs/agent-efficiency-upgrade-tier-1-tier-2
fulfills:
  - AC-2.1
  - AC-2.2
---
# FR-3: Semantic Code Navigation

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
trace/impact/callers actions via code MCP tool. Structured JSON output with file:line refs. Phase 3 per implementation order.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Extend code MCP handler with trace action (call graph)
- [x] #2 Extend code MCP handler with impact action (affected files)
- [x] #3 Extend code MCP handler with callers action
- [x] #4 Structured JSON output with file:line refs
- [x] #5 Max trace depth configurable (default: 10)
- [x] #6 Leverage existing ast_indexer
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Done: Added trace/impact/callers to code MCP handler. trace: BFS call graph traversal. impact: transitive dependency analysis. callers: direct incoming call analysis. All return structured JSON. Build + tests pass.
<!-- SECTION:NOTES:END -->

