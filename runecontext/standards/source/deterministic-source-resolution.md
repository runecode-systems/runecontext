---
schema_version: 1
id: source/deterministic-source-resolution
title: Deterministic Source Resolution
status: active
tags:
  - source
  - resolution
  - determinism
  - trust
---

# Deterministic Source Resolution

## Intent

Ensure the same RuneContext source configuration resolves to the same materialized tree and trust posture across environments.

## Requirements

- Source resolution must produce explicit source mode, source reference, and verification posture metadata.
- Pinned and verified source modes should be preferred over mutable or unverified references when stronger guarantees are needed.
- Resolution must fail closed when configured trust requirements are not met.
- Snapshot limits and containment checks must protect resolution workflows from unsafe or unbounded local inputs.

## Rationale

Context packs, assurance receipts, and validation results depend on stable source interpretation. If resolution is ambiguous or environment-dependent, higher-level RuneContext guarantees become untrustworthy.

## Implementation Notes

- Surface verification posture explicitly in machine output.
- Preserve a sharp distinction between embedded, path, and git-backed source modes.
