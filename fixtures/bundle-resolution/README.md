# Bundle Resolution Fixtures

These fixtures cover the alpha.2 bundle engine, precedence rules, and path guardrails.

- `valid-project/`: embedded RuneContext tree used for golden bundle-resolution outputs.
- `golden/*.yaml`: expected resolved bundle outputs, including rule match sets and diagnostics.
- `reject-*/`: fail-closed bundle fixtures for unknown parents, duplicate IDs, cycles, depth overflow, invalid patterns, and path escapes.

The bundle goldens focus on the reusable data shape needed by later context-pack generation work:

- linearized bundle order
- per-aspect ordered rule evaluation
- concrete per-rule match sets
- selected and excluded inventories
- rule provenance and diagnostics
