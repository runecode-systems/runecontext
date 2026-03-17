## Summary
Exercise project-level extension opt-in rejection.

## Problem
The change uses an extension without `allow_extensions: true`.

## Proposed Change
Keep the markdown valid so validation fails for the intended extension policy reason.

## Why Now
Project-level opt-in is part of the alpha.1 machine-readable profile.

## Assumptions
The rest of the change contract is valid.

## Out of Scope
Enabling extensions in the root config.

## Impact
This fixture proves cross-file extension enforcement.
