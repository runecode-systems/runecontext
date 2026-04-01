# Tasks

- Replace adapters/source/shared/flows.json with per-flow structured definitions under adapters/source/shared/flows/.
- Extend tools/syncadapters model and render pipeline to generate richer flow markdown plus workflow.json for each adapter pack.
- Update adapter render-host-native to read generated workflow.json from the located adapter pack and render the shared structured contract for shell-injection and embedded hosts.
- Tighten adapter root discovery so runtime rendering consumes generated adapter packs rather than adapters/source.
- Expand adapter render/sync/generator tests and dogfood runecontext-change-new so generated host-native commands create RuneContext changes and stop after the immediate CLI outcome.
