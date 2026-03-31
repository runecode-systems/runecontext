## Summary
Add staged per-hop upgrade execution and validation

## Problem
Even with staged multi-hop support, the apply flow still needs to distinguish plain version-bump-only upgrades from upgrades that require real migration hops. It must support both “no migration, just bump the project pin” and “run one or more real migrations, then bump the final pinned version” without treating every upgrade as if a migration edge is required.

## Proposed Change
Execute approved project upgrades through the staged upgrader, whether the plan requires zero migration hops or several real migration hops. When migrations are required, apply each hop against a copied project tree, validate and verify after every hop, then bump the final pinned project version and replace real files only after the full staged chain succeeds.

## Why Now
Once planning distinguishes compatibility from migration requirements, apply-time behavior must honor the same model. Without that alignment, compatible older projects cannot be upgraded through a safe version-bump-only path, and migration-required upgrades cannot clearly separate the migration work from the final pinned-version update.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Some upgrades will need only a plain pinned-version rewrite, while others will need dedicated migration hops plus hop-specific verification.

## Out of Scope
- Replacing preview/path planning contracts.
- Mutating real project files incrementally during intermediate hops.
- Skipping staged validation or hop-specific verification in the name of convenience.

## Impact
The change makes `runectx upgrade apply` trustworthy for both version-bump-only upgrades and future multi-version migrations by giving each declared hop a typed execution path, staged verification, and fail-closed rollback before the real project tree is touched.
