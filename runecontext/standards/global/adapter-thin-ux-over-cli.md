---
schema_version: 1
id: global/adapter-thin-ux-over-cli
title: Adapter Thin UX Over CLI
status: active
tags:
  - adapters
  - cli
  - ux
  - portability
---

# Adapter Thin UX Over CLI

## Intent

Keep adapter-specific experiences conversational and ergonomic without creating alternate RuneContext semantics.

## Requirements

- Adapters must treat the CLI and core contracts as the source of truth for semantics and mutation behavior.
- Adapter-native assets may gather context, ask clarifying questions, and present next steps, but they must not invent hidden state or alternate workflow rules.
- Host-specific generated files must remain additive tool UX layers rather than authoritative project data.
- Adapter flows should prefer reviewable command proposals and structured CLI outputs over opaque prompt-only behavior.

## Rationale

RuneContext is tool-agnostic by design. Adapters can improve usability, but the portable repository model breaks down if each host tool becomes its own workflow authority.

## Implementation Notes

- When adapters need richer guidance, encode it as explicit CLI contracts or generated flow instructions rather than host-only semantics.
- Use host-native prompts to streamline UX, not to redefine correctness.
