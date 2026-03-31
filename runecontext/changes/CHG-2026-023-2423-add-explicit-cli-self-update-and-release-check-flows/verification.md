# Verification

## Notes
Verification should cover command parsing, machine-readable output for CLI update preview/apply, explicit network boundary behavior, and user guidance when a newer CLI release exists. Tests should also confirm that bare project upgrade commands retain their meaning and that no hidden network access is introduced into validate, status, or other core local-first flows.

## Planned Checks
- `just test`

## Close Gate
Use the repository's standard verification flow before closing this change.
