# Verification

## Planned Checks
- `./bin/runectx validate --path .`
- `go test ./internal/cli`
- `go test ./internal/contracts`

## Close Gate
Close this change after human-output coverage proves the new layout is deterministic, ASCII-safe without color, color-aware when terminals allow it, and still leaves `status --json` unchanged.
