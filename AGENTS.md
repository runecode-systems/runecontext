# AGENTS.md

This file gives coding agents the practical rules for working safely and
effectively in the `runecontext` repository.

## Scope And Priority

- Applies to the whole repository rooted at the current checkout.
- Follow direct user instructions first, then this file, then nearby file-level conventions.
- Treat `docs/project_idea.md` as a historical design baseline, not an active spec.
- Do not edit `docs/project_idea.md` during normal feature work.
- If a narrow historical correction is unavoidable, record the rationale in
  `docs/implementation-plan/README.md`.

## Rule Files Present

- Existing repo-level agent file: `AGENTS.md` (this file).
- No Cursor rules were found in `.cursor/rules/`.
- No `.cursorrules` file was found.
- No Copilot instructions file was found at `.github/copilot-instructions.md`.

## Environment And Toolchain

- Primary language: Go.
- Go version: `go 1.25` with `toolchain go1.25.7` in `go.mod`.
- Canonical local workflow: Nix + `just`.
- Canonical command runner: `just`.
- CI runs the same logical checks as local commands in `.github/workflows/ci.yml`.
- Release builds are driven by Nix in `flake.nix` and `.github/workflows/release.yml`.

## Canonical Commands

- Show available tasks: `just --list`
- Build local CLI package: `just build`
- Format repo: `just fmt`
- Lint repo: `just lint`
- Run all tests: `just test`
- Run main local CI gate: `just ci`
- Run full Nix CI gate: `just nix-ci`
- Run flake checks only: `just check`
- Build release artifacts only: `just release` or `just release-check`

## What Those Commands Do

- `just build` runs `bash tools/build-local.sh`.
- `just fmt` runs `go run ./tools/gofmtcheck --write` and formats Nix files.
- `just lint` runs:
  - `go run ./tools/gofmtcheck`
  - `go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run`
  - `go vet ./...`
  - `go run ./tools/checksourcequality`
  - `just layout-check`
- `just test` runs `go test ./...`.
- `just ci` runs `just lint` and `just test`.
- `just nix-ci` runs lint, tests, release build checks, and flake checks.

## Fast Targeted Test Commands

- Run one package: `go test ./internal/cli`
- Run one package verbosely: `go test -v ./internal/contracts`
- Run one specific test: `go test -v ./internal/cli -run TestRunPromoteJSONGolden`
- Run tests by regex: `go test ./internal/contracts -run 'TestPromote|TestClose'`
- Run one test with count disabled: `go test -count=1 ./internal/cli -run TestRunInitDryRun`
- Run a single tool package test: `go test ./tools/checksourcequality -run TestName`

## When To Run Which Checks

- Small focused Go change: run the affected package tests first.
- CLI behavior change: run `go test ./internal/cli` and then `just lint`.
- Contract/resolution/change-workflow change: run `go test ./internal/contracts` and then `just lint`.
- Tooling or source-quality checker change: run the relevant tool tests, then `just lint`, then `go test ./...`.
- Before finishing a non-trivial change: prefer at least `just lint` and `just test`.

## Repository Layout

- `cmd/runectx/`: thin Go binary entrypoint.
- `internal/cli/`: CLI parsing, output contracts, and command behavior.
- `internal/contracts/`: core validation, resolution, change workflow, bundles, packs, promotion.
- `tools/`: repo-owned developer tooling such as `gofmtcheck` and `checksourcequality`.
- `fixtures/`: shared test fixtures and golden inputs.
- `docs/`: implementation plan, install docs, release docs, and policy docs.
- `schemas/` and `core/`: normative contract surfaces.

## Code Organization Conventions

- Keep `cmd/` extremely thin; business logic belongs in `internal/`.
- Keep CLI concerns in `internal/cli`, not mixed into contract logic.
- Keep canonical semantics in `internal/contracts`.
- Prefer small files grouped by responsibility instead of large mixed-purpose files.
- Follow existing naming clusters such as `cli_*.go`, `bundle_resolution_*.go`, and `change_*`.

## Formatting And Imports

- Use standard Go formatting; do not hand-format code against `gofmt`.
- Run `just fmt` or `go run ./tools/gofmtcheck --write` after Go edits.
- Keep imports grouped in normal Go order: standard library, blank line, external/internal imports.
- Keep imports sorted as `gofmt` expects.
- Avoid unused imports; CI will fail on them.

## Naming Conventions

- Use idiomatic Go names: exported identifiers in PascalCase, unexported in camelCase.
- Prefer clear domain names over abbreviations.
- Use typed string enums for closed sets, following patterns like
  `type LifecycleStatus string` with a `const` block.
- Name command-specific CLI files and helpers consistently with the command.
- Keep helper names descriptive: `buildValidateOutput`, `parsePromoteArgs`, `resolveSeedBundlePath`.

## Types And Data Modeling

- Prefer concrete structs for domain state and operation results.
- Use typed string enums for statuses, modes, severities, and failure classes.
- Keep zero values meaningful where possible.
- Avoid introducing interface abstractions unless there is a real boundary.
- Reuse existing result and diagnostic types before inventing new payload shapes.

## Error Handling

- Return errors; do not panic for expected failures.
- Add context with `fmt.Errorf("context: %w", err)` where it improves diagnosis.
- Use structured domain errors when the caller needs to branch on failure kind.
- Follow existing patterns like `ValidationError` and `SignedTagVerificationError`.
- CLI code should convert errors into stable user/machine-facing output, not leak raw panics.
- Fail closed on ambiguous, unsafe, or partially invalid states.

## Control Flow And Helper Extraction

- Prefer early returns over deep nesting.
- Extract helpers to clarify responsibilities, not just to silence line-count checks.
- Keep parsing, validation, mutation, and rendering as separate steps when possible.
- Preserve narrow command boundaries: `status` for workflow summary, `validate` for authoritative checks, `doctor` for diagnostics.

## Comments And Documentation

- Keep comments sparse and useful.
- Add package comments and exported-doc comments when required by lints.
- Do not add ornamental comments that restate code.
- Keep README and docs aligned with behavior when changing user-visible commands.
- Treat planning docs, schemas, and fixtures as part of the product contract.

## Testing Conventions

- Use table-driven tests when it improves clarity; simple direct tests are also common here.
- Use `t.Helper()` in shared test helpers.
- Use `t.Parallel()` where the file already follows that pattern and test isolation is safe.
- Prefer repo fixtures over ad hoc inline test trees when an existing fixture family fits.
- Reuse helpers like `repoFixtureRoot`, `fixtureRoot`, and command-specific helpers.
- Golden tests are common for CLI JSON and deterministic outputs; preserve their shape.
- For CLI tests, assert exit code plus stable output fields.

## Fixtures And Golden Files

- Check `fixtures/` before creating new test data.
- Extend an existing fixture family if the scenario matches.
- Keep deterministic outputs stable and reviewable.
- If output intentionally changes, update the corresponding golden fixtures in the same change.

## Source Quality Policy

- `just lint` is the canonical lint gate.
- The repo enforces extra source-quality limits via `tools/checksourcequality`.
- Tier 1 paths are `internal/**` and `tools/**`; they have the strictest file/function limits.
- Tier 2 paths are `cmd/**`.
- Do not treat baseline/config changes as casual escapes.
- Prefer refactoring over weakening thresholds.
- Inline suppressions are discouraged; `nolint` requires specificity and explanation.

## Lint Configuration To Respect

- `.golangci.yml` enables `funlen`, `gocognit`, `nolintlint`, and `revive`.
- `revive` enforces package comments, exported comments, and comment spacing.
- `nolintlint` requires specific linter names and an explanation.
- Cognitive complexity and function length are intentionally enforced; split logic cleanly.

## Protected / High-Trust Surfaces

- `flake.nix` and `flake.lock`: high-trust local tooling and release inputs.
- `.golangci.yml`: linter policy.
- `justfile`: canonical developer command surface.
- `.source-quality-config.json` and `.source-quality-baseline.json`: quality policy.
- `tools/checksourcequality/**` and `tools/gofmtcheck/**`: policy-enforcement tools.
- `schemas/` and `core/`: normative contract surfaces.
- `docs/implementation-plan/`: current milestone contract and planning guidance.

## Agent Reminders

- Prefer minimal, targeted changes that fit existing patterns.
- Do not silently broaden scope by editing protected policy files unless required.
- If you change command behavior, update tests and relevant docs in the same patch.
- If you change a protected surface, call it out explicitly in your summary.
- If you commit, use DCO sign-off: `git commit -s`.
