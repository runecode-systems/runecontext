---
schema_version: 1
id: architecture/derived-artifact-non-authority
title: Derived Artifact Non-Authority
status: active
tags:
  - generated
  - indexes
  - architecture
  - authority
---

# Derived Artifact Non-Authority

## Intent

Keep generated manifests and indexes useful for review and tooling without letting them become a second source of truth.

## Requirements

- Generated artifacts such as manifests and status indexes must be treated as derived outputs, not authoritative inputs.
- Validation and mutation logic must derive truth from authored repository artifacts when correctness matters.
- Generated files must remain reproducible from the same validated source state.
- Conflicts between authored files and derived outputs must be resolved in favor of the authored artifacts.

## Rationale

Reviewable indexes are valuable, but making generated files authoritative introduces hidden state and recovery hazards. RuneContext should always be able to reconstruct its derived views from the portable repo model.

## Implementation Notes

- Keep generation deterministic so diffs remain reviewable.
- Treat generated artifacts as summaries and caches, not workflow authorities.
