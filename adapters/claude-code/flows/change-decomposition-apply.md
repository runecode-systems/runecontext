# Claude Code Flow: change decomposition-apply

Map conversational decomposition apply to:

```sh
runectx change decomposition-apply <UMBRELLA_CHANGE_ID> --sub-change <CHANGE_ID> [--sub-change <CHANGE_ID> ...] [--depends-on <SUB_CHANGE_ID:CHANGE_ID> ...] [--path <project-root>]
```

Guided behavior:

- Execute only after explicit user confirmation of IDs and edges.
- If relationship intent is unclear, route back to decomposition plan first.
