# Typed Metadata Registry

The canonical CLI metadata registry lives in:

- `internal/cli/cli_metadata_registry.go`

## Model

The typed model includes:

- Command path and usage (`CommandMetadata`)
- Subcommands
- Flags with value kind (`none`, `text`, `enum`)
- Stable enum values for flags and positionals
- Positional argument metadata
- Optional dynamic suggestion provider names for text flags/positionals

## Derivations

From this one registry, RuneContext derives:

- Human-readable operation docs in `runecontext/operations/`
- Machine-readable completion metadata (`CLICompletionMetadata`)
- Static shell completion scripts (`runectx completion <shell>`)
- Shared dynamic suggestion surfaces (`runectx completion suggest ...`)

## Drift Prevention

Parity tests in `internal/cli` validate that registry usage/flag metadata stays
aligned with the executable CLI surface.
