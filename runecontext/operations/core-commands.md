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
- `completion <bash|zsh|fish>`

## Operation Boundaries

- `status`: workflow summary.
- `validate`: authoritative contract enforcement.
- `doctor`: diagnostics/environment/source posture.
- `change*`, `init`, `promote`, and assurance mutating flows: explicit,
  reviewable write operations.
- `completion`: read-only script generation derived from the CLI metadata
  registry.

## Adapter Mapping Rule

Adapters map user-facing conversational/tool-native actions to these same
operations and flags. Adapters must not redefine command semantics or invent
adapter-only operation meaning.
