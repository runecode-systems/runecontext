# Verification

## Notes
Verification should prove that the planner surfaces only real migration edges in the requested version interval, that zero-hop version-bump-only upgrades still work when no migration edge is present, and that fresh projects initialized at the installed version do not replay historical migrations. Apply-time coverage should confirm that staged execution runs only the real edges discovered by preview.

## Planned Checks
- `go test ./internal/cli -run 'TestRunUpgrade|TestBuildUpgrade|TestRunInit'`
- `just test`

## Close Gate
Close this change only after preview and apply agree on interval-based migration-edge discovery, no synthetic hops appear in the machine or human upgrade contracts, and fresh-project initialization remains free of historical migration replay.
