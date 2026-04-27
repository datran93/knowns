---
id: ijzk9n
title: 'FR-4: Contextual Auto-Validation'
status: done
priority: high
labels:
  - from-spec
createdAt: '2026-04-27T14:46:31.253Z'
updatedAt: '2026-04-27T15:00:33.013Z'
timeSpent: 0
spec: specs/agent-efficiency-upgrade-tier-1-tier-2
fulfills:
  - AC-3.1
  - AC-3.2
---
# FR-4: Contextual Auto-Validation

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Hook validation into task/doc/memory create/update. Non-blocking with 5-min TTL cache. Phase 1 per implementation order.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Create validation hook middleware for MCP write operations
- [x] #2 Add 5-min TTL cache for validation results
- [x] #3 Integrate with existing validate package
- [x] #4 Non-blocking async execution
- [x] #5 Tests for validation hook
- [x] #6 NFR-1: < 50ms latency
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Done: Created auto_validation.go with beforeCallTool hook (non-blocking async, 5-min TTL cache). Wired into MCP server via WithHooks. Config checked at call time (no blocking). Tests: 11 passed in mcp package."
<!-- SECTION:NOTES:END -->

