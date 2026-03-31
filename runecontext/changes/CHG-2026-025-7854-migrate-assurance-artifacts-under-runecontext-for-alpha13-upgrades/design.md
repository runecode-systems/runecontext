# Design

## Overview
Alpha.13 becomes the canonical assurance-layout release. All new assurance artifacts live under `runecontext/assurance/`. Upgrade apply performs a staged `alpha.12 -> alpha.13` migration that moves legacy root-level assurance artifacts into the canonical project tree, rewrites baseline backfill references that still point at `assurance/backfill/...`, validates the staged result, and replaces the live tree only after the final staged state succeeds.

## Canonical Layout
- `runecontext/assurance/baseline.yaml`
- `runecontext/assurance/backfill/*.json`
- `runecontext/assurance/receipts/context-packs/*.json`
- `runecontext/assurance/receipts/changes/*.json`
- `runecontext/assurance/receipts/promotions/*.json`
- `runecontext/assurance/receipts/verifications/*.json`

## Writer Rules
- `runectx init` creates a fresh alpha.13 project with the current canonical layout only.
- `assurance enable`, `assurance capture`, `assurance backfill`, and any shared receipt writers must emit only `runecontext/assurance/...` artifacts once alpha.13 is installed.
- CLI dry-run and explain output should also name the canonical path.

## Migration Rules
- The real migration edge is `0.1.0-alpha.12 -> 0.1.0-alpha.13`.
- Migration runs only inside the staged upgrade tree.
- Migration moves the entire legacy `assurance/` subtree into `runecontext/assurance/`.
- Migration rewrites baseline `value.imported_evidence[].path` entries that reference `assurance/backfill/...` to `runecontext/assurance/backfill/...`.
- If the canonical destination already contains conflicting files, migration should fail closed and leave the live tree untouched.

## Validation Direction
- Post-alpha.13 validation treats `runecontext/assurance/` as the sole canonical assurance location.
- Normal alpha.13 validation may fail on legacy root-level assurance layouts; dedicated legacy guidance is not required for this alpha release as long as `upgrade apply` can perform the staged migration.

## Shape Rationale
- Full mode was requested explicitly to deepen the change.
- Minimum mode is sufficient for the current size and risk signal.
