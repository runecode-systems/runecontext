# OpenCode Flow: change decomposition-plan

Use this conversational flow to compute an advisory decomposition graph for an
umbrella change and selected sub-changes.

## Inputs

- umbrella change ID
- one or more sub-change IDs
- zero or more dependency edges (`SUB_CHANGE_ID:CHANGE_ID`)
- optional project path

## Command Mapping

```sh
runectx change decomposition-plan <UMBRELLA_CHANGE_ID> --sub-change <CHANGE_ID> [--sub-change <CHANGE_ID> ...] [--depends-on <SUB_CHANGE_ID:CHANGE_ID> ...] [--path <project-root>]
```

## Review Checkpoint

- Confirm umbrella and sub-change IDs before execution.
- Keep advisory graph outputs (`graph_*`) user-visible for review.

## Guided Clarification And Iteration

- If the planned graph conflicts with user intent, gather clarifications on
  missing sub-changes or dependency edges and re-run planning.
- Preserve reviewability by proposing exact revised `--sub-change` and
  `--depends-on` arguments before each re-run.

## Apply Handoff

- When the user accepts the advisory graph, hand off to
  `runectx change decomposition-apply` with the same umbrella/sub-change/edge
  inputs used for the approved plan.
