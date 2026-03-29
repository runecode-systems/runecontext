# Tasks

- Add a versioned capability descriptor schema with closed objects, stable token enums, and explicit fail-closed behavior for unknown schema versions and values.
- Implement a canonical metadata builder that derives command/capability data from the CLI registry and derives compatibility, runtime layout, assurance, and resolution data from existing contracts and release metadata.
- Expose the descriptor through a dedicated `runectx metadata` command while preserving `runectx version --json` compatibility.
- Embed the exact same descriptor object in the release manifest and add parity tests so CLI and release artifacts cannot drift.
- Add fixtures and tests for schema validity, CLI output shape, fail-closed rejection, supported-version vs upgrade-edge reporting, and release-manifest parity.
