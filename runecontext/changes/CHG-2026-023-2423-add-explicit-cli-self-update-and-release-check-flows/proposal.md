## Summary
Add explicit CLI self-update and release-check flows

## Problem
Keep project upgrade preview/apply under runectx upgrade while adding explicit upgrade cli preview/apply commands for checking, downloading, and installing newer CLI releases.

## Proposed Change
Add an explicit CLI-update lane that stays separate from project-upgrade behavior. Bare `runectx upgrade` and `runectx upgrade apply` should remain about the managed project, while `runectx upgrade cli` and `runectx upgrade cli apply` should handle checking for newer CLI releases and installing them.

## Why Now
Once project upgrade semantics distinguish older compatible projects from newer-than-cli failures, the CLI needs a clear place to send users when their binary must be upgraded. The command surface also needs an explicit network-enabled path for release checks and binary install/update behavior.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Inferred `just test` from the repository's justfile test target.

## Out of Scope
- Changing bare project-upgrade commands to mean CLI self-update.
- Hidden background network access in core local project operations.

## Impact
The change gives users a clear and scriptable self-update story without overloading the meaning of project upgrade.
