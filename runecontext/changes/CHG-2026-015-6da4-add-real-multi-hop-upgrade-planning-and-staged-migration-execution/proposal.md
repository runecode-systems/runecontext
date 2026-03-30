## Summary
Add real multi-hop upgrade planning and staged migration execution

## Problem
The current upgrade flow only supports a single exact transition from the project's current `runecontext_version` to one requested target version. That model cannot express or safely execute multi-step upgrades across several declared migration edges, even when each individual hop is known and reviewable. It also leaves no structured place for per-hop migration logic and per-hop verification beyond a final version rewrite.

## Proposed Change
Track the upgrade enhancement as an umbrella over two linked deliverables:

- `CHG-2026-016-e00c-add-multi-hop-upgrade-path-planning-and-preview-contracts` for migration-path planning, preview contracts, and fail-closed path selection.
- `CHG-2026-017-d68f-add-staged-per-hop-upgrade-execution-and-validation` for staged apply-time execution, per-hop verification, and final atomic replacement.

## Why Now
Recent branch work exposed a practical gap between compatibility reporting and upgrade execution. The metadata and release-version work clarified that explicit upgrade edges are migration contracts, not just compatibility hints, and that future alpha-line or schema transitions will need a real path planner plus safe staged execution rather than ad hoc manual version edits.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- `runectx upgrade` remains the reviewable assessment path and should not mutate project files.
- `runectx upgrade apply` remains the only mutation entrypoint for version transitions.

## Out of Scope
- Replacing version-range compatibility checks with migration-path logic.
- Auto-bumping `runecontext_version` when no explicit migration path exists.
- Defining downstream-tool-specific migration semantics outside the CLI and project tree.

## Impact
The umbrella keeps path planning, migration execution, and verification rules explicitly linked while preserving the current user model: preview first, apply explicitly, and fail closed when no approved migration path exists.
