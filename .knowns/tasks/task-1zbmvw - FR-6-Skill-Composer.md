---
id: 1zbmvw
title: 'FR-6: Skill Composer'
status: done
priority: medium
labels:
  - from-spec
createdAt: '2026-04-27T14:46:31.277Z'
updatedAt: '2026-04-27T15:56:48.742Z'
timeSpent: 0
spec: specs/agent-efficiency-upgrade-tier-1-tier-2
fulfills:
  - AC-5.1
  - AC-5.2
---
# FR-6: Skill Composer

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Composite skill template system with $1, $2 positional args and built-in templates (full-review, implement-and-test). Phase 5 per implementation order.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Template system with skill composition syntax
- [x] #2 { "skill": "kn-xxx", "args": {...} } in steps array
- [x] #3 Variable interpolation: $1, $2 positional args
- [x] #4 Built-in composite templates: full-review, implement-and-test
- [x] #5 Sequential execution engine
- [x] #6 Tests for skill composer
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Done: Created skill_compose.go handler with run/list actions. Built-in templates: full-review, implement-and-test, review-and-commit. $1, $2 positional interpolation. Registered in server.go. Build pass.
<!-- SECTION:NOTES:END -->

