# Context Pack Fixtures

These fixtures cover the alpha.4 Branch Cut 1 context-pack engine and
determinism rules, plus the Branch Cut 2 pack-report output modes.

- `golden/child-reinclude.yaml`: single-bundle pack generated from the
  `bundle-resolution/valid-project` fixture, including selected inventories,
  persisted selector provenance, and a deterministic top-level pack hash.
- `golden/left-right.yaml`: ordered multi-bundle request pack showing preserved
  `requested_bundle_ids`, merged parent linearization, selected inventories, and
  excluded outputs.
- `reports/child-reinclude.json`: machine-readable report output for a
  deterministic context-pack build with `report_schema_version: 1` and
  `--explain`-style provenance details.
- `reports/child-reinclude.txt`: human-readable report output for the same pack,
  including summary metrics, warnings, and include/exclude explanation detail.

The context-pack goldens focus on the reusable shape needed by later alpha.4
work:

- ordered `requested_bundle_ids` distinct from resolved `context_bundle_ids`
- explicit `runecontext-canonical-json-v1` canonicalization token for the
  current emitted pack serializer profile
- compact persisted selector provenance with `bundle`, `aspect`, `rule`,
  `pattern`, and `kind`
- stable selected and excluded aspect inventories
- deterministic per-file content hashes
- UTF-8 text hashing parity across LF and CRLF checkouts
- deterministic top-level `pack_hash` with `generated_at` excluded from the
  canonical hash input
- arbitrary fixture timestamps that stay outside canonical pack hashing
- stable human-readable and machine-readable report outputs over the same pack,
  including advisory metrics, explain-style provenance detail, and an explicit
  machine-report schema version
