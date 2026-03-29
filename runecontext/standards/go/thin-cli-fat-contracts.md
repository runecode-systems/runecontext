---
schema_version: 1
id: go/thin-cli-fat-contracts
title: Thin CLI Fat Contracts
status: active
tags:
  - go
  - architecture
  - cli
  - contracts
---

# Thin CLI Fat Contracts

## Intent

Keep the Go codebase organized so command parsing and rendering stay thin while canonical semantics live in the contracts layer.

## Requirements

- `cmd/` should remain a thin binary entrypoint.
- `internal/cli` should focus on parsing, output shaping, and command wiring.
- `internal/contracts` should own validation, mutation semantics, and durable workflow rules.
- New behavior should prefer extracting contract-level helpers over embedding business logic in CLI handlers.

## Rationale

RuneContext's CLI needs to support direct use, machine interfaces, and adapters. A thin CLI over stronger contracts keeps semantics reusable and easier to test.

## Implementation Notes

- Preserve narrow command boundaries.
- Favor small responsibility-focused files over large mixed-purpose command implementations.
