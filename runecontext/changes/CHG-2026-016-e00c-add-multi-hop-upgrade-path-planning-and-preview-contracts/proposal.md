## Summary
Add multi-hop upgrade path planning and preview contracts

## Problem
`runectx upgrade` currently reasons about only one direct from-version to to-version transition. That prevents the CLI from explaining or approving an upgrade path when a project is multiple declared migration hops behind the installed release, even when each intermediate edge is explicit and supported.

## Proposed Change
Replace exact-edge-only upgrade planning with deterministic multi-hop path planning and extend preview output so users can review the full ordered migration chain without mutating the project tree.

## Why Now
The current branch and recent metadata work sharpened the distinction between broad compatibility and explicit migration edges. The upgrade preview surface needs to expose that distinction clearly before staged multi-hop apply logic is introduced.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- `runectx upgrade` remains the assessment path and acts as the dry-run view for migration planning.

## Out of Scope
- Mutating project files during preview.
- Staged file execution, rollback, or per-hop apply logic.
- Auto-approving version bumps when a migration path is missing.

## Impact
The change gives users and adapters a stable preview contract for upgrade paths, with explicit ordered hops and fail-closed path selection before any write-capable migration work runs.
