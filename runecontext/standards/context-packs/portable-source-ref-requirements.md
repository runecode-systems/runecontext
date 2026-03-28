---
schema_version: 1
id: context-packs/portable-source-ref-requirements
title: Portable Source Ref Requirements
status: active
tags:
  - context-packs
  - source
  - portability
  - provenance
---

# Portable Source Ref Requirements

## Intent

Ensure context packs describe their source in a form that remains meaningful outside the local machine that produced them.

## Requirements

- Context-pack source metadata must use portable source references.
- Local path-backed packs must reject non-portable source references instead of embedding machine-specific paths.
- Source mode and source verification posture must remain explicit in the pack metadata.
- Signer identity and resolved commit data should be preserved when available and relevant.

## Rationale

Context packs are designed for exchange and audit. A pack that only makes sense on the original machine undermines RuneContext's portability goals.

## Implementation Notes

- Prefer stable relative references and pinned identifiers where supported.
- Fail closed when a source reference cannot be represented portably.
