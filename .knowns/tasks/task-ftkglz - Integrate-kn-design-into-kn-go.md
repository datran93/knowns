---
id: ftkglz
title: Integrate kn-design into kn-go
status: done
priority: medium
labels:
  - from-spec
  - go-mode
  - skill
createdAt: '2026-04-28T16:51:24.658Z'
updatedAt: '2026-04-28T16:53:31.273Z'
timeSpent: 31
spec: specs/kn-design-skill
fulfills:
  - AC-5
  - AC-6
---
# Integrate kn-design into kn-go

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Modify kn-go skill to auto-invoke kn-design when design doesn't exist.

Requirements:
- FR-5: kn-go auto-invokes kn-design if design doesn't exist
- FR-6: kn-go skips design step if design already exists

Check designs/<spec-name>.md existence before invoking kn-design.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 1. Read current kn-go skill.md
- [x] #2 2. Add design check in Phase 1: if designs/<spec-name>.md doesn't exist, invoke kn-design
- [x] #3 3. Document the auto-invoke behavior in kn-go process
- [x] #4 4. Add Phase 0.5 for design phase
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Updated kn-go skill.md:
- Added "When NOT to Use" case for design review before planning
- Added Phase 0.5: Design Check before Phase 1
- Design check checks designs/<spec-name>.md existence
- If design doesn't exist, auto-invokes /kn-design specs/<name>
- If design exists, skips design generation
<!-- SECTION:NOTES:END -->

