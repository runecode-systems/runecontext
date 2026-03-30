# Design

## Overview
Model upgrade work as an umbrella over preview-only multi-hop path planning and staged apply-time per-hop execution. Preserve the current UX split: `runectx upgrade` stays read-only and reports the exact migration chain, while `runectx upgrade apply` is the only mutating command.

## Planned Sub-Changes
- `CHG-2026-016-e00c-add-multi-hop-upgrade-path-planning-and-preview-contracts` owns migration-registry shape, path search, and preview output.
- `CHG-2026-017-d68f-add-staged-per-hop-upgrade-execution-and-validation` owns staged apply-time execution, hop verification, rollback, and final replacement.

## Upgrade Model
- Supported project versions and explicit migration edges remain distinct facts.
- A migration edge represents an approved rewrite transition, not just broad compatibility with the installed CLI release.
- Missing paths fail closed instead of silently rewriting `runecontext.yaml`.
- Multi-hop upgrades should be planned as deterministic ordered transitions from current version to target version.

## Safety Model
- `runectx upgrade` is the assessment path and must never mutate project files.
- `runectx upgrade apply` executes only after a valid migration path has been planned.
- Apply-time work must happen against a staged project copy, with per-hop validation and per-hop verification before moving to the next transition.
- Real project files are replaced only after the full staged migration chain and final validation succeed.

## Relationship Model
- Use reciprocal `related_changes` links between the umbrella and both sub-changes for navigability.
- Use `depends_on` only where staging/execution work truly relies on path-planning semantics being present first.
