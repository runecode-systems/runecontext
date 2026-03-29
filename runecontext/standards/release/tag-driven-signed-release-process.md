---
schema_version: 1
id: release/tag-driven-signed-release-process
title: Tag Driven Signed Release Process
status: active
tags:
  - release
  - tags
  - signatures
  - process
---

# Tag Driven Signed Release Process

## Intent

Keep official releases reproducible, reviewable, and traceable to a signed source state.

## Requirements

- Official releases must be cut from the automated release workflow, not created manually in the GitHub UI.
- Release tags must match canonical release metadata and should be signed annotated tags.
- Protected review and approval gates should guard publication.
- A bad release should be superseded by a new release version rather than silently rewritten in place.

## Rationale

Signed tags and workflow-driven release publication give RuneContext a durable provenance story and reduce ambiguity about what constitutes an official release.

## Implementation Notes

- Keep release metadata, CI checks, signatures, and attestations aligned.
- Document repository and maintainer setup requirements alongside the process.
