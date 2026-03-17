# RuneContext

RuneContext is a portable, markdown-first, git-native project knowledge system.

This repository now includes the initial Go/Nix scaffold plus the first frozen
core contract docs for `v0.1.0-alpha.1` Epic 1:

- `cmd/runectx/` - the Go CLI entrypoint for `runectx`
- `internal/` - shared Go packages, including the contract-validation foundation under `internal/contracts/`
- `tools/releasebuilder/` - placeholder for future release helper tooling
- `core/` - normative core contract docs for terminology, boundaries, layout,
  and trust-boundary rules
- `adapters/` - tool-specific adapter packs and docs
- `schemas/` - placeholder area for hand-authored JSON Schema files
- `nix/` - flake support for dev shells, checks, and canonical release artifacts

Common commands:

- `go run ./cmd/runectx validate`

The alpha.1 `runectx validate` command uses a stable line-oriented `key=value`
output contract for automation without introducing full `--json` yet.
Consumers should parse each line by splitting on the first `=`.
- `nix develop`
- `just fmt`
- `just lint`
- `just test`
- `just check`
- `just release`

Release docs:

- `docs/install-verify.md`
- `docs/release-process.md`

## Install

RuneContext supports two straightforward installation paths today.

- Repo bundle: download and verify a release bundle, then copy or vendor the
  released files into your target repository. This is the canonical path.
- CLI binary: download and verify a platform `runectx` archive, then place the
  `runectx` binary on your `PATH` for validation and future CLI-assisted flows.

For full verification steps, use `docs/install-verify.md`.

Quick local install of the current source checkout:

```sh
go build -o runectx ./cmd/runectx
install -m 0755 runectx "$HOME/.local/bin/runectx"
```

Make sure `$HOME/.local/bin` is on your `PATH`.

## Uninstall

To remove a locally installed `runectx` binary:

```sh
rm -f "$HOME/.local/bin/runectx"
```

If you vendored RuneContext files into a project from a release bundle, remove
those copied files using your normal reviewable project workflow.

Core contract entrypoint:

- `core/README.md`

## Contributing

See `CONTRIBUTING.md`. DCO sign-off is required (`git commit -s`).

## Security

Please do not open public issues for security vulnerabilities. See
`SECURITY.md`.

## License

Apache-2.0. See `LICENSE` and `NOTICE`.
