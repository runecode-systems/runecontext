# OpenCode Adapter

This adapter is the first end-to-end tool adapter for alpha.7.

## Scope

- Keep OpenCode UX conversational while preserving RuneContext core semantics.
- Map adapter flows to explicit `runectx` operations and stable CLI flags.
- Keep synced files additive in OpenCode-native repo-local locations.

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

The validate-after-edit hook script remains part of the adapter source package:

- `automation/validate_after_authoritative_edit.sh`

Host-native `runectx adapter sync opencode` currently materializes only
`.opencode/skills/` and `.opencode/commands/` artifacts. It does not
materialize the automation script into the repository root.

## Host-Native Sync Artifacts

OpenCode sync also writes host-native additive artifacts for discoverability:

- Canonical flow assets: `.opencode/skills/runecontext-*.md`
- Discoverability shims: `.opencode/commands/runecontext-*.md`

OpenCode host-native artifacts use shell-output injection to keep prompt bodies
minimal and machine-oriented:

- ``!`runectx adapter render-host-native --role ... opencode <operation>` ``

All generated host-native artifacts include `runecontext-managed-artifact:
host-native-v1` so ownership is explicit for future uninstall and upgrade flows.
