---
id: ywdl06
title: 'Integrate external MCP tools: doc-researcher, github-reader, database-inspector'
status: done
priority: high
labels:
  - mcp
  - integration
  - feature
createdAt: '2026-04-27T09:48:42.812Z'
updatedAt: '2026-04-27T10:04:30.647Z'
timeSpent: 865
---
# Integrate external MCP tools: doc-researcher, github-reader, database-inspector

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Integrate 3 MCP tools from antigravity-kit into Knowns: doc-researcher (web search + scraping), github-reader (GitHub API), database-inspector (SQL/Redis inspection)
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Create research.go handler with search_latest_syntax tool
- [x] #2 Create github.go handler with get_file_content, search_repos, list_commits tools
- [x] #3 Create database.go handler with list_tables, inspect_schema, explain_query tools
- [x] #4 Update server.go to register 3 new handlers
- [x] #5 Update go.mod with required dependencies
- [x] #6 Add unit tests for each handler
- [x] #7 Update ai/mcp documentation
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
AC1-5 completed: Created research.go, github.go, database.go handlers. Updated server.go to register new handlers. Updated go.mod with go-github v62, oauth2, redis, gorm dependencies.
<!-- SECTION:NOTES:END -->

