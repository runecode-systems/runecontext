# Verification

## Planned Checks
- `./bin/runectx validate --path .`
- `go test ./internal/cli`
- `go test ./internal/contracts`

## Close Gate
Close this change after status tests cover zero-history projects, fewer-than-limit histories, large historical counts with hidden-item hints, and the human-only flag behavior needed to preserve the current machine contract.
