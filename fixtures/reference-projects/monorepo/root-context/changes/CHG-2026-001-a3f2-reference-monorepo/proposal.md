## Summary
Add root-level monorepo reference fixture content.

## Problem
Alpha.8 requires nested-root coverage for reference projects.

## Proposed Change
Include a valid root RuneContext tree in the monorepo fixture.

## Why Now
End-to-end tests need explicit nearest-ancestor discovery coverage.

## Assumptions
Service-level nested root is defined separately in this fixture.

## Out of Scope
Cross-repository monorepo discovery semantics.

## Impact
Provides deterministic root-node discovery coverage.
