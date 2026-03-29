# Design

## Overview
Extend `ProjectStatusSummary` and `ChangeStatusEntry` so status consumers can reason about associations, display verification posture, and sort or bound historical sections. This is a data-model enhancement for human rendering, not a redesign of the existing machine contract.

## Summary Additions
- Add reciprocal and directional relationship fields needed for human rendering, including `related_changes`, `depends_on`, `supersedes`, and `superseded_by`.
- Add `verification_status` so the human view can distinguish unfinished work from work that is ready to close or already verified.
- Add recency fields such as `created_at` and `closed_at` so historical sections can sort deterministically and bound previews by count.

## Contract Boundary
- Keep the current active, closed, and superseded grouping in the summary.
- Preserve the existing flat `runectx status --json` contract for now; the extra fields are for internal summary consumers and future additive output decisions.
- Ensure relationship fields reflect the already-validated change graph rather than inventing alternate hierarchy semantics in the CLI.

## Sequencing Role
- This change is the prerequisite for both the human renderer and the progressive-disclosure work.
