## Summary
Support nested monorepo source discovery.

## Problem
Nested packages need their own RuneContext root without being masked by the monorepo-wide config.

## Proposed Change
Select the nearest ancestor config and report the chosen path in structured metadata.

## Why Now
Monorepo support is part of the alpha.2 source discovery contract.

## Assumptions
Package-level roots intentionally override the broader monorepo default.

## Out of Scope
Cross-root bundle inheritance.

## Impact
Subtree-specific runs can resolve the correct RuneContext tree deterministically.
