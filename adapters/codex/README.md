# Codex Adapter

Codex adapter pack for conversational RuneContext workflows.

## Scope

- Keep Codex interactions conversational while preserving RuneContext core
  semantics.
- Map flows to explicit `runectx` operations and stable CLI contracts.
- Keep synced files additive in Codex-native repo-local locations.

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

Codex sync writes additive host-native artifacts:

- Canonical flow assets: `.agents/skills/runecontext-*.md`

Codex host-native integration remains skills-only.

Codex host-native artifacts currently keep static machine-oriented bodies.
Shell-output injection is not enabled yet for Codex.

All generated host-native artifacts include `runecontext-managed-artifact:
host-native-v1` so ownership is explicit for future uninstall and upgrade flows.
