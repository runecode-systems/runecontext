# Generic Flow: change decomposition-plan

Use explicit CLI command proposals:

```sh
runectx change decomposition-plan <UMBRELLA_CHANGE_ID> --sub-change <CHANGE_ID> [--sub-change <CHANGE_ID> ...] [--depends-on <SUB_CHANGE_ID:CHANGE_ID> ...] [--path <project-root>]
```

Keep plan output reviewable before mutation:

- Keep `graph_*` output user-visible.
- If clarification is needed, revise `--sub-change` and `--depends-on` inputs
  and re-run planning.
- After explicit user approval of the plan, map the same graph inputs to
  `runectx change decomposition-apply`.
