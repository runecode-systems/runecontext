## Summary
Allow change update to record completed verification state

## Problem
`change update --status verified` is currently inconsistent with validation: verified changes must have a completed `verification_status`, but `change update` does not let users set that field. That leaves open verified changes awkward to represent and pushes users toward terminal close even when they only want to record completed verification while keeping the change open.

## Proposed Change
Extend `change update` so it can record a completed verification status as part of a non-terminal lifecycle mutation, allowing open changes to validly reach `status=verified` without requiring immediate close.

## Why Now
The current workflow contradiction already shows up in daily use: the lifecycle model includes `verified`, but the CLI cannot cleanly move an open change into that state while satisfying `verification_status` invariants.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- `change update` remains non-terminal and must not write `closed` or `superseded`.

## Out of Scope
- Changing terminal close semantics.
- Auto-mutating promotion assessment.
- Allowing contradictory states such as `verified` plus `pending` verification.

## Impact
The change makes the documented lifecycle model executable through the CLI by letting open changes record completed verification state without conflating verification completion with terminal close.
