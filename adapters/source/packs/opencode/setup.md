# OpenCode Adapter Setup

Sync the OpenCode adapter with:

```sh
runectx adapter sync opencode --path <project-root>
```

Sync writes RuneContext-managed files to:

- `.opencode/skills/runecontext-*.md`
- `.opencode/commands/runecontext-*.md`

The `.opencode/skills/` files are canonical host-native flow assets.

The `.opencode/commands/` files are discoverability shims.

OpenCode files include shell-output injection calls to:

- `runectx adapter render-host-native`

All generated host-native files include the ownership marker:

- `runecontext-managed-artifact: host-native-v1`

Adapter sync remains local-only and never fetches from the network.
