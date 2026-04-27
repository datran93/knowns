---
title: Agent Efficiency Upgrade (Tier 1 + Tier 2)
description: Specification for Agent Efficiency Upgrade - 8 features across Tier 1 and Tier 2
createdAt: '2026-04-27T14:23:41.509Z'
updatedAt: '2026-04-27T14:41:53.549Z'
tags:
  - spec
  - approved
---

# Agent Efficiency Upgrade Spec

## Overview

Upgrade Knowns v2 with 8 coordinated features (Tier 1 + Tier 2) to dramatically improve agent productivity through smarter context management, faster search, and better workflow composition.

## Locked Decisions

- **D1**: All features are additive — no breaking changes to existing APIs or data models
- **D2**: Incremental rollout — each feature is independently deployable
- **D3**: Non-blocking by default — new features don't slow down existing operations
- **D4**: Backward compatible — existing workflows continue to work unchanged
- **D5**: Default: all features **ON** (opinionated defaults). Users who want conservative behavior can opt-out per-feature. Most users benefit from day-one activation.
- **D6**: Session checkpoint storage: **Knowns SQLite store** alongside existing data. Benefits from existing backup/cleanup mechanisms; no extra files to manage.
- **D7**: Multi-agent locking: **Optimistic** (lock-free with conflict detection). Lower overhead than pessimistic locking; most tasks don't conflict. On conflict, requester waits or picks different task.
- **D8**: Background indexer: **In-process goroutine** in the main CLI binary. Simpler deployment than separate daemon; existing runtimequeue pattern handles heartbeat/cleanup.

## Requirements

### Tier 1 Features

#### FR-1: Persistent Context Window
Agent nhận context từ memories khi session bắt đầu, không cần re-discover.

- **FR-1.1**: `kn-init` tự động load top N memories liên quan đến project (configurable, default: 5)
- **FR-1.2**: Memories được filtered theo relevance (recency + keyword match)
- **FR-1.3**: Output format: compact summary prepended vào session context

#### FR-2: Smart Session Resume
Lưu session checkpoint để agent có thể resume nhanh khi restart.

- **FR-2.1**: Tạo checkpoint sau mỗi task completion
- **FR-2.2**: Checkpoint lưu trong Knowns store (SQLite): lastTaskID, recentDocs[], pendingDecisions[], codeGraphSummary
- **FR-2.3**: Agent check và apply checkpoint khi session start
- **FR-2.4**: Checkpoint có TTL (configurable, default: 24h)

#### FR-3: Semantic Code Navigation
Navigate code structure với trace và impact analysis.

- **FR-3.1**: `trace` action: find call graph từ/to a symbol
- **FR-3.2**: `impact` action: find all files affected by a change
- **FR-3.3**: `callers` action: find who calls a function
- **FR-3.4**: Output format: structured JSON với file:line references

#### FR-4: Contextual Auto-Validation
Validate automatic khi có write operations.

- **FR-4.1**: Hook validation vào task/doc/memory create/update (non-blocking)
- **FR-4.2**: Validation results cached với 5-min TTL
- **FR-4.3**: Warning notifications hiển thị nhưng không block operations

### Tier 2 Features

#### FR-5: Embedding Model Router
Route queries đến appropriate search backend.

- **FR-5.1**: Query classifier: keyword-only → FTS5, semantic → ONNX, hybrid → combine
- **FR-5.2**: Configurable routing rules
- **FR-5.3**: Fallback: if embedding fails, auto-fallback to FTS5
- **FR-5.4**: Metrics collection for routing effectiveness

#### FR-6: Skill Composer
Kết hợp skills thành composite workflows.

- **FR-6.1**: Template system hỗ trợ skill composition
- **FR-6.2**: Syntax: `{ "skill": "kn-xxx", "args": {...} }` trong steps array
- **FR-6.3**: Variable interpolation: `$1`, `$2` positional args
- **FR-6.4**: Built-in composite templates: "full-review", "implement-and-test"
- **FR-6.5**: Execution engine run sequential steps

#### FR-7: Proactive Background Indexing
Incremental index updates khi storage thay đổi.

- **FR-7.1**: Debounced updates (default: 5s sau last change)
- **FR-7.2**: Background worker as in-process goroutine (not separate process)
- **FR-7.3**: Index status visible via `knowns status`
- **FR-7.4**: Manual reindex still available via `knowns reindex`

#### FR-8: Multi-Agent Awareness
Coordination giữa multiple agents.

- **FR-8.1**: Agent registry track active sessions (stored in Knowns SQLite)
- **FR-8.2**: Task locking: optimistic lock with conflict detection — prevent 2 agents same task
- **FR-8.3**: Broadcast mechanism khi task completed
- **FR-8.4**: Lock TTL với auto-release on agent disconnect (configurable, default: 5min)

## Non-Functional Requirements

### NFR-1: Performance
- Validation hook < 50ms latency
- Session resume < 2s
- Index update latency < 5s (debounced)

### NFR-2: Compatibility
- All new features optional via config flags (default: all ON — opinionated defaults)
- Default behavior upgraded for new users; existing users can opt-out per-feature

### NFR-3: Observability
- New features log to `~/.knowns/logs/`
- Metrics exposed via `knowns status --json`

## Acceptance Criteria

### Tier 1

- [ ] **AC-1.1**: `kn-init` outputs memory summary when project has existing memories (when FR-1 enabled)
- [ ] **AC-1.2**: Session resume applies last checkpoint when available (when FR-2 enabled)
- [ ] **AC-2.1**: `code trace` returns structured call graph JSON
- [ ] **AC-2.2**: `code impact` returns affected files list
- [ ] **AC-3.1**: Write operations trigger non-blocking validation (when FR-4 enabled)
- [ ] **AC-3.2**: Validation warnings visible but not blocking

### Tier 2

- [ ] **AC-4.1**: Query routing chooses correct backend (FTS5 vs embedding) (when FR-5 enabled)
- [ ] **AC-4.2**: Keyword queries bypass embedding for 60%+ latency improvement
- [ ] **AC-5.1**: Composite skill executes all steps sequentially (when FR-6 enabled)
- [ ] **AC-5.2**: Variable interpolation works with positional args
- [ ] **AC-6.1**: Background index updates within 5s of storage change (when FR-7 enabled)
- [ ] **AC-6.2**: Index status shows "ready" or "updating"
- [ ] **AC-7.1**: Agent registry shows active sessions (when FR-8 enabled)
- [ ] **AC-7.2**: Task lock prevents duplicate work (optimistic, not blocking)

## Scenarios

### Scenario 1: New Session with Memories
**Given** Project có memories từ previous sessions
**When** Agent run `kn-init` hoặc MCP session start (with FR-1 enabled)
**Then** Top 5 relevant memories prepended vào context
**And** Agent không cần re-discover project history

### Scenario 2: Session Resume After Crash
**Given** Agent crashed giữa chừng task (FR-2 enabled)
**When** Agent restart và reconnect
**Then** Last checkpoint applied automatically
**And** Agent resume từ point trước crash

### Scenario 3: Trace Unknown Symbol
**Given** Agent cần understand một function được call từ đâu
**When** `code trace` với symbol name
**Then** Full call graph returned (callers + callees)
**And** Each node includes file:line reference

### Scenario 4: Two Agents Same Task
**Given** Agent A đã lock task X (FR-8 enabled)
**When** Agent B request task X lock
**Then** Optimistic conflict detected: lock denied với message "Task X locked by Agent A"
**And** Agent B wait hoặc pick different task (no blocking)

### Scenario 5: Keyword Search Fast Path
**Given** User search với keyword-only query (FR-5 enabled)
**When** Query routed to FTS5 backend
**Then** Response time < 100ms
**And** No embedding computation triggered

### Scenario 6: Background Index Update
**Given** FR-7 enabled, user creates new doc
**When** 5s debounce elapses with no further changes
**Then** Background goroutine incrementally updates search index
**And** `knowns status` shows "index: updating" then "index: ready"

## Technical Notes

### Implementation Order (Dependency Graph)

```
Phase 1: FR-4 (Auto-Validation) — Simplest, no dependencies
Phase 2: FR-1 (Context Window) + FR-2 (Session Resume) — Both use memory store
Phase 3: FR-3 (Code Navigation) — Requires code graph existing
Phase 4: FR-5 (Model Router) — Requires search engine understanding
Phase 5: FR-6 (Skill Composer) — Requires template system
Phase 6: FR-7 (Background Indexing) — Requires Phase 4/5
Phase 7: FR-8 (Multi-Agent) — Requires runtime understanding
```

### Config Schema

```json
{
  "settings": {
    "agentEfficiency": {
      "persistentContext": {
        "enabled": true,
        "maxMemories": 5
      },
      "sessionResume": {
        "enabled": true,
        "checkpointTTLHours": 24
      },
      "codeNavigation": {
        "enabled": true,
        "maxTraceDepth": 10
      },
      "autoValidation": {
        "enabled": true,
        "cacheTTLSeconds": 300
      },
      "modelRouter": {
        "enabled": true,
        "fts5Threshold": 0.7
      },
      "skillComposer": {
        "enabled": true
      },
      "backgroundIndexing": {
        "enabled": true,
        "debounceSeconds": 5
      },
      "multiAgent": {
        "enabled": true,
        "lockTTLSeconds": 300
      }
    }
  }
}
```

## Open Questions

All resolved:

- [x] Default config: all features **ON** (opinionated defaults) (D5)
- [x] Session checkpoint storage: **Knowns SQLite store** (D6)
- [x] Multi-agent locking: **Optimistic** (lock-free w/ conflict detection) (D7)
- [x] Background indexer: **In-process goroutine** in main binary (D8)

## Related Research

@doc/research/system-upgrade-research-agent-efficiency-optimization
