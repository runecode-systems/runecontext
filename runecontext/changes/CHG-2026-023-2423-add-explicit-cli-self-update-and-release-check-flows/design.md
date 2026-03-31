# Design

## Overview
Keep bare `runectx upgrade` focused on the managed project and add an explicit CLI-update namespace under the same command family. `runectx upgrade cli` should be the explicit network-enabled preview/check surface for newer CLI releases, while `runectx upgrade cli apply` should download and install the selected newer CLI release. This preserves the current meaning of project upgrade while making CLI release checks and self-update behavior explicit. Any newer-release notification should come from explicit network-enabled flows or clearly opted-in status/version checks, not hidden background network access in core project commands.

## Command Model
- `runectx upgrade` remains the read-only project upgrade preview command.
- `runectx upgrade apply` remains the mutating project upgrade command.
- `runectx upgrade cli` checks for a newer CLI release and previews the selected install action.
- `runectx upgrade cli apply` downloads and installs the selected newer CLI release.

## Network Boundary
- CLI release discovery and self-update may use network access by design, but that access must stay explicit in the command surface and docs.
- Core local project commands such as `validate`, `status`, `doctor`, and read-only project inspection should not perform hidden release checks.
- Any optional release notification surfaced outside the CLI-update commands should require clear opt-in behavior.

## User Guidance
- When a project is newer than the installed CLI, project upgrade and validation flows should tell the user to run the explicit CLI-update flow.
- When a newer CLI release exists but the current project is still compatible, the user should be able to discover that through the explicit CLI-update flow without forcing a project change.

## Shape Rationale
- Full mode was requested explicitly to deepen the change.
- Minimum mode is sufficient for the current size and risk signal.
