## Summary
Exercise a decision reference to a missing originating change.

## Problem
The decision frontmatter points at a change that does not exist.

## Proposed Change
Keep the change markdown valid so project validation fails on the missing decision change reference.

## Why Now
This fixture proves strict fail-closed lineage checks.

## Assumptions
The rest of the change contract is valid.

## Out of Scope
Backfilling the missing change.

## Impact
The validator should report the missing referenced change.
