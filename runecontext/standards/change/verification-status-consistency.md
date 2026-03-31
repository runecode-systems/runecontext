---
schema_version: 1
id: change/verification-status-consistency
title: Verification Status Consistency
status: active
tags:
  - changes
  - verification
  - lifecycle
  - validation
---

# Verification Status Consistency

## Intent

Keep change verification state aligned with lifecycle state so workflow summaries remain trustworthy.

## Requirements

- Verified changes must record a completed verification status.
- Non-terminal `verified` changes are valid only when they record a completed `verification_status`.
- Terminal changes must not leave `verification_status` as `pending`.
- Closing workflows should require explicit verification outcomes when the current state is incomplete.
- Non-terminal mutation workflows should support recording completed verification state explicitly instead of forcing terminal close.
- Validation must reject impossible lifecycle and verification combinations.

## Rationale

Change verification is part of the durable workflow contract, not a cosmetic field. Inconsistent verification state makes status reporting and assurance evidence misleading.

## Implementation Notes

- Prefer explicit status transitions over implicit guessing.
- Keep verification output stable enough for CLI and adapter workflows to consume safely.
