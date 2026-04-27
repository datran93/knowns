---
id: 1tuujo
title: Agent Efficiency Config Schema + Defaults
status: done
priority: high
labels:
  - from-spec
createdAt: '2026-04-27T14:46:31.287Z'
updatedAt: '2026-04-27T16:26:06.615Z'
timeSpent: 0
spec: specs/agent-efficiency-upgrade-tier-1-tier-2
fulfills:
  - NFR-2
---
# Agent Efficiency Config Schema + Defaults

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add agentEfficiency config section with all 8 feature flags defaulting to ON (opinionated defaults). All features opt-out per-feature.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Add agentEfficiency config section to config.go
- [x] #2 All 8 feature flags with default: true (opinionated)
- [x] #3 Per-feature opt-out support
- [x] #4 Schema validated against spec
- [x] #5 Default values match spec Config Schema section
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Config schema implemented in internal/models/config.go (AgentEfficiencySettings with all 8 feature flags defaulting to ON) + config_store.go (defaults applied when nil). Build clean. All 5 ACs verified.
<!-- SECTION:NOTES:END -->

