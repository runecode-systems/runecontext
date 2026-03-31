# Verification

## Notes
Verification should cover four categories: older-but-compatible no-migration project upgrades, migration-required project upgrades, project-newer-than-cli failures, and machine-facing contract updates in validate/doctor/upgrade preview output. The design should also preserve the distinction between read-only preview and mutating apply.

## Planned Checks
- `just test`

## Close Gate
Use the repository's standard verification flow before closing this change.
