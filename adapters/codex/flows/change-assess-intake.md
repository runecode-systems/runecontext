# Codex Flow: change assess-intake

Map conversational intake assessment to:

```sh
runectx change assess-intake --title "<title>" --type <type> [--size <size>] [--bundle <bundle-id>] [--description "<text>"] [--path <project-root>]
```

Guided behavior:

- Ask each emitted `clarification_prompt_*` when `intake_readiness` indicates
  clarification or `clarification_needed=true`.
- Re-run intake assessment with updated inputs until guidance stabilizes.
- If `recommended_mode=full`, default to proposing `runectx change new --shape full`
  unless the user explicitly overrides.
- If `decomposition_signal=consider_decomposition`, propose follow-up:
  `change assess-decomposition` -> `change decomposition-plan` ->
  `change decomposition-apply` after plan approval.
