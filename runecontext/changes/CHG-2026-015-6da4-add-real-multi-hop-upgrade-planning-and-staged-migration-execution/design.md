# Design

## Overview
Model upgrade work as an umbrella over project-upgrade planning, staged apply-time execution, compatibility refinement, and explicit CLI self-update. Preserve the current UX split: bare `runectx upgrade` stays the read-only project-upgrade preview surface, `runectx upgrade apply` stays the only mutating project-upgrade command, and CLI binary update flows live under an explicit `runectx upgrade cli` namespace.

## Planned Sub-Changes
- `CHG-2026-016-e00c-add-multi-hop-upgrade-path-planning-and-preview-contracts` owns migration-registry shape, path search, and preview output.
- `CHG-2026-017-d68f-add-staged-per-hop-upgrade-execution-and-validation` owns staged apply-time execution, hop verification, rollback, and final replacement.
- `CHG-2026-022-58d8-refine-project-upgrade-compatibility-and-version-bump-only-semantics` owns the distinction between compatibility, upgradeability, and migration requirements.
- `CHG-2026-023-2423-add-explicit-cli-self-update-and-release-check-flows` owns CLI release checks and self-update behavior.

## Upgrade Model
- Project compatibility, project upgradeability, and explicit migration requirements remain distinct facts.
- A migration edge represents an approved rewrite transition that needs real migration logic, not just broad compatibility with the installed CLI release.
- Compatible older projects may remain pinned to an older `runecontext_version` until a user explicitly runs project upgrade.
- Compatible older projects should still be upgradeable even when no migration edge exists, in which case the plan is a version-bump-only upgrade.
- Multi-hop upgrades should be planned as deterministic ordered transitions across only the migration-required hops between the current version and the target version.
- Projects pinned to a newer version than the installed CLI supports fail closed with explicit guidance to upgrade the CLI binary.

## Command Model
- `runectx upgrade` remains the assessment path for project upgrades and must never mutate project files.
- `runectx upgrade apply` remains the only mutation entrypoint for project version transitions.
- `runectx upgrade cli` should be the explicit network-enabled preview/check surface for newer CLI releases.
- `runectx upgrade cli apply` should be the explicit network-enabled mutation surface for downloading and installing a newer CLI release.

## Safety Model
- Project apply-time work must happen against a staged project copy, with per-hop validation and per-hop verification before moving to the next transition.
- Plain version-bump-only project upgrades should still use the same staged safety boundary even when no migration hop runs.
- Real project files are replaced only after the full staged migration chain, final version update, and final validation succeed.
- Network-enabled CLI self-update behavior must stay explicit and must not change the semantics of core local project commands such as `validate`, `status`, or `doctor`.

## Relationship Model
- Use reciprocal `related_changes` links between the umbrella and all associated feature changes for navigability.
- Use `depends_on` only where planning or execution work truly relies on earlier upgrade semantics being present first.
