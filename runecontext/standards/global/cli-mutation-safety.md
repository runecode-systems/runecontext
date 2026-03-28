---
schema_version: 1
id: global/cli-mutation-safety
title: CLI Mutation Safety
status: active
tags:
  - cli
  - mutation
  - validation
  - safety
---

# CLI Mutation Safety

## Intent

Keep write-capable CLI operations reviewable, reversible during failure, and safe to run in normal local workflows.

## Requirements

- RuneContext write operations must validate inputs before mutating authoritative files.
- Multi-file mutations must be applied transactionally so partial writes do not become durable state.
- Post-write validation must run before a mutation is considered successful.
- Dry-run behavior should simulate the planned mutation and validate the would-be result instead of emitting prose-only intent.
- Commands must fail closed on ambiguous, unsafe, or partially invalid mutation plans.

## Rationale

RuneContext is meant to be a reviewable, git-native system. CLI convenience must not come at the cost of silent corruption, half-written workflow state, or hidden mutation behavior.

## Implementation Notes

- Specialized commands may collect extra workflow-specific evidence before mutating state.
- Manual file edits remain allowed, but `runectx validate` is the canonical safety net after manual changes.
