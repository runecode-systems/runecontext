# Verification

## Notes
Verification should prove both sides of the alpha.13 boundary: fresh alpha.13 projects must write only the canonical `runecontext/assurance` layout, and existing alpha.12 verified projects must migrate safely in a staged tree without touching the live repository until the final staged state validates. Coverage should also confirm that baseline imported-evidence backfill paths are rewritten during migration and that rollback leaves the live tree unchanged on failure.

## Planned Checks
- `go test ./internal/cli -run 'TestRunUpgradeApply|TestRunAssurance'`
- `go test ./internal/contracts -run 'TestAssurance|TestValidate'`
- `just test`

## Close Gate
Close this change only after fresh alpha.13 assurance writes use `runecontext/assurance`, staged `alpha.12 -> alpha.13` migration successfully moves baseline, receipts, and backfill artifacts into the canonical tree, imported-evidence backfill paths are rewritten, and staged rollback preserves the live project on failure.
