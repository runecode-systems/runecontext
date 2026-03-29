# Design

## Overview
Add explicit history-scaling behavior to human `runectx status` output so the command stays readable as projects accumulate large numbers of closed and superseded changes. The feature should complement the new renderer rather than forcing operators to read an ever-growing full history dump.

## Default Behavior
- Show all active or in-flight changes by default.
- Show closed and superseded sections as bounded recent previews by count.
- Emit hints when entries are hidden, such as `showing 5 of 143 closed changes; use --history all`.
- Treat zero historical entries and fewer-than-limit entries as normal cases rather than special failures.

## Display Controls
- Add `--history recent|all|none` for human output.
- Add `--history-limit N` to control the preview count.
- Add `--verbose` for richer row details without forcing full historical dumps.

## Ordering Rules
- Keep active work grouped and fully visible.
- Sort closed and superseded previews by recency descending using the summary's recency fields.
- Ensure relationship-aware ordering stays predictable when combined with bounded previews.

## Finalized Presentation Constraints
- History controls operate over the multiline tree renderer, so bounded previews must preserve tree readability.
- Hidden-history hints and relationship lines must use compact IDs by default and switch to full IDs only under `--verbose`.
- Wrapped hint lines must stay aligned and readable under the same controlled wrapping rules used by status rows.

## Machine Contract Boundary
- Keep `status --json` flat and fully stable by default.
- Reject or defer human-only history controls for `--json` until a future machine contract explicitly supports them.
