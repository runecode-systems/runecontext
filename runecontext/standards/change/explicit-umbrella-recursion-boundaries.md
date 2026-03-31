---
schema_version: 1
id: change/explicit-umbrella-recursion-boundaries
title: Explicit Umbrella Recursion Boundaries
status: active
tags:
  - changes
  - lifecycle
  - umbrella
  - recursion
---

# Explicit Umbrella Recursion Boundaries

## Intent

Allow coordinated umbrella lifecycle mutations without turning change relationships into an implicit broad mutation graph.

## Requirements

- Recursive lifecycle mutation must be explicit and opt-in rather than the default behavior.
- Only `project` umbrella changes may recursively propagate lifecycle mutations.
- Eligible recursive targets must be reciprocal `related_changes` entries whose change type is `feature`.
- Recursive mutation must not propagate to arbitrary related changes, dependencies, or unrelated graph neighbors.
- Recursive lifecycle operations must validate the full target set before writing and fail closed if any target is invalid.
- Recursive close and supersession flows must reject successor selections that overlap the recursive target set.

## Rationale

Umbrella changes often need coordinated bookkeeping across sub-changes, but `related_changes` is a navigability field, not a hidden hierarchy contract. Explicit target boundaries preserve reviewability and avoid unsafe cascades.

## Implementation Notes

- Keep recursive targeting deterministic and reviewable.
- Preserve transactional mutation behavior so umbrella and sub-change state does not drift on partial failure.
