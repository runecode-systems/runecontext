# OpenCode Adapter

This adapter is the first end-to-end tool adapter for alpha.7.

## Scope

- Keep OpenCode UX conversational while preserving RuneContext core semantics.
- Map adapter flows to explicit `runectx` operations and stable CLI flags.
- Keep tool-managed files under a namespaced managed subtree when synced with
  `runectx adapter sync opencode`.

## Conversational Flows

The OpenCode adapter provides conversational wrappers for:

- `change new`
- `change shape`
- `standard discover`
- `promote`

See `flows/conversational-parity.md` for CLI-parity mapping.

## Validation Hook

The adapter includes a validate-after-edit hook script:

- `automation/validate_after_authoritative_edit.sh`

The script enforces the alpha.7 authoritative-file boundary:

- It runs `runectx validate --path <project-root>` only after edits to authored
  authoritative RuneContext files.
- It skips generated artifacts, adapter-managed files, and unrelated repository
  code.
