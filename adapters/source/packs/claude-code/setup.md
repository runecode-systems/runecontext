# Claude Code Adapter Setup

This adapter is synced with:

```sh
runectx adapter sync claude-code --path <project-root>
```

The sync writes only to:

- `.claude/skills/runecontext-*.md`
- `.claude/commands/runecontext.md`

The `.claude/skills/` files are canonical host-native flow assets.

The `.claude/commands/runecontext.md` file is an optional discoverability shim.

Claude Code files include shell-output injection calls to:

- `runectx adapter render-host-native`

All generated host-native files include the ownership marker:

- `runecontext-managed-artifact: host-native-v1`

No implicit network fetches occur during sync.
