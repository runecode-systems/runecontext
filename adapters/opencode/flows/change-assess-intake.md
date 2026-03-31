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
