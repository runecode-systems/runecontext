---
schema_version: 1
id: assurance/fail-closed-verified-boundary
title: Fail Closed Verified Boundary
status: active
tags:
  - assurance
  - verified
  - validation
  - receipts
---

# Fail Closed Verified Boundary

## Intent

Keep verified assurance mode trustworthy by refusing to operate as verified when its required evidence is incomplete or invalid.

## Requirements

- Verified mode must require a valid assurance baseline.
- Verified-mode receipts and linkage must validate cleanly before the project is treated as healthy.
- Repositories not configured for verified mode must not carry verified-only receipt trees.
- Mutation and capture flows that promise verified evidence must fail if receipt emission fails.

## Rationale

Verified mode is only meaningful if it enforces stronger guarantees than plain mode. Silent downgrade or partial evidence would make the assurance tier misleading.

## Implementation Notes

- Keep plain and verified on the same authored repository model.
- Verified should add evidence requirements, not a separate source-of-truth tree.
