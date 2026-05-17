---
id: 14vi00
title: Implement batch embedding + parser pooling (v0.20.5)
status: done
priority: high
labels:
  - changelog
  - v0.20.5
  - performance
  - embedding
createdAt: '2026-05-11T10:45:33.063Z'
updatedAt: '2026-05-11T13:03:16.714Z'
timeSpent: 1865
assignee: '@me'
---
# Implement batch embedding + parser pooling (v0.20.5)

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Implement length-sorted content batching in onnx_runtime.go. Add adaptive batch sizes (64 for short content, 32 for long). Implement parser instance pools by extension in ast_indexer_parse.go to eliminate alloc/dealloc overhead per file.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented length-sorted adaptive batching (64 short / 32 long) in EmbedDocumentBatch. Implemented parser pooling per language extension in ast_indexer_parse.go to eliminate alloc/dealloc overhead.
<!-- SECTION:NOTES:END -->

