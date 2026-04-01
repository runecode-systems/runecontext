# Generic Flow: change decomposition-apply

Use explicit CLI command proposals:

```sh
runectx change decomposition-apply <UMBRELLA_CHANGE_ID> --sub-change <CHANGE_ID> [--sub-change <CHANGE_ID> ...] [--depends-on <SUB_CHANGE_ID:CHANGE_ID> ...] [--path <project-root>]
```

Apply only after explicit review:

- Confirm umbrella/sub-change IDs and dependency edges before execution.
- If relationships are uncertain, return to `change decomposition-plan` first.
