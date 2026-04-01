# Generic Flow: change assess-intake

Use explicit CLI command proposals:

```sh
runectx change assess-intake --title "<title>" --type <type> [--size <size>] [--bundle <bundle-id>] [--description "<text>"] [--path <project-root>]
```

Use advisory fields to drive guided loops:

- Ask emitted `clarification_prompt_*` when `intake_readiness=needs_clarification`
  or `clarification_needed=true`.
- Re-run assessment with updated inputs until guidance stabilizes.
- If `recommended_mode=full`, prefer `runectx change new --shape full` unless
  the user explicitly overrides.
- If `decomposition_signal=consider_decomposition`, propose follow-up with
  `change assess-decomposition`, then `change decomposition-plan`, then
  `change decomposition-apply` after plan review.
