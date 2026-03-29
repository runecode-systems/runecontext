# Tasks

- Add a human-oriented status renderer for non-JSON output without changing the existing machine envelope.
- Render grouped sections with ASCII hierarchy for related umbrella, sub-change, dependency, and supersession links.
- Finalize lifecycle-first multiline row styling and compact-ID defaults, with full IDs shown only under `--verbose`.
- Ensure title and relationship hint wrapping is renderer-controlled and tree-aligned for narrow terminals.
- Add optional semantic color with safe non-TTY and `NO_COLOR` behavior.
- Add human-output tests that cover plain ASCII output and color-disabled deterministic rendering.
