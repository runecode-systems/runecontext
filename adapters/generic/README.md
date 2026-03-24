# Generic Adapter

Generic adapter materials and tool-agnostic command-pack docs will live here.

## Scope

- The generic adapter is docs-first and host-agnostic.
- It does not own hidden runtime suggestion logic.
- Dynamic suggestions are implemented in shared CLI/completion surfaces and can
  be reused by richer tool adapters.

## Shared Suggestion Entry Points

- `runectx completion suggest change-ids`
- `runectx completion suggest bundle-ids`
- `runectx completion suggest promotion-targets`
- `runectx completion suggest adapter-names`

These providers are read-only, support nearest-root discovery, honor `--path`,
and soft-fail outside RuneContext projects.
