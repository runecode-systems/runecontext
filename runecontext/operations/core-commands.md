# Core Commands

RuneContext exposes a stable CLI operation surface through `runectx`:

- `help`
- `validate`
- `status`
- `change new|shape|close|reallocate`
- `generate indexes`
- `bundle resolve`
- `doctor`
- `init`
- `promote`
- `standard discover`
- `assurance enable|backfill|capture`
- `adapter sync <tool>`
- `completion <bash|zsh|fish>`
- `completion suggest <change-ids|bundle-ids|promotion-targets|adapter-names>`

## Operation Boundaries

- `status`: workflow summary.
- `validate`: authoritative contract enforcement.
- `doctor`: diagnostics/environment/source posture.
- `change*`, `init`, `promote`, and assurance mutating flows: explicit,
  reviewable write operations.
- `completion`: read-only script generation derived from the CLI metadata
  registry.
- `completion suggest`: read-only, repo-aware dynamic suggestions for command
  values.
- `adapter sync`: local-only materialization of tool adapter files into a
  namespaced managed subtree (`.runecontext/adapters/<tool>/managed`) plus a
  convenience sync manifest (`.runecontext/adapters/<tool>/sync-manifest.yaml`).

## Adapter Mapping Rule

Adapters map user-facing conversational/tool-native actions to these same
operations and flags. Adapters must not redefine command semantics or invent
adapter-only operation meaning.
