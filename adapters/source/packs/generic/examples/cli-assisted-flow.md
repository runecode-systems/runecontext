# Generic Adapter: CLI-Assisted Flow

Use completion and suggestion providers to reduce manual lookup.

1. Install shell completion: `runectx completion bash` (or `zsh` / `fish`).
2. Suggest change IDs: `runectx completion suggest --path <project-root> change-ids`.
3. Suggest promotion targets: `runectx completion suggest --path <project-root> promotion-targets`.
4. Run explicit mutation commands with reviewed arguments.
