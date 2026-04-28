---
id: iobg9r
title: Create kn-design skill file
status: done
priority: medium
labels:
  - from-spec
  - go-mode
  - skill
createdAt: '2026-04-28T16:51:15.724Z'
updatedAt: '2026-04-28T16:52:49.179Z'
timeSpent: 61
spec: specs/kn-design-skill
fulfills:
  - AC-1
  - AC-2
  - AC-3
  - AC-4
  - AC-7
---
# Create kn-design skill file

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Create the kn-design skill file at /Users/datran/Project/knowns/internal/instructions/skills/kn-design/skill.md

Requirements:
- FR-1: Design doc generation with Architecture Decisions, Component Breakdown, Data Flow sections
- FR-2: Codebase context awareness (scans existing patterns, checks architecture docs)
- FR-3: Interactive review (show draft, user requests changes, save on approval)
- FR-4: Standalone invocation via /kn-design specs/<name>

Design doc output: designs/<spec-name>.md
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 1. Read existing skill patterns (kn-plan, kn-spec) for structure reference
- [x] #2 2. Create directory: internal/instructions/skills/kn-design/
- [x] #3 3. Write skill.md with: FR-1 (Architecture Decisions, Component Breakdown, Data Flow sections), FR-2 (codebase context awareness), FR-3 (interactive review loop)
- [x] #4 4. Add frontmatter: name, description fields
- [x] #5 5. Add skill to skills directory listing
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Created /Users/datran/Project/knowns/internal/instructions/skills/kn-design/skill.md with:
- Frontmatter (name: kn-design, description)
- Phase 1: Validate spec (approved check)
- Phase 2: Gather codebase context (patterns, architecture docs)
- Phase 3: Draft design (Architecture Decisions, Component Breakdown, Data Flow)
- Phase 4: Interactive review (show draft, user feedback, iterate)
- Phase 5: Save design document (designs/<spec-name>.md)
- Phase 6: Report completion + suggest next steps
- Integration section for kn-go auto-invoke
<!-- SECTION:NOTES:END -->

