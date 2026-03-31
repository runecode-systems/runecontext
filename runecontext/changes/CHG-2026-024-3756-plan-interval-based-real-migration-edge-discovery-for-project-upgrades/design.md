# Design

## Overview
Refine project-upgrade planning so the explicit migration registry is interpreted as a sparse set of real migration edges rather than as a required synthetic hop chain from the project's current version. Upgrade preview remains the read-only planning surface. Upgrade apply remains the only mutating command and continues to use staged execution. Fresh projects created at the installed version initialize directly into that version's layout and do not replay historical migrations.

## Semantics
- Explicit migration edges represent only transitions that require real project-tree migration logic.
- Planning should collect only the explicit migration edges whose transitions fall between the current project version and the selected target version.
- Absence of migration edges inside the requested interval means the upgrade is version-bump-only, not unsupported.
- Fresh projects initialized at the installed version should never replay historical migration edges from earlier releases.

## Planning Rules
- If the selected target equals the current project version, preview should remain zero-hop and apply should be a no-op aside from any existing managed-artifact refresh logic.
- If one or more explicit migration edges fall inside the requested current-to-target interval, preview should report only those ordered edges.
- If no explicit migration edges fall inside the requested interval and the target version is otherwise compatible, preview should report a zero-hop version-bump-only plan.
- The planner must not invent fake bridge hops for versions that have no real migration logic.
- If explicit migration edges overlap ambiguously inside the interval, planning should fail closed.

## Apply Sequencing
- Upgrade apply continues to operate in a staged tree.
- Before each real migration edge runs, the staged project version may be advanced through plain version rewrites until the staged version reaches that edge's `from` version.
- Each real migration edge then runs its explicit apply and verify logic in the staged tree.
- After the final real migration edge, apply performs any remaining final version rewrite to the selected target and runs final staged validation before replacing the live tree.

## User Model
- New projects start with the current release layout for their initialized version.
- Existing older projects upgrade only through the real migration boundaries that lie between their pinned version and the chosen target.
- `hop_count` continues to mean the count of real migration edges that must run, not the number of ordinary version increments.

## Shape Rationale
- Large, ambiguous, or high-risk feature work should move to full mode early.
