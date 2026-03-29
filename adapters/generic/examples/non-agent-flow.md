# Generic Adapter: Non-Agent Flow

Workflow for developers who want static docs and direct command usage.

1. Use generated reference surfaces in `docs/reference/generated/runecontext-reference.{json,yaml}` for docs tooling and stable command/capability IDs.
2. Use `runectx completion metadata` for machine-readable command/flag metadata.
3. Keep all mutations explicit and reviewable with normal version control diffs.
