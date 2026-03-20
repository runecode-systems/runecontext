package contracts

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildContextPackReportRenderModes(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	report, err := index.BuildContextPackReport(ContextPackReportOptions{
		ContextPackOptions: ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)},
		Explain:            true,
	})
	if err != nil {
		t.Fatalf("build context pack report: %v", err)
	}
	if report.ReportSchemaVersion != contextPackReportSchemaVersion {
		t.Fatalf("expected report schema version %d, got %d", contextPackReportSchemaVersion, report.ReportSchemaVersion)
	}
	assertContextPackValidAgainstSchema(t, v, report.Pack)
	assertContextPackReportValidAgainstSchema(t, v, report)
	assertContextPackOutputMatchesGolden(t, report, ContextPackOutputModeMachine, fixturePath(t, "context-packs", "reports", "child-reinclude.json"))
	assertContextPackOutputMatchesGolden(t, report, ContextPackOutputModeHuman, fixturePath(t, "context-packs", "reports", "child-reinclude.txt"))
}

func TestBuildContextPackReportRenderRejectsNilMachineAndHumanReports(t *testing.T) {
	var report *ContextPackReport
	for _, mode := range []ContextPackOutputMode{ContextPackOutputModeHuman, ContextPackOutputModeMachine} {
		_, err := report.Render(mode)
		if err == nil || !strings.Contains(err.Error(), "report is unavailable") {
			t.Fatalf("expected nil report error for mode %q, got %v", mode, err)
		}
	}
}

func TestBuildContextPackReportWarnsWhenSelectedFilesExceedDefault(t *testing.T) {
	index := loadModifiedContextPackProject(t, func(root string) {
		rewriteBaseBundleForProjectGlob(t, root)
		for i := 0; i < 260; i++ {
			path := filepath.Join(root, "runecontext", "project", fmt.Sprintf("extra-%03d.md", i))
			if err := os.WriteFile(path, []byte("extra\n"), 0o644); err != nil {
				t.Fatalf("write extra project file: %v", err)
			}
		}
	})
	defer index.Close()
	report, err := index.BuildContextPackReport(ContextPackReportOptions{ContextPackOptions: ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)}})
	if err != nil {
		t.Fatalf("build context pack report: %v", err)
	}
	defaults := DefaultContextPackAdvisoryThresholds()
	assertContextPackWarningPresent(t, report.Warnings, "selected_files_threshold_exceeded", int64(defaults.SelectedFiles))
	assertContextPackReportValidAgainstSchema(t, NewValidator(schemaRoot(t)), report)
}

func TestBuildContextPackReportWarnsWhenReferencedBytesExceedDefault(t *testing.T) {
	index := loadModifiedContextPackProject(t, func(root string) {
		rewriteBaseBundleForProjectGlob(t, root)
		defaults := DefaultContextPackAdvisoryThresholds()
		big := strings.Repeat("a", int(defaults.ReferencedContentBytes)+1)
		path := filepath.Join(root, "runecontext", "project", "big.md")
		if err := os.WriteFile(path, []byte(big), 0o644); err != nil {
			t.Fatalf("write large project file: %v", err)
		}
	})
	defer index.Close()
	report, err := index.BuildContextPackReport(ContextPackReportOptions{ContextPackOptions: ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)}})
	if err != nil {
		t.Fatalf("build context pack report: %v", err)
	}
	assertContextPackWarningPresent(t, report.Warnings, "referenced_content_bytes_threshold_exceeded", DefaultContextPackAdvisoryThresholds().ReferencedContentBytes)
}

func TestBuildContextPackReportWarningOutputRoundTrip(t *testing.T) {
	index := loadModifiedContextPackProject(t, func(root string) {
		rewriteBaseBundleForProjectGlob(t, root)
		for i := 0; i < 260; i++ {
			path := filepath.Join(root, "runecontext", "project", fmt.Sprintf("warning-%03d.md", i))
			if err := os.WriteFile(path, []byte("warning\n"), 0o644); err != nil {
				t.Fatalf("write warning project file: %v", err)
			}
		}
	})
	defer index.Close()
	v := NewValidator(schemaRoot(t))
	report, err := index.BuildContextPackReport(ContextPackReportOptions{ContextPackOptions: ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)}})
	if err != nil {
		t.Fatalf("build warning report: %v", err)
	}
	assertContextPackReportValidAgainstSchema(t, v, report)
	data, err := report.Render(ContextPackOutputModeMachine)
	if err != nil {
		t.Fatalf("render warning report: %v", err)
	}
	var decoded ContextPackReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal warning report: %v", err)
	}
	if decoded.ReportSchemaVersion != contextPackReportSchemaVersion {
		t.Fatalf("expected report schema version %d, got %d", contextPackReportSchemaVersion, decoded.ReportSchemaVersion)
	}
	assertContextPackWarningPresent(t, decoded.Warnings, "selected_files_threshold_exceeded", int64(DefaultContextPackAdvisoryThresholds().SelectedFiles))
}

func TestBuildContextPackReportWarnsWhenProvenanceBytesExceedDefault(t *testing.T) {
	index := loadModifiedContextPackProject(t, func(root string) {
		rewriteBaseBundleForProjectGlob(t, root)
		for i := 0; i < 2200; i++ {
			path := filepath.Join(root, "runecontext", "project", fmt.Sprintf("provenance-%04d.md", i))
			if err := os.WriteFile(path, []byte("p\n"), 0o644); err != nil {
				t.Fatalf("write provenance project file: %v", err)
			}
		}
	})
	defer index.Close()
	report, err := index.BuildContextPackReport(ContextPackReportOptions{ContextPackOptions: ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)}})
	if err != nil {
		t.Fatalf("build context pack report: %v", err)
	}
	assertContextPackWarningPresent(t, report.Warnings, "provenance_bytes_threshold_exceeded", DefaultContextPackAdvisoryThresholds().ProvenanceBytes)
}

func TestBuildContextPackReportRetriesAfterFileChange(t *testing.T) {
	root, index := loadContextPackFixtureProject(t)
	defer index.Close()
	missionPath := filepath.Join(root, "runecontext", "project", "mission.md")
	mutated := false
	withContextPackReadHook(t, func(boundaryPath, path string, next contextPackReadProjectFileFunc) ([]byte, error) {
		data, err := next(boundaryPath, path)
		if err != nil {
			return nil, err
		}
		if shouldMutateContextPackMission(path, &mutated) {
			writeContextPackMissionContent(t, missionPath, "updated mission\n")
		}
		return data, nil
	}, func() {
		report, err := index.BuildContextPackReport(ContextPackReportOptions{ContextPackOptions: ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)}})
		if err != nil {
			t.Fatalf("build context pack report with retry: %v", err)
		}
		sum := sha256.Sum256([]byte("updated mission\n"))
		wantHash := fmt.Sprintf("%x", sum[:])
		if got := findContextPackSelectedFile(report.Pack.Selected.Project, "project/mission.md"); got == nil || got.SHA256 != wantHash {
			t.Fatalf("expected rebuilt mission hash %s, got %#v", wantHash, got)
		}
	})
	if !mutated {
		t.Fatal("expected mission file mutation during retry test")
	}
}

func TestBuildContextPackReportFailsWhenFilesKeepChanging(t *testing.T) {
	root, index := loadContextPackFixtureProject(t)
	defer index.Close()
	missionPath := filepath.Join(root, "runecontext", "project", "mission.md")
	flip := false
	mutations := 0
	withContextPackReadHook(t, func(boundaryPath, path string, next contextPackReadProjectFileFunc) ([]byte, error) {
		data, err := next(boundaryPath, path)
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(filepath.ToSlash(path), "project/mission.md") {
			mutations++
			writeContextPackMissionContent(t, missionPath, nextFlappingMissionContent(&flip))
		}
		return data, nil
	}, func() {
		_, err := index.BuildContextPackReport(ContextPackReportOptions{ContextPackOptions: ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)}})
		if err == nil || !strings.Contains(err.Error(), "changed during build") {
			t.Fatalf("expected changed-during-build error, got %v", err)
		}
	})
	if mutations == 0 {
		t.Fatal("expected mission file flapping during failure test")
	}
}

func TestBuildContextPackReportPropagatesNonTransientDigestErrors(t *testing.T) {
	_, index := loadContextPackFixtureProject(t)
	defer index.Close()
	reads := 0
	withContextPackReadHook(t, newPermissionFailingContextPackReadHook(&reads), func() {
		_, err := index.BuildContextPackReport(ContextPackReportOptions{ContextPackOptions: ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)}})
		assertContextPackPermissionFailure(t, err)
	})
	if reads < 2 {
		t.Fatalf("expected stability check to re-read mission file, got %d reads", reads)
	}
}

func TestBuildContextPackReportRetriesWhenSelectedFileDisappears(t *testing.T) {
	root, index := loadContextPackFixtureProject(t)
	defer index.Close()
	missionPath := filepath.Join(root, "runecontext", "project", "mission.md")
	removed := false
	restored := false
	original := []byte("mission restored\n")
	hook := newDisappearingMissionReadHook(t, missionPath, original, &removed, &restored)
	withContextPackReadHook(t, hook, func() {
		report, err := index.BuildContextPackReport(ContextPackReportOptions{ContextPackOptions: ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)}})
		if err != nil {
			t.Fatalf("expected disappearance to be retried, got %v", err)
		}
		sum := sha256.Sum256(original)
		wantHash := fmt.Sprintf("%x", sum[:])
		if got := findContextPackSelectedFile(report.Pack.Selected.Project, "project/mission.md"); got == nil || got.SHA256 != wantHash {
			t.Fatalf("expected restored mission hash %s, got %#v", wantHash, got)
		}
	})
	if !removed || !restored {
		t.Fatalf("expected disappearance and restoration during retry, removed=%v restored=%v", removed, restored)
	}
}

func TestNormalizeContextPackAdvisoryThresholdsPreservesExplicitZeroValues(t *testing.T) {
	thresholds := normalizeContextPackAdvisoryThresholds(ContextPackAdvisoryThresholds{SelectedFiles: 0, ReferencedContentBytes: 10, ProvenanceBytes: 0})
	if thresholds.SelectedFiles != 0 || thresholds.ProvenanceBytes != 0 {
		t.Fatalf("expected explicit zero thresholds to be preserved, got %#v", thresholds)
	}
	if thresholds.ReferencedContentBytes != 10 {
		t.Fatalf("expected referenced content threshold 10, got %#v", thresholds)
	}
}

func TestNormalizeContextPackAdvisoryThresholdsUsesDefaultsForZeroStruct(t *testing.T) {
	thresholds := normalizeContextPackAdvisoryThresholds(ContextPackAdvisoryThresholds{})
	if thresholds != DefaultContextPackAdvisoryThresholds() {
		t.Fatalf("expected zero struct to use defaults, got %#v", thresholds)
	}
}

func TestNormalizeContextPackAdvisoryThresholdsUsesDefaultsForNegativeFields(t *testing.T) {
	thresholds := normalizeContextPackAdvisoryThresholds(ContextPackAdvisoryThresholds{SelectedFiles: -1, ReferencedContentBytes: 5, ProvenanceBytes: -1})
	defaults := DefaultContextPackAdvisoryThresholds()
	if thresholds.SelectedFiles != defaults.SelectedFiles {
		t.Fatalf("expected negative selected files to fall back to default, got %#v", thresholds)
	}
	if thresholds.ReferencedContentBytes != 5 {
		t.Fatalf("expected explicit referenced content value 5, got %#v", thresholds)
	}
	if thresholds.ProvenanceBytes != defaults.ProvenanceBytes {
		t.Fatalf("expected negative provenance bytes to fall back to default, got %#v", thresholds)
	}
}

type contextPackReadProjectFileFunc func(boundaryPath, path string) ([]byte, error)

func loadModifiedContextPackProject(t *testing.T, mutate func(root string)) *ProjectIndex {
	t.Helper()
	v := NewValidator(schemaRoot(t))
	root := t.TempDir()
	copyDirForTest(t, fixturePath(t, "bundle-resolution", "valid-project"), root)
	mutate(root)
	index, err := v.ValidateProject(root)
	if err != nil {
		t.Fatalf("validate modified project: %v", err)
	}
	return index
}

func rewriteBaseBundleForProjectGlob(t *testing.T, root string) {
	t.Helper()
	path := filepath.Join(root, "runecontext", "bundles", "base.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read base bundle: %v", err)
	}
	updated := strings.Replace(string(data), "project/mission.md", "project/**", 1)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatalf("rewrite base bundle: %v", err)
	}
}

func loadContextPackFixtureProject(t *testing.T) (string, *ProjectIndex) {
	t.Helper()
	v := NewValidator(schemaRoot(t))
	root := t.TempDir()
	copyDirForTest(t, fixturePath(t, "bundle-resolution", "valid-project"), root)
	index, err := v.ValidateProject(root)
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	return root, index
}

func withContextPackReadHook(t *testing.T, hook func(boundaryPath, path string, next contextPackReadProjectFileFunc) ([]byte, error), run func()) {
	t.Helper()
	original := currentContextPackReadProjectFile()
	restore := setContextPackReadProjectFileHookForTest(func(boundaryPath, path string) ([]byte, error) {
		return hook(boundaryPath, path, original)
	})
	defer restore()
	run()
}

func shouldMutateContextPackMission(path string, mutated *bool) bool {
	if *mutated || !strings.HasSuffix(filepath.ToSlash(path), "project/mission.md") {
		return false
	}
	*mutated = true
	return true
}

func writeContextPackMissionContent(t *testing.T, missionPath, content string) {
	t.Helper()
	if err := os.WriteFile(missionPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write mission file: %v", err)
	}
}

func newPermissionFailingContextPackReadHook(reads *int) func(boundaryPath, path string, next contextPackReadProjectFileFunc) ([]byte, error) {
	return func(boundaryPath, path string, next contextPackReadProjectFileFunc) ([]byte, error) {
		data, err := next(boundaryPath, path)
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(filepath.ToSlash(path), "project/mission.md") {
			*reads = *reads + 1
			if *reads >= 2 {
				return nil, os.ErrPermission
			}
		}
		return data, nil
	}
}

func newDisappearingMissionReadHook(t *testing.T, missionPath string, restoredContent []byte, removed, restored *bool) func(boundaryPath, path string, next contextPackReadProjectFileFunc) ([]byte, error) {
	t.Helper()
	return func(boundaryPath, path string, next contextPackReadProjectFileFunc) ([]byte, error) {
		data, err := next(boundaryPath, path)
		if err != nil {
			return handleContextPackMissionDisappearance(t, path, missionPath, restoredContent, removed, restored, next, boundaryPath, err)
		}
		if isContextPackMissionPath(path) && !*removed {
			removeContextPackMissionFile(t, missionPath)
			*removed = true
		}
		return data, nil
	}
}

func handleContextPackMissionDisappearance(t *testing.T, path, missionPath string, restoredContent []byte, removed, restored *bool, next contextPackReadProjectFileFunc, boundaryPath string, err error) ([]byte, error) {
	t.Helper()
	if !isContextPackMissionMissing(path, *removed, *restored, err) {
		return nil, err
	}
	writeContextPackMissionContent(t, missionPath, string(restoredContent))
	*restored = true
	return next(boundaryPath, path)
}

func isContextPackMissionPath(path string) bool {
	return strings.HasSuffix(filepath.ToSlash(path), "project/mission.md")
}

func isContextPackMissionMissing(path string, removed, restored bool, err error) bool {
	return isContextPackMissionPath(path) && removed && !restored && errors.Is(err, os.ErrNotExist)
}

func removeContextPackMissionFile(t *testing.T, missionPath string) {
	t.Helper()
	if err := os.Remove(missionPath); err != nil {
		t.Fatalf("remove mission file: %v", err)
	}
}

func assertContextPackPermissionFailure(t *testing.T, err error) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected permission error, got %v", err)
	}
	if strings.Contains(err.Error(), "changed during build") {
		t.Fatalf("expected non-transient error propagation, got %v", err)
	}
}

func nextFlappingMissionContent(flip *bool) string {
	content := "flip-a\n"
	if *flip {
		content = "flip-b\n"
	}
	*flip = !*flip
	return content
}

func assertContextPackOutputMatchesGolden(t *testing.T, report *ContextPackReport, mode ContextPackOutputMode, goldenPath string) {
	t.Helper()
	actual, err := report.Render(mode)
	if err != nil {
		t.Fatalf("render context pack report: %v", err)
	}
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("missing golden fixture %s\n%s", goldenPath, string(actual))
		}
		t.Fatalf("read golden fixture %s: %v", goldenPath, err)
	}
	if normalizeContextPackOutputForGoldenCompare(string(actual)) != normalizeContextPackOutputForGoldenCompare(string(golden)) {
		t.Fatalf("output mismatch for %s\nexpected:\n%s\nactual:\n%s", goldenPath, string(golden), string(actual))
	}
}

func normalizeContextPackOutputForGoldenCompare(content string) string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return strings.TrimSpace(normalized)
}

func assertContextPackWarningPresent(t *testing.T, warnings []ContextPackAdvisory, code string, threshold int64) {
	t.Helper()
	for _, warning := range warnings {
		if warning.Code == code {
			if warning.Threshold != threshold {
				t.Fatalf("expected warning %s threshold %d, got %d", code, threshold, warning.Threshold)
			}
			return
		}
	}
	t.Fatalf("expected warning %s in %#v", code, warnings)
}

func assertContextPackReportValidAgainstSchema(t *testing.T, v *Validator, report *ContextPackReport) {
	t.Helper()
	data, err := report.Render(ContextPackOutputModeMachine)
	if err != nil {
		t.Fatalf("render machine report: %v", err)
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("unmarshal report JSON: %v", err)
	}
	if err := v.ValidateValue("context-pack-report.schema.json", "generated-context-pack-report.json", value); err != nil {
		t.Fatalf("expected generated context pack report to satisfy schema: %v\n%s", err, string(data))
	}
}

func findContextPackSelectedFile(items []ContextPackSelectedFile, path string) *ContextPackSelectedFile {
	for i := range items {
		if items[i].Path == path {
			return &items[i]
		}
	}
	return nil
}
