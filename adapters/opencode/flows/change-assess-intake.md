# OpenCode Flow: change assess-intake

Use this conversational flow to gather read-only intake readiness and shaping
guidance for a proposed change.

## Inputs

- title
- type (`project|feature|bug|standard|chore`)
- optional size (`small|medium|large`)
- optional bundle IDs
- optional description
- optional project path

## Command Mapping

```sh
runectx change assess-intake --title "<title>" --type <type> [--size <size>] [--bundle <bundle-id>] [--description "<text>"] [--path <project-root>]
```

## Review Checkpoint

- Keep this advisory/read-only output visible in the conversation.
- Surface `recommended_mode`, `intake_readiness`, and `decomposition_signal`
  before proposing mutations.

## Guided Clarification Loop

- If `intake_readiness=needs_clarification` or `clarification_needed=true`, ask
  each emitted `clarification_prompt_*` explicitly and wait for user answers.
- Re-run `change assess-intake` with the updated title/type/size/bundle/
  description inputs until readiness no longer requires clarification or the
  user asks to proceed with explicit tradeoffs.
- Keep each advisory result user-visible and avoid hidden adapter-only state.

## Guided Decomposition Handoff

- If `recommended_mode=full`, default to `--shape full` when proposing
  `runectx change new`, unless the user explicitly overrides.
- If `decomposition_signal=consider_decomposition`, propose a decomposition
  sequence after change creation: `change assess-decomposition`, then
  `change decomposition-plan`, then `change decomposition-apply` once reviewed.
