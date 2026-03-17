## Summary
Exercise a spec ID mismatch with an ancestor path containing the word specs.

## Problem
Substring-based path matching can select the wrong specs segment.

## Proposed Change
Keep the project valid except for the spec ID mismatch.

## Why Now
The validator should use segment-aware path matching.

## Assumptions
The rest of the fixture remains valid.

## Out of Scope
Fixing the spec ID.

## Impact
This fixture proves the path-matching logic is root-aware.
