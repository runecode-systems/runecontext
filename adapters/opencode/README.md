# OpenCode Adapter

This adapter is the first end-to-end tool adapter for alpha.7.

## Scope

- Keep OpenCode UX conversational while preserving RuneContext core semantics.
- Map adapter flows to explicit `runectx` operations and stable CLI flags.
- Keep tool-managed files under a namespaced managed subtree when synced with
  `runectx adapter sync opencode`.

## Capability Declaration

- Prompts: supported
- Shell access: supported
- Hooks: supported
- Dynamic suggestions: supported via shared completion metadata/providers
- Structured output: supported

See `capabilities.yaml` for machine-readable declarations.

## Conversational Flows

The OpenCode adapter provides conversational wrappers for:

- `change new`
- `change shape`
- `standard discover`
- `promote`

See `flows/conversational-parity.md` for CLI-parity mapping.

Detailed flow playbooks:

- `flows/change-new.md`
- `flows/change-shape.md`
- `flows/standard-discover.md`
- `flows/promote.md`

## Validation Hook

The adapter includes a validate-after-edit hook script:

- `automation/validate_after_authoritative_edit.sh`

The script enforces the alpha.7 authoritative-file boundary:

- It runs `runectx validate --path <project-root>` only after edits to authored
  authoritative RuneContext files.
- It skips generated artifacts, adapter-managed files, and unrelated repository
  code.

## Host-Native Sync Artifacts

OpenCode sync also writes host-native additive artifacts for discoverability:

- Canonical flow assets: `.opencode/skills/runecontext-*.md`
- Discoverability shims: `.opencode/commands/runecontext-*.md`

All generated host-native artifacts include `runecontext-managed-artifact:
host-native-v1` so ownership is explicit for future uninstall and upgrade flows.
