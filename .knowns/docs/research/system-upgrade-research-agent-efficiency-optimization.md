---
title: 'System Upgrade Research: Agent Efficiency Optimization'
description: Research document phân tích hệ thống Knowns và đề xuất các idea nâng cấp agent efficiency
createdAt: '2026-04-27T14:11:16.976Z'
updatedAt: '2026-04-27T14:11:16.976Z'
tags:
  - research
  - agent
  - efficiency
  - upgrade
  - planning
---

# System Upgrade Research: Agent Efficiency Optimization

## Executive Summary

Phân tích toàn bộ hệ thống Knowns và đề xuất các idea nâng cấp giúp agent hoạt động hiệu quả hơn. Document này serve như một research baseline trước khi chọn các feature cụ thể để spec và implement.

---

## 1. Current System Architecture

### 1.1 Core Components

| Component | Path | Purpose |
|-----------|------|---------|
| MCP Server | `internal/mcp/server.go` | Protocol server exposing all operations as MCP tools |
| Handlers | `internal/mcp/handlers/` | 17 MCP tool handlers (task, doc, search, code, memory, etc.) |
| Storage | `internal/storage/` | SQLite-based persistent storage for all entities |
| Search | `internal/search/` | Hybrid search: semantic (ONNX) + keyword (sqlite FTS5) |
| Code Index | `internal/search/ast_indexer*.go` | Tree-sitter based code analysis and graph construction |
| Skills | `internal/instructions/skills/` | 16 built-in skills embedded in binary |
| Permissions | `internal/permissions/` | Guard middleware for MCP tool access control |
| Runtime | `internal/runtimequeue/` | Background job queue with daemon heartbeat |

### 1.2 Data Model

```
Store (SQLite)
├── docs/          # Markdown docs with metadata
├── tasks/         # Task entities with AC, plan, notes
├── memories/      # Layered memory (project/global)
├── templates/     # Code generation templates
├── time_entries/  # Time tracking
└── versions/      # Version history for all entities
```

### 1.3 MCP Tools Summary (17 tools)

| Category | Tools |
|----------|-------|
| Project | detect, set, current, status |
| Tasks | list, get, create, update, delete, history, board |
| Docs | list, get, create, update, delete, history |
| Memory | add, get, list, update, delete, promote, demote |
| Search | search, retrieve, resolve |
| Code | search, symbols, deps, graph |
| Time | start, stop, add, report |
| Templates | list, get, run, create |
| Research | search_latest_syntax (NEW) |
| GitHub | get_file_content, search_repos, list_commits (NEW) |
| Database | list_tables, inspect_schema, run_query, redis_* (NEW) |
| Validation | validate |

---

## 2. Identified Improvement Areas

### 2.1 High Impact Ideas

#### Idea 1: Persistent Context Window

**Problem:** Mỗi session agent đều bắt đầu từ zero context. Memory có nhưng không được leverage hiệu quả trong context building.

**Current State:**
- `runtimememory` chỉ là session-scoped và cleared mỗi session
- Memory store persistent nhưng không tự động inject vào session init
- `kn-init` chỉ đọc docs, không đọc memory

**Proposal:**
```go
// Trong kn-init, tự động inject relevant memories vào context
mcp_knowns_memory({ "action": "list", "layer": "project" })
// ↓ Lọc top 5 memories liên quan đến project hiện tại
// ↓ Inject vào system prompt hoặc prepend vào conversation
```

**Expected Impact:** Giảm 20-30% token cho những project đã làm việc nhiều lần.

---

#### Idea 2: Smart Session Resume

**Problem:** Khi agent restart trong cùng project, nó phải re-discover mọi thứ từ đầu.

**Current State:**
- Session history không persist
- Không có "session checkpoint" để resume nhanh

**Proposal:**
```go
// Tạo session checkpoint sau mỗi task completion
// Lưu: current task state, recent findings, pending decisions
type SessionCheckpoint struct {
    LastTaskID    string
    RecentDocs    []string  // paths accessed gần đây
    PendingDecisions []string
    CodeGraphSummary string  // compressed graph snapshot
}
```

**Expected Impact:** Resume session nhanh hơn 50%, giảm repetitive discovery.

---

#### Idea 3: Proactive Background Indexing

**Problem:** Search index chỉ được rebuild khi user chạy `knowns reindex` hoặc sau sync. Không có incremental update khi docs/tasks thay đổi.

**Current State:**
- Index sync là batch operation (manual hoặc post-sync)
- Code graph được rebuild hoàn toàn mỗi lần

**Proposal:**
- Background incremental indexer chạy khi storage thay đổi
- Debounced update (3-5s sau last change)
- Separate concerns: index service vs query service

**Expected Impact:** Search latency giảm từ O(n) xuống O(1) cho incremental updates.

---

#### Idea 4: Multi-Agent Awareness

**Problem:** Hiện tại Knowns chỉ hỗ trợ 1 agent tại một thời điểm. Không có coordination giữa multiple agents cùng làm việc.

**Current State:**
- Runtime queue chỉ là job queue, không có coordination
- Locking mechanism đơn giản (per-project)

**Proposal:**
```go
// Agent registry để track các agent đang hoạt động
type AgentSession struct {
    AgentID      string
    StartedAt     time.Time
    CurrentTask   string
    LastHeartbeat time.Time
    Capabilities  []string
}

// Broadcast khi có task completed để các agent khác biết
// Collision detection khi 2 agents cùng nhắm task
```

**Expected Impact:** Enable concurrent agent workflows, tránh duplicate work.

---

### 2.2 Medium Impact Ideas

#### Idea 5: Semantic Code Navigation

**Problem:** `mcp__knowns__code` chỉ hỗ trợ search và deps graph. Không có navigation commands như "find where X is called" hoặc "trace this value through call chain".

**Current State:**
- Code search là keyword/semantic based
- Không có flow analysis (data flow, control flow)

**Proposal:**
```go
// Thêm tools:
mcp_knowns_code({ "action": "trace", "symbol": "DoThing", "direction": "both" })
// Returns: full call graph từ symbol

mcp_knowns_code({ "action": "impact", "file": "auth.go" })
// Returns: all files that depend on auth.go
```

**Expected Impact:** Giảm thời gian understand code structure từ 30min → 5min cho large codebases.

---

#### Idea 6: Skill Composer

**Problem:** 16 skills hiện tại là atomic và standalone. Không có cách combine chúng thành composite workflows.

**Current State:**
- Mỗi skill là single-purpose
- User phải manually chain nhiều commands

**Proposal:**
```go
// Cho phép tạo composite skill từ existing skills
mcp_knowns_templates({ "action": "create",
  "name": "full-review",
  "steps": [
    { "skill": "kn-research", "args": { "topic": "$1" }},
    { "skill": "kn-review", "args": { "taskId": "$2" }},
    { "skill": "kn-test", "args": { "taskId": "$2" }}
  ]
})
```

**Expected Impact:** Giảm 40% manual chaining cho common workflows.

---

#### Idea 7: Contextual Auto-Validation

**Problem:** Validation chỉ chạy khi user hoặc skill gọi nó. Không có proactive validation.

**Current State:**
- Validation là explicit action
- Broken refs không được detect cho đến khi check

**Proposal:**
- Hook validation vào every write operation (non-blocking)
- Background validator cho structural integrity
- Cache validation results với 5-min TTL

**Expected Impact:** Catch broken refs sớm, giảm downstream errors.

---

#### Idea 8: Embedding Model Router

**Problem:** Tất cả semantic search dùng chung model. Không có routing cho different query types.

**Current State:**
- Single model (multilingual-e5-small default)
- Không có hot/cold model switching
- No query classification

**Proposal:**
```go
// Query classification + routing
type QueryRouter struct {
    // Fast path: keyword queries → FTS5 (no embedding)
    // Semantic path: concept queries → ONNX embedding
    // Hybrid path: combine both scores
}

func (r *QueryRouter) Route(query string) QueryPlan {
    if isKeywordOnly(query) {
        return QueryPlan{Type: "fts5", Query: query}
    }
    if isSemanticConcept(query) {
        return QueryPlan{Type: "embedding", Query: query}
    }
    return QueryPlan{Type: "hybrid", Query: query}
}
```

**Expected Impact:** Search latency giảm 60% cho keyword queries, improve relevance cho semantic.

---

### 2.3 Lower Impact (But Still Valuable)

#### Idea 9: Version Branch Diff

**Problem:** Version store lưu history nhưng không có diff visualization.

**Current State:**
- Chỉ raw content stored
- Không có change visualization

**Proposal:**
- Side-by-side diff view cho docs/tasks
- Timeline view cho evolution của entity

#### Idea 10: Permission Presets Manager

**Problem:** Permissions system có presets nhưng không có UI để manage.

**Current State:**
- Presets là compile-time constants
- Không có runtime override

**Proposal:**
- CRUD cho custom permission presets
- Per-session permission overrides

#### Idea 11: Cross-Project References

**Problem:** Knowns chỉ hoạt động trong single project. Không có cross-project refs.

**Current State:**
- `@doc/` refs chỉ trong project hiện tại
- Không có way để reference external projects

**Proposal:**
```go
// External reference syntax
@external(project-id:path/to/doc)
```

---

## 3. Implementation Complexity Analysis

| Idea | Complexity | Dependencies | Priority |
|------|------------|--------------|----------|
| #1 Persistent Context Window | Medium | Memory store, kn-init | HIGH |
| #2 Smart Session Resume | Medium | Version store, runtime | HIGH |
| #3 Proactive Indexing | High | Search service, debouncer | MEDIUM |
| #4 Multi-Agent Awareness | High | Runtime queue, locking | MEDIUM |
| #5 Semantic Code Navigation | Medium | Code graph, ast_indexer | HIGH |
| #6 Skill Composer | Medium | Template system | MEDIUM |
| #7 Contextual Auto-Validation | Low | Validation service | HIGH |
| #8 Embedding Model Router | Medium | Search engine, ONNX | MEDIUM |

---

## 4. Recommended Priority Order

### Tier 1 (Do First - Quick Wins)
1. **#7 Auto-Validation Hook** - Low effort, high impact
2. **#1 Persistent Context Window** - Medium effort, immediate value
3. **#5 Semantic Code Navigation** - Medium, strong ROI

### Tier 2 (Next Quarter)
4. **#2 Smart Session Resume** - Enables continuity
5. **#8 Embedding Model Router** - Performance critical
6. **#6 Skill Composer** - UX improvement

### Tier 3 (Strategic)
7. **#3 Proactive Indexing** - Infrastructure
8. **#4 Multi-Agent Awareness** - Future-proofing

---

## 5. Open Questions

- [ ] User có muốn tất cả ideas hay chỉ subset?
- [ ] Có priority cụ thể nào khác ngoài complexity không?
- [ ] Budget/bandwidth cho implementation là bao nhiêu?
- [ ] Có specific project nào đang blocker mà we should address first?

---

## 6. Next Steps

1. User review và select ideas ưu tiên
2. For selected ideas → `/kn-spec <idea-name>` để detail spec
3. Sau spec approved → `/kn-plan --from @doc/specs/<name>` để generate tasks
