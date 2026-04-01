# OpenCode Flow: change assess-decomposition

Use this conversational flow to gather read-only decomposition signals for an
existing change.

## Inputs

- change ID
- optional project path

## Command Mapping

```sh
runectx change assess-decomposition <CHANGE_ID> [--path <project-root>]
```

## Review Checkpoint

- Confirm the target `CHANGE_ID` before execution.
- Keep advisory outputs (`decomposition_signal`, `eligible_sub_change_*`, and
  `prerequisite_change_*`) user-visible.

## Guided Clarification Loop

- If `clarification_needed=true`, ask each `clarification_prompt_*` and confirm
  any ambiguous scope before proposing write operations.
- Re-run `change assess-decomposition` after major clarifications when the
  target graph assumptions change.

## Guided Decomposition Progression

- If decomposition is indicated, propose `change decomposition-plan` first and
  keep `graph_*` output user-visible for review.
- Only propose `change decomposition-apply` after the reviewed plan and edge
  list are explicitly accepted.
