# Design

## Overview
Add a structured `runectx change update` command for non-terminal lifecycle movement so users can advance a change from `proposed` to `planned`, `implemented`, and `verified` without invoking terminal close behavior.

## Command Contract
- Support `runectx change update <CHANGE_ID> --status planned|implemented|verified [--path PATH]` as the primary lifecycle mutation surface for non-terminal changes.
- Keep `runectx change close` as the only command that writes `closed` or `superseded` and the only command that runs close-time promotion assessment.
- Preserve stable JSON and human output contracts so adapters and scripts can use the command safely.

## Mutation Rules
- Honor existing lifecycle ordering rules and reject backward transitions.
- Reject updates to terminal changes instead of rewriting history.
- Leave `promotion_assessment` unchanged for non-terminal lifecycle updates.
- Keep relationship edits safe and reviewable if they are included under the same command family.

## Discussion Capture
- This change exists because dogfooding the status UI work exposed a missing workflow step: implemented changes needed to move beyond `proposed`, but the CLI only offered `shape` and terminal `close` actions.
- The command should make the lifecycle intent explicit so users do not confuse change progression with promotion acceptance.
