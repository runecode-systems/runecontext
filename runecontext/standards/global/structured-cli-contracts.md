---
schema_version: 1
id: global/structured-cli-contracts
title: Structured CLI Contracts
status: active
tags:
  - cli
  - machine-interface
  - adapters
  - contracts
---

# Structured CLI Contracts

## Intent

Keep RuneContext operations machine-friendly, stable, and reusable across direct CLI use, shell completion, and adapter-driven workflows.

## Requirements

- CLI operations must expose stable machine-facing outputs with a shared envelope and failure taxonomy.
- New command behavior should prefer explicit structured fields over output text that requires fragile parsing.
- Read-only advisory flows should emit reusable candidate data and next-step hints without hidden session state.
- Completion metadata, command docs, and adapter-native command surfaces should derive from the same canonical operation metadata where possible.

## Rationale

RuneContext is consumed by humans, scripts, and conversational tools. A stable structured contract prevents drift across those entrypoints and makes thin adapters possible.

## Implementation Notes

- Prefer additive fields over breaking renames.
- Explain and dry-run output should remain compatible with the same core contract shape.
