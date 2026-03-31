## Summary
Refine project upgrade compatibility and version-bump-only semantics

## Problem
Separate compatibility from migration edges so older-but-compatible projects can be upgraded without synthetic migration hops, while projects newer than the installed CLI fail closed with explicit upgrade-your-cli guidance.

## Proposed Change
Refine project upgrade semantics so the planner and apply path distinguish three cases clearly: compatible older projects that may remain pinned until a user upgrades them, compatible older projects that can be upgraded with only a version bump, and projects pinned to a newer version than the installed CLI supports. Explicit migration edges should be reserved for transitions that require real migration logic rather than being used as synthetic permission to bump a project pin.

## Why Now
The alpha.12 dogfood branch exposed that the current implementation treats a compatible alpha.11 project as unsupported under an alpha.12 CLI because no explicit migration edge exists. That behavior is too strict for normal team usage and makes it hard to communicate the right action to users.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Inferred `just test` from the repository's justfile test target.

## Out of Scope
- CLI binary self-update and release discovery flows.
- Changing the read-only nature of `runectx upgrade` preview.

## Impact
The change lets teams use newer compatible `runectx` binaries without forcing immediate project upgrades, while still requiring older CLIs to fail closed once a project pin moves beyond what they support.
