---
schema_version: 1
id: assurance/portable-receipt-linkage
title: Portable Receipt Linkage
status: active
tags:
  - assurance
  - receipts
  - linkage
  - portability
---

# Portable Receipt Linkage

## Intent

Ensure assurance receipts point to durable RuneContext subjects in a portable and validation-friendly way.

## Requirements

- Receipt subjects must identify the intended RuneContext artifact family clearly and consistently.
- Receipt linkage must validate against current repository artifacts without relying on home-directory caches or deployment-local metadata.
- Portable receipts should contain the minimum durable identifiers needed for independent verification.
- Family-specific linkage rules must remain explicit and schema-backed.

## Rationale

Receipts are only useful if they can be understood and validated away from the machine that created them. Portable linkage keeps assurance evidence durable across teams and environments.

## Implementation Notes

- Use stable subject naming conventions.
- Reject receipts whose subject identity does not match the family-specific contract.
