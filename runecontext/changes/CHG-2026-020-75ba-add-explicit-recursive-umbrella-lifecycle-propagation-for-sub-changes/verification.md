# Verification

## Notes
Run go test ./internal/cli -run Change and just test. Cover non-recursive default behavior, recursive umbrella update and close success, rejection when a recursive target is not an eligible feature sub-change, and rollback when one sub-change blocks the requested cascade.

## Planned Checks
- `just test`

## Close Gate
Use the repository's standard verification flow before closing this change.
