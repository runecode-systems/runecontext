## Summary
Resolve RuneContext content from a git source.

## Problem
Linked repositories need deterministic source selection metadata for later audit-oriented flows.

## Proposed Change
Support pinned commits and mutable refs with explicit structured metadata and warnings.

## Why Now
Alpha.2 requires linked-source resolution before context packs can record provenance.

## Assumptions
Signed-tag verification remains out of scope for this branch.

## Out of Scope
Trusted-signer verification.

## Impact
Linked source resolution becomes testable with reusable fixtures.
