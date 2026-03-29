## Summary
Add human-friendly status rendering with ASCII hierarchy and color

## Problem
The current non-JSON `runectx status` output is technically structured but not human-friendly. It reads as a contract dump instead of a console interface, which makes it hard to distinguish active work from historical context and impossible to visually understand umbrella, dependency, or supersession relationships at a glance.

## Proposed Change
Introduce a dedicated human renderer for `runectx status` that presents grouped sections, ASCII relationship trees, and optional semantic color while leaving `--json` output unchanged.

## Why Now
The status command is the most obvious day-to-day read path for understanding change posture. Improving it now makes the repository easier to use immediately and gives future projects a more scalable, understandable default experience.

## Assumptions
- Inferred size "medium" from the change type because no explicit size was provided.
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Inferred `just test` from the repository's justfile test target.
- ASCII output remains the baseline so the command stays usable in plain terminals, logs, and copy-pasted reviews.

## Out of Scope
- Changing the stable flat machine contract for `status --json`.
- Replacing validated change-relationship semantics with a CLI-only hierarchy model.

## Impact
The new renderer should make status output easier to scan, explain relationships directly in the console, and improve readability without introducing adapter-only semantics.
