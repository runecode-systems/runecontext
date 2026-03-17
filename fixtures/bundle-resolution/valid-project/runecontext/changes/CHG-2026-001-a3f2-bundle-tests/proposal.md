## Summary

Add bundle-engine precedence and guardrail tests.

## Problem

Bundle resolution semantics need deterministic fixtures.

## Proposed Change

Add a dedicated bundle-resolution fixture tree and golden outputs.

## Why Now

Alpha.2 depends on deterministic bundle semantics.

## Assumptions

N/A

## Out of Scope

Context-pack hashing and emission.

## Impact

Tests can verify precedence, diagnostics, and path guardrails.
