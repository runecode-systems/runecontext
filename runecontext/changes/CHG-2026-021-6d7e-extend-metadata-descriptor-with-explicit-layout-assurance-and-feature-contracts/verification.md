# Verification

## Planned Checks
- `go test ./internal/cli ./internal/contracts`
- `./bin/runectx validate --path .`
- `just sync-metadata`
- `just test`
- `just ci`

## Close Gate
Close this change only after the descriptor `v3` schema, CLI output, release-manifest embedding, generated docs/reference artifact, and metadata golden fixtures all agree; fail-closed unknown-version, legacy-field, and unknown-token tests pass; direct-support and upgrade-entry semantics stay machine-distinct; and the documented field meanings match the implemented contract.
