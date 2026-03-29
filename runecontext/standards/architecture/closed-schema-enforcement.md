---
schema_version: 1
id: architecture/closed-schema-enforcement
title: Closed Schema Enforcement
status: active
tags:
  - schemas
  - validation
  - contracts
  - architecture
---

# Closed Schema Enforcement

## Intent

Prevent silent contract drift by rejecting unknown machine-readable fields unless RuneContext explicitly allows them.

## Requirements

- Authoritative YAML and JSON artifacts must validate against closed schemas.
- Unknown fields must fail closed unless they live in an explicitly permitted `extensions` block.
- Extensions must remain non-authoritative and must not alter RuneContext core semantics.
- Schema updates must accompany any intentional expansion of the machine-readable contract.

## Rationale

Closed schemas make mistakes obvious, keep automation reliable, and protect cross-tool portability. Typos and informal extensions should never quietly change repository meaning.

## Implementation Notes

- When a field needs to become official, promote it into the schema and implementation together.
- Use warnings for non-authoritative extensions only when the schema intentionally allows them.
