# Verification

## Planned Checks
- `go test ./internal/contracts`
- `go test ./internal/cli`
- `./bin/runectx change update CHG-2026-014-cf14-add-structured-lifecycle-status-update-command-for-non-terminal-changes --status planned --path . --dry-run`
- `./bin/runectx validate --path .`

## Close Gate
Verify that non-terminal lifecycle updates work without mutating promotion assessment, and that invalid backward or terminal transitions fail with stable diagnostics.
