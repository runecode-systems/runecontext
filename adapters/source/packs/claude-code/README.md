# Claude Code Adapter

Claude Code adapter pack for conversational RuneContext workflows.

## Scope

- Keep Claude Code interactions conversational while preserving RuneContext core
  semantics.
- Map flows to explicit `runectx` operations and stable CLI contracts.
- Keep synced files additive in Claude Code-native repo-local locations.

## Capability Declaration

- Prompts: supported
- Shell access: supported
- Hooks: supported
- Dynamic suggestions: supported via shared completion metadata/providers
- Structured output: supported (fallback to explicit CLI contract when disabled)

See `capabilities.yaml` and `flows/conversational-parity.md`.

## Adapter Assets

- Setup guide: `setup.md`
- Conversation playbooks:
  - `flows/change-new.md`
  - `flows/change-shape.md`
  - `flows/standard-discover.md`
  - `flows/promote.md`

## Host-Native Sync Artifacts

Claude Code sync writes additive host-native artifacts:

- Canonical flow assets: `.claude/skills/runecontext-*.md`
- Optional discoverability shim: `.claude/commands/runecontext.md`

Claude Code host-native artifacts use shell-output injection to keep prompt
bodies minimal and machine-oriented:

- ``!`runectx adapter render-host-native --role ... claude-code <operation>` ``

All generated host-native artifacts include `runecontext-managed-artifact:
host-native-v1` so ownership is explicit for future uninstall and upgrade flows.
