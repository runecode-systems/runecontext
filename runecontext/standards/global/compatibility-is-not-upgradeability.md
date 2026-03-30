---
schema_version: 1
id: global/compatibility-is-not-upgradeability
title: Compatibility Is Not Upgradeability
status: active
tags:
  - compatibility
  - upgrade
  - metadata
  - cli
---

# Compatibility Is Not Upgradeability

## Intent

Keep broad version compatibility distinct from explicit supported migration paths so upgrade tooling stays honest and fail-closed.

## Requirements

- Tooling must distinguish between versions that are broadly supported by the current release and version transitions that have an explicit registered upgrade path.
- Machine-readable metadata should publish compatibility coverage and explicit upgrade edges as separate fields.
- Upgrade preview and apply flows must fail closed when no registered upgrade path exists for the requested transition, even if both endpoint versions are otherwise recognized.
- Docs, adapters, and release metadata must not collapse compatibility and upgradeability into one ambiguous notion of support.

## Rationale

Compatibility answers whether a repository version is understood by the current release. Upgradeability answers whether RuneContext knows how to move that repository safely to another version. Conflating the two encourages unsafe migrations and misleading automation.

## Implementation Notes

- Prefer additive machine-readable fields over prose-only explanations.
- Keep upgrade-edge publication deterministic so downstream tools can reason about migrations safely.
