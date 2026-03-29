---
schema_version: 1
id: cli/completion-and-metadata-from-canonical-registry
title: Completion And Metadata From Canonical Registry
status: active
tags:
  - cli
  - completion
  - metadata
  - registry
---

# Completion And Metadata From Canonical Registry

## Intent

Keep command metadata, completion behavior, and adapter suggestion surfaces aligned by deriving them from one canonical registry.

## Requirements

- Shell completion, machine-readable completion metadata, and adapter-facing command metadata should derive from the same typed command registry.
- New commands and flags must update the canonical registry instead of maintaining parallel metadata sources.
- Dynamic suggestions should remain read-only and repository-aware.
- Metadata consumers must not redefine command semantics outside the canonical registry.

## Rationale

Parallel command models drift over time and create inconsistent UX across shells, scripts, and adapters. One registry keeps the CLI surface reviewable and synchronized.

## Implementation Notes

- Add suggestion providers and enum metadata centrally.
- Prefer derived documentation over hand-maintained shadow command references.
