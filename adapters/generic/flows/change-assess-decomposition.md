# Generic Flow: change assess-decomposition

Use explicit CLI command proposals:

```sh
runectx change assess-decomposition <CHANGE_ID> [--path <project-root>]
```

Use advisory fields to drive guided loops:

- Ask emitted `clarification_prompt_*` when `clarification_needed=true`.
- Keep `decomposition_signal`, `eligible_sub_change_*`, and
  `prerequisite_change_*` visible to the user.
- Re-run assessment when clarifications materially change decomposition scope.
- Propose `change decomposition-plan` before `change decomposition-apply` so the
  graph is reviewed before mutation.
