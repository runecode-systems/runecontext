# Codex Adapter Setup

This adapter is synced with:

```sh
runectx adapter sync codex --path <project-root>
```

The sync writes only to:

- `.runecontext/adapters/codex/managed/`
- `.runecontext/adapters/codex/sync-manifest.yaml`
- `.agents/skills/runecontext-*.md`

Codex host-native integration is skills-only.

All generated host-native files include the ownership marker:

- `runecontext-managed-artifact: host-native-v1`

No implicit network fetches occur during sync.
