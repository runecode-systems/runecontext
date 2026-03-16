# RuneContext

RuneContext is a portable, markdown-first, git-native project knowledge system.

This repository now includes the initial Go/Nix scaffold plus the first frozen
core contract docs for `v0.1.0-alpha.1` Epic 1:

- `cmd/runectx/` - placeholder for the future Go CLI entrypoint
- `internal/` - placeholder for shared Go packages
- `tools/releasebuilder/` - placeholder for future release helper tooling
- `core/` - normative core contract docs for terminology, boundaries, layout,
  and trust-boundary rules
- `adapters/` - tool-specific adapter packs and docs
- `schemas/` - placeholder area for hand-authored JSON Schema files
- `nix/` - flake support for dev shells, checks, and canonical release artifacts

Common commands:

- `nix develop`
- `just fmt`
- `just lint`
- `just test`
- `just check`
- `just release`

Core contract entrypoint:

- `core/README.md`

## Contributing

See `CONTRIBUTING.md`. DCO sign-off is required (`git commit -s`).

## Security

Please do not open public issues for security vulnerabilities. See
`SECURITY.md`.

## License

Apache-2.0. See `LICENSE` and `NOTICE`.
