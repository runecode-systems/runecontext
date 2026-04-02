## Summary
Stage generated adapter packs in tests and release builds

## Problem
Update test, build, release, and install flows to generate adapter packs into staging layouts instead of reading committed rendered adapters from the repo tree.

## Proposed Change
Track and deliver Stage generated adapter packs in tests and release builds while keeping the intent and standards linkage reviewable.

## Why Now
The work needs stable intent, standards linkage, and verification planning before it moves further.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Inferred `just test` from the repository's justfile test target.

## Out of Scope
Work outside the scoped change tracked here.

## Impact
The change keeps intent, assumptions, and standards linkage reviewable.
