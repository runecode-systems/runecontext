# Core Commands

RuneContext exposes a stable CLI operation surface through `runectx`:

- `help`
- `version`
- `validate`
- `status`
- `change new|shape|close|reallocate`
- `generate indexes`
- `bundle resolve`
- `doctor`
- `init`
- `upgrade|upgrade apply`
- `promote`
- `standard discover`
- `assurance enable|backfill|capture`
- `adapter sync <tool>`
- `completion <bash|zsh|fish>`
- `completion suggest <change-ids|bundle-ids|promotion-targets|adapter-names|adapter-names-shell-injection>`
- `completion metadata`

## Operation Boundaries

- `status`: workflow summary.
- `validate`: authoritative contract enforcement.
- `doctor`: diagnostics/environment/source posture.
- `version`: installed CLI release identification.
- `upgrade`: preview-first project upgrade planning.
- `change*`, `init`, `promote`, and assurance mutating flows: explicit,
  reviewable write operations.
- `upgrade apply`: explicit transactional mutation after preview.
- `completion`: read-only script generation derived from the CLI metadata
  registry.
- `completion suggest`: read-only, repo-aware dynamic suggestions for command
  values.
- `completion metadata`: read-only machine-readable completion metadata derived
  from the typed command registry.
- `adapter sync`: local-only materialization of repo-local host-native adapter
  artifacts in tool-native locations (`.opencode/*`, `.claude/*`, `.agents/*`)
  with explicit ownership markers and no `.runecontext/adapters` mirror tree.

## Adapter Mapping Rule

Adapters map user-facing conversational/tool-native actions to these same
operations and flags. Adapters must not redefine command semantics or invent
adapter-only operation meaning.
