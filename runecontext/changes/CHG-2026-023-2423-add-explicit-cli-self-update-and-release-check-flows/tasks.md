# Tasks

- Add `runectx upgrade cli` as an explicit preview/check command for available newer CLI releases.
- Add `runectx upgrade cli apply` as the explicit mutating self-update command for downloading and installing a selected newer CLI release.
- Define machine-facing output contracts for CLI update availability, selected release, install action, and failure guidance.
- Keep network access explicit and bounded to CLI-update flows or clearly opted-in checks.
- Document how CLI self-update relates to project upgrade and how users should respond when their project requires a newer CLI version.
