# RuneContext - Portable, markdown-first project knowledge, standards, and context bundles

[![CI](https://github.com/runecode-systems/runecontext/actions/workflows/ci.yml/badge.svg)](https://github.com/runecode-systems/runecontext/actions/workflows/ci.yml) [![Status: pre-alpha](https://img.shields.io/badge/status-pre--alpha-orange)](docs/implementation-plan/README.md) [![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

RuneContext is a portable, markdown-first, git-native system for project knowledge, standards, changes, and reusable context bundles. It is AI-tooling-agnostic by design, so teams can use the same source artifacts with different tools or manual workflows without turning the format into a product-specific silo.

## Status

RuneContext is pre-alpha and not production-ready. The current repository has the alpha.1 foundation in place: frozen core contracts, versioned schemas, fixtures, Go validation code, and release/dev scaffolding. Broader source resolution, change workflow, deterministic context-pack generation, assurance flows, adapters, and the full CLI surface remain in progress toward `v0.1.0`.

## Why RuneContext

- Portable, repo-native project memory that survives tools, sessions, and branches.
- Reusable standards referenced by path instead of repeatedly copied into specs and changes.
- Low-ceremony shaping and change tracking without heavyweight process theater.
- Deterministic context bundles that can later resolve into auditable context packs.
- One source format that works for AI tooling, automation, and manual review.

## Design Center

- Markdown-first and git-native by default.
- Portable core artifacts with adapters as UX layers, not alternate sources of truth.
- Reviewable suggestions and generated outputs over hidden automatic mutation.
- Deterministic resolution, closed contracts, and fail-closed validation.
- Policy-neutral project knowledge that stays separate from runtime authority.
- Progressive assurance so lightweight use stays lightweight.

## What's Implemented Today

- Normative core contracts in `core/` for terminology, boundaries, layout, and trust rules.
- Versioned schemas in `schemas/` for `runecontext.yaml`, bundles, change status, context packs, specs, and decisions.
- Contract fixtures in `fixtures/` for schema validation, markdown structure, and cross-artifact traceability.
- A Go validation foundation in `internal/contracts/` plus a narrow `runectx validate [path]` command in `cmd/runectx/`.
- Nix, `just`, and GitHub Actions scaffolding for repeatable development, checks, and release work.

Still incremental / not implemented end-to-end yet:

- Embedded, linked, monorepo, and signed-tag source resolution.
- Full context bundle engine and deterministic context-pack generation.
- Change authoring, shaping, promotion, and assurance workflows.
- Thin adapter packs and the broader CLI/automation surface planned for later alphas.

## What The MVP Includes

`v0.1.0` is the RuneContext MVP for this repository. It is planned to include:

- Portable markdown/yaml/json-first source artifacts.
- Embedded and linked source modes, including signed-tag verification support.
- Deterministic context bundle resolution and context-pack hashing.
- Minimum and full change shapes with stable change IDs and traceability.
- `Plain` and `Verified` assurance tiers.
- A small CLI surface for validation, change flows, bundle resolution, and assurance operations.
- Thin adapters as the primary day-to-day UX.
- Repo-first releases, reviewable updates, and compatibility artifacts for deeper external integrations, including RuneCode.

## Roadmap

| Release | Focus |
| --- | --- |
| `v0.1.0-alpha.1` | Core model, naming, file contracts, schemas, canonical data rules, and validation foundation |
| `v0.1.0-alpha.2` | Source resolution, storage modes, monorepo support, and bundle semantics |
| `v0.1.0-alpha.3` | Change workflow, standards linkage, traceability, and history preservation |
| `v0.1.0-alpha.4` | Deterministic context packs, generated indexes, and promotion assessment |
| `v0.1.0-alpha.5` | Plain/Verified assurance, baselines, receipts, and backfill |
| `v0.1.0-alpha.6` | Minimal CLI, validation, doctoring, and machine-facing command contracts |
| `v0.1.0-alpha.7` | Generic and tool-specific adapters plus adapter-pack UX |
| `v0.1.0-alpha.8` | Release/install/update hardening and end-to-end MVP readiness fixtures |
| `v0.1.0` | Stabilization, compatibility freeze, and MVP acceptance sign-off |

## Repository Layout

- `core/` - normative core contracts for RuneContext terminology, boundaries, layout, and trust rules.
- `schemas/` - hand-authored JSON Schemas and machine-readable contract docs.
- `fixtures/` - shared fixture families for schema, markdown, and traceability validation.
- `cmd/runectx/` - the Go CLI entrypoint for `runectx`.
- `internal/` - shared Go implementation packages, including contract validation.
- `adapters/` - tool-specific adapter packs and docs as that surface grows.
- `docs/` - design, planning, installation, and release-process documentation.
- `nix/` - canonical dev-shell, check, and release-artifact definitions.

## Install / Try It Now

The long-term canonical install path is a reviewable repo bundle from GitHub Releases. The simplest way to try the current pre-alpha is to build `runectx` from source.

Quick local install from the current checkout:

```sh
go build -o runectx ./cmd/runectx
install -m 0755 runectx "$HOME/.local/bin/runectx"
runectx validate
```

If you want to try the validator without installing it first:

```sh
go run ./cmd/runectx validate
```

Current alpha.1 note:

- `runectx validate` emits stable line-oriented `key=value` output for automation before broader `--json` support lands.

Common local commands:

```sh
nix develop
just fmt
just lint
just test
just check
```

For release verification and maintainer workflow details, see `docs/install-verify.md` and `docs/release-process.md`.

## Uninstall

To remove a locally installed `runectx` binary:

```sh
rm -f "$HOME/.local/bin/runectx"
```

If you copied or vendored RuneContext files into another repository from a release bundle, remove those files using your normal reviewable project workflow.

## Docs

- `docs/project_idea.md` - the full design idea and product rationale.
- `docs/implementation-plan/README.md` - the alpha-by-alpha implementation plan.
- `core/README.md` - the entrypoint for normative core contracts.
- `schemas/README.md` - schema inventory and contract notes.
- `docs/install-verify.md` - install and verification guidance.
- `docs/release-process.md` - maintainer release workflow.

## Contributing

See `CONTRIBUTING.md`. DCO sign-off is required (`git commit -s`).

## Security

Please do not open public issues for security vulnerabilities. See `SECURITY.md`.

## License

Apache-2.0. See `LICENSE` and `NOTICE`.
