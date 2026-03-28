# Design

## Overview
Capture the status UX enhancement as an umbrella over summary-model expansion, a dedicated human renderer, and explicit scaling controls. The human path should become easier to read without weakening the stable flat `--json` contract that scripts and adapters already consume.

## Planned Sub-Changes
- `CHG-2026-011-d50b-extend-status-summaries-with-relationship-and-recency-metadata` owns summary-field expansion for relationships, verification state, and recency.
- `CHG-2026-012-f67a-add-human-friendly-status-rendering-with-ascii-hierarchy-and-color` owns the human renderer, grouping model, ASCII tree output, and terminal color behavior.
- `CHG-2026-013-1f97-add-progressive-disclosure-and-history-controls-to-status` owns default bounded history previews, display-control flags, and hidden-item hints for large projects.

## UX Direction
- Default non-JSON `runectx status` should become a human-first console view rather than a raw contract dump.
- Group sections by meaning, with in-flight work emphasized and historical sections de-emphasized.
- Show change associations with ASCII tree structure first (`|-` and `\-`), then reinforce meaning with optional color when the terminal allows it.
- Keep symbols sparse and never rely on color alone to carry meaning.
- Use lifecycle-first multiline rows with compact IDs in default output and full IDs only under `--verbose`.
- Apply renderer-controlled wrapping to titles and relationship hints so output remains readable in narrow terminals.
- Preserve the flat machine contract for `--json` until a future structured contract is intentionally designed.

## Scaling Defaults
- Show all active or in-flight changes by default.
- Show closed and superseded history as bounded count-based previews rather than date-window slices.
- Emit clear hints when entries are hidden, for example `showing 5 of 143 closed changes; use --history all`.
- Historical preview sections must behave cleanly for zero entries, fewer-than-limit entries, and very large histories.

## Relationship Model
- Use reciprocal `related_changes` links between the umbrella and all three sub-changes for navigability.
- Reserve `depends_on` for real sequencing only; rendering and disclosure work should depend on summary metadata expansion rather than on umbrella semantics alone.
