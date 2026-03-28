---
schema_version: 1
id: release/repo-first-canonical-distribution
title: Repo First Canonical Distribution
status: active
tags:
  - release
  - distribution
  - install
  - provenance
---

# Repo First Canonical Distribution

## Intent

Preserve the reviewable repo bundle as the canonical RuneContext release artifact even when convenience binaries are available.

## Requirements

- Official releases must publish canonical repo bundle artifacts.
- Standalone CLI binaries are convenience delivery formats and must not replace the repo bundle as the primary audit surface.
- Release verification guidance must keep the repo-bundle lane explicit.
- Release metadata should keep bundle, schema, adapter, and binary assets tied to the same release provenance set.

## Rationale

RuneContext is a repo-native system. The full release bundle is the artifact that best preserves portability, reviewability, and long-term audit value.

## Implementation Notes

- Keep install docs split clearly between lightweight and fully verified paths.
- Use the same signed release set to tie all published artifacts together.
