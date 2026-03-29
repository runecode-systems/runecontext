---
schema_version: 1
id: change/immutable-terminal-change-records
title: Immutable Terminal Change Records
status: active
tags:
  - changes
  - lifecycle
  - history
  - audit
---

# Immutable Terminal Change Records

## Intent

Preserve closed and superseded changes as stable historical records instead of mutable workflow scratchpads.

## Requirements

- Once a change reaches a terminal lifecycle state, its record should be treated as finalized history.
- Follow-up work should be represented with new changes and explicit relationships instead of rewriting the old record's intent.
- Terminal lifecycle transitions must capture the required closing or supersession metadata.
- Tooling should avoid encouraging casual edits to terminal workflow state.

## Rationale

Terminal changes are part of the project's durable narrative. Stable history supports audits, traceability, and trustworthy promotion and assurance evidence.

## Implementation Notes

- Use explicit successor and related-change links for continuation work.
- Reserve specialized commands for terminal lifecycle operations that need stronger auditing.
