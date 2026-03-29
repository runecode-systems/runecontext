## Summary
Extend status summaries with relationship and recency metadata

## Problem
The current `ProjectStatusSummary` and `ChangeStatusEntry` surfaces are too thin to support a useful human-oriented status view. They expose lifecycle buckets and a few labels, but they do not carry enough relationship, verification, or recency data to build hierarchy, sort recent history, or explain why one change appears beneath another.

## Proposed Change
Extend the status summary model to include the fields needed by the human renderer while keeping the current flat `--json` output contract intact. The expanded summary should carry relationship links, verification state, and recency data without forcing a nested machine schema migration.

## Why Now
The renderer and scaling work both depend on richer summary data. Capturing the model expansion separately keeps the sequencing explicit and lets downstream presentation work build on a stable, well-tested summary surface.

## Assumptions
- Inferred size "medium" from the change type because no explicit size was provided.
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Inferred `just test` from the repository's justfile test target.
- The existing lifecycle grouping into active, closed, and superseded sections remains useful and should stay in place.

## Out of Scope
- Replacing the current flat `--json` envelope with nested machine output.
- Final human layout decisions that belong in the renderer change.

## Impact
The richer summary model gives the renderer and history controls the data they need while preserving backward-compatible machine output for scripts and adapters.
