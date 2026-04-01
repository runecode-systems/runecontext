# Codex Adapter Setup

This adapter is synced with:

```sh
runectx adapter sync codex --path <project-root>
```

The sync writes only to:

- `.agents/skills/runecontext-*.md`

Codex host-native integration is skills-only.

Codex host-native files currently use static machine-oriented bodies.

All generated host-native files include the ownership marker:

- `runecontext-managed-artifact: host-native-v1`

No implicit network fetches occur during sync.
