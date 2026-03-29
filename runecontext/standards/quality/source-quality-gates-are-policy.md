---
schema_version: 1
id: quality/source-quality-gates-are-policy
title: Source Quality Gates Are Policy
status: active
tags:
  - quality
  - lint
  - policy
  - go
---

# Source Quality Gates Are Policy

## Intent

Treat source-quality enforcement as reviewed repository policy rather than a convenience hurdle to bypass.

## Requirements

- `just lint` must remain the canonical local source-quality gate.
- Protected quality surfaces such as linter config, source-quality config, and checker code must be reviewed as policy changes.
- When checks fail, the default response should be refactoring or clarifying code rather than weakening limits casually.
- Exceptions should be narrow, explicit, and justified.

## Rationale

Quality gates define what the repository will continue to permit. Treating them as policy keeps complexity creep visible and prevents short-term convenience from normalizing long-term debt.

## Implementation Notes

- Call out protected-surface edits explicitly in reviews and summaries.
- Prefer deleting exceptions after refactors rather than normalizing them as permanent debt.
