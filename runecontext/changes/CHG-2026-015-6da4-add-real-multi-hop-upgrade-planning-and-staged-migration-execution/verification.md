# Verification

## Planned Checks
- `./bin/runectx validate --path .`
- `go test ./internal/cli -run Upgrade`
- `just test`

## Close Gate
Close the umbrella only after path planning, staged execution, compatibility/version-bump-only semantics, and CLI self-update flows all pass reviewable verification, and all associated feature changes land with reciprocal relationships and clean project validation.
