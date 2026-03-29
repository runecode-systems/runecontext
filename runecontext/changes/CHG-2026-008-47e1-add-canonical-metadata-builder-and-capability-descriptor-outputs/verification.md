# Verification

## Planned Checks
- `go test ./internal/cli ./internal/contracts`
- `./bin/runectx validate --path .`
- `just test`

## Close Gate
Close this change only after schema fixtures, fail-closed tests, CLI/release parity tests, and compatibility-reporting tests all pass.
