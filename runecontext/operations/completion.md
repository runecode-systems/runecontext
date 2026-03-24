# Completion Model

RuneContext completion is generated from the canonical typed command registry.

## Command

- `runectx completion bash`
- `runectx completion zsh`
- `runectx completion fish`
- `runectx completion suggest <provider>`

Providers:

- `change-ids`
- `bundle-ids`
- `promotion-targets`
- `adapter-names`

## What Is Included

- Static command and subcommand completion
- Static flag completion
- Enum completion for stable flag values (for example `--mode`, `--type`)
- Enum completion for stable positional enums (for example
  `completion <bash|zsh|fish>`)
- Dynamic repo-aware suggestions for selected text arguments (for example
  change IDs, bundle IDs, promotion targets)

## Repo-Aware Suggestion Behavior

- Suggestion providers are shared CLI/completion infrastructure, not
  `adapters/generic`-specific hidden behavior.
- Suggestions are read-only and never mutate project state.
- Discovery uses nearest-root behavior by default and honors explicit `--path`.
- Outside RuneContext projects, providers soft-fail by returning no suggestions.
- Adapter names are discovered from the repository `adapters/` directory.

## Consistency Guarantees

- Completion metadata is derived from one typed registry, not parsed from help
  text.
- Golden tests validate generated Bash/Zsh/Fish scripts.
- Parity tests validate metadata and CLI surface alignment.
