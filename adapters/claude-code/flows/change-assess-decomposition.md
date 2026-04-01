# Claude Code Flow: change assess-decomposition

Map conversational decomposition assessment to:

```sh
runectx change assess-decomposition <CHANGE_ID> [--path <project-root>]
```

Guided behavior:

- Ask emitted `clarification_prompt_*` when `clarification_needed=true`.
- Keep decomposition advisory fields user-visible.
- Re-run assessment when clarified scope materially changes graph planning.
- Prefer `change decomposition-plan` before `change decomposition-apply`.
