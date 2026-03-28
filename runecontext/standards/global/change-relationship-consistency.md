---
schema_version: 1
id: global/change-relationship-consistency
title: Change Relationship Consistency
status: active
tags:
  - changes
  - relationships
  - graph
  - traceability
---

# Change Relationship Consistency

## Intent

Keep change-to-change relationships navigable, semantically clear, and validation-safe across the repository.

## Requirements

- `related_changes` must be reciprocal across all referenced change records.
- `depends_on` must remain directional and must not introduce dependency cycles.
- `supersedes` and `superseded_by` must remain bidirectionally consistent and lifecycle-valid.
- Changes must not reference themselves in relationship fields.
- Relationship mutations should preserve the canonical meaning of each link instead of inventing alternate hierarchy semantics.

## Rationale

Change graphs are part of RuneContext's durable workflow history. Navigation, sequencing, and replacement semantics only stay trustworthy when links are applied consistently and validated together.

## Implementation Notes

- Use `related_changes` for navigability and `depends_on` only for prerequisite ordering.
- For umbrella and sub-change workflows, prefer explicit graph wiring over hidden structural assumptions.
