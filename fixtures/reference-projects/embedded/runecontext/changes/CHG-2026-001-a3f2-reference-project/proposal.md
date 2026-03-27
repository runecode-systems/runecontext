## Summary
Define a reference fixture used by install and upgrade validation tests.

## Problem
Alpha.8 needs stable end-to-end fixtures across source modes.

## Proposed Change
Provide a minimal valid embedded RuneContext project as a reusable reference.

## Why Now
Release/install/upgrade hardening requires deterministic fixture coverage.

## Assumptions
Fixture content remains small and hand-authored.

## Out of Scope
Production migration behavior.

## Impact
Improves repeatable test coverage for CLI and contract validation.
