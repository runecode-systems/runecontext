## Summary
Migrate assurance artifacts under `runecontext/` for alpha.13 upgrades.

## Problem
The canonical project profile and layout metadata already define `runecontext/assurance/` as the project assurance path, but the current implementation still writes and validates verified assurance artifacts under a legacy root-level `assurance/` tree. Alpha.13 is the right point to correct that drift with a real project migration rather than leaving two layouts in circulation.

## Proposed Change
Make alpha.13 the canonical assurance-layout boundary. Fresh alpha.13 projects and all alpha.13 assurance writers emit only `runecontext/assurance/` artifacts. Upgrade apply performs a real staged `alpha.12 -> alpha.13` migration that moves the legacy root `assurance/` tree into `runecontext/assurance/`, rewrites baseline `imported_evidence` backfill paths to the canonical location, validates the staged result, and only then replaces the live project tree.

## Why Now
This is the first real staged project migration on top of the new upgrade framework, and the project has already been bumped to alpha.13 specifically to carry this migration logic. Fixing the assurance path now keeps the on-disk layout aligned with the documented canonical profile before more verified projects normalize the legacy placement.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Inferred `just test` from the repository's justfile test target.

## Out of Scope
- Adding dedicated runtime guidance for legacy root-level assurance layouts beyond making `upgrade apply` work.
- Supporting dual canonical assurance layouts after alpha.13.
- Inventing synthetic upgrade hops for versions that have no real migration logic.

## Impact
The change gives RuneContext its first real project-tree migration while restoring consistency between the implemented assurance layout and the documented canonical project profile. After alpha.13, all RuneContext-owned project artifacts remain under `runecontext/` except the root `runecontext.yaml` file.
