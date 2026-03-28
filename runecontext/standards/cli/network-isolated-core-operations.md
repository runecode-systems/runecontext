---
schema_version: 1
id: cli/network-isolated-core-operations
title: Network-Isolated Core Operations
status: active
tags:
  - cli
  - offline
  - install
  - portability
---

# Network-Isolated Core Operations

## Intent

Keep normal RuneContext project operations reliable in local and offline environments.

## Requirements

- Core project operations such as init, validate, status, change workflows, bundle resolution, and assurance validation must work without hidden network access.
- Any network-enabled behavior must be explicit in the command surface and documentation.
- Installed or vendored release contents should be sufficient for day-to-day repository management.
- Failures caused by unavailable network resources must not silently change core semantics.

## Rationale

RuneContext is designed for portable repo-first use. Developers and automated environments should be able to manage a project from reviewed local assets without depending on live services.

## Implementation Notes

- Upgrade and release retrieval flows may be network-enabled by design, but their boundaries should stay explicit.
- Keep local-first behavior visible in docs and adapter guidance.
