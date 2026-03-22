package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateChangeMinimum(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	if got, want := result.Mode, ChangeModeMinimum; got != want {
		t.Fatalf("expected mode %q, got %q", want, got)
	}
	assertMinimumChangeFiles(t, root, result.ID)
	assertMinimumChangeSkippedDesign(t, root, result.ID)
	assertValidatedWorkflowProject(t, v, root)
}

func assertMinimumChangeFiles(t *testing.T, root, changeID string) {
	t.Helper()
	changeDir := filepath.Join(root, "runecontext", "changes", changeID)
	requireFileContent(t, filepath.Join(changeDir, "status.yaml"), strings.Join([]string{"schema_version: 1", "id: CHG-2026-001-aabb-add-cache-invalidation", "title: Add cache invalidation", "status: proposed", "type: feature", "size: small", "verification_status: pending", "context_bundles:", "  - base", "related_specs: []", "related_decisions: []", "related_changes: []", "depends_on: []", "informed_by: []", "supersedes: []", "superseded_by: []", "created_at: \"2026-03-18\"", "closed_at: null", "promotion_assessment:", "  status: pending", "  suggested_targets: []", ""}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "proposal.md"), strings.Join([]string{"## Summary", "Add cache invalidation", "", "## Problem", "The repository needs a reviewable RuneContext change record for this work.", "", "## Proposed Change", "Track Add cache invalidation through the minimum RuneContext change artifacts.", "", "## Why Now", "The work needs stable intent, standards linkage, and verification planning before it moves further.", "", "## Assumptions", "- Inferred `just test` from the repository's justfile test target.", "", "## Out of Scope", "Work outside the scoped change tracked here.", "", "## Impact", "The change keeps intent, assumptions, and standards linkage reviewable.", ""}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "standards.md"), strings.Join([]string{"## Applicable Standards", "- `standards/global/base.md`: Selected from the current context bundles.", "", "## Resolution Notes", "Generated from the current context bundle selection; review any automatic refresh before committing.", ""}, "\n"))
}

func assertMinimumChangeSkippedDesign(t *testing.T, root, changeID string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, "runecontext", "changes", changeID, "design.md")); !os.IsNotExist(err) {
		t.Fatalf("expected minimum mode to skip design.md, got err=%v", err)
	}
}

func TestCreateProjectChangeAutoShapes(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	_, result := mustCreateChange(t, root, defaultProjectChangeOptions("Launch payments platform", []byte{0xaa, 0xbb}))
	if got, want := result.Mode, ChangeModeFull; got != want {
		t.Fatalf("expected mode %q, got %q", want, got)
	}
	assertProjectAutoShapeFiles(t, root, result.ID)
}

func assertProjectAutoShapeFiles(t *testing.T, root, changeID string) {
	t.Helper()
	changeDir := filepath.Join(root, "runecontext", "changes", changeID)
	requireFileContent(t, filepath.Join(changeDir, "design.md"), strings.Join([]string{"# Design", "", "## Overview", "Shape Launch payments platform before implementation so scope, standards linkage, and verification stay reviewable.", "", "## Shape Rationale", "- Project work uses deeper intake because bad defaults compound.", "", "## Project Intake Checklist", "- Mission and target users.", "- Stack and runtime constraints.", "- Deployment and security constraints.", "- Success criteria.", "- Non-goals.", "", "## Ask More When", "- Mission and target users.", "- Stack and runtime constraints.", "- Deployment and security constraints.", "- Success criteria.", "- Non-goals.", ""}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "verification.md"), strings.Join([]string{"# Verification", "", "## Planned Checks", "- `just test`", "", "## Close Gate", "Use the repository's standard verification flow before closing this change.", ""}, "\n"))
}

func TestCreateChangeCleansUpDirectoryOnValidationFailure(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()
	originalValidate := validateProjectAfterChangeMutation
	t.Cleanup(func() { validateProjectAfterChangeMutation = originalValidate })
	validateProjectAfterChangeMutation = func(*Validator, string) (*ProjectIndex, error) { return nil, fmt.Errorf("forced validation failure") }
	_, err := CreateChange(v, loaded, defaultFeatureChangeOptions("Add cache invalidation", []byte{0xaa, 0xbb}))
	if err == nil || !strings.Contains(err.Error(), "forced validation failure") {
		t.Fatalf("expected forced validation failure, got %v", err)
	}
	changeDir := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-aabb-add-cache-invalidation")
	if _, statErr := os.Stat(changeDir); !os.IsNotExist(statErr) {
		t.Fatalf("expected failed create to clean up %s, got err=%v", changeDir, statErr)
	}
}

func TestCreateChangeRejectsSymlinkedMutationPath(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()
	originalLstat := lstatPath
	t.Cleanup(func() { lstatPath = originalLstat })
	changesRoot := filepath.Clean(filepath.Join(root, "runecontext", "changes"))
	lstatPath = func(path string) (os.FileInfo, error) {
		if filepath.Clean(path) == changesRoot {
			return fakeFileInfo{name: filepath.Base(path), mode: os.ModeSymlink}, nil
		}
		return os.Lstat(path)
	}
	_, err := CreateChange(v, loaded, defaultFeatureChangeOptions("Add cache invalidation", []byte{0xaa, 0xbb}))
	if err == nil || !strings.Contains(err.Error(), "symlinked targets") {
		t.Fatalf("expected symlink rejection, got %v", err)
	}
}

func TestCreateChangeAllowsNoSelectableStandards(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	configureWorkflowWithNoSelectableStandards(t, root)
	v, result := mustCreateDefaultFeatureChange(t, root)
	changeDir := filepath.Join(root, "runecontext", "changes", result.ID)
	if got := len(result.ApplicableStandards); got != 0 {
		t.Fatalf("expected no applicable standards, got %d (%#v)", got, result.ApplicableStandards)
	}
	wantAssumption := "No selectable standards are defined in the project yet; the Applicable Standards section is rendered as N/A."
	if !containsStringValue(result.Assumptions, wantAssumption) {
		t.Fatalf("expected assumptions to include %q, got %#v", wantAssumption, result.Assumptions)
	}
	requireFileContent(t, filepath.Join(changeDir, "standards.md"), strings.Join([]string{"## Applicable Standards", "N/A", "", "## Resolution Notes", "Generated from the current context bundle selection; review any automatic refresh before committing.", ""}, "\n"))
	assertValidatedWorkflowProject(t, v, root)
}

func TestCreateChangeUsesConservativeFallbackAssumptionWhenStandardsExist(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	writeWorkflowBundle(t, root, filepath.Join("bundles", "base.yaml"), []string{"schema_version: 1", "id: base", "includes:", "  project:", "    - project/mission.md", ""})
	v, result := mustCreateDefaultFeatureChange(t, root)
	wantAssumption := "Used all non-draft standards as a conservative fallback because no standards were selected through context bundles."
	if !containsStringValue(result.Assumptions, wantAssumption) {
		t.Fatalf("expected assumptions to include %q, got %#v", wantAssumption, result.Assumptions)
	}
	if got, want := len(result.ApplicableStandards), 2; got != want {
		t.Fatalf("expected %d applicable standards, got %d (%#v)", want, got, result.ApplicableStandards)
	}
	changeDir := filepath.Join(root, "runecontext", "changes", result.ID)
	requireFileContent(t, filepath.Join(changeDir, "standards.md"), strings.Join([]string{"## Applicable Standards", "- `standards/global/base.md`: Selected from the current context bundles.", "- `standards/security/review.md`: Selected from the current context bundles.", "", "## Resolution Notes", "Generated from the current context bundle selection; review any automatic refresh before committing.", ""}, "\n"))
	assertValidatedWorkflowProject(t, v, root)
}

func TestShapeChangeRefreshesStandardsAfterNoSelectableStandardsCreate(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	configureWorkflowWithNoSelectableStandards(t, root)
	v, created := mustCreateDefaultFeatureChange(t, root)
	restoreBaseBundleToSelectSecurityStandard(t, root)
	rewriteWorkflowStandardStatus(t, root, filepath.Join("standards", "security", "review.md"), "active")
	shapeResult := mustShapeChange(t, v, root, created.ID, ChangeShapeOptions{})
	if got, want := shapeResult.StandardsRefreshAction, "updated"; got != want {
		t.Fatalf("expected standards refresh %q, got %q", want, got)
	}
	if got, want := len(shapeResult.AddedStandards), 1; got != want || shapeResult.AddedStandards[0] != "standards/security/review.md" {
		t.Fatalf("expected one added standard, got %#v", shapeResult.AddedStandards)
	}
	changeDir := filepath.Join(root, "runecontext", "changes", created.ID)
	requireFileContent(t, filepath.Join(changeDir, "standards.md"), strings.Join([]string{"## Applicable Standards", "- `standards/security/review.md`: Selected from the current context bundles.", "", "## Standards Added Since Last Refresh", "- `standards/security/review.md`: Newly selected during standards refresh.", "", "## Resolution Notes", "Generated from the current context bundle selection; review any automatic refresh before committing.", ""}, "\n"))
	assertValidatedWorkflowProject(t, v, root)
}

func configureWorkflowWithNoSelectableStandards(t *testing.T, root string) {
	t.Helper()
	for _, rel := range []string{filepath.Join("standards", "global", "base.md"), filepath.Join("standards", "security", "review.md")} {
		rewriteWorkflowStandardStatus(t, root, rel, "draft")
	}
	writeWorkflowBundle(t, root, filepath.Join("bundles", "base.yaml"), []string{"schema_version: 1", "id: base", "includes:", "  project:", "    - project/mission.md", ""})
	writeWorkflowBundle(t, root, filepath.Join("bundles", "security.yaml"), []string{"schema_version: 1", "id: security", "includes:", "  project:", "    - project/mission.md", ""})
}

func rewriteWorkflowStandardStatus(t *testing.T, root, standardRelPath, status string) {
	t.Helper()
	path := filepath.Join(root, "runecontext", standardRelPath)
	rewriteFile(t, path, func(text string) string {
		for _, current := range []string{"active", "draft", "deprecated"} {
			updated := strings.Replace(text, "status: "+current, "status: "+status, 1)
			if updated != text {
				return updated
			}
		}
		t.Fatalf("expected status field in %s", path)
		return text
	})
}

func restoreBaseBundleToSelectSecurityStandard(t *testing.T, root string) {
	t.Helper()
	writeWorkflowBundle(t, root, filepath.Join("bundles", "base.yaml"), []string{"schema_version: 1", "id: base", "includes:", "  project:", "    - project/mission.md", "  standards:", "    - standards/security/review.md", ""})
}

func writeWorkflowBundle(t *testing.T, root, bundleRelPath string, lines []string) {
	t.Helper()
	path := filepath.Join(root, "runecontext", bundleRelPath)
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatalf("write bundle %s: %v", path, err)
	}
}

func containsStringValue(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func TestShapeChangeCreatesSupplementalDocsAndRefreshesStandards(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	rewriteShapedStatusBundle(t, root, result.ID, "security")
	shapeResult := mustShapeChange(t, v, root, result.ID, ChangeShapeOptions{Tasks: []string{"Implement cache invalidation flow.", "Add regression coverage."}, References: []string{"docs/cache.md", "issue-42"}})
	assertShapeRefreshResult(t, shapeResult)
	assertShapedChangeFiles(t, root, result.ID)
}

func rewriteShapedStatusBundle(t *testing.T, root, changeID, bundle string) {
	t.Helper()
	statusPath := filepath.Join(root, "runecontext", "changes", changeID, "status.yaml")
	rewriteFile(t, statusPath, func(text string) string { return strings.Replace(text, "- base", "- "+bundle, 1) })
}

func assertShapeRefreshResult(t *testing.T, result *ChangeOperationResult) {
	t.Helper()
	if got, want := result.StandardsRefreshAction, "updated"; got != want {
		t.Fatalf("expected standards refresh %q, got %q", want, got)
	}
	if !result.ReviewDiffRequired {
		t.Fatalf("expected updated standards refresh to require review diff")
	}
}

func assertShapedChangeFiles(t *testing.T, root, changeID string) {
	t.Helper()
	changeDir := filepath.Join(root, "runecontext", "changes", changeID)
	requireFileContent(t, filepath.Join(changeDir, "design.md"), strings.Join([]string{"# Design", "", "## Overview", "Shape Add cache invalidation before implementation so scope, standards linkage, and verification stay reviewable.", "", "## Shape Rationale", "- Full mode was requested explicitly to deepen the change.", "- Minimum mode is sufficient for the current size and risk signal.", ""}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "verification.md"), strings.Join([]string{"# Verification", "", "## Planned Checks", "- `just test`", "", "## Close Gate", "Use the repository's standard verification flow before closing this change.", ""}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "tasks.md"), strings.Join([]string{"# Tasks", "", "- Implement cache invalidation flow.", "- Add regression coverage.", ""}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "references.md"), strings.Join([]string{"# References", "", "- docs/cache.md", "- issue-42", ""}, "\n"))
	requireFileContent(t, filepath.Join(changeDir, "standards.md"), strings.Join([]string{"## Applicable Standards", "- `standards/security/review.md`: Selected from the current context bundles.", "", "## Standards Added Since Last Refresh", "- `standards/security/review.md`: Newly selected during standards refresh.", "", "## Resolution Notes", "Generated from the current context bundle selection; review any automatic refresh before committing.", ""}, "\n"))
}

func TestShapeChangeIsIdempotentAndLeavesStandardsUnchanged(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	firstShape := mustShapeChange(t, v, root, result.ID, ChangeShapeOptions{})
	assertUnchangedShapeResult(t, firstShape, "first")
	secondShape := mustShapeChange(t, v, root, result.ID, ChangeShapeOptions{})
	assertUnchangedShapeResult(t, secondShape, "second")
	if len(secondShape.ChangedFiles) != 0 {
		t.Fatalf("expected idempotent shape to leave files unchanged, got %#v", secondShape.ChangedFiles)
	}
}

func assertUnchangedShapeResult(t *testing.T, result *ChangeOperationResult, label string) {
	t.Helper()
	if got, want := result.StandardsRefreshAction, "unchanged"; got != want {
		t.Fatalf("expected %s standards refresh %q, got %q", label, want, got)
	}
	if result.ReviewDiffRequired {
		t.Fatalf("expected %s standards refresh to avoid review diff requirement", label)
	}
}

func TestShapeChangeRejectsTerminalChange(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	mustCloseChange(t, v, root, result.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := ShapeChange(v, loaded, result.ID, ChangeShapeOptions{})
	if err == nil || !strings.Contains(err.Error(), "terminal status") {
		t.Fatalf("expected terminal-shape rejection, got %v", err)
	}
}
