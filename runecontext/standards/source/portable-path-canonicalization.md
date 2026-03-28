---
schema_version: 1
id: source/portable-path-canonicalization
title: Portable Path Canonicalization
status: active
tags:
  - source
  - portability
  - paths
  - validation
---

# Portable Path Canonicalization

## Intent

Keep path-bearing RuneContext artifacts portable across operating systems, machines, and repository locations.

## Requirements

- Stored RuneContext paths must use forward-slash notation.
- Authoritative paths should be relative to the RuneContext content root unless a contract explicitly says otherwise.
- Absolute, drive-qualified, UNC, or traversal-based paths must not appear where portability is required.
- Path normalization must preserve containment boundaries and reject escape attempts.

## Rationale

Portable project knowledge fails if file references depend on one machine's filesystem semantics. Canonical path rules keep validation, hashing, and cross-platform collaboration stable.

## Implementation Notes

- Normalize early at input boundaries.
- Reject path values that cannot be represented portably instead of attempting lossy repair.
