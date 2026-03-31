# Design

## Overview
Separate project compatibility, project upgradeability, and migration requirements. Older-but-compatible projects should remain valid under newer CLIs while still being upgradeable. `runectx upgrade` must stay the read-only project preview surface. If a project is older than the installed CLI target but no migration logic is required, preview should report an upgradeable version-bump-only plan and `runectx upgrade apply` should only rewrite `runecontext_version`. Explicit migration hops must represent real migration logic only. If the project is newer than the installed CLI, planning and validation should fail closed with explicit guidance to upgrade the CLI binary rather than suggesting project downgrade.

## Semantics
- Compatibility answers whether the installed CLI may operate on the project safely as-is.
- Upgradeability answers whether the project can move forward to the chosen target version under the current CLI.
- Migration requirements answer whether any real migration hop must run between the current project version and the target version.

## Planning Rules
- A compatible older project may remain unchanged until a user explicitly runs project upgrade.
- A compatible older project with no migration-required interval should preview as upgradeable with `hop_count=0` and a plan action that bumps the pinned `runecontext_version`.
- A compatible older project with one or more migration-required intervals should preview the ordered migration hops and any final pinned-version update required to reach the target.
- A project pinned to a newer version than the installed CLI supports should fail closed and tell the user to upgrade the CLI binary.

## Diagnostics Direction
- `validate` and `doctor` should report that compatible older projects can be upgraded without treating them as unsupported.
- `validate` and `doctor` should report that newer-than-cli projects require a newer CLI binary rather than implying that the project itself is invalid.

## Shape Rationale
- Full mode was requested explicitly to deepen the change.
- Minimum mode is sufficient for the current size and risk signal.
