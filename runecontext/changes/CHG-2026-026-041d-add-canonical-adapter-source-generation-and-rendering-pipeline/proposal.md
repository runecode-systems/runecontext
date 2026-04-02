## Summary
Add canonical adapter source generation and rendering pipeline

## Problem
Introduce a smaller canonical adapter source model and a deterministic generator that renders adapter packs into build/generated/adapters for local and staged use.

## Proposed Change
Track and deliver Add canonical adapter source generation and rendering pipeline while keeping the intent and standards linkage reviewable.

## Why Now
The work needs stable intent, standards linkage, and verification planning before it moves further.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Inferred `just test` from the repository's justfile test target.

## Out of Scope
Work outside the scoped change tracked here.

## Impact
The change keeps intent, assumptions, and standards linkage reviewable.
