# RuneContext - Portable, markdown-first project knowledge, standards, and context bundles

[![CI](https://github.com/runecode-systems/runecontext/actions/workflows/ci.yml/badge.svg)](https://github.com/runecode-systems/runecontext/actions/workflows/ci.yml) [![Status: alpha.5](https://img.shields.io/badge/status-alpha.5-orange)](docs/implementation-plan/README.md) [![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

RuneContext is a portable, markdown-first, git-native system for project knowledge, standards, changes, and reusable context bundles. It is AI-tooling-agnostic by design, so teams can use the same source artifacts with different tools or manual workflows without turning the format into a product-specific silo.

## Status

RuneContext is still pre-MVP and not production-ready. The current repository now includes the alpha.1 contract foundation, the alpha.2 source-resolution and bundle-engine slice, the alpha.3 change-workflow slice, the alpha.4 context-pack slice, and the alpha.5 broadened CLI slice: frozen core contracts, versioned schemas, fixtures, Go validation code, embedded/git/path source resolution, monorepo discovery, signed-tag verification with explicit trusted-signer input, deterministic bundle semantics, standards linkage, traceability, stable change IDs, deterministic context-pack generation and hashing, generated indexes/manifests, reviewable promotion assessment, local-first `init` scaffolding, `bundle resolve`, `doctor`, explicit `promote`, advisory `standard discover`, and shared machine-facing CLI contracts with `--json`, `--non-interactive`, `--dry-run`, and `--explain`. Verified assurance flows, adapters, completion/sync surfaces, and release/install/update hardening remain in progress toward `v0.1.0`.

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
- Versioned schemas in `schemas/` for `runecontext.yaml`, bundles, change status, context packs, specs, decisions, and standards.
- Contract fixtures in `fixtures/` for schema validation, markdown structure, cross-artifact traceability, source resolution, bundle resolution, and change workflow.
- A Go validation, resolution, change-workflow, context-pack, promotion, and broadened CLI implementation in `internal/contracts/`, `internal/cli/`, and `cmd/runectx/`.
- Source resolution for embedded projects, linked git sources by pinned commit, linked git sources by signed tag, opt-in mutable refs, local path sources, and nearest-ancestor monorepo discovery.
- Change authoring and history-preserving workflow operations for stable change IDs, minimum/full shaping, standards linkage, lifecycle validation, and fail-closed close/reallocate behavior.
- Deterministic context bundle loading and evaluation with inheritance linearization, cycle/depth rejection, ordered include/exclude precedence, concrete per-rule match inventories, and fail-closed path/symlink containment.
- Deterministic context-pack building and reporting with explicit whole-second `generated_at`, stable `pack_hash` output, normalized LF/CRLF hashing, persisted provenance, explanation output, advisory thresholds, and fail-closed rebuild checks.
- Stable generated `manifest.yaml`, `indexes/changes-by-status.yaml`, and `indexes/bundles.yaml` artifacts plus reviewable close-time promotion assessment suggestions and explicit promotion transitions.
- CLI support for `runectx init`, `runectx validate`, `runectx status`, `runectx change new`, `runectx change shape`, `runectx change close`, `runectx change reallocate`, `runectx bundle resolve`, `runectx doctor`, `runectx promote`, and `runectx standard discover`.
- Shared machine-facing CLI behavior across commands with stable `key=value` output by default plus `--json`, `--non-interactive`, `--dry-run` for write flows, and incremental `--explain` support.
- Nix, `just`, and GitHub Actions scaffolding for repeatable development, checks, and release work.

Still incremental / not implemented end-to-end yet:

- Verified assurance flows, baselines, receipts, and backfill.
- Assurance enablement/backfill commands and the broader alpha.6 assurance surface.
- Shell completion, adapter sync, thin adapter packs, and the later alpha adapter-management and automation surface.
- Release/install/update hardening and network-enabled init/update flows.

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
| `v0.1.0-alpha.5` | Broadened CLI, local-first `init`, promotion/resolve flows, validation, doctoring, and machine-facing command contracts |
| `v0.1.0-alpha.6` | Plain/Verified assurance, baselines, receipts, and backfill |
| `v0.1.0-alpha.7` | Generic and tool-specific adapters plus adapter-pack UX |
| `v0.1.0-alpha.8` | Release/install/update hardening and end-to-end MVP readiness fixtures |
| `v0.1.0` | Stabilization, compatibility freeze, and MVP acceptance sign-off |

## Repository Layout

- `core/` - normative core contracts for RuneContext terminology, boundaries, layout, and trust rules.
- `schemas/` - hand-authored JSON Schemas and machine-readable contract docs.
- `fixtures/` - shared fixture families for schema, markdown, traceability, source/bundle resolution, and change-workflow validation.
- `cmd/runectx/` - the Go CLI entrypoint for `runectx`.
- `internal/` - shared Go implementation packages for validation, source resolution, and change workflow.
- `adapters/` - tool-specific adapter packs and docs as that surface grows.
- `docs/` - design, planning, installation, and release-process documentation.
- `tools/` - repository-owned quality and maintenance tooling.
- `nix/` - canonical dev-shell, check, and release-artifact definitions.

## Install / Try The CLI

The recommended way to use RuneContext is through the `runectx` CLI. Even in the current alpha.5 pre-MVP state, the CLI is the main executable entrypoint for local init, validation, status/change workflow, bundle resolution, diagnostics, and promotion flows, while the library foundation also covers deterministic context packs, generated indexes, and promotion assessment.

The long-term canonical install path is a reviewable repo bundle from GitHub Releases. Today, the simplest way to dogfood the current checkout is `just build`, which assembles a repo-local package with the built CLI plus the RuneContext files it needs for full local operation.

Recommended local dogfood build from the current checkout:

```sh
just build
bin/runectx help
```

`just build` writes a local package to `build/local/runecontext/` and creates a repo-local launcher at `bin/runectx` for that build.

Use the built CLI against a RuneContext project:

```sh
bin/runectx init --path /tmp/example --dry-run
bin/runectx validate /path/to/project
bin/runectx status /path/to/project
bin/runectx bundle resolve base --path /path/to/project
bin/runectx doctor --path /path/to/project
```

Current CLI scope:

- `runectx init [--mode embedded|linked] [--seed-bundle NAME] [--path PATH]`
  - scaffolds a local-first RuneContext project for embedded or linked workflows
  - supports `--dry-run`, `--json`, `--non-interactive`, and `--explain`
- `runectx validate [--ssh-allowed-signers PATH] [path]`
  - validates source settings, bundle semantics, markdown contracts, standards/spec/decision metadata, and cross-artifact traceability
  - supports embedded, git, path, and monorepo project layouts
  - supports signed-tag verification when you provide explicit trusted signer material with `--ssh-allowed-signers`
  - supports `--json`, `--non-interactive`, and `--explain`
- `runectx status [path]`
  - reports active, closed, and superseded changes plus bundle/root metadata
  - supports `--json`, `--non-interactive`, and `--explain`
- `runectx change new --title TITLE --type TYPE ...`
  - creates a minimum change by default and auto-shapes large or project work
  - supports `--json`, `--non-interactive`, `--dry-run`, and `--explain`
- `runectx change shape CHANGE_ID ...`
  - materializes `design.md` and `verification.md` by default
  - supports `--json`, `--non-interactive`, `--dry-run`, and `--explain`
- `runectx change close CHANGE_ID ...`
  - closes or supersedes a change without moving it off its stable path
  - supports `--json`, `--non-interactive`, `--dry-run`, and `--explain`
- `runectx change reallocate CHANGE_ID [--path PATH]`
  - reallocates a rare colliding change ID before merge and rewrites only local in-change references
  - supports `--json`, `--non-interactive`, `--dry-run`, and `--explain`
- `runectx bundle resolve [--path PATH] <bundle-id>...`
  - resolves bundles through the same deterministic core used by validation and context-pack work
  - supports `--json`, `--non-interactive`, and `--explain`
- `runectx doctor [--path PATH] [path]`
  - reports environment, install, and source-posture diagnostics separately from authoritative validation
  - supports `--json`, `--non-interactive`, and `--explain`
- `runectx promote CHANGE_ID [--accept | --complete] [--target TYPE:PATH] [--path PATH]`
  - is the explicit durable promotion workflow for advancing reviewable promotion state
  - supports `--json`, `--non-interactive`, `--dry-run`, and `--explain`
- `runectx standard discover [--path PATH] [--change CHANGE_ID] [--confirm-handoff] [--target TYPE:PATH]`
  - emits advisory standards candidates and reusable promotion-target data without mutating project state
  - supports `--json`, `--non-interactive`, and `--explain`
- The CLI emits stable line-oriented `key=value` output by default and supports one shared machine-facing `--json` envelope and failure taxonomy across commands.
- For invalid command results, `root` identifies the selected project root and `error_path` is emitted when a specific failing artifact path is available.

If you only want a quick in-tree hack loop from this checkout, `go run` still works:

```sh
go run ./cmd/runectx validate
go run ./cmd/runectx status
```

If you prefer to inspect or vendor release contents directly, the long-term canonical distribution remains a reviewable repo bundle from GitHub Releases; the CLI binary is a convenience entrypoint on top of that release model.

Common local commands:

```sh
nix develop
just build
just fmt
just lint
just test
just check
```

`just lint` includes the repo's source-quality gate in addition to formatting,
lint, and vet checks. See `docs/source-quality.md` for the current policy.

For release verification and maintainer workflow details, see `docs/install-verify.md` and `docs/release-process.md`.

## Uninstall

To remove the repo-local dogfood build output:

```sh
rm -rf build/local bin/runectx
```

To remove a separately installed `runectx` binary:

```sh
rm -f "$HOME/.local/bin/runectx"
```

If you copied or vendored RuneContext files into another repository from a release bundle, remove those files using your normal reviewable project workflow.

## Docs

- `docs/project_idea.md` - the original design idea and historical product rationale baseline; do not edit during normal feature work.
- `docs/implementation-plan/README.md` - the alpha-by-alpha implementation plan.
- `core/README.md` - the entrypoint for normative core contracts.
- `schemas/README.md` - schema inventory and contract notes.
- `docs/source-quality.md` - source-quality policy and protected review surfaces.
- `docs/install-verify.md` - install and verification guidance.
- `docs/release-process.md` - maintainer release workflow.

## Contributing

See `CONTRIBUTING.md`. DCO sign-off is required (`git commit -s`).

## Security

Please do not open public issues for security vulnerabilities. See `SECURITY.md`.

## License

Apache-2.0. See `LICENSE` and `NOTICE`.
