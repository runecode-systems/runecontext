# Verification

## Planned Checks
- `./bin/runectx validate --path .`
- `go test ./internal/cli -run Upgrade`
- `just test`

## Close Gate
Close the umbrella only after path planning, staged execution, and rollback coverage all pass, and both sub-changes land with reciprocal relationships and clean project validation.
