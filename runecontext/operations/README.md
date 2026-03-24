# RuneContext Operations Reference

This directory is the canonical in-repository operations reference for the
`runectx` CLI and adapter-facing operation semantics.

## Scope

- Documents the stable operation surface exposed by `runectx`.
- Documents how command metadata is represented in a typed command registry.
- Documents how shell completion and machine-readable completion metadata are
  derived from that registry.
- Documents shared repo-aware suggestion providers used by shell completion and
  adapter-native UX.

## Source Of Truth

The canonical operation/CLI metadata source is the typed command registry in
`internal/cli/cli_metadata_registry.go`.

Operations docs and completion outputs are derived from that registry. This
prevents command/flag drift between help text, adapters, and completion.

## Files

- `core-commands.md`: operation and command/subcommand reference.
- `metadata-registry.md`: typed metadata model and derivation rules.
- `completion.md`: completion generation model for Bash/Zsh/Fish.
