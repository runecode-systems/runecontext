# RuneContext (Design Idea)

RuneContext is a proposed portable spec, standards, and project-knowledge system that RuneCode can integrate with deeply without making the format RuneCode-specific.

This document captures the design decisions, recommendations, and constraints from the RuneContext discussion. It is intended to be detailed enough that another developer can implement RuneContext and RuneCode's RuneContext integration from it.

## Executive Summary

- RuneContext should be a portable, markdown-first, git-native system for project details, standards, context bundles, changes, and stable specs.
- RuneContext should not be "RuneCode's private internal doc format". It should stand on its own and be usable with RuneCode, other agentic tools, or manual workflows.
- RuneContext should keep the best parts of Agent OS:
  - markdown/git-native artifacts
  - low ceremony
  - shaping before implementation
  - standards referenced by file path instead of copied into specs
  - reusable inheritance for standards collections
- RuneContext should borrow selected OpenSpec ideas:
  - change-oriented workflow
  - explicit change status tracking
  - lightweight promotion of durable knowledge into stable project docs
- RuneContext should use `context bundles` as the primary reusable context-selection concept. The folder name should be `bundles/`.
- A context bundle should be able to choose which aspects of overall project context to include, inherit, and exclude in one place.
- The default change shape should be minimal and reviewable:
  - `status.yaml`
  - `proposal.md`
  - `standards.md`
- A fuller change shape should be materialized only when the work needs it:
  - `design.md`
  - `tasks.md`
  - `references.md`
  - `verification.md`
- `proposal.md` should be the canonical reviewable intent artifact for "what changes and why".
- `standards.md` should always be present, tied to the project's standards knowledge base, and automatically maintained by tooling.
- RuneContext should preserve historical change/spec information without burying it. Closing a change should be a status transition, not a file move to a hard-to-find archive tree.
- RuneContext should support both:
  - an embedded mode where RuneContext lives inside a project repo
  - a linked mode where a project points at a dedicated RuneContext repo
- Linked-mode sources should prefer immutable refs such as commit SHAs or signed tags. Mutable refs must require explicit opt-in and warnings.
- Tool-specific command packs should be the primary UX. A small CLI should exist for power users, automation, and parity testing, but it should not be the canonical definition of RuneContext.
- RuneCode should be able to resolve RuneContext into deterministic, auditable `context packs` for agent runs, including a top-level pack hash over the canonicalized resolved output, but RuneContext itself must remain policy-neutral and portable.
- RuneContext should support optional assurance tiers:
  - `Plain` for lightweight standalone use
  - `Verified` for verifiable tracing with generated assurance artifacts
- When RuneCode is used for normal audited/project-operating workflows, `Verified` assurance must be enabled.
- Additional assurance artifacts should only be generated when assurance is enabled so Plain mode stays low-friction.

## Why RuneContext Exists

RuneContext exists because the desired system needs all of the following at once:

- platform-agnostic plain-text source artifacts
- compatibility with multiple agentic coding tools
- reusable standards with inheritance
- standards referenced by path instead of repeatedly copied
- low-ceremony workflow that still captures enough structure to guide good implementation
- first-class compatibility with local RuneCode use, remote/server RuneCode use, and non-RuneCode use
- a format that is safe for RuneCode to consume without turning markdown into an authority for runtime permissions or trust decisions

Vanilla Agent OS is philosophically close, but RuneContext should be fully custom so it can be designed around RuneCode's needs without a compatibility layer against a forked external format.

## Best Ideas To Keep From Agent OS

RuneContext should retain these Agent OS qualities as hard design goals:

- Markdown-first, git-native artifacts
  - Project knowledge should live in normal markdown and yaml/json files that can be versioned, reviewed, branched, and edited with any tool.
- Standards referenced by file path
  - Specs and change docs should reference standards by path rather than embedding standard content.
  - Updating a standard should not require editing every spec that depends on it.
- Lightweight shaping before implementation
  - There should be a deliberate shaping/design step before execution, but it must not feel like heavyweight process theater.
- Reusable standards collections
  - Projects need a way to define reusable collections of standards that can be inherited and combined.
- Tool portability
  - The underlying artifacts should not depend on one IDE, one runtime, or one commercial agent platform.

## OpenSpec Ideas To Mix In

RuneContext should borrow these ideas from OpenSpec:

- Change-oriented workflow
  - Work should begin as a named change with explicit intent.
- First-class lifecycle artifacts
  - Each change should always have core intent artifacts and should materialize design, tasks, references, and verification artifacts when the work needs them.
- Explicit status tracking
  - A change should have an explicit lifecycle state rather than relying on folder location alone.
- Stable pathing and traceability
  - A change should have a durable identifier that can be referenced from specs, decisions, commits, PRs, and audit records.

RuneContext should not copy OpenSpec's archive approach exactly. RuneContext should preserve history at stable paths instead of physically moving old changes into an archive tree that becomes harder to browse and audit.

## Research Findings That Should Shape RuneContext

The design should reflect what users consistently value in adjacent tools and what they repeatedly dislike.

### What Users Consistently Value Most

- Persistent repo-native memory
  - Users want specs, standards, and project memory to live in git rather than only in chat history.
- Low ceremony
  - Users strongly prefer systems that feel lightweight and practical over systems that feel like process theater.
- Better alignment before coding
  - Users value being pushed to clarify intent, scope, assumptions, and acceptance criteria before implementation begins.
- Brownfield friendliness
  - Existing repositories and existing conventions matter more than greenfield purity.
- Reusable standards and conventions
  - Users value shared standards libraries, especially when they can be reused and inherited across projects.
- Tool and agent portability
  - Users do not want project knowledge locked to one IDE or one coding agent.
- Reviewable intent artifacts
  - Users want artifacts that explain what is changing and why, not just the resulting code diff.
- Verification support
  - Users value explicit checks, acceptance criteria, and a path to human verification.
- Context hygiene
  - Users want systems that reduce context rot and survive long-running work across multiple sessions.

### What Users Consistently Like Least

- Too much ceremony, too many files, or workflow "theater"
- Confusing onboarding that requires videos, community help, or tribal knowledge to use basic flows
- Brittle agent/tool integrations
- Weak brownfield and monorepo support
- Drift between specs, standards, and real code behavior
- Systems that interrupt for trivial decisions but silently make important ones
- Excessive token, latency, or workflow overhead
- Install, upgrade, or platform friction
- Security and privacy concerns
- Anything that makes history, recovery, or auditing harder

### Design Implications For RuneContext

- Make the happy path very small.
- Use progressive disclosure instead of forcing every change into the full artifact set.
- Treat brownfield and existing repository structure as first-class inputs.
- Keep one canonical source of truth; generated views are derived, not authoritative.
- Preserve intent, traceability, and verification, but only ask questions when the answers materially affect behavior.
- Keep adapters thin and keep the core model stable.
- Prefer explicit reviewable suggestions over hidden automatic mutation.
- Treat linked-source integrity, canonical hashing, and policy neutrality as first-class design requirements, not implementation details.
- Keep assurance progressive so standalone users can stay lightweight while higher-assurance users opt into stronger provenance.

## Goals

- Keep all core knowledge artifacts portable and tool-agnostic.
- Keep important state project-local or repo-local. Avoid hidden global state that is required for correctness.
- Make the system manually editable.
- Make the system easy for RuneCode to read, update, and propose changes to.
- Support both local developer workflows and remote/server workflows.
- Keep the user-facing process small and easy to learn.
- Preserve history in a way that is easy for both humans and tools to find, inspect, reference, and audit.
- Make standards/context selection deterministic so RuneCode can generate an auditable resolved context.
- Make reviewable intent artifacts first-class so both humans and RuneCode can trace what changed and why.
- Preserve enough structured traceability that a lineage/index view can be generated later for historical changes, specs, standards, and decisions.
- Let RuneCode integrate deeply without making RuneContext itself a policy engine.
- Keep linked sources and resolved context packs reproducible and integrity-verifiable.
- Keep Plain mode low-friction while allowing Verified mode to add stronger provenance when needed.

## Non-Goals

RuneContext should not become any of the following:

- a giant orchestration engine
- a proprietary IDE workflow
- a permissions or capability system
- a replacement for RuneCode's runtime security model
- a hidden global configuration system
- a system that requires RuneCode to be useful
- a workflow that requires every change to materialize the full artifact set
- a system that silently auto-promotes all change details into stable project docs
- a system that forces all users into cryptographic assurance workflows even when they only want lightweight standalone project memory

## Product Decomposition

RuneContext should be designed as three layers, even if those layers are not split into three repositories on day one.

### 1. RuneContext Core

RuneContext Core is the portable format and workflow model.

It owns:

- the on-disk file/folder layout
- markdown/yaml/json conventions
- schemas for core files
- context bundle resolution semantics
- change lifecycle semantics
- generated manifest/index conventions

It does not own:

- IDE-specific command shims
- RuneCode runtime permissions or trust rules
- provider integrations

### 2. RuneContext Adapters

RuneContext Adapters are thin tool-specific layers that make RuneContext comfortable to use in various agentic environments.

They own:

- slash command docs
- skill/prompt files
- tool-specific bootstrap or installer helpers
- minimal translation between a tool's UX and RuneContext's core operations

They should not redefine the core format.

### 3. RuneCode Integration

RuneCode Integration is the RuneCode-specific layer that resolves RuneContext into auditable, deterministic inputs for runs.

It owns:

- resolving context bundles and changes into a `context pack`
- hashing/signing or otherwise binding the resolved pack into audit/provenance flows
- integrating standards/changes with RuneCode approvals and PR generation
- mapping RuneContext tags/metadata into RuneCode-specific runtime behavior where appropriate

It should not mutate RuneContext's portable format into a RuneCode-only format.

## Why The Three Layers Need To Exist

These layers solve different problems:

- Core solves portability and source-of-truth format.
- Adapters solve tool UX.
- RuneCode Integration solves security-aware runtime resolution and audit.

If these layers are collapsed into one thing, several problems appear quickly:

- the format becomes RuneCode-specific
- adapter churn destabilizes the source format
- non-RuneCode users get a worse experience
- runtime security and audit needs start leaking into authoring files as if markdown were authoritative policy

The separation is mainly conceptual and packaging-oriented. Users should not need to think about it often.

## Packaging, Repositories, And Releases

Do not start with three GitHub repositories and three separate binaries/downloads.

Recommended starting structure:

- `runecontext` repository
  - contains RuneContext Core and RuneContext Adapters
- `runecode` repository
  - contains RuneCode and its RuneContext integration

Recommended release model:

- `runecontext` releases:
  - schemas
  - reference docs
  - optional `runectx` CLI
  - adapter packs
- `runecode` releases:
  - RuneCode binary
  - built-in RuneContext resolver/integration

Recommended repository layout:

```text
runecontext/
  core/
  adapters/
    claude-code/
    opencode/
    codex/
    generic/
  docs/
  schemas/
  cli/

runecode/
  cmd/
  internal/
  integrations/runecontext/
```

This preserves a clean product boundary without early operational overhead.

## Releases and Installation

RuneContext should use a repo-first, CLI-assisted release and installation model.

It should not be:

- CLI-first
- global-home-directory-first
- bash-installer-first
- template-only

### Recommended Release Method

RuneContext should ship as:

- a versioned GitHub repository
- GitHub Releases containing:
  - release ZIP/tarball bundles
  - checksums and, when available, signatures
  - release notes and changelog
  - schema bundle
  - adapter packs
  - optional `runectx` binaries
- a compatibility matrix describing supported RuneCode <-> RuneContext versions

Reasoning:

- RuneContext is fundamentally a markdown/git-native project asset system
- users should be able to inspect, copy, vendor, and review the files directly
- the release artifact should preserve that transparency instead of hiding behavior behind an installer

### Recommended Installation Model

RuneContext should support three installation lanes.

#### Lane 1: Manual Repo Install (Canonical)

Users should always be able to:

- download a tagged release ZIP/tarball and copy `runecontext/` into a repo
- clone or vendor a tagged RuneContext release into their repository
- point `runecontext.yaml` at a pinned external RuneContext repository

This is the canonical install model because it is:

- transparent
- git-native
- easy to audit
- compatible with users who do not want RuneCode or any package-manager dependency

#### Lane 2: `runectx` Convenience Install

The `runectx` CLI should exist as a convenience layer for technical users, automation, and updates.

Recommended commands:

- `runectx init`
- `runectx update`
- `runectx validate`
- `runectx doctor`
- `runectx adapter sync <tool>`

Important rule:

- the CLI should help manage repo files, but must not become the source of truth
- all meaningful state should remain in reviewable project files

#### Lane 3: RuneCode-Managed Integration

RuneCode should be able to:

- scaffold RuneContext into a repository
- sync/update a supported RuneContext version
- propose reviewed diffs for upgrades or adapter changes

RuneCode should treat RuneContext as a pinned dependency, not as a mutable hidden external install.

### Default User Experience

The intended user flow should be:

- install RuneContext into the repository
- sync/install the appropriate tool adapter
- use the tool-native command pack for daily work
- use `runectx` for updates, validation, debugging, and automation

This preserves the desired model:

- repo files are the product
- adapters are the daily UX
- `runectx` is the helper

### Update Strategy

RuneContext updates should be:

- explicit
- versioned
- diff-first
- reviewable before merge

Recommended update behavior:

- `runectx update` should stage or propose file changes as a reviewable diff
- updates must not silently overwrite local project customizations
- adapter sync/update should be namespaced and merge-aware rather than destructive

### Package Manager Strategy

Package managers should be convenience channels, not the canonical distribution method.

Recommended rollout:

- phase 1:
  - GitHub repo + release artifacts
  - optional `runectx` binaries
- phase 2:
  - package-manager convenience such as npm
- phase 3:
  - additional channels such as Homebrew, Scoop, or winget if adoption warrants them

Important rule:

- RuneContext should not require a package manager in order to be usable

### What To Avoid

Avoid the following release/install patterns:

- required global installs into home-directory state
- bash-only installers
- overwriting `.claude`, `.github`, or other existing instruction files in place
- template-only distribution as the primary install path
- hidden runtime-manager dependencies that are required for normal operation
- silent auto-updates or non-reviewable file replacement

### RuneCode Compatibility Expectations

RuneCode integration should:

- pin supported RuneContext versions
- read `runecontext_version` from the project-root `runecontext.yaml` and compare it against its supported range before deeper resolution
- expose sync/upgrade as explicit, reviewable operations
- record which RuneContext version and source revision a run used
- fail clearly when the loaded RuneContext version is unsupported or out of compatibility range

## Optional Assurance And Verifiable Tracing

RuneContext should support progressive assurance rather than forcing every user into the same operational mode.

### Assurance Tiers

RuneContext should start with two assurance tiers:

- `Plain`
  - default lightweight mode for standalone/manual/tool-assisted use
  - no additional assurance artifacts are required
- `Verified`
  - enables generated assurance artifacts, baseline recording, and stronger verifiable tracing

Future extension:

- `Anchored`
  - may be added later once RuneCode's audit anchoring capabilities are mature enough to support it cleanly

### Tier Selection Rules

- `Plain` should be the default for users who want RuneContext on its own without extra ceremony
- `Verified` should be required when RuneCode is used for normal audited/project-operating workflows
- a Plain-mode project opened by RuneCode should trigger a migration/enablement flow before normal audit-bound write workflows proceed
- the active tier must be persisted in the project-root `runecontext.yaml` as the version-controlled `assurance_tier`

### Keep Friction Low

The hand-authored source of truth should stay the same in both tiers.

Users should primarily continue editing:

- `proposal.md`
- `standards.md`
- `status.yaml`
- `bundles/*.yaml`
- `specs/`
- `decisions/`

Additional assurance files should only be generated when `Verified` mode is enabled.

### Verified-Mode Generated Artifacts

When `Verified` mode is enabled, RuneContext should generate assurance artifacts rather than requiring users to hand-author them.

Recommended generated area:

```text
runecontext/
  assurance/
    baseline.yaml
    receipts/
      context-packs/
      changes/
      promotions/
      verifications/
```

Recommended generated artifact families:

- baseline record
  - captures the adoption point for verification going forward
- context-pack receipts
  - bind resolved pack hashes, source revisions, and generation metadata
- change-event receipts
  - bind important lifecycle events to change IDs and relevant context
- promotion receipts
  - record what durable knowledge was promoted where
- verification receipts
  - record verification outcomes and associated evidence references

These files should be machine-generated and schema-versioned.

Receipt identity and concurrency rules:

- every assurance receipt should carry a `receipt_id` and `receipt_hash`
- receipt filenames should be content-addressed or otherwise collision-resistant rather than allocated from a shared mutable counter
- receipt generation should not require a central mutable index
- concurrent branches should merge receipt files by normal file union wherever possible

### Transitioning From Plain To Verified

RuneContext should support a clean adoption path from Plain mode to Verified mode.

Recommended transition flow:

1. choose an adoption commit
2. generate a baseline record for the current RuneContext state
3. record the current repo commit SHA and relevant resolved context information
4. enable Verified mode for future operations
5. begin generating assurance receipts for new context resolutions, changes, promotions, and verification events

This creates a clear trust boundary: a project has a visible point where verifiable tracing starts.

### Backfilling Earlier Work

RuneContext should automate as much trust-building for earlier work as practical.

Recommended approach:

- inspect git history, existing specs, decisions, and change-related files
- reconstruct historical links where possible
- generate imported/backfilled lineage and provenance records automatically
- attach those backfilled records to the adoption baseline so they remain distinguishable from natively captured Verified evidence

Important rule:

- backfilled evidence should increase trust and usability for historical work
- backfilled evidence must remain explicitly distinguishable from evidence captured natively after Verified mode is enabled

Recommended provenance classes:

- `captured_verified`
  - generated natively after Verified mode is enabled
- `imported_git_history`
  - reconstructed automatically from git history and project artifacts during migration

This allows the system to provide sufficient trust and continuity for earlier work without pretending that backfilled evidence is identical to live-captured evidence.

### Trust Model For Earlier Work

RuneContext should take the strongest practical path for earlier work:

- automate backfill wherever possible
- use git history as the primary reconstruction source
- trust imported historical lineage as repo-history-derived provenance
- reserve the strongest cryptographic confidence for post-baseline `captured_verified` evidence

In other words:

- earlier work can become much more trustworthy through automated backfill
- going forward, Verified mode should make trust and auditability rock solid

### RuneCode Expectations

RuneCode should integrate with these tiers as follows:

- Plain mode
  - acceptable for standalone/manual RuneContext use
  - not acceptable for normal RuneCode-operated audited workflows
- Verified mode
  - required for RuneCode-managed project-operating workflows
  - RuneCode should consume and bind assurance artifacts into its own audit/provenance history

RuneCode should remain the authoritative runtime security/audit system, while RuneContext assurance artifacts serve as portable, evidence-friendly inputs.

## RuneCode Context And Integration Constraints

This section explains what RuneCode is as a product, its core tenets, and the constraints RuneContext must respect when integrating with it.

### What RuneCode Is

RuneCode is a QubesOS-inspired, security-first agentic automation platform for software engineering.

Its product mission is to let users run agentic coding workflows with:

- least-privilege defaults
- explicit approvals for elevated risk
- tight isolation boundaries
- cryptographic provenance
- tamper-evident auditability

The product framing from `agent-os/product/mission.md` and the current spec suite is consistent:

- isolation and cryptographic provenance are co-equal first-class pillars
- work runs in tightly scoped isolates
- capabilities are deny-by-default
- actions, decisions, and outputs are auditable

### RuneCode Core Tenets Relevant To RuneContext

RuneContext must be designed to fit these RuneCode tenets.

#### Security-First And Deny-By-Default

From the product mission, initial MVP spec suite, and policy engine direction:

- RuneCode is security-first, not convenience-first
- isolation and cryptographic provenance are foundational, not add-ons
- capability expansion requires explicit signed inputs and policy authorization
- deny-by-default is a hard invariant for network, filesystem, shell, and secrets access

RuneContext must not encourage designs that assume permissive ambient access.

#### Isolation Is The Primary Runtime Boundary

From the MVP spec suite and workflow specs:

- microVMs are the preferred primary isolation boundary
- containers are reduced-assurance and explicit opt-in only
- no host filesystem mounts into isolates
- data moves via explicit, hash-addressed artifacts

RuneContext should align with artifact-based handoff and review, not implicit shared mutable state.

#### Policy Is Deterministic And Signed

From `policy-engine-v0` and `protocol-schemas-v0`:

- effective behavior is governed by signed manifests, typed policy inputs, and deterministic policy evaluation
- approval profiles affect when explicit human approval is needed, but never convert `deny -> allow`
- unknown or malformed policy-visible inputs fail closed

RuneContext must never be treated as a hidden policy side channel.

#### Cross-Boundary Communication Is Typed And Schema-Validated

From `protocol-schemas-v0`, `broker-local-api-v0`, and related specs:

- cross-boundary communication must be structured, schema-validated, versioned, and hash-addressable
- freeform model output does not directly trigger privileged actions
- approvals, policy decisions, manifests, audit events, and artifact provenance all use typed shared object families

RuneContext should remain a human-oriented knowledge layer, not a replacement for typed runtime contracts.

#### Auditability Is Non-Negotiable

From the audit log, audit anchoring, and related specs:

- proposals, validations, authorizations, executions, gate results, and approvals are meant to be recorded in a tamper-evident audit trail
- audit verification is first-class
- posture changes and degraded modes must be visible rather than hidden

RuneContext should provide reviewable artifacts that RuneCode can bind into audit history.

#### The Workflow Runner Is Untrusted

From `workflow-workspace-roles-gates-v0` and the current tech stack:

- the TS/Node workflow runner is treated as untrusted at runtime
- stable integration surfaces are typed schemas and broker/local APIs, not framework internals
- workspace roles run offline; egress is isolated behind dedicated gateway roles

RuneContext integration must assume that human-readable planning context and untrusted orchestration are not equivalent to trusted runtime authority.

#### Extensibility Must Not Weaken Safety

From `workflow-extensibility-v0` and `bridge-runtime-protocol-v0`:

- custom workflows must be schema-validated and hash-bound
- extension points must not widen capabilities beyond signed manifests and policy
- bridge runtimes must remain in explicit LLM-only posture
- provider-specific behavior must not gain workspace/file/patch capabilities implicitly

RuneContext should help author and organize intent, but should not become an executable plugin surface that bypasses RuneCode's typed runtime extension model.

### RuneCode Architecture Traits That Matter To RuneContext

The current product/tech-stack direction implies these practical integration traits:

- Go-based local control plane for trusted security-critical components
- TypeScript/Node runner treated as untrusted
- SQLite plus append-only files for local durable state and indexing
- TUI-first local UX for runs, approvals, diffs, artifacts, and audit timelines
- local-first, single-user MVP scope

Important note:

- RuneCode's current MVP direction is local-first and single-user on a single machine
- RuneContext should still remain portable enough that it does not block future remote/server RuneCode execution or non-RuneCode workflows

### How RuneCode Needs To Interface With RuneContext

RuneContext should integrate deeply with RuneCode, but only in specific ways.

#### 1. RuneContext Is A Knowledge Layer, Not A Runtime Authority

RuneCode should consume RuneContext as:

- project knowledge
- standards knowledge
- context bundle selection
- change intent and traceability
- promotion suggestions

RuneCode must not consume RuneContext as:

- an authority for capabilities
- an authority for approvals
- an executable runtime workflow definition
- a substitute for signed manifests, typed policy inputs, or typed broker/runtime protocols

#### 2. RuneCode Must Resolve RuneContext Deterministically

RuneCode should resolve RuneContext inputs into a deterministic `context pack` that can be:

- hashed
- bound into audit history
- reproduced locally or remotely from the same refs and inputs

That resolution must include:

- selected context bundle(s)
- selected standards
- active change ID
- active intent artifact (`proposal.md`)
- relevant project-level context

#### 3. RuneCode Must Bind Intent Into Audit History

RuneCode should bind the active change ID and `proposal.md` into auditable run history so future reviewers can see:

- what changed
- why it changed
- which standards were in scope
- what assumptions were recorded

This supports a later lineage view without requiring RuneContext v1 to ship one.

#### 4. RuneCode Must Keep Context Bundles Separate From Approval Profiles

RuneContext `context bundles` and RuneCode approval profiles are different concepts.

- context bundles select reusable project context inputs
- approval profiles control when otherwise-allowed actions need explicit human approval

RuneContext must not overload or blur these concepts.

#### 5. RuneCode May Use RuneContext Tags, But Only As Advisory Inputs

RuneCode may interpret standard or change metadata such as:

- trust-boundary tags
- cross-boundary tags
- security sensitivity tags

But those tags are advisory inputs for planning, presentation, search, and context assembly workflows. They are not themselves runtime permissions.

Explicit boundary rule:

- RuneContext tags must never be direct inputs to policy-engine allow/deny evaluation
- RuneContext tags must never directly select or change an approval profile
- if a trusted RuneCode component wants to map RuneContext tags to stricter review suggestions, it must do so through a trusted-side allowlist/mapping layer and treat RuneContext content as untrusted input

#### 6. RuneCode Must Keep Promotion Reviewable

RuneCode should automatically assess durable knowledge and propose promotion updates, but it should keep those proposals reviewable.

RuneContext should support:

- explicit promotion targets
- explicit "no promotion needed" outcomes
- stable references between changes, specs, and decisions

RuneCode should not silently fold all historical change detail into stable docs.

#### 7. RuneContext Should Not Compete With RuneCode's Typed Workflow Model

RuneCode's typed runtime workflow model is moving toward schema-validated process definitions and typed protocol objects.

RuneContext should complement that by providing:

- human-readable authoring and review artifacts
- durable standards and project memory
- reviewable change intent

RuneContext should not try to replace:

- typed `ProcessDefinition` objects
- signed manifests
- policy decisions
- broker/runtime protocols

#### 8. RuneContext Should Fit Offline Workspace Roles And Gateway Separation

RuneCode's workflow direction separates:

- offline workspace roles
- explicit gateway roles for egress
- LLM-only bridge runtimes

RuneContext should therefore assume:

- work is planned with explicit role separation in mind
- standards and change docs may discuss egress/auth/provider behavior, but must not imply direct workspace+egress authority in a single runtime surface
- provider/tooling integrations remain bounded by RuneCode's gateway and bridge constraints

### Practical RuneContext Requirements Implied By RuneCode

Given RuneCode's current product and spec direction, RuneContext should satisfy these integration requirements:

- every meaningful run can point to a stable change ID
- every meaningful run can point to a stable intent artifact
- standards selection is deterministic and reviewable
- traceability survives long enough for future audit/lineage tooling
- linked-repo and embedded-repo resolution can be pinned and audited
- resolved context packs carry canonical top-level hashes
- RuneContext delivery into isolates uses typed broker/runtime transport plus hash-addressed artifacts rather than host mounts
- no RuneContext artifact silently widens runtime capabilities
- no adapter or command pack invents a RuneCode-only hidden source of truth
- RuneCode-operated audited workflows require Verified assurance mode

## Usage Scenarios

### Scenario 1: Local Developer Workflow Using RuneCode

In this scenario:

- the project repo contains `runecontext/`, or a project pointer file refers to a linked RuneContext repo
- the developer uses RuneCode locally
- the project is operating in Verified assurance mode

Expected flow:

1. RuneCode opens the project.
2. RuneCode detects embedded or linked RuneContext configuration.
3. RuneCode resolves the selected context bundle plus the active change into a deterministic context pack.
4. RuneCode uses that pack for planning, implementation, verification, and promotion assessment.
5. If RuneCode discovers a missing standard or context gap, it drafts:
   - a new or updated standard file
   - any needed context bundle edits
   - any related change/spec updates
6. RuneCode binds the active change intent into auditable history.
7. RuneCode presents the diff for approval or commits it to a branch/PR depending on the active workflow.

Key rule:

- RuneCode should not make hidden out-of-band changes to RuneContext's source of truth unless the user explicitly allows that behavior.

### Scenario 2: Remote Server Workflow Using RuneCode

In this scenario:

- this represents a future or extended RuneCode deployment posture beyond the current local-first MVP target
- RuneCode runs on a remote server or automation host
- the server has access to the application repo and either embedded RuneContext content or a linked RuneContext repo/ref
- the project is operating in Verified assurance mode

Expected flow:

1. RuneCode checks out the project repo and resolves RuneContext at pinned refs.
2. RuneCode resolves the selected context bundle and change on the server exactly as it would locally.
3. RuneCode binds the resolved context pack into audit/provenance records.
4. RuneCode performs planning or implementation.
5. RuneCode submits changes through approved mechanisms such as a branch and PR.

Important properties:

- local and remote resolution must produce the same result from the same inputs
- the linked/embedded source ref used for resolution should be recorded in audit data
- `type: path` sources are not valid for remote/CI reproducible runs
- approvals should use RuneCode's normal approval system, not a RuneContext-specific one

### Scenario 3: Non-RuneCode Workflow Using RuneContext Directly

In this scenario:

- the developer does not use RuneCode at all
- the developer may use Claude Code, Codex, another tool, or manual editing only

Expected flow:

- the developer primarily uses a RuneContext adapter command pack for their tool of choice, or edits RuneContext markdown/yaml directly
- the `runectx` CLI is available for power users, automation, and direct workflows
- adapters provide tool-specific prompt/command docs as convenience, not as the source of truth
- the project can still use standards, context bundles, changes, and specs without RuneCode

Key rule:

- RuneContext must remain useful even if the only tooling available is git, a text editor, and optional helper scripts/CLI

## Terminology

Use the following terms consistently:

- `standard`
  - a reusable normative document describing a rule, practice, or convention
- `project context`
  - the broader durable project knowledge layer that includes project files, standards, context bundles, changes, specs, and decisions
- `context bundle`
  - a named reusable selection of project-context inputs across one or more aspect families that may inherit from other context bundles
- `context pack`
  - a resolved, deterministic artifact generated from one or more context bundles plus optional change/project inputs for runtime use
- `change`
  - a proposed or in-flight body of work with lifecycle state and stable identity
- `spec`
  - a stable, current document describing a feature/subsystem after changes are promoted into long-lived project knowledge
- `decision`
  - an ADR-like durable architectural or policy decision

Disambiguation rule:

- use `project context` for the overall durable knowledge layer
- use `context bundle` for reusable selectors
- use `context pack` for resolved runtime input
- use `model context window` when referring to LLM token context
- avoid using bare `context` alone in normative specifications when a more specific term is available

### Why `context bundle` Was Chosen

The folder should be named `bundles/`.

The user-facing term should be `context bundle` rather than `profile` because RuneCode already uses approval-profile terminology elsewhere, and the two concepts should stay distinct.

The product name `RuneContext` makes `work context` too collision-prone inside the system. Using `context bundle` keeps the reusable selector concept distinct from the broader project context and from runtime context assembly.

## Storage Modes

RuneContext must support both embedded and linked modes.

### Project Root Configuration

Every project should have a root `runecontext.yaml` that declares the project-level RuneContext compatibility version and assurance tier.

Required top-level fields:

- `schema_version`
- `runecontext_version`
- `assurance_tier`
- `source`

Recommended embedded-mode example:

```yaml
schema_version: 1
runecontext_version: 0.1
assurance_tier: plain
source:
  type: embedded
  path: runecontext
```

Rules:

- `runecontext_version` is the project-level compatibility version RuneCode uses for in-band support checks
- `assurance_tier` must be one of `plain` or `verified` in v1
- future versions may add `anchored`
- RuneCode should read `runecontext_version` and fail clearly when the project is outside its supported range

### Embedded Mode

RuneContext lives inside the project repository, typically at `./runecontext/`.

Benefits:

- simplest mental model
- easy to branch and review together with code
- ideal for smaller teams or project-specific standards

### Linked Mode

The project repo points to a dedicated RuneContext repository.

Benefits:

- shared standards/spec governance across multiple repos
- easier central management of reusable standards
- allows a dedicated RuneContext repo to evolve independently

Recommended immutable git pointer file:

```yaml
schema_version: 1
runecontext_version: 0.1
assurance_tier: plain
source:
  type: git
  url: git@github.com:org/project-runecontext.git
  commit: 0123456789abcdef0123456789abcdef01234567
  subdir: runecontext
```

Also support a signed-tag form when the implementation can verify the tag and record the resolved commit:

```yaml
schema_version: 1
runecontext_version: 0.1
assurance_tier: plain
source:
  type: git
  url: git@github.com:org/project-runecontext.git
  signed_tag: v1.2.3
  expect_commit: 0123456789abcdef0123456789abcdef01234567
  subdir: runecontext
```

Mutable refs are allowed only via explicit opt-in and should be clearly warned about:

```yaml
schema_version: 1
runecontext_version: 0.1
assurance_tier: plain
source:
  type: git
  url: git@github.com:org/project-runecontext.git
  ref: main
  allow_mutable_ref: true
  subdir: runecontext
```

Also support a local path form for developer-local workflows:

```yaml
schema_version: 1
runecontext_version: 0.1
assurance_tier: plain
source:
  type: path
  path: ../project-runecontext
```

Notes:

- the exact filename can be `runecontext.yaml` at project root
- cache/materialization behavior is implementation detail, not part of the portable authoring model
- correctness must not depend on hidden home-directory config
- git sources should prefer immutable refs (`commit`) or verifiable signed tags
- mutable refs like `main` should require explicit opt-in and should produce warnings because they are not reproducible over time
- the resolved commit SHA must always be recorded in the context pack, even when the source is selected by signed tag or mutable ref
- signed-tag verification is an advanced mode; if implemented, it must verify against explicitly configured trusted signer keys on the trusted side, record the resolved signer identity/fingerprint plus resolved commit in the context pack, and fail closed on untrusted signer or commit mismatch
- `type: path` is for developer-local convenience only; it is non-auditable and should be treated as `unverified_local_source`
- `type: path` should be invalid for remote/CI contexts unless a trusted wrapper explicitly downgrades the run to non-reproducible/unverified mode
- `type: path` sources and any symlinks they contain must resolve within the declared local source tree
- path-sourced projects should be snapshotted at resolution time before hashing to reduce TOCTOU risk
- path-sourced context packs may be hashed for local debugging, but must not be treated or recorded as verified provenance in RuneCode audit flows

### Monorepo Support

RuneContext should support monorepos explicitly.

Rules:

- multiple `runecontext.yaml` files may exist in one monorepo
- tooling should discover the nearest ancestor `runecontext.yaml` from the current working directory unless an explicit project root/path is provided
- a root-level RuneContext may serve the entire monorepo
- nested RuneContext files may provide package- or subtree-specific overrides
- tooling should report which `runecontext.yaml` was selected so the applied project context is always visible

## Core On-Disk Layout

Recommended embedded layout:

```text
runecontext/
  project/
    mission.md
    roadmap.md
    stack.md
  standards/
    global/
    security/
    backend/
    frontend/
  bundles/
    base.yaml
    go-control-plane.yaml
    trust-boundary-heavy.yaml
  changes/
    CHG-2026-001-a3f2-auth-gateway/
      proposal.md
      standards.md
      design.md
      tasks.md
      references.md
      verification.md
      status.yaml
  specs/
    auth-gateway.md
    workflow-engine.md
  decisions/
    DEC-0001-trust-boundary-model.md
  commands/
    discover-standards.md
    select-bundle.md
    propose-change.md
    shape-change.md
    close-change.md
  schemas/
    standard.schema.json
    bundle.schema.json
    change-status.schema.json
    context-pack.schema.json
  assurance/
    baseline.yaml
    receipts/
  manifest.yaml
```

Notes:

- `manifest.yaml` should be generated or refreshed by tooling. It should not be the sole source of truth.
- the authoritative source of truth should remain the markdown/yaml files themselves.
- generated indexes/manifests may exist for speed, browsing, audit, or tooling convenience.
- the example change folder above shows the full shape; minimum mode only requires `status.yaml`, `proposal.md`, and `standards.md`.
- `commands/` contains the canonical in-project human-readable command reference and adapter source material, but not executable workflow definitions or authoritative machine contracts.
- `assurance/` is optional and should only be generated when Verified mode is enabled.

### Machine-Readable Schema Versioning

For machine-readable RuneContext files such as `runecontext.yaml`, `bundles/*.yaml`, `changes/*/status.yaml`, generated context packs, and any generated assurance receipts:

- `schema_version` is required
- unknown `schema_version` values must fail clearly rather than being guessed at permissively
- for known schema versions, implementations should preserve unknown fields when round-tripping and may ignore them unless the active schema version defines stricter behavior

Machine-readable YAML profile:

- no anchors or aliases
- no duplicate keys
- no implicit type coercion beyond the active schema
- no custom tags
- UTF-8 only
- canonical hashes are computed over the normalized JSON data model derived from this restricted YAML profile, not from raw YAML bytes
- generated machine-readable artifacts may be stored as JSON on disk if an implementation prefers, as long as the canonical data model and hash rules remain identical

### Generated Artifact Commit Policy

Recommended defaults:

- hand-authored core files should be committed
- `manifest.yaml` is optional and regenerable; tooling should work whether it is committed or gitignored
- generated context packs should normally be generated on demand and not committed
- `assurance/baseline.yaml` should be committed when Verified mode is enabled
- `assurance/receipts/context-packs/` should normally be treated as ephemeral and gitignored
- `assurance/receipts/changes/`, `assurance/receipts/promotions/`, and `assurance/receipts/verifications/` are committable evidence artifacts for standalone Verified mode; RuneCode-managed environments may additionally or alternatively store equivalent evidence in RuneCode's audit system

## Core File Types

### Project Files

The `project/` folder holds durable project-wide context.

Recommended files:

- `mission.md`
  - what the project is and why it exists
- `roadmap.md`
  - major work areas, future direction, and current/next priorities
- `stack.md`
  - implementation stack and platform constraints

These files are portable project knowledge, not RuneCode runtime policy.

### Standards

The `standards/` tree holds reusable normative documents.

Standards should be organized by domain, for example:

- `standards/global/`
- `standards/security/`
- `standards/backend/`
- `standards/frontend/`

Recommended standard frontmatter:

```yaml
---
schema_version: 1
id: security/trust-boundary-interfaces
title: Trust Boundary Interfaces
tags: [security, trusted, cross-boundary]
status: active
suggested_context_bundles: [trust-boundary-heavy]
---
```

Field guidance:

- `id`
  - required, stable unique identifier
  - should match the path under `standards/` without the `.md` extension (for example `standards/security/trust-boundary-interfaces.md` -> `security/trust-boundary-interfaces`)
- `title`
  - required human-readable title
- `tags`
  - optional but recommended
- `status`
  - recommended; examples: `active`, `draft`, `deprecated`
- `suggested_context_bundles`
  - optional helper metadata for tooling suggestions only
- `replaced_by`
  - optional path or id-like reference for deprecated standards that have a clear successor
- `aliases`
  - optional former IDs retained temporarily during rename/migration workflows

Validation rules:

- standard IDs must be unique across the full `standards/` tree
- tooling should reject standards whose `id` does not match the path-relative convention unless an explicit migration/compatibility mode is in effect

Resolution and migration rules:

- context-bundle resolution uses paths under `standards/` as the authoritative selection mechanism
- standard `id` is primarily for indexing, traceability, and migration support
- if a standard is renamed, tooling should rewrite affected path references in bundles/specs/change docs and may populate `aliases` during a compatibility window
- deprecated standards may still resolve when directly selected for compatibility, but tooling should emit a warning
- when `replaced_by` is present, tooling should surface that suggested migration target
- migration/compatibility mode is a temporary tooling-assisted state used during renames/imports where alias metadata and rewritten references may coexist until the migration is completed

Important rule:

- `suggested_context_bundles` is not authoritative membership.
- authoritative inclusion/exclusion lives in `bundles/*.yaml`.

### Context Bundles

The `bundles/` folder defines reusable context selectors.

Required and optional fields:

- `schema_version` - required
- `id` - required
- `includes` - required
- `extends` - optional
- `excludes` - optional

Recommended aspect-aware base schema:

```yaml
schema_version: 1
id: trust-boundary-heavy
extends:
  - base
  - go-control-plane
includes:
  project:
    - project/mission.md
    - project/stack.md
  standards:
    - standards/security/**
    - standards/global/deterministic-check-write-tools.md
  decisions:
    - decisions/DEC-0001-trust-boundary-model.md
excludes:
  standards:
    - standards/frontend/**
```

Rules:

- `extends` references other context bundle `id` values, never file paths
- `includes` and `excludes` are aspect-aware maps keyed by RuneContext aspect family
- each aspect entry may use exact paths and glob-style patterns
- `includes` is required even if a bundle only extends parents
- context bundles may select from multiple aspects of RuneContext in one place, including `project/`, `standards/`, `specs/`, and `decisions/`
- active `changes/` are generally selected separately at runtime rather than baked into reusable bundles
- tooling should record the concrete file set matched by each glob at resolution time
- when a glob's matched file set changes between resolutions, tooling should emit a visible diff/warning
- implementations may support a stricter pinned-glob mode that requires explicit acceptance when the matched file set changes
- bundle patterns are always relative to the selected aspect root inside the RuneContext tree
- absolute paths, drive-qualified paths, and any pattern containing `..` path traversal segments are invalid
- after normalization and symlink resolution, matched files must remain inside both the RuneContext root and the selected aspect root
- files that escape those roots through symlinks or traversal must be rejected

### Changes

The `changes/` tree is the core workflow surface.

Each change must have a stable ID in the folder name. Recommended format:

- `CHG-YYYY-NNN-RAND-short-slug`

Example:

- `CHG-2026-001-a3f2-auth-gateway`

Allocation and uniqueness rules:

- `NNN` is a repo-local monotonic counter scoped per calendar year
- `RAND` is a short collision-resistant suffix generated by tooling (recommended 4-6 lowercase hex or base32 characters)
- change creation tooling must validate that the ID is unique across the entire `changes/` tree before writing files
- if a collision is detected, tooling must allocate a new ID or fail clearly and require regeneration
- because `RAND` is part of the stable ID, reallocation should be rare; if it is required, tooling must rewrite dependent local references in the same change before merge

Concurrent workflow guidance:

- teams may create changes concurrently on different branches
- generated manifests/indexes should be sorted deterministically to minimize merge conflicts
- if two branches somehow allocate the same change ID, one branch must be reallocated before merge and local references must be rewritten atomically

### Minimum And Full Change Shapes

RuneContext should use progressive disclosure.

Every substantive work item should get a change ID and start in minimum mode.

Minimum mode files:

- `status.yaml`
  - lifecycle state, work type/size, traceability, and machine-readable metadata
- `proposal.md`
  - the canonical reviewable intent artifact for what changes and why
- `standards.md`
  - the always-present link from the change to the project's standards knowledge base

Full mode adds these files to the minimum set:

- `design.md`
  - shaping and implementation approach
- `tasks.md`
  - detailed implementation checklist or task breakdown
- `references.md`
  - external references, related changes, issues, links, docs
- `verification.md`
  - verification notes, checks, and/or user acceptance notes

Important rule:

- the full artifact set is the maximum shape, not the default minimum
- RuneCode and other tooling may materialize full-mode files when the change crosses a complexity/risk threshold

Recommended `status.yaml` shape:

```yaml
schema_version: 1
id: CHG-2026-001-a3f2-auth-gateway
title: Add auth gateway
status: proposed
type: feature
size: medium
context_bundles:
  - go-control-plane
related_specs:
  - specs/auth-gateway.md
related_decisions: []
related_changes: []
depends_on: []
informed_by: []
supersedes: []
superseded_by: []
created_at: 2026-03-15
closed_at: null
verification_status: pending
promotion_assessment:
  status: suggested
  suggested_targets:
    - target_type: spec
      target_path: specs/auth-gateway.md
      summary: Promote externally visible auth-gateway behavior and accepted constraints
```

Field guidance:

- `id`
  - required stable change identifier
- `title`
  - required human-readable summary
- `status`
  - required lifecycle status
- `type`
  - required base enum: `project`, `feature`, `bug`, `standard`, `chore`
  - custom values should use an `x-` prefix, for example `x-migration`
- `size`
  - recommended; examples: `small`, `medium`, `large`
- `context_bundles`
  - resolved or selected context bundle IDs
- `related_specs`
  - stable spec paths touched or informed by the change
- `related_decisions`
  - durable decision paths touched or informed by the change
- `related_changes`
  - non-hierarchical links to other relevant changes
- `depends_on`
  - changes that must land first or whose outputs are assumed
- `informed_by`
  - older changes whose reasoning or outputs influenced this one
- `verification_status`
  - required base enum: `pending`, `passed`, `failed`, `skipped`
- `schema_version`
  - required version marker for machine-readable file compatibility
- `promotion_assessment`
  - durable knowledge assessment status and structured suggested targets

Recommended `promotion_assessment.status` values:

- `pending`
- `none`
- `suggested`
- `accepted`
- `completed`

Validation rules:

- when `status = superseded`, `superseded_by` must be present and non-empty
- supersession links should be bidirectionally consistent: if change A lists change B in `superseded_by`, change B should list change A in `supersedes`
- `promotion_assessment.suggested_targets` entries should include at minimum:
  - `target_type` (`spec`, `standard`, or `decision`)
  - `target_path`
  - `summary`

### Proposal.md Structure

`proposal.md` must use a strict structure even in minimum mode.

Required sections:

- `Summary`
- `Problem`
- `Proposed Change`
- `Why Now`
- `Assumptions`
- `Out of Scope`
- `Impact`

Parsing and validation rules:

- the required sections must appear in the exact order shown above
- tooling should parse them as level-2 markdown headings (`##`)
- every required section must be present
- every required section must either contain content or an explicit `N/A`
- additional custom sections are allowed only after the required section block so parsers can treat the required core as stable

Design intent:

- `proposal.md` is the canonical reviewable intent artifact
- it must be readable by humans and easy for tooling to summarize
- later changes, specs, and decisions should be able to point back to the originating change ID and proposal

### Standards.md Structure

`standards.md` must always be present.

It should be automatically maintained by RuneContext tooling and RuneCode integration so users do not have to manually keep it in sync.

Recommended sections:

- `Applicable Standards`
- `Standards Added Since Last Refresh`
- `Standards Considered But Excluded` (optional)
- `Resolution Notes` (optional)

Important rules:

- the canonical section listing applicable standards may be normalized/regenerated by tooling
- users may still edit the file manually when needed
- the system should make updates reviewable rather than silently drifting the file

### Automatic Standards Maintenance

Recommended behavior:

- on `change new`
  - infer likely context bundle(s), resolve standards, and create/populate `standards.md`
- on `change shape`
  - refresh standards based on any changed scope, assumptions, or context bundle selection
- during RuneCode planning/implementation
  - if new standards become relevant, propose a `standards.md` update in the same reviewed diff

The user should not need to manually add standards to `standards.md` for normal operation.

Important review rule:

- automatic maintenance of `standards.md` must always produce a reviewable diff before commit/merge
- tooling may refresh the file automatically in the working tree or proposed diff, but it must not silently rewrite and commit it without review

### Stable Specs

The `specs/` folder holds stable, current subsystem/feature specs after relevant change knowledge is promoted into durable project docs.

Specs should:

- remain stable current-state references
- reference originating change IDs and later revising change IDs where useful
- reference standards by file path rather than copying standard content

Recommended spec traceability metadata:

- `originating_changes`
- `revised_by_changes`

### Decisions

The `decisions/` folder holds durable architecture/policy/ADR-style decisions.

These should be used when a decision needs long-lived discoverability separate from a single change's working files.

Recommended decision traceability metadata:

- `originating_changes`
- `related_changes`

### Traceability And Future Lineage

RuneContext should preserve enough structured traceability to support a future generated lineage/index view, but that lineage view does not need to be implemented in the first version.

Minimum requirement:

- changes, specs, and decisions must carry enough IDs/links that a later tool can reconstruct decision trees and historical relationships across artifacts

## Context Bundle Semantics

This section is normative for v1.

### Required Semantics

- `schema_version` is required.
- `id` is required and must be unique.
- `includes` is required.
- `extends` is optional and ordered.
- `excludes` is optional.

### Resolution Model

Do not model context bundle inheritance as naive set union.

Instead, use ordered rule application.

Recommended algorithm:

1. Load the requested context bundle.
2. Compute the parent bundle order using depth-first, left-to-right traversal of `extends`, with duplicate ancestor bundle IDs removed after their first appearance.
3. Detect and reject inheritance cycles.
4. Build an ordered list of rules from:
   - earlier parents first
   - later parents next
   - child last
5. Within each individual bundle, apply the aspect-aware rules:
    - all `includes` first
    - then all `excludes`
6. Evaluate candidate paths within each aspect family against the full ordered rule list.
7. The last matching rule wins.

This linearization rule is normative for v1 and applies equally to diamond inheritance patterns.

### Consequences Of The Resolution Model

- later parent bundles override earlier parent bundles
- the child bundle overrides all parent bundles
- a child `exclude` can remove something included by a parent
- a child `include` can re-include something excluded by a parent
- repeated rules are harmless and may be normalized in generated output

### Example 1: Parent Excludes, Child Re-Includes

Parent:

```yaml
schema_version: 1
id: base
includes:
  standards:
    - standards/global/**
excludes:
  standards:
    - standards/global/legacy.md
```

Child:

```yaml
schema_version: 1
id: child
extends:
  - base
includes:
  standards:
    - standards/global/legacy.md
```

Result:

- `standards/global/legacy.md` is included because the child's later include wins.

### Example 2: Parent Includes, Child Excludes

Parent:

```yaml
schema_version: 1
id: base
includes:
  standards:
    - standards/security/**
```

Child:

```yaml
schema_version: 1
id: app-lite
extends:
  - base
includes:
  standards:
    - standards/global/**
excludes:
  standards:
    - standards/security/heavy-audit.md
```

Result:

- `standards/security/heavy-audit.md` is excluded because the child's later exclude wins.

### Validation Rules

- cycles in `extends` are invalid
- unknown parent IDs are invalid
- unknown path matches should be allowed during authoring if they are still patterns, but exact paths that do not exist should be warnable
- maximum inheritance depth is `8`
- implementations must reject bundle graphs deeper than `8` rather than choosing implementation-specific limits
- the limit of `8` is a deliberate human-comprehension and performance guardrail and may be revisited after real-world usage data
- during runtime/context-pack generation, exact referenced paths that are missing, unreadable, corrupted, or out of bounds must fail resolution closed
- if a selected file changes between enumeration, snapshot, hashing, or delivery, the pack must be regenerated or rejected rather than silently degraded

## Deterministic Resolved Output

RuneContext tooling should be able to emit a flattened resolved result for any requested context bundle.

Recommended output artifact name:

- `context pack`

The context pack is a generated artifact, not hand-edited source.

Recommended shape:

```yaml
schema_version: 1
canonicalization: rfc8785-jcs
pack_hash_alg: sha256
pack_hash: <hash-of-canonical-pack>
id: go-control-plane
requested_bundle_ids:
  - go-control-plane
resolved_from:
  source_mode: git
  source_ref: refs/tags/v0.1.0-alpha.1
  source_commit: 0123456789abcdef0123456789abcdef01234567
  source_verification: verified_signed_tag
  context_bundle_ids:
    - base
    - go-control-plane
selected:
  project:
    - path: project/mission.md
      sha256: <hash>
      selected_by:
        - bundle: base
          aspect: project
          rule: include
          pattern: project/mission.md
          kind: exact
  standards:
    - path: standards/global/deterministic-check-write-tools.md
      sha256: <hash>
      selected_by:
        - bundle: base
          aspect: standards
          rule: include
          pattern: standards/global/**
          kind: glob
        - bundle: go-control-plane
          aspect: standards
          rule: include
          pattern: standards/global/deterministic-check-write-tools.md
          kind: exact
    - path: standards/security/trust-boundary-interfaces.md
      sha256: <hash>
      selected_by:
        - bundle: go-control-plane
          aspect: standards
          rule: include
          pattern: standards/security/trust-boundary-interfaces.md
          kind: exact
  decisions:
    - path: decisions/DEC-0001-trust-boundary-model.md
      sha256: <hash>
      selected_by:
        - bundle: go-control-plane
          aspect: decisions
          rule: include
          pattern: decisions/DEC-0001-trust-boundary-model.md
          kind: exact
excluded:
  standards:
    - path: standards/frontend/example.md
      last_rule:
        bundle: go-control-plane
        aspect: standards
        rule: exclude
        pattern: standards/frontend/example.md
        kind: exact
generated_at: 2026-03-15T00:00:00Z
```

Minimum requirements for the resolved output:

- final selected context inputs by aspect family
- hashes of selected files
- enough provenance to understand why a file was included or excluded
- stable ordering
- a top-level pack hash over the canonicalized resolved pack
- the resolved source commit SHA and source verification posture

Canonicalization and integrity rules:

- the context pack must have a canonical serialization for hashing; RuneContext should use RFC 8785 JCS over the normalized JSON representation of the pack
- `pack_hash` must be computed over the canonicalized pack payload
- `generated_at` should remain a required emitted field for auditability, but it should stay outside the canonical `pack_hash` input so identical resolved content hashes the same across regenerations
- RuneCode should bind and/or sign `pack_hash` in audit/provenance flows

Request-identity rule:

- the normal authored workflow should still prefer one top-level bundle or an authored composite bundle
- when a caller supplies more than one top-level bundle, the generated pack should preserve the ordered request separately from the resolved bundle linearization so tool/runtime consumers do not need a schema refactor later

Size and portability rules:

- tooling should warn when a context pack exceeds advisory size thresholds
- recommended advisory defaults are `256` selected files and `1 MiB` total referenced content bytes
- tooling should also warn when provenance metadata grows beyond advisory thresholds such as `32` `selected_by` entries per selected item or `256 KiB` total provenance metadata
- RuneCode integration may enforce stricter hard limits for model-facing runs

RuneCode default escalation path:

- RuneCode should treat `256` selected files, `1 MiB` referenced content bytes, and `256 KiB` provenance metadata as default hard ceilings for direct model-facing context injection
- if those limits are exceeded, RuneCode should refuse direct model-facing injection and require bundle narrowing or generation of a derived summarized artifact rather than silently truncating the pack

Provenance compaction rule:

- the context pack should carry compact deterministic provenance sufficient for explanation and verification
- if fuller provenance would exceed pack metadata limits, implementations should keep the pack compact and place fuller provenance in a separate receipt when Verified mode is enabled

RuneCode should be able to consume this deterministically.

For RuneCode-operated isolated roles, the generated context pack should be delivered as hash-addressed artifacts plus a typed descriptor and must be re-verified inside the isolate before use.

## Standards Membership And Authoring Model

Standards must remain easy to edit manually.

Authoritative rule:

- context bundle membership is defined by `bundles/*.yaml`
- not by standard frontmatter alone

This keeps one clear source of truth for inclusion/exclusion.

### Manual Authoring Flow

Users should always be able to:

- add a new standard markdown file
- edit or remove a standard markdown file
- edit context bundle yaml files directly to include or exclude that standard

### RuneCode-Assisted Flow

RuneCode should be able to:

- draft a new standard
- update or deprecate an existing standard
- propose matching `bundles/*.yaml` edits in the same diff
- include change/spec updates that reference the new standard

Approval model:

- local workflow: RuneCode may present an inline approval for the draft diff
- protected/remote workflow: RuneCode may open a branch/PR
- RuneCode must not silently mutate the source of truth outside approved workflow

### Suggested Context Bundles Metadata

`suggested_context_bundles` in standard frontmatter may be used to help tooling propose context-bundle edits, but it must remain advisory.

Do not implement automatic tag-query-based bundle membership in v1.

Reason:

- explicit bundle files are easier to audit, reason about, and review
- auto-membership adds hidden behavior too early

## Change Lifecycle

RuneContext should adopt a lightweight change lifecycle.

Recommended lifecycle states:

- `proposed`
- `planned`
- `implemented`
- `verified`
- `closed`
- `superseded`

Guidance:

- `proposed`
  - change exists, intent captured
- `planned`
  - design/tasks are shaped enough for execution
- `implemented`
  - implementation landed or diff exists
- `verified`
  - verification is complete and `verification_status` should normally be `passed` or an explicitly accepted exception state such as `skipped`
- `closed`
  - change is complete and any durable knowledge has been promoted
- `superseded`
  - replaced by later change(s)

### Promotion

Promotion means selectively moving durable knowledge learned during a change into long-lived project knowledge surfaces such as:

- `specs/`
- `standards/`
- `decisions/`

Promotion exists because not every detail in a change should become permanent project truth.

Examples of change-local details that should often remain only in the change folder:

- temporary rollout details
- implementation exploration that did not become durable architecture
- one-off debugging notes

Important rule:

- preserve the full historical change folder
- only promote durable truth

Why RuneContext should not auto-promote everything:

- stable docs become noisy and harder to read
- transient implementation notes get mistaken for permanent truth
- maintenance burden rises quickly

### Promotion Assessment

RuneContext Core may support lightweight promotion suggestion behavior, but it should not silently auto-promote.

RuneCode should automatically assess durable knowledge as part of its workflow and propose promotion updates when appropriate.

Recommended behavior:

- on `change shape`
  - optional low-confidence promotion hints are allowed
- on `change close`
  - a promotion assessment should always run
  - if no durable knowledge should be promoted, record that explicitly
  - if durable knowledge should be promoted, propose target updates in `specs/`, `standards/`, and/or `decisions/`
  - the close-time status should settle to `none` or `suggested`; later explicit
    promotion workflows may advance reviewable promotions to `accepted` and
    `completed`

This keeps the burden low while still preserving important project knowledge.

Promotion assessment output should be structured.

Minimum suggested-target fields:

- `target_type`
- `target_path`
- `summary`

### Archive/Promotion Rule

Do not physically archive closed changes into a separate hard-to-browse tree in v1.

Instead:

- keep the change folder at its stable path
- update `status.yaml`
- promote durable knowledge into `specs/`, `standards/`, and/or `decisions/` as needed
- let generated indexes surface active vs closed vs superseded views

This preserves stable paths and makes history easier to audit.

### Historical Traceability Requirements

Historical information must be easy to find.

Minimum expectations:

- every change has a stable ID
- specs and decisions can reference originating change IDs
- future changes can reference older changes that informed or constrained them
- generated indexes can filter by status
- closed changes remain directly readable at their original paths
- no information required for audit should be hidden behind relocation conventions alone
- enough structured references should exist that a lineage/index view can be generated later, even if RuneContext does not ship that view in v1

## Minimal Process And User Experience

RuneContext should keep the visible process small.

The main mental model should be four nouns:

- `project`
- `standards`
- `bundles`
- `changes`

The main lifecycle should be four verbs:

- `propose`
- `shape`
- `implement`
- `close`

The broader underlying workflow can still cover:

- discover standards
- maintain project context
- select context bundle
- propose change
- shape/design change
- generate tasks
- implement
- verify
- close

But users should not have to learn a large command surface just to benefit from the system.

### Progressive Disclosure

RuneContext should probe deeply enough to get at the heart of a new project, feature, or bug, but it should not force unnecessary process.

Recommended model:

- every substantive work item gets a change ID
- every change starts in minimum mode
- full-mode files are only materialized when the work actually needs them

### Branching Logic By Work Type And Size

#### New Project

New project work should almost always use deeper intake because bad defaults compound.

Recommended intake topics:

- mission and target users
- stack/runtime constraints
- deployment and security constraints
- success criteria
- non-goals

#### New Feature

- `small`
  - minimum mode is usually enough
- `medium`
  - minimum mode is often enough if scope is clear; otherwise move to full mode
- `large` or `high-risk`
  - full mode is usually appropriate

#### Bug

- `simple/localized`
  - minimum mode is usually enough, but verification remains mandatory
- `unclear root cause`, `behavioral ambiguity`, or `security/schema/API impact`
  - shape the change and move to full mode

#### Standard Or Process Change

- often small in scope
- may still require explicit standards and references handling because it affects future work broadly

### When To Ask More Vs Less

Ask more questions when the answer materially changes:

- user-facing behavior
- API or interface shape
- migrations or rollout
- trust boundaries or risk profile
- verification and acceptance criteria

Infer defaults when the repository already makes the answer obvious, such as:

- naming and placement conventions
- likely context bundle
- standard verification commands
- related standard families

If the system infers a meaningful assumption, it should record that assumption in `proposal.md` so it remains reviewable.

## Invocation Surfaces And Command Architecture

RuneContext's canonical surface should be:

- the on-disk model
- schemas
- resolution semantics
- operation contracts

The CLI should be important but not canonical.

### Primary UX

Users should primarily be encouraged to use their coding tool's RuneContext command pack.

Examples:

- Claude Code adapter commands
- OpenCode adapter commands
- Codex adapter skills/commands

The user-facing message should be:

- use RuneContext through your tool's command pack for normal day-to-day work
- use the `runectx` CLI for automation, power use, debugging, and server workflows

### Why The CLI Still Matters

The CLI remains valuable for:

- automation and CI
- server-side execution
- non-agent workflows
- adapter debugging
- parity testing between adapters and the core model
- providing a universal machine-facing interface

### RuneCode Integration Surface

RuneCode should not depend only on shelling out to the CLI.

Recommended approach:

- RuneCode uses a direct resolver/library/API when available
- parity between the direct integration and CLI behavior should be tested explicitly

### Recommended Minimal CLI Surface

If a standalone CLI is built, keep the front door small.

Primary commands:

- `runectx init`
- `runectx status`
- `runectx change new`
- `runectx change shape`
- `runectx bundle resolve`
- `runectx change close`

Secondary/admin commands:

- `runectx validate`
- `runectx doctor`
- `runectx standard discover`
- `runectx promote`
- `runectx assurance enable verified`
- `runectx assurance backfill`

Adapters may expose friendlier tool-native commands, but they should map back to this small set of underlying operations.

### Command Semantics

#### `runectx init`

- scaffold RuneContext in embedded or linked mode
- create the root `runecontext.yaml` with project version and assurance tier
- create base folders and starter files
- optionally seed a base context bundle and project files

#### `runectx status`

- summarize current RuneContext state
- list active/closed/superseded changes
- list available context bundles
- report the active `runecontext_version` and `assurance_tier`
- report stale generated artifacts or missing required files

#### `runectx change new`

- create a unique change ID and folder
- initialize `status.yaml`, `proposal.md`, and `standards.md`
- infer likely context bundle(s)
- resolve and populate initial standards

#### `runectx change shape`

- deepen a change when needed
- materialize/update the full-mode files when the change warrants them:
  - `design.md`
  - `tasks.md`
  - `references.md`
  - `verification.md`
- refresh `standards.md`
- record assumptions when operating non-interactively

#### `runectx bundle resolve`

- resolve one or more context bundles deterministically
- emit a human-readable or JSON context pack
- include the canonical pack hash and resolved source revision in the output
- explain why each standard was included or excluded when requested

#### `runectx change close`

- validate close readiness
- record verification state
- run promotion assessment
- propose promotion targets where needed
- update lifecycle status without burying the change in an archive tree

#### `runectx validate`

- validate schemas, references, bundle resolution, required artifact completeness, proposal structure, ID uniqueness, bidirectional supersession consistency, status invariants, and source-integrity posture warnings

#### `runectx doctor`

- diagnose install/setup problems
- report adapter sync status, missing files, unsupported version combinations, and source-integrity posture issues

#### `runectx standard discover`

- inspect project materials and propose candidate standards additions or updates
- remain advisory and reviewable

#### `runectx promote`

- explicitly promote durable knowledge from a change into `specs/`, `standards/`, and/or `decisions/`
- may also be used as a sub-step of `change close`

#### `runectx assurance enable verified`

- enable Verified mode for a project
- update `runecontext.yaml` to persist `assurance_tier: verified`
- generate the initial assurance baseline
- record the adoption commit and initial verification posture

#### `runectx assurance backfill`

- inspect git history and existing RuneContext artifacts
- generate imported/backfilled provenance records where possible
- clearly mark imported records as distinct from natively captured Verified evidence

### Universal Machine-Facing Flags

Every machine-facing command should support:

- `--json`
  - structured machine-readable output
- `--non-interactive`
  - never prompt; infer defaults or fail clearly
- `--dry-run`
  - show what would happen without writing changes
- `--explain`
  - provide reasoning/provenance for decisions such as bundle resolution, standards selection, or promotion suggestions

## Adapters

RuneContext adapters should be thin.

Their job is to make RuneContext usable inside specific tools without changing RuneContext's source format.

They should be the primary end-user UX for normal daily use.

Recommended initial adapters:

- `generic`
  - plain markdown command docs usable with any agent/tool
- `claude-code`
  - command/skill wrappers for Claude Code
- `opencode`
  - command wrappers for OpenCode
- `codex`
  - skill/prompt wrappers for Codex-style tools

Adapter responsibilities:

- explain the RuneContext workflow in the host tool's native style
- point the agent to source files and schemas
- provide consistent command naming where useful
- derive from or stay aligned with the canonical human-readable command reference in `runecontext/commands/` when that directory is present

Adapter non-responsibilities:

- storing project truth in adapter-only files
- redefining bundle resolution
- changing lifecycle semantics

### Adapter Capability Model

Adapters should all target the same underlying RuneContext operations, but they may differ in UX depending on host capabilities.

Relevant capability dimensions include:

- interactive prompting
- structured output support
- shell access
- file editing support
- background task/subagent support
- approval hooks

Adapters should support:

- full mode when the host tool supports richer interactions
- compatibility mode when the host tool is more constrained

Important rule:

- adapters may differ in UX, but not in core semantics, file formats, or lifecycle rules

## RuneCode Integration Details

RuneCode should be the best runtime for RuneContext, not the only runtime.

### Required RuneCode Capabilities

RuneCode integration should provide:

- deterministic context bundle resolution
- context pack generation
- context pack hashing/signing or other audit binding
- typed artifact-based delivery of context packs into isolates with in-isolate re-verification
- approval-aware drafting of standards, context bundles, changes, and spec updates
- local/remote parity in RuneContext resolution
- binding the active change ID and `proposal.md` intent artifact into auditable run history
- automatic promotion assessment with reviewable promotion proposals
- enforcement that normal RuneCode-operated workflows run only against Verified-mode RuneContext projects

### Context Pack Delivery Into Isolates

RuneCode must deliver RuneContext into isolated workspace roles in a way that matches its artifact-based trust model.

Required runtime flow:

1. a trusted RuneCode component resolves the context bundle(s), active change, and project inputs into a canonical context pack
2. the trusted side snapshots the selected RuneContext files into hash-addressed artifacts
3. the trusted side emits a typed descriptor (via broker/local API message or equivalent typed runtime object) that includes:
   - artifact identifiers for the pack and/or referenced content
   - the expected top-level `pack_hash`
   - the expected source revision and per-file hashes
4. the isolate receives the descriptor through RuneCode's typed cross-boundary transport, not through host filesystem mounts
5. inside the isolate, the receiving side re-verifies the top-level `pack_hash` and the referenced file hashes before making the content available to the untrusted workflow runner

Important rules:

- RuneContext content must cross the trusted/untrusted boundary via typed broker/runtime transport and hash-addressed artifacts
- host filesystem mounts are not a valid RuneContext delivery path for isolated roles
- any missing artifact, broker mismatch, pack-hash mismatch, or per-file hash mismatch must fail closed

### Recommended RuneCode-Specific Enhancements

These belong in RuneCode integration, not RuneContext Core:

- trust-boundary tags
  - RuneCode may interpret standard tags like `trusted`, `untrusted`, or `cross-boundary`
- evidence binding
  - RuneCode may attach test, audit, approval, or execution evidence to change verification flows
- run-mode-aware context assembly
  - RuneCode may combine context bundles, the active change, and project knowledge into different context packs depending on task type
- audit/provenance integration
  - RuneCode may store the source ref and hashes used to resolve the pack

### Reviewable Intent In RuneCode History

When RuneCode uses RuneContext, it should surface a compact reviewable intent view in auditable history based on the active change.

That history should be able to show:

- the change ID
- what is changing
- why it is changing
- which standards were in scope
- which assumptions were recorded

Future decisions and later changes should be able to reference earlier change IDs so that audit history can walk back through the relevant decision tree.

### Policy Neutrality Rule

RuneContext itself must not grant permissions or capabilities.

Markdown/yaml in RuneContext is guidance and project knowledge. RuneCode's runtime policy, trust, approval, and capability enforcement must remain separate and authoritative.

Normative enforcement rule:

- RuneContext content may influence context assembly, presentation, summarization, and review suggestions
- RuneContext content must not directly influence policy-engine allow/deny results
- RuneContext content must not directly choose approval profiles or widen capabilities
- any trusted-side mapping from RuneContext metadata to stricter review posture must be explicit, allowlisted, and auditable

### LLM Input Trust Boundary

RuneContext content is also untrusted model input.

Important rules:

- linked or compromised RuneContext content may contain prompt-injection attempts disguised as standards, decisions, or change guidance
- RuneCode must treat RuneContext text as untrusted input for LLM behavior, just as it treats it as untrusted input for policy behavior
- RuneCode's protection against RuneContext-based prompt injection should rely on typed policy, broker/runtime boundaries, approval gates, and isolate constraints rather than trusting the text itself
- future implementations may add prompt-hygiene scanning or content-safety heuristics, but those are supplementary defenses rather than the primary trust boundary

## Generated Indexes And Manifests

RuneContext should support generated indexes/manifests to improve browsing and auditing, but these should not replace the source files as the true source of truth.

Recommended generated outputs:

- overall `manifest.yaml`
  - inventory of standards, bundles, changes, specs, decisions
- optional change index
  - `indexes/changes-by-status.yaml`
- optional bundle inventory
  - `indexes/bundles.yaml` with resolved parents and referenced patterns

Deferred for later:

- a richer lineage/index view connecting changes, promoted specs, decisions, and standards across history

RuneContext should preserve enough traceability fields now that such a lineage/index view can be generated later for historical artifacts without retrofitting old changes.

Important rule:

- no manual-only index should become the only authoritative inventory that must be edited by hand for correctness
- generated indexes/manifests should use stable ordering and merge-friendly formatting to reduce multi-branch conflicts
- `manifest.yaml` should be treated as optional and regenerable; implementations should not require it to be committed for correctness
- generated index artifacts should use closed schemas so RuneCode and other
  tooling can validate them without treating them as source-of-truth files

## Standards Referencing Rule

This should be a hard design rule:

- change docs and stable specs should reference standards by path
- they should not copy entire standards into the spec/change body except for short quoted excerpts when necessary

Reason:

- standards remain reusable
- standard updates do not force churn across many change/spec files
- audit/review can clearly see which standards were intended to apply

## Recommended Implementation Plan

Implement RuneContext in phases.

### Phase 1: RuneContext Core

Build:

- core folder layout and file conventions
- schemas for standards, context bundles, change status, and context pack
- strict `proposal.md` and `standards.md` structures
- minimum and full change shape semantics
- embedded and linked repo resolution
- context bundle resolver with deterministic ordering
- immutable linked-source handling and mutable-ref warnings
- context-pack canonicalization and top-level hashing
- Plain and Verified assurance-tier model
- optional generated assurance artifact schemas and baseline format
- automatic standards maintenance for change creation/shaping
- promotion assessment semantics and traceability fields
- generated manifest/index support

Acceptance criteria:

- a project can use embedded RuneContext
- a project can point to a linked RuneContext repo via `runecontext.yaml`
- project-root `runecontext.yaml` carries `runecontext_version` and `assurance_tier`
- context bundles resolve deterministically with the required precedence rules
- linked git sources record resolved commit SHAs and mutable refs warn clearly
- context packs include top-level canonical hashes
- Plain mode works without generating extra assurance artifacts
- Verified mode can generate baseline and receipt artifacts when enabled
- changes have stable IDs and status lifecycle
- minimum mode works with `status.yaml`, `proposal.md`, and `standards.md`
- full mode works by materializing deeper files only when needed
- closed changes remain directly accessible at their original paths
- enough traceability is captured to support future lineage/index generation

### Phase 2: RuneContext Adapters And Minimal CLI

Build:

- generic adapter docs
- at least one agent-specific adapter
- adapter-first command packs as the primary end-user UX
- minimal CLI for init, status, change creation/shaping/closing, standards discovery, promotion, and context resolution
- CLI support for enabling Verified mode and automated backfill

Acceptance criteria:

- a non-RuneCode user can create and maintain RuneContext artifacts manually or with the CLI
- adapters do not introduce tool-specific source-of-truth files
- adapters map to the same underlying RuneContext operations across host tools
- the same project artifacts remain readable and useful without the adapter installed
- enabling Verified mode is optional for standalone users

### Phase 3: RuneCode Integration

Build:

- RuneCode resolver for embedded and linked RuneContext
- deterministic context pack assembly
- audit/provenance binding of resolved context packs
- approval-aware flows for adding/updating/removing standards and context-bundle membership
- binding of `proposal.md`/change intent into auditable run history
- automatic promotion assessment and reviewable promotion proposals
- migration flow from Plain mode to Verified mode with baseline generation/backfill

Acceptance criteria:

- local RuneCode and remote RuneCode produce the same resolved context pack from the same inputs
- RuneCode can draft standards and corresponding bundle edits in a reviewable diff
- RuneCode can operate against a dedicated RuneContext repo or embedded RuneContext content
- RuneCode can surface reviewable intent artifacts in audit history
- RuneCode binds the canonical context-pack hash into audit/provenance flows
- RuneCode requires Verified assurance mode for normal audited workflows
- RuneCode does not treat RuneContext markdown as authoritative runtime policy

## Explicit Design Decisions From The Discussion

This section captures the most important decisions verbatim in implementation terms.

### Naming

- The reusable context-selection unit is called a `context bundle`.
- The folder name is `bundles/`.
- `profiles` should not be the primary term because it collides with RuneCode approval-profile language.

### Context Bundle Fields

- `schema_version` is required.
- `id` is required.
- `includes` is required.
- `extends` is optional.
- `excludes` is optional.
- `includes` and `excludes` are aspect-aware maps that can target multiple RuneContext aspect families in one bundle.

### Context Bundle Merge/Inheritance Behavior

When a context bundle extends another context bundle:

- everything should be merged through ordered rule application
- `includes` and `excludes` should both be honored
- aspect families are resolved independently under the same ordered precedence model
- parent ordering uses depth-first, left-to-right traversal with duplicate ancestors removed after first appearance
- the child's values take precedence over the parent's values
- later parents take precedence over earlier parents
- the final evaluation rule is "last matching rule wins"

### Linked Sources And Integrity

- linked RuneContext repositories should prefer immutable commit SHAs or verifiable signed tags
- mutable refs must require explicit opt-in and warning
- local path sources are developer-local only and should be treated as unverified/non-auditable
- the resolved source commit must be recorded in the context pack

### Project Root Configuration

- the project-root `runecontext.yaml` should carry `runecontext_version` and `assurance_tier`
- RuneCode should use `runecontext_version` for compatibility checks

### Change IDs

- change IDs should use the form `CHG-YYYY-NNN-RAND-short-slug`
- the short random suffix reduces concurrent-branch collision risk while keeping IDs readable

### Context Pack Integrity

- context packs must include a top-level canonical hash
- RuneCode should bind and/or sign that hash in audit/provenance flows

### Isolate Delivery

- RuneCode should deliver context packs into isolates through typed transport plus hash-addressed artifacts
- the isolate should re-verify the delivered pack before use

### Policy Neutrality

- RuneContext content may guide context assembly and presentation
- RuneContext content must not directly influence policy-engine allow/deny results
- RuneContext content must not directly choose approval profiles or widen capabilities

### Assurance Tiers

- RuneContext starts with `Plain` and `Verified` tiers
- `Plain` is the default lightweight standalone mode
- `Verified` enables generated assurance artifacts and stronger verifiable tracing
- RuneCode-operated normal audited workflows require `Verified`
- additional assurance files are only generated when `Verified` is enabled
- a future `Anchored` tier may be added later when RuneCode's anchoring capabilities are ready

### Standards Updates

- users must be able to edit standards and context bundles manually
- RuneCode must be able to draft new standards or remove/update existing standards
- RuneCode must be able to automatically propose corresponding bundle membership edits
- final mutation of the source of truth should happen through an approved flow, such as inline approval or PR acceptance
- auto-maintained files like `standards.md` must always surface reviewable diffs before commit/merge

### Minimum And Full Change Shapes

- every substantive work item gets a change ID
- minimum mode is:
  - `status.yaml`
  - `proposal.md`
  - `standards.md`
- full mode adds:
  - `design.md`
  - `tasks.md`
  - `references.md`
  - `verification.md`
- the full artifact set is not the default requirement for all work

### Proposal.md

- `proposal.md` is the canonical reviewable intent artifact
- it must use a strict structure with:
  - `Summary`
  - `Problem`
  - `Proposed Change`
  - `Why Now`
  - `Assumptions`
  - `Out of Scope`
  - `Impact`

### Standards.md

- `standards.md` must always be present
- it should be automatically maintained by tooling
- users should not be required to manually keep it updated for normal operation

### History Preservation

- historical spec/change information should not be forgotten, lost, or buried
- closing a change should not require physically moving it into a remote archive tree in v1
- historical information must remain easy for humans and RuneCode to find and audit

### Traceability

- future changes, specs, and decisions should be able to reference older change IDs
- RuneContext must preserve enough traceability to support future generated lineage/index views
- the first version does not need to implement that lineage/index view yet

### Promotion

- promotion means selectively moving durable knowledge into `specs/`, `standards/`, and/or `decisions/`
- RuneCode should automatically assess and propose promotion when appropriate
- RuneContext Core may support promotion suggestions, but should not silently auto-promote all change details

### Storage Flexibility

- end users must be able to keep RuneContext inside a project's repo
- end users must also be able to keep RuneContext in its own dedicated repo
- RuneCode must be able to consume either form
- non-RuneCode users must be able to use either form

### Invocation Surface

- tool-specific command packs should be the primary user-facing UX
- the CLI should exist for power users, automation, debugging, and parity testing
- the on-disk model, schemas, and operation semantics remain the canonical surface

## Final Recommendation

Build RuneContext as a portable project-knowledge/spec system with:

- Agent OS-style markdown standards and low-ceremony shaping
- OpenSpec-style change lifecycle and status tracking
- explicit context bundles with deterministic inheritance
- immutable or verifiable linked sources plus hashed context packs
- optional assurance tiers with Verified mode for stronger provenance when needed
- stable-path history preservation
- optional adapters for multiple tools
- deep RuneCode integration through resolved, auditable context packs

The design center is:

- portable by default
- deterministic when resolved
- auditable in RuneCode
- still fully usable without RuneCode
