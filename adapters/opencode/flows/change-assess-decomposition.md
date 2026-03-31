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
