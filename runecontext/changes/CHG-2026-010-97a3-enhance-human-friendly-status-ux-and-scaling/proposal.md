## Summary
Enhance human-friendly status UX and scaling

## Problem
`runectx status` currently emits the same flat key-value contract for both humans and machines, which makes the non-JSON console view hard to scan and hard to trust once a project accumulates many changes. The command does not visually expose umbrella or sub-change relationships, does not distinguish in-flight work from historical noise well, and does not scale gracefully when closed or superseded history grows large.

## Proposed Change
Track the status UX work as an umbrella over three linked deliverables:

- `CHG-2026-011-d50b-extend-status-summaries-with-relationship-and-recency-metadata` for the relationship, verification, and recency fields needed by richer human rendering.
- `CHG-2026-012-f67a-add-human-friendly-status-rendering-with-ascii-hierarchy-and-color` for grouped human output, ASCII relationship trees, and optional terminal color.
- `CHG-2026-013-1f97-add-progressive-disclosure-and-history-controls-to-status` for default bounded history previews, explicit display controls, and scaling hints.

## Why Now
The repository is already using RuneContext actively enough that raw `key=value` status output obscures meaning instead of clarifying it. Capturing the redesign now keeps the machine contract stable while giving the human experience a deliberate plan before more history and relationship complexity accumulate.

## Assumptions
- Inferred size "large" from the change type because no explicit size was provided.
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Inferred `just test` from the repository's justfile test target.
- The flat `runectx status --json` contract remains the stable machine-facing surface for now.

## Out of Scope
- Replacing the current flat `--json` envelope with nested machine output.
- Adding alternate workflow semantics outside the existing change/status model.
- Redesigning unrelated CLI commands in the same effort.

## Impact
The umbrella keeps summary-model changes, rendering work, and scaling controls explicitly linked so the eventual status experience remains readable, relationship-aware, and backward-compatible for machine consumers.
