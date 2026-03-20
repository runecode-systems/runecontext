---
name: source-quality-violation-triage-and-fix
description: Triage and fix RuneContext source-quality violations with the narrowest safe change, then rerun the gate.
argument-hint: "[optional paths, rule names, or failing command output]"
disable-model-invocation: true
---

Use this workflow when `tools/checksourcequality`, `golangci-lint`, `go vet`, or
related source-quality review feedback must be fixed.

## Read first

- `docs/source-quality.md`
- `.source-quality-baseline.json`
- `.source-quality-config.json`
- `.golangci.yml`
- `justfile`
- `CONTRIBUTING.md`

If the work touches a failing file directly, read that file before choosing a
fix. If the work touches checker behavior or policy, read the relevant files in
`tools/checksourcequality/` or `tools/gofmtcheck/` before editing.

## Procedure

1. Identify the failing surface.
   - `go run ./tools/checksourcequality`
   - `go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run`
   - `go vet ./...`
   - review comments about source-quality policy, checker logic, or protected
     surfaces
2. Classify each finding before editing.
   - missing package or exported docs
   - comment-quality false positive or true positive
   - function or file budget violation
   - suppression or reviewed-exception issue
   - protected-surface policy drift
3. Prefer the narrowest safe fix.
   - refactor code before raising limits
   - improve docs before widening exceptions
   - tighten checker heuristics instead of weakening policy when a false
     positive is real
   - use checked-in baseline or config changes only when the exception is
     justified and reviewable
4. Keep exceptions disciplined.
   - do not rely on reviewer memory or PR comments as the active exception path
   - baseline or config changes must stay narrow, justified, and easy to audit
   - Tier 1 suppressions must not become casual inline escapes
5. Re-run verification after fixes.
   - targeted command(s) first
   - `just lint`
   - `go test ./...`
6. Report.
   - what failed
   - what changed
   - whether any protected-surface files changed
   - final command results

## Guardrails

- Do not weaken thresholds, exclusions, or protected-surface rules casually.
- Prefer fixing code, docs, or heuristics before adding reviewed exceptions.
- If a change touches `.source-quality-baseline.json`,
  `.source-quality-config.json`, `.golangci.yml`, `justfile`,
  `tools/checksourcequality/**`, or `tools/gofmtcheck/**`, call out that a
  protected surface changed.
- Do not commit or push unless explicitly requested.
