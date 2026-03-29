---
schema_version: 1
id: context-packs/explicit-generated-at-and-stable-hashing
title: Explicit Generated At And Stable Hashing
status: active
tags:
  - context-packs
  - determinism
  - hashing
  - provenance
---

# Explicit Generated At And Stable Hashing

## Intent

Keep context-pack generation reproducible, reviewable, and suitable for durable evidence capture.

## Requirements

- Context-pack generation must require explicit `generated_at` input instead of hidden wall-clock defaults.
- Timestamps must be normalized to a stable portable form suitable for reproducible output.
- Pack hashing and canonicalization must remain explicit and stable.
- Equivalent selected content must produce equivalent pack identity and hash outputs.

## Rationale

Context packs are meant to be auditable artifacts. Hidden time defaults or unstable serialization would make diffs noisy and evidence hard to trust.

## Implementation Notes

- Keep canonicalization and hash algorithm names surfaced in the pack itself.
- Reject inputs that cannot be normalized safely.
