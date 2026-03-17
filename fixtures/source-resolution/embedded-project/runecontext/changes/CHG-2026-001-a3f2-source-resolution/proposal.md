## Summary
Add deterministic source-resolution fixtures and tests.

## Problem
Alpha.2 needs source discovery and resolution behavior that is exercised by fixtures rather than prose alone.

## Proposed Change
Add resolver tests for embedded, git, path, and monorepo cases with structured metadata outputs.

## Why Now
Later bundle and context-pack work depends on the selected source metadata being stable.

## Assumptions
The source-resolution result shape can be shared across later audit-oriented milestones.

## Out of Scope
Signed-tag verification.

## Impact
This keeps source-resolution behavior reviewable and reusable across future parity suites.
