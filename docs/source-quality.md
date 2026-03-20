# Source Quality Policy

This repository uses source-quality checks to keep the Go codebase reviewable,
split along clear boundaries, and resistant to slow complexity creep.

The current enforcement scope is intentionally narrow:

- Go source and Go tests only.
- Repository-owned checker/config surfaces only.
- No Markdown or YAML lint policy is enforced here yet.

## Enforcement Surfaces

The active Go quality gates are:

- `go run ./tools/gofmtcheck`
- `go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run`
- `go vet ./...`
- `go run ./tools/checksourcequality`
- `just lint`

`just lint` is the canonical local entry point for these checks.

## Default Limits

`tools/checksourcequality` uses these defaults:

- Tier 1 source files: 250 SLOC
- Tier 1 test files: 500 SLOC
- Tier 2 source files: 400 SLOC
- Tier 2 test files: 800 SLOC
- Tier 1 functions: 40 lines, cognitive complexity 10
- Tier 2 functions: 60 lines, cognitive complexity 15

Current path mapping in this repo is:

- Tier 1: `internal/**`, `tools/**`
- Tier 2: `cmd/**`

This keeps the core implementation and checker code under the strictest review
budget.

## Policy Priorities

When a quality check fails, prefer the narrowest safe fix in this order:

1. Refactor code to reduce size, branching, or mixed responsibilities.
2. Improve docs or comments when the issue is missing source-facing context.
3. Tighten checker heuristics if the failure is a real false positive.
4. Use a checked-in reviewed exception only when the code should stay as-is for
   a justified reason.

Do not treat the baseline or config as a casual escape hatch.

## Reviewed Exceptions

Reviewed exceptions are tracked in `.source-quality-baseline.json`.

Rules for baseline entries:

- Keep each entry file-specific.
- Raise only the minimum limit needed.
- Record a concrete rationale.
- Add follow-up text when the exception should be removed later.
- Prefer deleting baseline entries after refactors instead of normalizing them as
  permanent debt.

Tier 1 inline suppressions are not the normal exception path. If a Tier 1 file
needs an exception, it should be represented in checked-in policy/config, not in
ad hoc comments.

## Protected Surfaces

Changes to the following files or directories are protected-source-quality
surfaces and should be called out explicitly in reviews:

- `.source-quality-baseline.json`
- `.source-quality-config.json`
- `.golangci.yml`
- `justfile`
- `tools/checksourcequality/**`
- `tools/gofmtcheck/**`

These files define or enforce policy. Changes here can alter what the repo will
permit going forward.

## Expectations For Fixes

When fixing quality violations:

- prefer splitting large files by responsibility instead of moving code around
  mechanically
- prefer extracting helpers that clarify boundaries, not helpers that only hide
  line count
- keep exported docs and package comments accurate rather than ornamental
- avoid commented-out code; keep rationale in prose and let history preserve old
  implementations
- do not weaken thresholds casually to make current code pass

## Verification Flow

For targeted quality work, run the smallest relevant command first, then confirm
the full gate:

1. targeted checker command
2. `just lint`
3. `go test ./...`

If a change modifies a protected surface, call that out clearly in the change
summary.
