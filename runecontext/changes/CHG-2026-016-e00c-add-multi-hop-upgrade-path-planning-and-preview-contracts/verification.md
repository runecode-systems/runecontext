# Verification

## Planned Checks
- `./bin/runectx validate --path .`
- `go test ./internal/cli -run Upgrade`
- `just test`

## Close Gate
Close this change only after preview output covers direct-edge and multi-hop paths, missing paths fail closed, and the planner exposes deterministic ordered hops suitable for staged apply-time execution.
