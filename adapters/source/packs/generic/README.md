# Generic Adapter

The generic adapter is the host-agnostic baseline for teams that prefer explicit,
reviewable CLI flows.

## Scope

- The generic adapter is docs-first and host-agnostic.
- It does not own hidden runtime suggestion logic.
- Dynamic suggestions are implemented in shared CLI/completion surfaces and can
  be reused by richer tool adapters.

## Capability Declaration

- Prompts: optional (fallback to static guidance and command proposals)
- Shell access: optional (fallback to user-run command steps)
- Hooks: optional (fallback to explicit `runectx validate` runs)
- Dynamic suggestions: optional (fallback to `runectx completion suggest` output)
- Structured output: optional (fallback to explicit CLI flag/value sets)

See `capabilities.yaml` for machine-readable declarations.

## Example Flows

- Manual flow: `examples/manual-flow.md`
- CLI-assisted flow: `examples/cli-assisted-flow.md`
- Non-agent flow: `examples/non-agent-flow.md`

## Shared Suggestion Entry Points

- `runectx completion suggest change-ids`
- `runectx completion suggest bundle-ids`
- `runectx completion suggest promotion-targets`
- `runectx completion suggest adapter-names`

These providers are read-only, support nearest-root discovery, honor `--path`,
and soft-fail outside RuneContext projects.
