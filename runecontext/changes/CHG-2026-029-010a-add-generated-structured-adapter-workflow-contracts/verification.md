# Verification

## Notes
Verify shared flow contract loading and generation, assert generated flow docs include outcome/guardrails/workflow/stop-condition/next-command sections for every operation, verify adapter render-host-native output includes the structured workflow contract and no-question rule for OpenCode/Claude/Codex surfaces, run adapter sync dogfood checks in this repo, and pass just ci.

## Planned Checks
- `just test`

## Close Gate
Use the repository's standard verification flow before closing this change.
