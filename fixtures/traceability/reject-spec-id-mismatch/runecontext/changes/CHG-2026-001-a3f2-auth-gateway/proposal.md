## Summary
Exercise a spec frontmatter id mismatch.

## Problem
The spec path and frontmatter id disagree.

## Proposed Change
Keep the change markdown valid so the project fails on the spec metadata mismatch.

## Why Now
The validator should reject path/id drift immediately.

## Assumptions
The rest of the fixture remains valid.

## Out of Scope
Reconciling the spec id.

## Impact
This fixture proves strict path-matched ids.
