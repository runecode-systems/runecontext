# Completion Model

RuneContext completion is generated from the canonical typed command registry.

## Command

- `runectx completion bash`
- `runectx completion zsh`
- `runectx completion fish`

## What Is Included

- Static command and subcommand completion
- Static flag completion
- Enum completion for stable flag values (for example `--mode`, `--type`)
- Enum completion for stable positional enums (for example
  `completion <bash|zsh|fish>`)

## Consistency Guarantees

- Completion metadata is derived from one typed registry, not parsed from help
  text.
- Golden tests validate generated Bash/Zsh/Fish scripts.
- Parity tests validate metadata and CLI surface alignment.
