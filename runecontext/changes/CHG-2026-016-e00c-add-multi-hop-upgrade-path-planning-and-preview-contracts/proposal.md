## Summary
Add multi-hop upgrade path planning and preview contracts

## Problem
`runectx upgrade` currently treats explicit migration hops as the only way to move a project forward. That makes preview fail closed even when a project is merely older-but-compatible and only needs its pinned `runecontext_version` bumped. It also makes it hard to tell users the difference between “this project can be upgraded now without migrations”, “this project needs real migration hops”, and “this project is newer than your CLI and you must upgrade the binary first.”

## Proposed Change
Refine upgrade planning so it distinguishes compatibility from migration requirements, uses deterministic ordered paths only for real migration hops, and supports zero-hop version-bump-only project upgrades when the project is compatible with the installed CLI but no migration edge is required.

## Why Now
The current branch and recent dogfood behavior sharpened the distinction between broad compatibility and explicit migration edges. Preview needs to expose that distinction clearly so teams can keep compatible older projects pinned until they explicitly upgrade them, while still failing closed when a project requires a newer CLI release.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- `runectx upgrade` remains the assessment path and acts as the dry-run view for migration planning.

## Out of Scope
- Mutating project files during preview.
- Staged file execution, rollback, or per-hop apply logic.
- CLI binary self-update flows.

## Impact
The change gives users and adapters a stable preview contract for project upgrade readiness, including version-bump-only upgrades, migration-required upgrades, and project-newer-than-cli failures, before any write-capable migration work runs.
