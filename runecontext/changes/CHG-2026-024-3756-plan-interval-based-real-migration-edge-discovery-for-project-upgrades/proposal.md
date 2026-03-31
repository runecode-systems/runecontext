## Summary
Plan interval-based real migration-edge discovery for project upgrades.

## Problem
The current upgrade framework supports real staged migrations, but the planner still needs an explicit way to discover only the real migration edges that fall between a project's current version and the selected target version. The planner must not invent synthetic hops to bridge compatible version gaps, and fresh projects created at the installed version must not replay historical migrations.

## Proposed Change
Refine project-upgrade planning so upgrade preview and apply treat explicit migration edges as sparse real migration boundaries inside a version interval. When a user upgrades an older project, the planner should collect only the ordered explicit migration edges whose transitions fall between the current project version and the target version. When no real migration edge falls inside the interval, upgrade remains a zero-hop staged version-bump-only operation. Fresh projects initialized at the installed version continue to start directly in that version's canonical layout without replaying any earlier migration edges.

## Why Now
Alpha.13 is the first release that will carry a real on-disk project migration. That makes the planner semantics user-visible: users upgrading from older versions must run the real migration edge when it falls inside their requested interval, but must not see fake hops or historical migrations unrelated to their project's current version.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Inferred `just test` from the repository's justfile test target.

## Out of Scope
- Implementing the assurance artifact layout rewrite itself.
- Adding dedicated runtime guidance for legacy pre-alpha.13 assurance layouts beyond what upgrade preview/apply already expose.
- Replaying historical migrations for projects initialized directly at the installed version.

## Impact
The change keeps upgrade semantics honest: explicit migration hops represent only real migration logic, fresh projects start directly in the installed version's layout, and direct upgrades from older compatible versions to alpha.13 can still run the real migration edges that actually lie inside the requested interval.
