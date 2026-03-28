## Applicable Standards
- `standards/global/cli-mutation-safety.md`: Standard authoring commands should use the same transactional, validate-after-write mutation safety model as other structured CLI write flows.
- `standards/global/structured-cli-contracts.md`: New standard authoring commands should expose stable machine-facing contracts that adapters and shell tooling can reuse.
- `standards/go/thin-cli-fat-contracts.md`: The CLI surface should remain thin while standard authoring semantics live in the contracts layer.

## Resolution Notes
These standards focus the work on safe standard mutations, stable machine contracts, and keeping authoring semantics in the contracts layer rather than the CLI shell.
