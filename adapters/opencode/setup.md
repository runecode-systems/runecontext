# OpenCode Adapter Setup

Sync the OpenCode adapter with:

```sh
runectx adapter sync opencode --path <project-root>
```

Sync writes RuneContext-managed files to:

- `.runecontext/adapters/opencode/managed/`
- `.runecontext/adapters/opencode/sync-manifest.yaml`
- `.opencode/skills/runecontext-*.md`
- `.opencode/commands/runecontext-*.md`

The `.opencode/skills/` files are canonical host-native flow assets.

The `.opencode/commands/` files are discoverability shims.

All generated host-native files include the ownership marker:

- `runecontext-managed-artifact: host-native-v1`

Adapter sync remains local-only and never fetches from the network.
