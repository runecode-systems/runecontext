# Design

## Overview
Model this work as an umbrella over two linked improvements to change lifecycle mutations. First, allow non-terminal lifecycle updates to carry completed verification state so an open change can validly reach status verified without forcing immediate terminal close. Second, add explicit recursive propagation controls for umbrella project changes so change update and change close can optionally cascade to associated feature sub-changes only. Recursive propagation must be opt-in via a flag such as --recursive, must never target all related_changes indiscriminately, and must fail closed if any targeted sub-change cannot take the requested transition.

## Planned Sub-Changes
- `CHG-2026-019-c1af-allow-change-update-to-record-completed-verification-state` owns the non-terminal verification-state mutation fix.
- `CHG-2026-020-75ba-add-explicit-recursive-umbrella-lifecycle-propagation-for-sub-changes` owns the explicit recursive propagation model.

## Lifecycle Rules
- `change update` remains the non-terminal lifecycle command.
- `change close` remains the terminal lifecycle command.
- Validation must continue rejecting contradictory lifecycle and verification combinations.
- Recursive propagation must not weaken lifecycle validation or relationship semantics.

## Relationship Rules
- Recursive propagation must be scoped to project umbrella changes and their associated feature sub-changes only.
- `related_changes` remains a navigability field and must not become an implicit "mutate everything here" set.
- Recursive target selection should use explicit umbrella/sub-change semantics rather than generic graph traversal.

## Safety Rules
- Non-recursive behavior remains the default.
- Recursive behavior requires explicit user intent, for example `--recursive`.
- Recursive operations must be transactional across all affected change records and fail closed on partial invalid cascades.
