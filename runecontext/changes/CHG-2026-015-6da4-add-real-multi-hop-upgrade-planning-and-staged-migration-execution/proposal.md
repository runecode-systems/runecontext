## Summary
Add real multi-hop upgrade planning and staged migration execution

## Problem
The current upgrade flow still conflates project compatibility, project upgradeability, and explicit migration requirements. That makes it hard to support teams that intentionally keep a project pinned to an older but still compatible `runecontext_version`, while also making it hard to distinguish plain version bumps from upgrades that require real migration logic. The command surface also lacks an explicit CLI self-update lane, so there is no clean place to tell users when their binary is too old for a project or when a newer `runectx` release is available.

## Proposed Change
Track the upgrade enhancement as an umbrella over four linked deliverables:

- `CHG-2026-016-e00c-add-multi-hop-upgrade-path-planning-and-preview-contracts` for migration-path planning, preview contracts, and fail-closed path selection.
- `CHG-2026-017-d68f-add-staged-per-hop-upgrade-execution-and-validation` for staged apply-time execution, per-hop verification, and final atomic replacement.
- `CHG-2026-022-58d8-refine-project-upgrade-compatibility-and-version-bump-only-semantics` for separating compatibility from migration requirements, allowing version-bump-only project upgrades without synthetic migration edges, and failing closed when the project is newer than the CLI.
- `CHG-2026-023-2423-add-explicit-cli-self-update-and-release-check-flows` for explicit CLI release checks and self-update commands that stay separate from project upgrade behavior.

## Why Now
Recent dogfood work on the alpha.12 branch exposed a practical gap between release versioning, project version pins, and the project-upgrade UX. Teams need to be able to keep older but still compatible projects unchanged until they explicitly run project upgrade, while newer projects must fail closed under older CLIs with clear instructions to upgrade the binary. At the same time, CLI self-update needs to be explicit and network-bounded rather than folded into the normal local project workflow.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- `runectx upgrade` remains the reviewable assessment path and should not mutate project files.
- `runectx upgrade apply` remains the only mutation entrypoint for version transitions.
- Explicit migration hops represent real migration logic only; plain version bumps should not require synthetic hops.
- CLI binary self-update should be explicit under the upgrade command family without changing the meaning of bare project-upgrade commands.

## Out of Scope
- Hidden background network calls in normal core project commands.
- Requiring all compatible older projects to upgrade immediately just because a newer CLI exists.
- Defining downstream-tool-specific migration semantics outside the CLI and project tree.

## Impact
The umbrella keeps project upgrade planning, staged migration execution, compatibility semantics, and CLI self-update rules explicitly linked while preserving a clear user model: preview project upgrades first, apply them explicitly, use explicit CLI update commands when the binary itself should change, and fail closed when a project requires a newer CLI release.
