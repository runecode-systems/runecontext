# Post-MVP Grouping

These groups are intentionally deferred until after `v0.1.0`.

Signed-tag verification is not listed here because it is part of the MVP.

Post-MVP work should follow the same quality bar as the MVP plan:

- new semantics land with unit tests and golden fixtures in the same milestone
- new user-facing flows land with integration or end-to-end coverage
- RuneCode-facing contract changes land with companion parity fixtures

## `v0.2.0-alpha.1` - Anchored Assurance

Primary outcome: add the future `Anchored` assurance tier once RuneCode's audit
anchoring model is mature enough to support it cleanly.

### Epics

- [ ] Issue: define the `anchored` assurance-tier contract and migration rules.
- [ ] Issue: define anchor record schemas and storage conventions.
- [ ] Issue: define how anchored records relate to Verified baseline and receipt
  artifacts.
- [ ] Issue: define failure, degraded-mode, and verification semantics for
  missing or broken anchors.

### RuneCode Companion-Track Checkpoints

- RuneCode can bind RuneContext verified evidence into its own anchoring flow.
- RuneCode can verify that anchored evidence strengthens provenance without
  mutating RuneContext's portable source model.

## `v0.2.0-alpha.2` - Rich Lineage And Historical Views

Primary outcome: build the richer generated lineage and index views explicitly
deferred by the idea document.

### Epics

- [ ] Issue: design the generated lineage graph model connecting changes,
  decisions, specs, and standards across history.
- [ ] Issue: build richer lineage views on top of the alpha.3 artifact-level and
  heading-fragment traceability model without introducing new mandatory authored
  fields.
- [ ] Issue: implement a generated lineage/index view over existing traceability
  fields.
- [ ] Issue: add filters for active, closed, superseded, and promoted history.
- [ ] Issue: define merge-friendly serialization and regeneration rules for the
  lineage view.

### RuneCode Companion-Track Checkpoints

- RuneCode can surface lineage views directly in audit/history UIs without
  needing new authoring semantics.

## `v0.2.0-alpha.3` - Distribution Convenience Channels

Primary outcome: add non-canonical convenience distribution channels without
changing the repo-first install model.

Note: Windows CI portability testing is part of v0.1.0-alpha.8, but Windows
binary packaging and distribution channels (Scoop, winget) remain deferred to
this phase pending adoption demand.

Even if convenience channels are added later, they remain subordinate to the
canonical GitHub release artifacts and must not reintroduce implicit network
fetches into normal adapter sync flows.

### Epics

- [ ] Issue: define npm packaging for schema bundle, adapters, or helper CLI.
- [ ] Issue: define Homebrew distribution for `runectx` binaries across supported
  architectures (amd64 and arm64).
- [ ] Issue: define Scoop and winget packaging and distribution strategy if
  Windows adoption demand warrants it (Windows CI portability parity is v0.1.x;
  distribution channels are post-MVP).
- [ ] Issue: document how convenience channels stay subordinate to GitHub
  release artifacts.

### RuneCode Companion-Track Checkpoints

- RuneCode can continue to treat RuneContext as a pinned dependency regardless
  of convenience channel availability.

## `v0.2.0-alpha.4` - Optional Advanced Guardrails

Primary outcome: implement useful advanced safeguards that the idea document
describes as optional or supplementary rather than mandatory for v1.

### Epics

- [ ] Issue: implement stricter pinned-glob mode requiring explicit acceptance
  when bundle match sets change.
- [ ] Issue: implement optional prompt-hygiene or content-safety heuristics for
  RuneContext text used in model-facing flows.
- [ ] Issue: implement richer provenance-compaction controls when a context pack
  needs separate receipt detail beyond the MVP compact form.
- [ ] Issue: consider optional deeper section-fragment semantics beyond
  heading-slug refs only if real-world authoring pressure appears after MVP.

### RuneCode Companion-Track Checkpoints

- RuneCode can test advanced guardrails without depending on them for baseline
  correctness.
