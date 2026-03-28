# Design

## Overview
Render non-JSON `runectx status` output through a dedicated human-oriented formatter instead of the current raw `key=value` line emitter. The renderer should use the richer summary data to present meaningful sections, show change relationships with ASCII structure, and optionally reinforce status meaning with color.

## Layout Direction
- Present a compact header with root, selected config path, RuneContext version, assurance tier, and bundle summary.
- Group changes by meaning, with sections such as `In Flight`, `Recently Completed`, and `Replaced`.
- Use lifecycle-first rows in the form `[lifecycle] CHG-YYYY-NNN [type size]` with title content rendered on controlled continuation lines.
- Keep compact IDs in default human output and reserve full IDs for `--verbose`.

## Relationship Rendering
- Build section-local ASCII trees for umbrella or dependency associations using `|-` and `\-` connectors.
- Show explicit dependency or supersession hints when a simple tree is not enough to explain the link.
- Avoid inventing hidden hierarchy; render the validated change graph in a predictable, readable order.
- If association ownership is ambiguous, fall back to flat rows with explicit relationship hints instead of forcing a tree.

## Wrap Behavior
- Do not rely on terminal auto-wrap for core rows.
- Wrap title and hint lines using renderer-controlled widths so wrapped output remains aligned with the association tree.
- Keep relationship hint IDs compact by default and expand to full IDs only under `--verbose`.

## Color Behavior
- Start with ASCII-only output, then add color only when the terminal supports it.
- Honor `NO_COLOR` and avoid color on non-TTY outputs.
- Use color to reinforce meaning, not to carry meaning alone, and keep symbols sparse.

## Contract Boundary
- `--json` should continue to use the existing flat machine envelope.
- The renderer should consume summary data rather than re-reading files directly so tests and behavior stay deterministic.
