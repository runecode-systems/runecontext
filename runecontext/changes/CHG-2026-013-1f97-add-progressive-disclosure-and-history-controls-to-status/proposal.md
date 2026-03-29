## Summary
Add progressive disclosure and history controls to status

## Problem
Even with a better renderer, `runectx status` will become cumbersome if it always prints the full closed and superseded history for long-lived projects. The human view needs progressive disclosure so it stays actionable by default while still letting operators expand history intentionally when they need deeper context.

## Proposed Change
Add bounded historical previews and explicit display controls to human `runectx status` output. The default should always show all active work, preview closed and superseded history by count, and explain when additional history is hidden.

## Why Now
Status output readability will degrade as soon as a project accumulates substantial closed history. Planning the scaling behavior now avoids a redesign that works only for small projects.

## Assumptions
- Inferred size "medium" from the change type because no explicit size was provided.
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Inferred `just test` from the repository's justfile test target.
- Count-based previews are easier to reason about than date-window heuristics for the initial rollout.

## Out of Scope
- Changing the default machine-facing `status --json` payload to a bounded or nested shape.
- Archival or search workflows beyond the status command itself.

## Impact
The command stays readable for large projects, preserves full visibility into active work, and gives users clear affordances for expanding or suppressing historical sections when needed.
