# Verification

## Planned Checks
- `./bin/runectx validate --path .`
- `go test ./internal/cli -run Upgrade`
- `just test`

## Close Gate
Close this change only after staged multi-hop apply succeeds for supported paths, rolls back cleanly on any intermediate failure, and preserves the rule that only apply mutates project files.
