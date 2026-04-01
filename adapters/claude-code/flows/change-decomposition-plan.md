# Claude Code Flow: change decomposition-plan

Map conversational decomposition planning to:

```sh
runectx change decomposition-plan <UMBRELLA_CHANGE_ID> --sub-change <CHANGE_ID> [--sub-change <CHANGE_ID> ...] [--depends-on <SUB_CHANGE_ID:CHANGE_ID> ...] [--path <project-root>]
```

Guided behavior:

- Keep `graph_*` output visible for review.
- If graph intent is unclear, gather clarifications and re-run with revised
  `--sub-change`/`--depends-on` values.
- Use the approved plan inputs when mapping to decomposition apply.
