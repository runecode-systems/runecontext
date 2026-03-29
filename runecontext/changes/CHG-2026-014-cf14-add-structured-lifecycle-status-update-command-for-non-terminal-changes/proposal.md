## Summary
Add structured lifecycle status update command for non-terminal changes

## Problem
The repository currently has lifecycle states such as `proposed`, `planned`, `implemented`, and `verified`, but no normal CLI mutation surface for moving a non-terminal change through those states. That gap forced manual status edits for the human-friendly `runectx status` work and blurred the boundary between lifecycle advancement and close-time promotion assessment.

## Proposed Change
Add `runectx change update <CHANGE_ID> --status planned|implemented|verified [--path PATH]` as the dedicated non-terminal lifecycle update command, with room for safe relationship edits under the same structured mutation surface. Keep `change close` as the only terminal lifecycle and promotion-assessment entrypoint.

## Why Now
Recent dogfooding exposed that implemented changes can remain stuck at `proposed` because the CLI has no supported way to record ordinary lifecycle advancement before close.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Inferred `just test` from the repository's justfile test target.

## Out of Scope
- Terminal transitions to `closed` or `superseded`.
- Auto-running or mutating promotion assessment during non-terminal lifecycle updates.

## Impact
The change gives users a reviewable CLI path for ordinary change progression while keeping lifecycle status, verification status, and promotion assessment as separate concepts.
