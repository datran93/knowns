---
id: 7vqhyh
title: 'FR-5: Embedding Model Router'
status: done
priority: medium
labels:
  - from-spec
createdAt: '2026-04-27T14:46:31.270Z'
updatedAt: '2026-04-27T15:54:04.826Z'
timeSpent: 0
spec: specs/agent-efficiency-upgrade-tier-1-tier-2
fulfills:
  - AC-4.1
  - AC-4.2
---
# FR-5: Embedding Model Router

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Route queries to FTS5 vs ONNX backend. Keyword queries bypass embedding for less than 100ms latency. Phase 4 per implementation order.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Query classifier: keyword-only vs semantic vs hybrid
- [x] #2 FTS5 fast path for keyword-only queries (< 100ms)
- [x] #3 ONNX semantic routing for semantic queries
- [x] #4 Fallback to FTS5 if embedding fails
- [x] #5 Configurable routing rules in agentEfficiency.modelRouter
- [x] #6 Metrics: track routing effectiveness
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Done: router.go + query_classifier.go already existed. search handler (line 131-156) integrates router via shouldUseModelRouter(). FTS5 fast path, ONNX semantic, fallback, metrics all implemented. Tests pass.
<!-- SECTION:NOTES:END -->

