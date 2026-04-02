## Summary
Stop committing rendered adapter packs and adopt build generated outputs

## Problem
Remove rendered adapter packs from git, keep generated outputs under build/generated, and document which generated surfaces remain committed versus ephemeral.

## Proposed Change
Track and deliver Stop committing rendered adapter packs and adopt build generated outputs while keeping the intent and standards linkage reviewable.

## Why Now
The work needs stable intent, standards linkage, and verification planning before it moves further.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Inferred `just test` from the repository's justfile test target.

## Out of Scope
Work outside the scoped change tracked here.

## Impact
The change keeps intent, assumptions, and standards linkage reviewable.
