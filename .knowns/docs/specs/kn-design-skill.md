---
title: kn-design Skill
description: Specification for kn-design skill - design step from spec before planning
createdAt: '2026-04-28T16:46:05.836Z'
updatedAt: '2026-04-28T16:50:17.960Z'
tags:
  - spec
  - approved
  - skill
  - design
  - workflow
---

# kn-design Skill Specification

## Overview

The `kn-design` skill produces a design document from an approved spec before planning begins. It extracts architecture decisions, component breakdown, and data flow from the spec + codebase context, then presents a draft for interactive review. Once saved, the design informs `/kn-plan` or `/kn-go` execution.

## Cross-Domain Impact

- CLI: `/kn-design` command (new skill invocation)
- MCP: Design document read/creation operations
- Storage: `designs/<spec-name>.md` file in project
- Workflow: Adds design step between spec approval and planning

## Locked Decisions

- D1: Both standalone (`/kn-design specs/<name>`) and integrated into `kn-go` (auto-invokes if design doesn't exist yet)
- D2: `designs/<spec-name>.md` as separate doc in designs folder
- D3: Full design doc — architecture decisions + component breakdown + data flow
- D4: Interactive review — skill shows draft, user requests changes, then saves
- D5: Spec + codebase context for alignment (existing code patterns and architecture)

## Requirements

### FR-1: Design Document Generation

The skill takes a spec path and produces a design document containing:
- **Architecture decisions**: Key choices that constrain implementation (why decisions)
- **Component breakdown**: What each component does, its responsibilities
- **Data flow**: How data moves through the system (inputs, processes, outputs)

### FR-2: Codebase Context Awareness

Before drafting, the skill:
- Scans related existing code patterns in the codebase
- Checks referenced architecture docs
- Ensures design aligns with existing patterns and conventions

### FR-3: Interactive Review

The skill:
1. Shows draft design to user
2. Allows user to request changes/refinements
3. Updates draft based on feedback
4. Saves only after user approves

### FR-4: Standalone Invocation

User can call `/kn-design specs/<name>` explicitly to produce a design.

### FR-5: Integration with kn-go

When `kn-go` runs on a spec without an existing design doc:
1. Auto-invokes `kn-design` to produce design first
2. Then proceeds with planning
3. If design already exists, skips generation and uses existing

## Acceptance Criteria

- [ ] AC-1: `/kn-design specs/<name>` generates design at `designs/<spec-name>.md`
- [ ] AC-2: Design document contains: Architecture Decisions, Component Breakdown, Data Flow sections
- [ ] AC-3: Skill shows draft design for user review before saving
- [ ] AC-4: User can request changes and see updated draft
- [ ] AC-5: `kn-go` auto-invokes `kn-design` if design doesn't exist
- [ ] AC-6: `kn-go` skips design step if design already exists
- [ ] AC-7: Design accounts for existing code patterns and architecture

## Scenarios

### Scenario 1: Standalone Design Generation

**Given** a spec exists at `specs/my-feature` with approved status
**When** user runs `/kn-design specs/my-feature`
**Then** skill produces draft design, shows it for review, and saves to `designs/my-feature.md` after approval

### Scenario 2: kn-go Auto-Design

**Given** a spec exists at `specs/my-feature` but no design at `designs/my-feature.md`
**When** user runs `/kn-go specs/my-feature`
**Then** `kn-go` auto-invokes `kn-design` to generate design first, then proceeds with planning

### Scenario 3: Design Already Exists

**Given** a spec exists at `specs/my-feature` AND design exists at `designs/my-feature.md`
**When** user runs `/kn-go specs/my-feature`
**Then** `kn-go` skips design generation and uses existing design

### Scenario 4: Interactive Refinement

**Given** skill has produced a draft design
**When** user reviews draft and requests changes
**Then** skill updates draft based on feedback and re-presents for review

## Technical Notes

- Design documents are stored in `designs/` folder (not in specs folder)
- File naming matches spec name: `designs/<spec-name>.md`
- `kn-go` checks for `designs/<spec-name>.md` existence as gate before auto-invoking
- Interactive review loop continues until user approves or cancels
- **Git-tracked**: designs folder is git-tracked (same as spec docs)
- **No versioning**: design is immutable once approved; if requirements change, regenerate
