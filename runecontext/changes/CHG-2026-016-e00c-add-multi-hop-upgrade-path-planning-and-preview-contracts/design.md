# Design

## Overview
Replace the current single-edge target check with deterministic project-upgrade planning that separates compatibility, upgradeability, and migration requirements. The preview command must stay read-only and compute either a zero-hop version-bump-only plan or a concrete ordered sequence of migration-required hops from the project `runecontext_version` to the chosen target version. Planning should preserve the distinction between supported project versions and explicit migration edges, report hop_count, ordered hop transitions, and per-hop plan actions when migrations are required, and fail closed with explicit CLI-upgrade guidance when the project is newer than the installed binary.

## Planning Rules
- Migration edges should be modeled explicitly and searched as a graph rather than treated as exact target membership checks only.
- Explicit migration edges should represent transitions that need real migration logic, not plain permission to bump `runecontext_version`.
- Compatible older projects should be able to preview an upgrade without any migration hop when only a version bump is required.
- Path selection must be deterministic so preview output and apply behavior are stable across runs.
- Preview should support direct-edge and multi-hop migration paths using the same planner.
- Projects newer than the installed CLI should fail closed with clear guidance to upgrade the CLI binary.

## Preview Contract
- `runectx upgrade` remains read-only and should act as the reviewable dry-run surface for upgrade planning.
- Preview output should include at least the current version, target version, state, hop count, ordered hop transitions when present, and plan actions that clearly distinguish version-bump-only upgrades from migration-required upgrades.
- Preview should keep existing flat structured output conventions so scripts and adapters can consume the new fields without a separate rendering path.

## Migration Registry Direction
- The registry should be able to describe ordered migration transitions, not only whether a direct edge exists.
- Registry entries should be reusable by apply-time migration execution without duplicating transition knowledge in multiple places.
- The planner should be able to decide that no migration edge is required for a compatible version-bump-only upgrade.
