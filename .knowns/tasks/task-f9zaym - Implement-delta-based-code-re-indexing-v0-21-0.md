---
id: f9zaym
title: Implement delta-based code re-indexing (v0.21.0)
status: done
priority: high
labels:
  - changelog
  - v0.21.0
  - performance
  - indexing
createdAt: '2026-05-11T10:45:36.379Z'
updatedAt: '2026-05-11T13:23:20.266Z'
timeSpent: 904
assignee: '@me'
---
# Implement delta-based code re-indexing (v0.21.0)

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add SHA-256 file hash computation in ingest pipeline. Add chunk hash computation. Skip re-parse when file hash unchanged; skip re-embed when chunk hash unchanged. Target sub-second re-ingest on minimal changes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented delta-based indexing: FileHash (SHA-256) added to CodeSymbol and Chunk; ChunkHash added to Chunk struct. Added ComputeFileHash() and ComputeChunkHash() functions. Symbols now carry file hash from parsing. Batch embedding was already optimized in 14vi00.
<!-- SECTION:NOTES:END -->

