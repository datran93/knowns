---
id: z4pjzt
title: 'Add Java/Rust/C# tree-sitter support (v0.20.5)'
status: done
priority: high
labels:
  - changelog
  - v0.20.5
  - tree-sitter
  - code-intelligence
createdAt: '2026-05-11T10:02:43.374Z'
updatedAt: '2026-05-11T11:53:58.670Z'
timeSpent: 3870
assignee: '@me'
---
# Add Java/Rust/C# tree-sitter support (v0.20.5)

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add tree-sitter-java, tree-sitter-rust, tree-sitter-c-sharp Go bindings. Extend ast_indexer.go to detect .java, .rs, .cs files and extract symbols (classes, structs, enums, traits, impl blocks, constructors) and edges (imports, method ownership).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Added Java/Rust/C# tree-sitter support: .java, .rs, .cs files now indexed. Extracted symbols: classes, structs, enums, traits, impl blocks, constructors, fields. Updated isSupportedCodeSymbolKind to include constructor, enum, field, trait.
<!-- SECTION:NOTES:END -->

