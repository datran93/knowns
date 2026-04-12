## v0.18.3 - Runtime Adapter Install, MCP-First Retrieval & Search Improvements

### ✨ Added

- Unified runtime adapter install package for global hook/plugin setup across claude-code, codex, kiro, and opencode
- `runtime-memory` CLI command for direct runtime memory hook invocation
- Kiro native runtime-memory hook (`.kiro/hooks/knowns-runtime-memory.kiro.hook`)
- Specs: multi-store semantic memory retrieval and unified runtime adapter install
- Design system reference doc for SVG diagrams, Web UI, and knowns.sh visual consistency
- README SVG diagrams: architecture, capabilities, cover, knowledge-graph, mcp-integration, multi-platform, template-pipeline, workflow

### 🔄 Changed

- All agent skills and compatibility shims now prefer MCP `retrieve` tool over CLI, with CLI as fallback
- Search engine improvements: chunker source-type awareness, vecstore type filtering, sqlite store optimizations, and sync refinements
- Init flow updated with runtime discovery, adapter install integration, and MCP-first guidance generation
- Docs refreshed: commands, configuration, semantic-search, mcp-integration, user-guide, developer-guide, web-ui

### 🐛 Fixed

- Search type filtering now applies at the vector store level so `--type memory` queries are not crowded out by unrelated doc/task chunks

---

### Contributors

@howznguyen

**Full Changelog**: https://github.com/knowns-dev/knowns/compare/v0.18.2...v0.18.3

## v0.18.2 - Shared Runtime Queue, Script Updates & Stability Fixes

### ✨ Added

- Shared global runtime queue for background semantic/code indexing work, with project-scoped queues, lease tracking, idle shutdown, and runtime status reporting
- Installer metadata persisted under `~/.knowns/install.json` for official shell and PowerShell installs
- Script-managed self-update flow for `knowns update`, including runtime version awareness and safer runtime restart behavior after upgrades
- MCP and runtime process logs under `~/.knowns/logs`, including per-process MCP logs and bounded log rotation

### 🔄 Changed

- MCP write paths now keep source-of-truth writes synchronous while enqueueing follow-up indexing, watch-triggered work, and reindex tasks to the shared runtime
- `knowns watch` and related background indexing flows now reuse shared runtime infrastructure instead of spawning redundant heavyweight work
- Shell installer now defaults to `~/.knowns/bin`, reducing the need for `sudo` and avoiding password prompts for script-managed installs on macOS and Linux
- Background indexing now uses queue pacing, debounce windows, duplicate coalescing, and lower ONNX thread defaults to reduce CPU spikes

### 🐛 Fixed

- Fixed runtime memory injection scoring so prompt hooks only inject memories with actual prompt relevance instead of matching only on runtime/context tokens
- Fixed Windows and race-sensitive test behavior by forcing test commands and MCP e2e flows to run with inline runtime mode instead of leaving shared runtime processes behind
- Fixed MCP/background work stability issues caused by inline indexing pressure, stale runtime reuse, and oversized shared log growth

---

### Contributors

@howznguyen

**Full Changelog**: https://github.com/knowns-dev/knowns/compare/v0.18.1...v0.18.2

## v0.18.1 - Release Docs, MCP Schema & Update Guidance

### ✨ Added

- Public docs for the `v0.18.0` feature set, including workspace-aware browser mode, code intelligence commands, code graph usage, and chat/runtime UX updates
- Installer README examples for pinning a specific version with `KNOWNS_VERSION` on both shell and PowerShell installers

### 🔄 Changed

- README and command guides now describe the shipped `knowns code ingest`, `knowns code watch`, `knowns code search`, `knowns code deps`, and `knowns code symbols` workflows
- Browser and web UI docs now cover workspace switching, project scanning, code watcher mode, and richer graph/chat capabilities

### 🐛 Fixed

- Fixed MCP tool schema for `knowns_retrieve` so `sourceTypes` declares array item types and no longer triggers `array schema missing items`
- Fixed Homebrew upgrade guidance to use `brew upgrade knowns-dev/tap/knowns` consistently in the CLI notifier and release/install messaging

---

### Contributors

@howznguyen

**Full Changelog**: https://github.com/knowns-dev/knowns/compare/v0.18.0...v0.18.1

## v0.18.0 - Workspace Switching, Code Intelligence & Chat UX

### ✨ Added

- AST-based code intelligence with symbol indexing, relationship edges, neighbor discovery, and dedicated tests
- New code-focused CLI and MCP capabilities for code search, code graph exploration, ingest, and file watching/sync flows
- Dedicated Code Graph experience in the web UI with new graph toolbar, legend, styling, and detail panel improvements
- Welcome page and chat timeline dialog to improve session onboarding and chat history navigation
- Workspace picker and server-side workspace switching improvements, including broader browser/server test coverage
- Search benchmarking utilities and new research/spec docs for retrieval, graph UX, and chat runtime upgrades

### 🔄 Changed

- Upgraded chat runtime and message rendering to better support tool calls, shell output, sub-agents, and richer thread interactions
- Expanded search engine and server routes to support code-aware retrieval alongside existing knowledge/doc search flows
- Updated browser/server integration, route wiring, and configuration plumbing to support workspace-aware behavior end to end
- Refined agent skill instructions and internal guidance for planning, implementation, review, research, and debugging workflows

### 🐛 Fixed

- OpenCode client/runtime handling is more robust, with added tests around client behavior and browser command flows
- Graph, search, validation, storage, and API route handling received targeted fixes alongside the new workspace/code features
- Publish workflow and package metadata were adjusted to keep release/build behavior aligned with the new runtime changes

---

### Contributors

@howznguyen

**Full Changelog**: https://github.com/knowns-dev/knowns/compare/v0.17.1...v0.18.0
