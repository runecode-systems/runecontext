## Summary
Exercise a missing related spec reference failure.

## Problem
The change points at a spec path that does not exist.

## Proposed Change
Keep the markdown valid so project validation fails for the intended traceability reason.

## Why Now
The CLI should surface the missing spec diagnostic cleanly.

## Assumptions
The rest of the change shape is valid.

## Out of Scope
Fixing the broken related spec path.

## Impact
This fixture isolates a single project-level failure.
