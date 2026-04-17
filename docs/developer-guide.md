# Knowns Developer Guide

Technical notes for contributors working on the current Go codebase.

---

## Tech Stack

| Layer | Technology |
| ----- | ---------- |
| Runtime | Go |
| CLI | Cobra |
| Storage | File-based `.knowns/` data |
| Web server | Go HTTP server |
| Web UI | React app embedded into the binary |
| AI integration | MCP server + generated instruction files |
| Search | Keyword + optional semantic search (ONNX Runtime + embedding models) |
| Code intelligence | AST-based indexing (Go, TypeScript, JavaScript, Python) |

---

## Codebase Map

Main repo structure:

```text
cmd/knowns/              CLI entrypoint
internal/cli/            Cobra commands (including update, download_setup)
internal/models/         Domain models and config structs
internal/storage/        Persistence for tasks, docs, versions, config, memories
internal/server/         HTTP server, routes, browser backend
internal/mcp/            MCP server implementation
  internal/mcp/handlers/ MCP tool handler implementations (code.go, search.go, etc.)
internal/search/         Hybrid/semantic/keyword search engine
  engine.go, chunker.go, index.go, vecstore.go, sqlite_vecstore.go, sync.go (runtime queue integration), types.go
internal/validate/       Validation engine for refs and integrity
internal/registry/       Global project registry (workspace discovery)
internal/runtimequeue/   Shared background work queue for indexing, watch, reindex jobs
  Routes by project, stores state under <project>/.knowns/runtime/queue.json
internal/runtimememory/   Runtime memory hook injection for session context
internal/runtimeinstall/ Runtime installation helpers (Claude Code, Codex, Kiro, OpenCode)
internal/codegen/        Code generation and skill syncing
internal/util/           Utilities (version, update notifier, install metadata)
install/                 Platform install scripts (install.sh, install.ps1, uninstall.sh, uninstall.ps1)
ui/                      React UI source and embedded assets
tests/                   End-to-end and integration coverage
```

---

## Important Runtime Behavior

### Browser UI

- `knowns browser` starts the local web server
- default port is `6420` unless overridden by `settings.serverPort` or `--port`; fallback tries 6420, 6421, 6422
- the browser only auto-opens when `--open` is passed
- auto-registers the project in the global registry on startup
- auto-ingests code if semantic search is configured but no code chunks exist

### Sync

- `knowns sync` handles instruction-file and skill syncing
- `knowns agents --sync` remains as a compatibility path for instruction files
- platform filtering is done with `--platform` or `settings.platforms`

### Search

- `knowns search <query>` performs search
- `knowns retrieve <query>` retrieves ranked context with citations
- `knowns search --reindex` rebuilds the index (routes through runtime queue if daemon is running)
- `knowns search --status-check` shows semantic-search status
- `knowns search --install-runtime` downloads ONNX Runtime; install scripts invoke it automatically after binary install
- `knowns search --setup` enables semantic search after ONNX and model are ready
- Semantic search requires ONNX Runtime (`~/.knowns/lib/`) and an embedding model (`~/.knowns/models/`)
- Search index is stored in `.knowns/.search/` as a SQLite database

### Self-Update

- `knowns update` detects install method (Homebrew/npm/script) and runs the appropriate upgrade
- For script-managed installs (`~/.knowns/bin/`), downloads and replaces the binary directly
- After update, runs `knowns sync` to refresh MCP configs (`.mcp.json`, `.kiro/settings/mcp.json`, `opencode.json`)
- `internal/util/update_notifier.go` provides background version checking

---

## Config Model

Project config is defined in `internal/models/config.go`.

Important fields include:

- `name`, `id`, `createdAt`
- `settings.defaultAssignee`, `settings.defaultPriority`, `settings.defaultLabels`
- `settings.statuses`, `settings.statusColors`, `settings.visibleColumns`
- `settings.semanticSearch`
- `settings.serverPort`
- `settings.platforms`
- `settings.autoSyncOnUpdate`
- `settings.enableChatUI`
- `settings.opencodeServer`, `settings.opencodeModels`

Supported platform IDs: `claude-code`, `opencode`, `codex`, `kiro`, `gemini`, `copilot`, `agents`.

---

## CLI Conventions

- Root commands are registered under `internal/cli/`
- command help is the source of truth for public syntax
- shorthand behavior exists for some groups, for example `knowns task <id>` and `knowns doc <path>`
- prefer updating docs from actual `knowns ... --help` output to avoid drift

---

## Development Commands

```bash
make build
make dev
make test
make test-e2e
make lint
make ui
```

Use `go run ./cmd/knowns --help` and subcommand help while updating docs or CLI behavior.

---

## Documentation Rule

This repo has accumulated older Node/TypeScript-era docs. If code and docs disagree, trust the Go implementation and CLI help first, then update docs to match.

---

## Related

- [Architecture](../ARCHITECTURE.md) - Higher-level system overview
- [Configuration](./configuration.md) - Runtime config fields
- [Command Reference](./commands.md) - Current CLI surface
