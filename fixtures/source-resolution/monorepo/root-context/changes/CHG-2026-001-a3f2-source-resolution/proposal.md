## Summary
Support nearest-ancestor source discovery in monorepos.

## Problem
Monorepos may contain both shared and package-specific RuneContext roots.

## Proposed Change
Use nearest-ancestor `runecontext.yaml` discovery unless the caller provides an explicit root.

## Why Now
Source discovery is part of the alpha.2 storage and bundle foundation.

## Assumptions
Nested packages may override the monorepo root when they contain their own config.

## Out of Scope
Bundle precedence behavior beyond source discovery.

## Impact
Monorepo runs can resolve the intended project context without hidden heuristics.
