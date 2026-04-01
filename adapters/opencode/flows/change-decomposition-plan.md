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
