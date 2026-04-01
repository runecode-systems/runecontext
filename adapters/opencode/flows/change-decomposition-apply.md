# OpenCode Flow: change decomposition-apply

Use this conversational flow to apply a decomposition graph rewrite across an
umbrella change and selected sub-changes.

## Inputs

- umbrella change ID
- one or more sub-change IDs
- zero or more dependency edges (`SUB_CHANGE_ID:CHANGE_ID`)
- optional project path

## Command Mapping

```sh
runectx change decomposition-apply <UMBRELLA_CHANGE_ID> --sub-change <CHANGE_ID> [--sub-change <CHANGE_ID> ...] [--depends-on <SUB_CHANGE_ID:CHANGE_ID> ...] [--path <project-root>]
```

## Review Checkpoint

- Confirm umbrella and sub-change IDs before execution.
- Review changed status files and relationship rewrites before commit.

## Guided Clarification Guardrails

- If the user is unsure about relationships or missing prerequisites, pause and
  return to `change decomposition-plan` for another advisory pass.
- Keep command proposals explicit and user-confirmed; do not apply hidden
  relationship edits.
