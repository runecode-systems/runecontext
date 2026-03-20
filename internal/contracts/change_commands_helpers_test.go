package contracts

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeFileInfo struct {
	name string
	mode os.FileMode
}

func (f fakeFileInfo) Name() string       { return f.name }
func (f fakeFileInfo) Size() int64        { return 0 }
func (f fakeFileInfo) Mode() os.FileMode  { return f.mode }
func (f fakeFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (f fakeFileInfo) IsDir() bool        { return f.mode.IsDir() }
func (f fakeFileInfo) Sys() any           { return nil }

func copyChangeWorkflowTemplate(t *testing.T) string {
	t.Helper()
	src := fixturePath(t, "change-workflow", "template-project")
	dst := t.TempDir()
	copyDirTree(t, src, dst)
	return dst
}

func mustLoadWorkflowProject(t *testing.T, root string) (*Validator, *LoadedProject) {
	t.Helper()
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	return v, loaded
}

func mustReloadWorkflowProject(t *testing.T, v *Validator, root string) *LoadedProject {
	t.Helper()
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	return loaded
}

func mustCreateChange(t *testing.T, root string, options ChangeCreateOptions) (*Validator, *ChangeOperationResult) {
	t.Helper()
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()
	result, err := CreateChange(v, loaded, options)
	if err != nil {
		t.Fatalf("create change: %v", err)
	}
	return v, result
}

func mustCreateDefaultFeatureChange(t *testing.T, root string) (*Validator, *ChangeOperationResult) {
	t.Helper()
	return mustCreateChange(t, root, defaultFeatureChangeOptions("Add cache invalidation", []byte{0xaa, 0xbb}))
}

func defaultFeatureChangeOptions(title string, entropy []byte) ChangeCreateOptions {
	return ChangeCreateOptions{Title: title, Type: "feature", Size: "small", ContextBundles: []string{"base"}, Now: time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC), Entropy: bytes.NewReader(entropy)}
}

func defaultProjectChangeOptions(title string, entropy []byte) ChangeCreateOptions {
	return ChangeCreateOptions{Title: title, Type: "project", ContextBundles: []string{"base"}, Now: time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC), Entropy: bytes.NewReader(entropy)}
}

func mustCloseChange(t *testing.T, v *Validator, root, changeID string, options ChangeCloseOptions) *ChangeOperationResult {
	t.Helper()
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := CloseChange(v, loaded, changeID, options)
	if err != nil {
		t.Fatalf("close change: %v", err)
	}
	return result
}

func mustShapeChange(t *testing.T, v *Validator, root, changeID string, options ChangeShapeOptions) *ChangeOperationResult {
	t.Helper()
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := ShapeChange(v, loaded, changeID, options)
	if err != nil {
		t.Fatalf("shape change: %v", err)
	}
	return result
}

func mustReallocateChange(t *testing.T, v *Validator, root, changeID string, entropy []byte) *ChangeReallocationResult {
	t.Helper()
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := ReallocateChange(v, loaded, changeID, ChangeReallocateOptions{Entropy: bytes.NewReader(entropy)})
	if err != nil {
		t.Fatalf("reallocate change: %v", err)
	}
	return result
}

func assertValidatedWorkflowProject(t *testing.T, v *Validator, root string) {
	t.Helper()
	validated, err := v.ValidateProject(root)
	if err != nil {
		t.Fatalf("validate generated project: %v", err)
	}
	_ = validated.Close()
}

func writeExistingChangeWithoutOptionalFields(t *testing.T, root string) string {
	t.Helper()
	changeID := "CHG-2026-001-a3f2-auth-gateway"
	changeDir := filepath.Join(root, "runecontext", "changes", changeID)
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatalf("mkdir change dir: %v", err)
	}
	writeExistingChangeFiles(t, changeDir)
	return changeID
}

func writeExistingChangeFiles(t *testing.T, changeDir string) {
	t.Helper()
	for path, body := range existingChangeFileBodies(changeDir) {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write file %s: %v", path, err)
		}
	}
}

func existingChangeFileBodies(changeDir string) map[string]string {
	return map[string]string{
		filepath.Join(changeDir, "status.yaml"):  existingChangeStatusBody(),
		filepath.Join(changeDir, "proposal.md"):  existingChangeProposalBody(),
		filepath.Join(changeDir, "standards.md"): existingChangeStandardsBody(),
	}
}

func existingChangeStatusBody() string {
	return strings.Join([]string{"schema_version: 1", "id: CHG-2026-001-a3f2-auth-gateway", "title: Add auth gateway", "status: implemented", "type: feature", "verification_status: pending", "context_bundles:", "  - base", "related_specs: []", "related_decisions: []", "related_changes: []", "depends_on: []", "informed_by: []", "supersedes: []", "superseded_by: []", "closed_at: null", "promotion_assessment:", "  status: pending", "  suggested_targets: []", ""}, "\n")
}

func existingChangeProposalBody() string {
	return strings.Join([]string{"## Summary", "Add auth gateway", "", "## Problem", "The repository needs a durable auth gateway change record.", "", "## Proposed Change", "Close the existing auth gateway change cleanly.", "", "## Why Now", "This regression test exercises missing optional status fields.", "", "## Assumptions", "N/A", "", "## Out of Scope", "Any auth implementation work.", "", "## Impact", "The rewritten status should remain schema-valid without placeholder strings.", ""}, "\n")
}

func existingChangeStandardsBody() string {
	return strings.Join([]string{"## Applicable Standards", "- `standards/global/base.md`: Selected from the current context bundles.", "", "## Resolution Notes", "This change keeps standards linkage valid while exercising missing optional metadata.", ""}, "\n")
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

func mustReadBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func assertFileBytesEqual(t *testing.T, path string, want []byte) {
	t.Helper()
	if got := mustReadBytes(t, path); !bytes.Equal(got, want) {
		t.Fatalf("expected %s to remain unchanged\nwant: %q\ngot:  %q", path, string(want), string(got))
	}
}

func containsMutation(items []FileMutation, path, action string) bool {
	for _, item := range items {
		if item.Path == path && item.Action == action {
			return true
		}
	}
	return false
}
