---
schema_version: 1
id: bundles/deterministic-bundle-linearization
title: Deterministic Bundle Linearization
status: active
tags:
  - bundles
  - determinism
  - resolution
  - context
---

# Deterministic Bundle Linearization

## Intent

Ensure bundle inheritance and rule evaluation produce the same resolved context every time.

## Requirements

- Bundle resolution order must be deterministic and reviewable.
- Include and exclude rule precedence must remain explicit and stable.
- Resolution diagnostics should explain significant rule and extension effects without changing semantics.
- Inheritance depth and traversal safety limits must remain enforced.

## Rationale

Bundles shape what context and standards are considered. If linearization or rule precedence drifts, context packs and downstream workflows become difficult to audit.

## Implementation Notes

- Keep match inventories and rule references concrete.
- Prefer explicit diagnostics over hidden fallback behavior when bundle inputs are surprising.
