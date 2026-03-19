package contracts

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateChangeMinimum(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	defer loaded.Close()
	result, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Add cache invalidation",
		Type:           "feature",
		Size:           "small",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xaa, 0xbb}),
	})
	if err != nil {
		t.Fatalf("create change: %v", err)
	}
	if got, want := result.Mode, ChangeModeMinimum; got != want {
		t.Fatalf("expected mode %q, got %q", want, got)
	}
	changeDir := filepath.Join(root, "runecontext", "changes", result.ID)
	requireFileContent(t, filepath.Join(changeDir, "status.yaml"), strings.Join([]string{
		"schema_version: 1",
		"id: CHG-2026-001-aabb-add-cache-invalidation",
		"title: Add cache invalidation",
		"status: proposed",
		"type: feature",
		"size: small",
		"verification_status: pending",
		"context_bundles:",
		"  - base",
		"related_specs: []",
		"related_decisions: []",
		"related_changes: []",
		"depends_on: []",
		"informed_by: []",
		"supersedes: []",
		"superseded_by: []",
		"created_at: \"2026-03-18\"",
		"closed_at: null",
		"promotion_assessment:",
		"  status: pending",
		"  suggested_targets: []",
		"",
	}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "proposal.md"), strings.Join([]string{
		"## Summary",
		"Add cache invalidation",
		"",
		"## Problem",
		"The repository needs a reviewable RuneContext change record for this work.",
		"",
		"## Proposed Change",
		"Track Add cache invalidation through the minimum RuneContext change artifacts.",
		"",
		"## Why Now",
		"The work needs stable intent, standards linkage, and verification planning before it moves further.",
		"",
		"## Assumptions",
		"- Inferred `just test` from the repository's justfile test target.",
		"",
		"## Out of Scope",
		"Work outside the scoped change tracked here.",
		"",
		"## Impact",
		"The change keeps intent, assumptions, and standards linkage reviewable.",
		"",
	}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "standards.md"), strings.Join([]string{
		"## Applicable Standards",
		"- `standards/global/base.md`: Selected from the current context bundles.",
		"",
		"## Resolution Notes",
		"Generated from the current context bundle selection; review any automatic refresh before committing.",
		"",
	}, "\n"))
	if _, err := os.Stat(filepath.Join(changeDir, "design.md")); !os.IsNotExist(err) {
		t.Fatalf("expected minimum mode to skip design.md, got err=%v", err)
	}
	validated, err := v.ValidateProject(root)
	if err != nil {
		t.Fatalf("validate generated project: %v", err)
	}
	_ = validated.Close()
}

func TestCreateProjectChangeAutoShapes(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	defer loaded.Close()
	result, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Launch payments platform",
		Type:           "project",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xaa, 0xbb}),
	})
	if err != nil {
		t.Fatalf("create project change: %v", err)
	}
	if got, want := result.Mode, ChangeModeFull; got != want {
		t.Fatalf("expected mode %q, got %q", want, got)
	}
	changeDir := filepath.Join(root, "runecontext", "changes", result.ID)
	requireFileContent(t, filepath.Join(changeDir, "design.md"), strings.Join([]string{
		"# Design",
		"",
		"## Overview",
		"Shape Launch payments platform before implementation so scope, standards linkage, and verification stay reviewable.",
		"",
		"## Shape Rationale",
		"- Project work uses deeper intake because bad defaults compound.",
		"",
		"## Project Intake Checklist",
		"- Mission and target users.",
		"- Stack and runtime constraints.",
		"- Deployment and security constraints.",
		"- Success criteria.",
		"- Non-goals.",
		"",
		"## Ask More When",
		"- Mission and target users.",
		"- Stack and runtime constraints.",
		"- Deployment and security constraints.",
		"- Success criteria.",
		"- Non-goals.",
		"",
	}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "verification.md"), strings.Join([]string{
		"# Verification",
		"",
		"## Planned Checks",
		"- `just test`",
		"",
		"## Close Gate",
		"Use the repository's standard verification flow before closing this change.",
		"",
	}, "\n"))
}

func TestShapeChangeCreatesSupplementalDocsAndRefreshesStandards(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	result, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Add cache invalidation",
		Type:           "feature",
		Size:           "small",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xaa, 0xbb}),
	})
	loaded.Close()
	if err != nil {
		t.Fatalf("create change: %v", err)
	}
	statusPath := filepath.Join(root, "runecontext", "changes", result.ID, "status.yaml")
	rewriteFile(t, statusPath, func(text string) string {
		return strings.Replace(text, "- base", "- security", 1)
	})
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	defer loaded.Close()
	shapeResult, err := ShapeChange(v, loaded, result.ID, ChangeShapeOptions{
		Tasks:      []string{"Implement cache invalidation flow.", "Add regression coverage."},
		References: []string{"docs/cache.md", "issue-42"},
	})
	if err != nil {
		t.Fatalf("shape change: %v", err)
	}
	if got, want := shapeResult.StandardsRefreshAction, "updated"; got != want {
		t.Fatalf("expected standards refresh %q, got %q", want, got)
	}
	if !shapeResult.ReviewDiffRequired {
		t.Fatalf("expected updated standards refresh to require review diff")
	}
	changeDir := filepath.Join(root, "runecontext", "changes", result.ID)
	requireFileContent(t, filepath.Join(changeDir, "design.md"), strings.Join([]string{
		"# Design",
		"",
		"## Overview",
		"Shape Add cache invalidation before implementation so scope, standards linkage, and verification stay reviewable.",
		"",
		"## Shape Rationale",
		"- Full mode was requested explicitly to deepen the change.",
		"- Minimum mode is sufficient for the current size and risk signal.",
		"",
	}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "verification.md"), strings.Join([]string{
		"# Verification",
		"",
		"## Planned Checks",
		"- `just test`",
		"",
		"## Close Gate",
		"Use the repository's standard verification flow before closing this change.",
		"",
	}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "tasks.md"), strings.Join([]string{
		"# Tasks",
		"",
		"- Implement cache invalidation flow.",
		"- Add regression coverage.",
		"",
	}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "references.md"), strings.Join([]string{
		"# References",
		"",
		"- docs/cache.md",
		"- issue-42",
		"",
	}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "standards.md"), strings.Join([]string{
		"## Applicable Standards",
		"- `standards/security/review.md`: Selected from the current context bundles.",
		"",
		"## Standards Added Since Last Refresh",
		"- `standards/security/review.md`: Newly selected during standards refresh.",
		"",
		"## Resolution Notes",
		"Generated from the current context bundle selection; review any automatic refresh before committing.",
		"",
	}, "\n"))
}

func TestShapeChangeIsIdempotentAndLeavesStandardsUnchanged(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	result, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Add cache invalidation",
		Type:           "feature",
		Size:           "small",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xaa, 0xbb}),
	})
	loaded.Close()
	if err != nil {
		t.Fatalf("create change: %v", err)
	}
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	firstShape, err := ShapeChange(v, loaded, result.ID, ChangeShapeOptions{})
	loaded.Close()
	if err != nil {
		t.Fatalf("first shape change: %v", err)
	}
	if got, want := firstShape.StandardsRefreshAction, "unchanged"; got != want {
		t.Fatalf("expected first standards refresh %q, got %q", want, got)
	}
	if firstShape.ReviewDiffRequired {
		t.Fatalf("expected unchanged standards refresh to avoid review_diff_required on first shape")
	}
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project again: %v", err)
	}
	defer loaded.Close()
	secondShape, err := ShapeChange(v, loaded, result.ID, ChangeShapeOptions{})
	if err != nil {
		t.Fatalf("second shape change: %v", err)
	}
	if got, want := secondShape.StandardsRefreshAction, "unchanged"; got != want {
		t.Fatalf("expected second standards refresh %q, got %q", want, got)
	}
	if secondShape.ReviewDiffRequired {
		t.Fatalf("expected idempotent shape to avoid review diff requirement")
	}
	if len(secondShape.ChangedFiles) != 0 {
		t.Fatalf("expected idempotent shape to leave files unchanged, got %#v", secondShape.ChangedFiles)
	}
}

func TestShapeChangeRejectsTerminalChange(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	result, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Add cache invalidation",
		Type:           "feature",
		Size:           "small",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xaa, 0xbb}),
	})
	loaded.Close()
	if err != nil {
		t.Fatalf("create change: %v", err)
	}
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	if _, err := CloseChange(v, loaded, result.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)}); err != nil {
		loaded.Close()
		t.Fatalf("close change: %v", err)
	}
	loaded.Close()
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload closed project: %v", err)
	}
	defer loaded.Close()
	_, err = ShapeChange(v, loaded, result.ID, ChangeShapeOptions{})
	if err == nil || !strings.Contains(err.Error(), "terminal status") {
		t.Fatalf("expected terminal-shape rejection, got %v", err)
	}
}

func TestCloseChangeWritesClosedStatus(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	result, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Add cache invalidation",
		Type:           "feature",
		Size:           "small",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xaa, 0xbb}),
	})
	loaded.Close()
	if err != nil {
		t.Fatalf("create change: %v", err)
	}
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	defer loaded.Close()
	closeResult, err := CloseChange(v, loaded, result.ID, ChangeCloseOptions{
		VerificationStatus: "passed",
		ClosedAt:           time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("close change: %v", err)
	}
	if len(closeResult.ContextBundles) != 1 || closeResult.ContextBundles[0] != "base" {
		t.Fatalf("expected close result to retain context bundles, got %#v", closeResult.ContextBundles)
	}
	if len(closeResult.ApplicableStandards) != 1 || closeResult.ApplicableStandards[0] != "standards/global/base.md" {
		t.Fatalf("expected close result to retain applicable standards, got %#v", closeResult.ApplicableStandards)
	}
	requireFileContent(t, filepath.Join(root, "runecontext", "changes", result.ID, "status.yaml"), strings.Join([]string{
		"schema_version: 1",
		"id: CHG-2026-001-aabb-add-cache-invalidation",
		"title: Add cache invalidation",
		"status: closed",
		"type: feature",
		"size: small",
		"verification_status: passed",
		"context_bundles:",
		"  - base",
		"related_specs: []",
		"related_decisions: []",
		"related_changes: []",
		"depends_on: []",
		"informed_by: []",
		"supersedes: []",
		"superseded_by: []",
		"created_at: \"2026-03-18\"",
		"closed_at: \"2026-03-20\"",
		"promotion_assessment:",
		"  status: pending",
		"  suggested_targets: []",
		"",
	}, "\n"))
}

func TestCloseChangeWritesSupersededStatusAndReciprocalLink(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	defer loaded.Close()
	if _, err := CloseChange(v, loaded, "CHG-2026-001-a3f2-auth-gateway", ChangeCloseOptions{
		VerificationStatus: "skipped",
		ClosedAt:           time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		SupersededBy:       []string{"CHG-2026-002-b4c3-auth-revision"},
	}); err != nil {
		t.Fatalf("supersede change: %v", err)
	}
	statusPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read superseded status: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "status: superseded") || !strings.Contains(text, "closed_at: \"2026-03-18\"") {
		t.Fatalf("unexpected superseded status contents:\n%s", text)
	}
	successorData, err := os.ReadFile(filepath.Join(root, "runecontext", "changes", "CHG-2026-002-b4c3-auth-revision", "status.yaml"))
	if err != nil {
		t.Fatalf("read successor status: %v", err)
	}
	if !strings.Contains(string(successorData), "supersedes:\n  - CHG-2026-001-a3f2-auth-gateway") {
		t.Fatalf("expected reciprocal supersedes link, got:\n%s", string(successorData))
	}
}

func TestCloseChangeRejectsTerminalSuccessorWithoutReciprocalLink(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	first, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Add cache invalidation",
		Type:           "feature",
		Size:           "small",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xaa, 0xbb}),
	})
	if err != nil {
		loaded.Close()
		t.Fatalf("create first change: %v", err)
	}
	second, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Revise cache invalidation",
		Type:           "feature",
		Size:           "small",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xcc, 0xdd}),
	})
	loaded.Close()
	if err != nil {
		t.Fatalf("create second change: %v", err)
	}
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	if _, err := CloseChange(v, loaded, second.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 19, 0, 0, 0, 0, time.UTC)}); err != nil {
		loaded.Close()
		t.Fatalf("close successor: %v", err)
	}
	loaded.Close()
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project for supersede: %v", err)
	}
	defer loaded.Close()
	_, err = CloseChange(v, loaded, first.ID, ChangeCloseOptions{VerificationStatus: "skipped", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC), SupersededBy: []string{second.ID}})
	if err == nil || !strings.Contains(err.Error(), "cannot be updated with a reciprocal supersedes link") {
		t.Fatalf("expected terminal successor rejection, got %v", err)
	}
}

func TestBuildProjectStatusSummaryLeavesMissingOptionalSizeEmpty(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	writeExistingChangeWithoutOptionalFields(t, root)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	defer loaded.Close()
	summary, err := BuildProjectStatusSummary(v, loaded)
	if err != nil {
		t.Fatalf("build status summary: %v", err)
	}
	if len(summary.Active) != 1 {
		t.Fatalf("expected one active change, got %#v", summary.Active)
	}
	if got := summary.Active[0].Size; got != "" {
		t.Fatalf("expected missing size to remain empty, got %q", got)
	}
}

func TestCloseChangeOmitsMissingOptionalFieldsWhenRewritingStatus(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	changeID := writeExistingChangeWithoutOptionalFields(t, root)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	defer loaded.Close()
	if _, err := CloseChange(v, loaded, changeID, ChangeCloseOptions{
		VerificationStatus: "passed",
		ClosedAt:           time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("close change: %v", err)
	}
	statusData, err := os.ReadFile(filepath.Join(root, "runecontext", "changes", changeID, "status.yaml"))
	if err != nil {
		t.Fatalf("read rewritten status: %v", err)
	}
	text := strings.ReplaceAll(string(statusData), "\r\n", "\n")
	if strings.Contains(text, "<nil>") {
		t.Fatalf("expected rewritten status to avoid <nil> placeholders, got:\n%s", text)
	}
	if strings.Contains(text, "created_at:") {
		t.Fatalf("expected missing created_at to stay omitted, got:\n%s", text)
	}
	if strings.Contains(text, "size:") {
		t.Fatalf("expected missing size to stay omitted, got:\n%s", text)
	}
	if !strings.Contains(text, "closed_at: \"2026-03-20\"") {
		t.Fatalf("expected closed_at to be written, got:\n%s", text)
	}
}

func TestCloseChangePreservesDefaultPromotionAssessmentStatusWhenEmpty(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	changeID := writeExistingChangeWithEmptyPromotionAssessment(t, root)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	defer loaded.Close()
	if _, err := CloseChange(v, loaded, changeID, ChangeCloseOptions{
		VerificationStatus: "passed",
		ClosedAt:           time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("close change: %v", err)
	}
	statusData, err := os.ReadFile(filepath.Join(root, "runecontext", "changes", changeID, "status.yaml"))
	if err != nil {
		t.Fatalf("read rewritten status: %v", err)
	}
	text := strings.ReplaceAll(string(statusData), "\r\n", "\n")
	if strings.Contains(text, "<nil>") {
		t.Fatalf("expected rewritten status to avoid <nil> placeholders, got:\n%s", text)
	}
	if !strings.Contains(text, "promotion_assessment:\n  status: pending\n  suggested_targets: []") {
		t.Fatalf("expected empty promotion assessment to preserve pending default, got:\n%s", text)
	}
}

func TestReallocateChangeUpdatesLocalMarkdownReferences(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	result, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Add cache invalidation",
		Type:           "feature",
		Size:           "small",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xaa, 0xbb}),
	})
	loaded.Close()
	if err != nil {
		t.Fatalf("create change: %v", err)
	}
	proposalPath := filepath.Join(root, "runecontext", "changes", result.ID, "proposal.md")
	rewriteFile(t, proposalPath, func(text string) string {
		return text + "\nSee changes/" + result.ID + " and changes/" + result.ID + "/proposal.md#summary and changes/" + result.ID + "/standards.md#applicable-standards for the local review flow.\n"
	})
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	defer loaded.Close()
	reallocated, err := ReallocateChange(v, loaded, result.ID, ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xcc, 0xdd})})
	if err != nil {
		t.Fatalf("reallocate change: %v", err)
	}
	if got, want := reallocated.OldID, result.ID; got != want {
		t.Fatalf("expected old ID %q, got %q", want, got)
	}
	if got, want := reallocated.ID, "CHG-2026-002-ccdd-add-cache-invalidation"; got != want {
		t.Fatalf("expected new ID %q, got %q", want, got)
	}
	if got, want := reallocated.RewrittenReferenceCount, 3; got != want {
		t.Fatalf("expected rewritten reference count %d, got %d", want, got)
	}
	if len(reallocated.Warnings) != 0 {
		t.Fatalf("expected successful reallocation without warnings, got %#v", reallocated.Warnings)
	}
	if _, err := os.Stat(filepath.Join(root, "runecontext", "changes", result.ID)); !os.IsNotExist(err) {
		t.Fatalf("expected old change path to disappear, got err=%v", err)
	}
	newChangeDir := filepath.Join(root, "runecontext", "changes", reallocated.ID)
	statusData, err := os.ReadFile(filepath.Join(newChangeDir, "status.yaml"))
	if err != nil {
		t.Fatalf("read reallocated status: %v", err)
	}
	if !strings.Contains(string(statusData), "id: "+reallocated.ID) {
		t.Fatalf("expected status ID rewrite, got:\n%s", string(statusData))
	}
	proposalData, err := os.ReadFile(filepath.Join(newChangeDir, "proposal.md"))
	if err != nil {
		t.Fatalf("read reallocated proposal: %v", err)
	}
	proposalText := strings.ReplaceAll(string(proposalData), "\r\n", "\n")
	if strings.Contains(proposalText, result.ID) {
		t.Fatalf("expected old change ID refs to be rewritten, got:\n%s", proposalText)
	}
	if !strings.Contains(proposalText, "changes/"+reallocated.ID+"/proposal.md#summary") {
		t.Fatalf("expected proposal self-reference rewrite, got:\n%s", proposalText)
	}
	if !strings.Contains(proposalText, "changes/"+reallocated.ID+" and") {
		t.Fatalf("expected change-root reference rewrite, got:\n%s", proposalText)
	}
	if !strings.Contains(proposalText, "changes/"+reallocated.ID+"/standards.md#applicable-standards") {
		t.Fatalf("expected standards self-reference rewrite, got:\n%s", proposalText)
	}
	validated, err := v.ValidateProject(root)
	if err != nil {
		t.Fatalf("validate reallocated project: %v", err)
	}
	_ = validated.Close()
	if len(reallocated.ChangedFiles) == 0 {
		t.Fatalf("expected changed files to be reported")
	}
}

func TestReallocateChangeRejectsExternalReferences(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	defer loaded.Close()
	_, err = ReallocateChange(v, loaded, "CHG-2026-001-a3f2-auth-gateway", ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xaa, 0xbb})})
	if err == nil || !strings.Contains(err.Error(), "alpha.3 reallocation only rewrites local references inside the change") {
		t.Fatalf("expected external-reference rejection, got %v", err)
	}
	statusData, readErr := os.ReadFile(filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml"))
	if readErr != nil {
		t.Fatalf("read original status after failed reallocation: %v", readErr)
	}
	if !strings.Contains(string(statusData), "id: CHG-2026-001-a3f2-auth-gateway") {
		t.Fatalf("expected failed reallocation to leave original status intact, got:\n%s", string(statusData))
	}
}

func TestReallocateChangeRejectsTerminalChange(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	result, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Add cache invalidation",
		Type:           "feature",
		Size:           "small",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xaa, 0xbb}),
	})
	loaded.Close()
	if err != nil {
		t.Fatalf("create change: %v", err)
	}
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	if _, err := CloseChange(v, loaded, result.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)}); err != nil {
		loaded.Close()
		t.Fatalf("close change: %v", err)
	}
	loaded.Close()
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload closed project: %v", err)
	}
	defer loaded.Close()
	_, err = ReallocateChange(v, loaded, result.ID, ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xcc, 0xdd})})
	if err == nil || !strings.Contains(err.Error(), "terminal status") {
		t.Fatalf("expected terminal-status rejection, got %v", err)
	}
}

func TestReallocateChangeRejectsExistingBackupPath(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	result, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Add cache invalidation",
		Type:           "feature",
		Size:           "small",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xaa, 0xbb}),
	})
	loaded.Close()
	if err != nil {
		t.Fatalf("create change: %v", err)
	}
	backupPath := filepath.Join(root, "runecontext", ".reallocate-"+result.ID+"-backup")
	if err := os.MkdirAll(backupPath, 0o755); err != nil {
		t.Fatalf("mkdir backup path: %v", err)
	}
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	defer loaded.Close()
	_, err = ReallocateChange(v, loaded, result.ID, ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xcc, 0xdd})})
	if err == nil || !strings.Contains(err.Error(), "backup path") {
		t.Fatalf("expected backup-path rejection, got %v", err)
	}
}

func TestReallocateChangeRejectsSymlinksInChangeDirectory(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	result, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Add cache invalidation",
		Type:           "feature",
		Size:           "small",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xaa, 0xbb}),
	})
	loaded.Close()
	if err != nil {
		t.Fatalf("create change: %v", err)
	}
	changeDir := filepath.Join(root, "runecontext", "changes", result.ID)
	if err := os.Symlink("proposal.md", filepath.Join(changeDir, "proposal-link.md")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	defer loaded.Close()
	_, err = ReallocateChange(v, loaded, result.ID, ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xcc, 0xdd})})
	if err == nil || !strings.Contains(err.Error(), "does not support symlinks") {
		t.Fatalf("expected symlink rejection, got %v", err)
	}
}

func TestReallocateChangeSurfacesRollbackFailures(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	result, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Add cache invalidation",
		Type:           "feature",
		Size:           "small",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xaa, 0xbb}),
	})
	loaded.Close()
	if err != nil {
		t.Fatalf("create change: %v", err)
	}
	originalRename := renamePath
	originalValidate := validateProjectAfterChangeMutation
	t.Cleanup(func() {
		renamePath = originalRename
		validateProjectAfterChangeMutation = originalValidate
	})
	validateProjectAfterChangeMutation = func(*Validator, string) (*ProjectIndex, error) {
		return nil, fmt.Errorf("forced validation failure")
	}
	backupPath := filepath.Join(root, "runecontext", ".reallocate-"+result.ID+"-backup")
	originalChangePath := filepath.Join(root, "runecontext", "changes", result.ID)
	renamePath = func(oldPath, newPath string) error {
		if oldPath == backupPath && newPath == originalChangePath {
			return fmt.Errorf("forced rollback rename failure")
		}
		return os.Rename(oldPath, newPath)
	}
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	defer loaded.Close()
	_, err = ReallocateChange(v, loaded, result.ID, ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xcc, 0xdd})})
	if err == nil || !strings.Contains(err.Error(), "manual recovery may be required") || !strings.Contains(err.Error(), "forced validation failure") || !strings.Contains(err.Error(), "forced rollback rename failure") {
		t.Fatalf("expected rollback failure details, got %v", err)
	}
}

func TestReallocateChangeReturnsCleanupWarning(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	result, err := CreateChange(v, loaded, ChangeCreateOptions{
		Title:          "Add cache invalidation",
		Type:           "feature",
		Size:           "small",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xaa, 0xbb}),
	})
	loaded.Close()
	if err != nil {
		t.Fatalf("create change: %v", err)
	}
	originalRemoveAll := removeAllPath
	t.Cleanup(func() {
		removeAllPath = originalRemoveAll
	})
	backupPath := filepath.Join(root, "runecontext", ".reallocate-"+result.ID+"-backup")
	removeAllPath = func(path string) error {
		if path == backupPath {
			return fmt.Errorf("forced cleanup failure")
		}
		return os.RemoveAll(path)
	}
	loaded, err = v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	defer loaded.Close()
	reallocated, err := ReallocateChange(v, loaded, result.ID, ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xcc, 0xdd})})
	if err != nil {
		t.Fatalf("reallocate change with cleanup warning: %v", err)
	}
	if len(reallocated.Warnings) != 1 || !strings.Contains(reallocated.Warnings[0], "forced cleanup failure") {
		t.Fatalf("expected cleanup warning, got %#v", reallocated.Warnings)
	}
}

func TestRewriteStatusChangeIDRefsUpdatesLocalLists(t *testing.T) {
	raw := map[string]any{
		"related_changes": []any{"CHG-2026-001-a3f2-auth-gateway", "CHG-2026-002-b4c3-auth-revision"},
		"depends_on":      []any{"CHG-2026-001-a3f2-auth-gateway"},
		"informed_by":     []any{"CHG-2026-009-c0de-auth-notes"},
		"supersedes":      []any{"CHG-2026-001-a3f2-auth-gateway"},
		"superseded_by":   []any{},
	}
	rewriteStatusChangeIDRefs(raw, "CHG-2026-001-a3f2-auth-gateway", "CHG-2026-010-dd44-auth-gateway")
	for _, key := range []string{"related_changes", "depends_on", "supersedes"} {
		if got := strings.Join(extractStringList(raw[key]), ","); strings.Contains(got, "CHG-2026-001-a3f2-auth-gateway") {
			t.Fatalf("expected %s to rewrite old ID, got %q", key, got)
		}
	}
	if got := strings.Join(extractStringList(raw["informed_by"]), ","); got != "CHG-2026-009-c0de-auth-notes" {
		t.Fatalf("expected unrelated IDs to remain untouched, got %q", got)
	}
}

func TestStatusDocumentFromMapRejectsInvalidPromotionAssessmentStatus(t *testing.T) {
	_, err := statusDocumentFromMap(map[string]any{
		"schema_version":      1,
		"id":                  "CHG-2026-001-a3f2-auth-gateway",
		"title":               "Add auth gateway",
		"status":              "proposed",
		"type":                "feature",
		"verification_status": "pending",
		"context_bundles":     []any{"base"},
		"related_specs":       []any{},
		"related_decisions":   []any{},
		"related_changes":     []any{},
		"depends_on":          []any{},
		"informed_by":         []any{},
		"supersedes":          []any{},
		"superseded_by":       []any{},
		"closed_at":           nil,
		"promotion_assessment": map[string]any{
			"status": "not-valid",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "promotion_assessment.status") {
		t.Fatalf("expected invalid promotion assessment status error, got %v", err)
	}
}

func copyChangeWorkflowTemplate(t *testing.T) string {
	t.Helper()
	src := fixturePath(t, "change-workflow", "template-project")
	dst := t.TempDir()
	copyDirTree(t, src, dst)
	return dst
}

func writeExistingChangeWithoutOptionalFields(t *testing.T, root string) string {
	t.Helper()
	changeID := "CHG-2026-001-a3f2-auth-gateway"
	changeDir := filepath.Join(root, "runecontext", "changes", changeID)
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatalf("mkdir change dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "status.yaml"), []byte(strings.Join([]string{
		"schema_version: 1",
		"id: CHG-2026-001-a3f2-auth-gateway",
		"title: Add auth gateway",
		"status: implemented",
		"type: feature",
		"verification_status: pending",
		"context_bundles:",
		"  - base",
		"related_specs: []",
		"related_decisions: []",
		"related_changes: []",
		"depends_on: []",
		"informed_by: []",
		"supersedes: []",
		"superseded_by: []",
		"closed_at: null",
		"promotion_assessment:",
		"  status: pending",
		"  suggested_targets: []",
		"",
	}, "\n")), 0o644); err != nil {
		t.Fatalf("write status: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte(strings.Join([]string{
		"## Summary",
		"Add auth gateway",
		"",
		"## Problem",
		"The repository needs a durable auth gateway change record.",
		"",
		"## Proposed Change",
		"Close the existing auth gateway change cleanly.",
		"",
		"## Why Now",
		"This regression test exercises missing optional status fields.",
		"",
		"## Assumptions",
		"N/A",
		"",
		"## Out of Scope",
		"Any auth implementation work.",
		"",
		"## Impact",
		"The rewritten status should remain schema-valid without placeholder strings.",
		"",
	}, "\n")), 0o644); err != nil {
		t.Fatalf("write proposal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "standards.md"), []byte(strings.Join([]string{
		"## Applicable Standards",
		"- `standards/global/base.md`: Selected from the current context bundles.",
		"",
		"## Resolution Notes",
		"This change keeps standards linkage valid while exercising missing optional metadata.",
		"",
	}, "\n")), 0o644); err != nil {
		t.Fatalf("write standards: %v", err)
	}
	return changeID
}

func writeExistingChangeWithEmptyPromotionAssessment(t *testing.T, root string) string {
	t.Helper()
	changeID := writeExistingChangeWithoutOptionalFields(t, root)
	statusPath := filepath.Join(root, "runecontext", "changes", changeID, "status.yaml")
	rewriteFile(t, statusPath, func(text string) string {
		return strings.Replace(text, "promotion_assessment:\n  status: pending\n  suggested_targets: []", "promotion_assessment: {}", 1)
	})
	return changeID
}

func requireFileContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	got := strings.ReplaceAll(string(data), "\r\n", "\n")
	if got != want {
		t.Fatalf("unexpected content for %s\nwant:\n%s\n---\ngot:\n%s", path, want, got)
	}
}
