## Summary
Add nested service-level monorepo reference fixture content.

## Problem
Nested RuneContext roots need explicit validated fixture coverage.

## Proposed Change
Include a minimal valid service-level RuneContext tree for nearest-ancestor tests.

## Why Now
Alpha.8 requires monorepo reference fixture end-to-end coverage.

## Assumptions
Root and service fixtures coexist in one monorepo tree.

## Out of Scope
Cross-workspace discovery beyond this monorepo shape.

## Impact
Enables deterministic nested-root validation and CLI discovery tests.
