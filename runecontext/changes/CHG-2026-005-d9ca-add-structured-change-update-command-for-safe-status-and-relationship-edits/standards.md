## Applicable Standards
- `standards/global/cli-mutation-safety.md`: Status and relationship editing must use validated, transactional writes with fail-closed behavior.
- `standards/global/change-relationship-consistency.md`: Relationship updates must preserve reciprocal and directional semantics across change graphs.
- `standards/global/structured-cli-contracts.md`: The update command should expose a stable machine-oriented mutation surface for direct CLI and adapter use.

## Resolution Notes
These standards define the safe-editing boundary for the new structured change update command.
