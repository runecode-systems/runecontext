# runectx

This directory contains the Go CLI entrypoint for the current alpha.5 command surface.

## Current Scope

- `runectx init [--json] [--non-interactive] [--dry-run] [--explain] [--mode embedded|linked] [--seed-bundle NAME] [--path PATH]`
  - scaffolds a local-first RuneContext project for embedded or linked workflows
  - supports plan reporting during `--dry-run` and keeps network-enabled init/update work out of alpha.5
- `runectx status [--json] [--non-interactive] [--explain] [path]`
  - reports active, closed, and superseded changes plus bundle and root metadata
- `runectx change [--json] [--non-interactive] [--dry-run] [--explain] new --title TITLE --type TYPE ...`
  - creates a minimum change by default and auto-shapes large or project work
  - writes `status.yaml`, `proposal.md`, and `standards.md`
- `runectx change [--json] [--non-interactive] [--dry-run] [--explain] shape CHANGE_ID ...`
  - materializes `design.md` and `verification.md` by default
  - creates `tasks.md` and `references.md` only when non-empty content is provided
  - fails closed for terminal changes rather than mutating historical artifacts
- `runectx change [--json] [--non-interactive] [--dry-run] [--explain] close CHANGE_ID ...`
  - closes or supersedes a change without moving it off its stable path
  - fails closed if a missing reciprocal `supersedes` link would require mutating
    an already-terminal successor change
- `runectx change [--json] [--non-interactive] [--dry-run] [--explain] reallocate CHANGE_ID [--path PATH]`
  - reallocates a rare colliding change ID before merge
  - rejects terminal changes, stages the rewrite outside the live `changes/`
    tree, and fails closed when external artifacts reference the old ID
  - rewrites local change-path references inside the change and returns warnings
    instead of ambiguous failures when only backup cleanup needs manual follow-up
- `runectx validate [--json] [--non-interactive] [--explain] [--ssh-allowed-signers PATH] [path]`
  - validates the root `runecontext.yaml`
  - validates change status files, markdown contracts, standards/spec/decision
    frontmatter, and project-level traceability
  - fails closed on schema, YAML-profile, and cross-artifact reference errors
  - supports explicit SSH allowed-signers input for signed-tag verification
  - emits stable line-oriented `key=value` output for success, invalid-state
    failures, and usage errors, plus a shared `--json` envelope for
    machine-facing automation
- `runectx bundle [--json] [--non-interactive] [--explain] resolve [--path PATH] <bundle-id>...`
  - resolves bundles through the same deterministic core used by validation and context-pack work
  - reports requested bundles, resolved linearization order, and diagnostics
- `runectx doctor [--json] [--non-interactive] [--explain] [--path PATH] [path]`
  - reports environment, install, and source-posture diagnostics separately from authoritative validation
  - includes lightweight environment warnings such as missing local `git`
- `runectx promote [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID [--accept | --complete] [--target TYPE:PATH] [--path PATH]`
  - is the explicit durable promotion workflow for advancing reviewable promotion state
  - preserves reviewable targets and supports machine-readable promotion transitions
- `runectx standard [--json] [--non-interactive] [--explain] discover [--path PATH] [--change CHANGE_ID] [--confirm-handoff] [--target TYPE:PATH]`
  - emits advisory standards candidates and reusable promotion-target data without mutating project state
  - supports explicit interactive handoff planning into `runectx promote` while keeping `--non-interactive` discovery advisory-only

## Output Contract

The CLI supports two machine-facing output modes.

- default line-oriented contract (`key=value` per line)

- success on stdout
  - `result=ok`
  - `command=<command>`
  - `root=<absolute-path>`
  - command-specific metadata such as change IDs, lifecycle counts, standards
    refresh details, warnings, and changed file paths/actions
- validation failure on stderr
  - `result=invalid`
  - `command=<command>`
  - `root=<absolute-path>`
  - `error_path=<path>` when available
  - `error_message=<message>`
- usage/command errors on stderr
  - `result=usage_error`
  - `command=<name>` when available
  - `error_message=<message>`
  - `usage=<command-usage>`

Values escape `\\`, `\n`, `\r`, `\t`, `\0`, and `\=` so each field stays on
one line. Consumers should split each output record on the first `=`
character and then reverse those escape sequences.

- JSON envelope mode (`--json`)
  - emits one JSON object with a shared envelope:
    - `schema_version`
    - `result`
    - `command`
    - `exit_code`
    - `failure_class`
    - `data` (key/value payload mirroring the line-oriented fields)
  - failure classes currently used across commands are `none`, `invalid`, and `usage`

Cross-command machine flags currently supported:

- `--json`
- `--non-interactive`
- `--explain`
- `--dry-run` (write commands only)

`--dry-run` runs write operations against a temporary project clone, validates
the would-be project state, and returns planned mutations without persisting
changes to the caller's repository.

`--dry-run` fails closed when cloning encounters absolute symlinks or relative
symlinks that resolve outside the selected project root.

`--explain` is accepted across the current alpha.5 command surface. Richer
explanation payloads are still incremental for some operations, and commands may
emit an `explain_warning` field when detailed explain data is not yet available.

## Current Constraint

- In the current repository-first alpha.5 implementation, schemas are discovered
  relative to the source checkout. Release-oriented schema embedding or explicit
  schema-root overrides remain future work.
- Whole-project validation follows the content root declared in `runecontext.yaml`
  (`source.path` for embedded/path sources and `source.subdir` for git sources)
  rather than assuming a fixed `runecontext/` directory.
- Explicit path arguments are respected as explicit roots even when the caller
  passes `.`; the CLI does not silently climb to an ancestor RuneContext root in
  that case.

- `runectx init` in alpha.5 is intentionally local-first. Network-enabled
  install, update, and release-hardening workflows remain later-milestone work.

The CLI remains pre-MVP and intentionally scoped. Later milestones still add
assurance enable/backfill flows, shell completion, adapter sync surfaces,
release/install/update hardening, and broader adapter UX.
