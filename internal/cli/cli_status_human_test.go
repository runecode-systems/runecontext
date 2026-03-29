package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func TestRunStatusHumanOutputUsesSectionedAsciiLayout(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	firstID := runCLIChangeNewForTest(t, projectRoot, "Add cache invalidation")
	secondID := runCLIChangeNewForTest(t, projectRoot, "Revise cache invalidation")
	runCLIChangeClose(t, projectRoot, firstID, []string{"--verification-status", "skipped", "--superseded-by", secondID, "--closed-at", "2026-03-20", "--path", projectRoot})
	runCLIChangeClose(t, projectRoot, secondID, []string{"--verification-status", "passed", "--closed-at", "2026-03-21", "--path", projectRoot})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"status", projectRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("status command failed: %d (%s)", code, stderr.String())
	}
	out := stdout.String()
	shortSecondID := compactChangeID(secondID)
	shortFirstID := compactChangeID(firstID)
	for _, token := range []string{
		"RuneContext Status",
		"In Flight (0)",
		"Recently Completed (1)",
		"Replaced (1)",
		shortSecondID,
		shortFirstID,
	} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected human status output to contain %q, got:\n%s", token, out)
		}
	}
	if strings.Index(out, shortSecondID) > strings.Index(out, shortFirstID) {
		t.Fatalf("expected more recent closed entry to appear first, got:\n%s", out)
	}
	if strings.Contains(out, "|--") || strings.Contains(out, "`--") {
		t.Fatalf("expected default non-verbose output to avoid detail trees, got:\n%s", out)
	}
	if strings.Contains(out, "result=") || strings.Contains(out, "active_count=") {
		t.Fatalf("expected human renderer output, got key=value contract dump:\n%s", out)
	}
}

func TestRunStatusHistoryRecentUsesDefaultBoundedPreview(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	ids := []string{
		runCLIChangeNewForTest(t, projectRoot, "One"),
		runCLIChangeNewForTest(t, projectRoot, "Two"),
		runCLIChangeNewForTest(t, projectRoot, "Three"),
	}
	for i, id := range ids {
		date := []string{"2026-03-20", "2026-03-21", "2026-03-22"}[i]
		runCLIChangeClose(t, projectRoot, id, []string{"--verification-status", "passed", "--closed-at", date, "--path", projectRoot})
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"status", "--history-limit", "2", projectRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("status command failed: %d (%s)", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "showing 2 of 3 closed changes; use --history all to show more") {
		t.Fatalf("expected hidden-history hint, got:\n%s", out)
	}
	if strings.Contains(out, compactChangeID(ids[0])) {
		t.Fatalf("expected oldest closed entry to be hidden by default recent mode, got:\n%s", out)
	}
}

func TestRunStatusHistoryAllShowsAllEntries(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	ids := []string{
		runCLIChangeNewForTest(t, projectRoot, "One"),
		runCLIChangeNewForTest(t, projectRoot, "Two"),
	}
	for i, id := range ids {
		date := []string{"2026-03-20", "2026-03-21"}[i]
		runCLIChangeClose(t, projectRoot, id, []string{"--verification-status", "passed", "--closed-at", date, "--path", projectRoot})
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"status", "--history", "all", projectRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("status command failed: %d (%s)", code, stderr.String())
	}
	out := stdout.String()
	for _, id := range ids {
		if !strings.Contains(out, compactChangeID(id)) {
			t.Fatalf("expected history all to include %q, got:\n%s", id, out)
		}
	}
	if strings.Contains(out, "showing ") {
		t.Fatalf("expected no hidden-history hint in all mode, got:\n%s", out)
	}
}

func TestRunStatusHistoryNoneHidesHistoricalSections(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	id := runCLIChangeNewForTest(t, projectRoot, "One")
	runCLIChangeClose(t, projectRoot, id, []string{"--verification-status", "passed", "--closed-at", "2026-03-20", "--path", projectRoot})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"status", "--history", "none", projectRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("status command failed: %d (%s)", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Recently Completed (0)") {
		t.Fatalf("expected hidden closed section count, got:\n%s", out)
	}
	if !strings.Contains(out, "showing 0 of 1 closed changes; use --history all to show more") {
		t.Fatalf("expected hidden-history hint for none mode, got:\n%s", out)
	}
}

func TestRunStatusVerboseShowsRelationshipDetails(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	firstID := runCLIChangeNewForTest(t, projectRoot, "Add cache invalidation")
	secondID := runCLIChangeNewForTest(t, projectRoot, "Revise cache invalidation")
	runCLIChangeClose(t, projectRoot, firstID, []string{"--verification-status", "skipped", "--superseded-by", secondID, "--closed-at", "2026-03-20", "--path", projectRoot})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"status", "--verbose", projectRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("status command failed: %d (%s)", code, stderr.String())
	}
	out := stdout.String()
	for _, token := range []string{"superseded by:", "path:"} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected verbose tree details token %q, got:\n%s", token, out)
		}
	}
}

func TestRenderHumanStatusColorToggleIsDeterministic(t *testing.T) {
	projectRoot := fixtureRoot(t, "valid-project")
	absRoot, validator, loaded, err := loadProjectForCLI(projectRoot, true)
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	defer loaded.Close()

	summary, err := contracts.BuildProjectStatusSummary(validator, loaded)
	if err != nil {
		t.Fatalf("build status summary: %v", err)
	}

	ascii := renderHumanStatus(absRoot, loaded, summary, statusRenderOptions{color: false})
	if strings.Contains(ascii, "\x1b[") {
		t.Fatalf("expected ASCII-only rendering without ANSI escapes, got:\n%s", ascii)
	}
	colored := renderHumanStatus(absRoot, loaded, summary, statusRenderOptions{color: true})
	if !strings.Contains(colored, "\x1b[") {
		t.Fatalf("expected ANSI escapes when color is enabled, got:\n%s", colored)
	}
}

func TestBuildStatusSummaryProvidesRelationshipMetadataForRenderer(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, fixtureRoot(t, "valid-project"), root)
	absRoot, validator, loaded, err := loadProjectForCLI(root, true)
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	defer loaded.Close()
	if _, err := contracts.CloseChange(validator, loaded, "CHG-2026-001-a3f2-auth-gateway", contracts.ChangeCloseOptions{
		VerificationStatus: "skipped",
		ClosedAt:           time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		SupersededBy:       []string{"CHG-2026-002-b4c3-auth-revision"},
	}); err != nil {
		t.Fatalf("close change: %v", err)
	}
	loaded.Close()

	_, _, reloaded, err := loadProjectForCLI(absRoot, true)
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	defer reloaded.Close()
	summary, err := contracts.BuildProjectStatusSummary(validator, reloaded)
	if err != nil {
		t.Fatalf("build summary: %v", err)
	}
	out := renderHumanStatus(absRoot, reloaded, summary, statusRenderOptions{color: false, verbose: true})
	for _, token := range []string{
		"depends on:",
		"related:",
		"superseded by:",
		"created:",
		"closed:",
	} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected rendered relationship/recency token %q, got:\n%s", token, out)
		}
	}
}

func TestRenderHumanStatusNestsProjectAssociationsInFlight(t *testing.T) {
	summary := nestedProjectAssociationsSummary()
	out := renderHumanStatus("/tmp/project", nil, summary, statusRenderOptions{color: false})
	assertNestedProjectAssociations(t, out)
}

func nestedProjectAssociationsSummary() *contracts.ProjectStatusSummary {
	return &contracts.ProjectStatusSummary{
		SelectedConfigPath: "/tmp/runecontext.yaml",
		RuneContextVersion: "0.1.0-alpha.10",
		AssuranceTier:      "verified",
		Active:             nestedProjectAssociationEntries(),
	}
}

func nestedProjectAssociationEntries() []contracts.ChangeStatusEntry {
	dependentEntries := nestedProjectAssociationDependentEntries()
	return []contracts.ChangeStatusEntry{
		{
			ID:             "CHG-2026-010-97a3-enhance-human-friendly-status-ux-and-scaling",
			Title:          "Enhance human-friendly status UX and scaling",
			Type:           "project",
			Size:           "large",
			Status:         "implemented",
			RelatedChanges: []string{"CHG-2026-011", "CHG-2026-012", "CHG-2026-013"},
		},
		{
			ID:             "CHG-2026-011",
			Title:          "Summary metadata expansion",
			Type:           "feature",
			Size:           "medium",
			Status:         "implemented",
			RelatedChanges: []string{"CHG-2026-010-97a3-enhance-human-friendly-status-ux-and-scaling"},
		},
		dependentEntries[0],
		dependentEntries[1],
	}
}

func nestedProjectAssociationDependentEntries() []contracts.ChangeStatusEntry {
	return []contracts.ChangeStatusEntry{
		{
			ID:             "CHG-2026-012",
			Title:          "Human renderer",
			Type:           "feature",
			Size:           "medium",
			Status:         "implemented",
			RelatedChanges: []string{"CHG-2026-010-97a3-enhance-human-friendly-status-ux-and-scaling"},
			DependsOn:      []string{"CHG-2026-011"},
		},
		{
			ID:             "CHG-2026-013",
			Title:          "History controls",
			Type:           "feature",
			Size:           "medium",
			Status:         "implemented",
			RelatedChanges: []string{"CHG-2026-010-97a3-enhance-human-friendly-status-ux-and-scaling"},
			DependsOn:      []string{"CHG-2026-011", "CHG-2026-012"},
		},
	}
}

func assertNestedProjectAssociations(t *testing.T, out string) {
	t.Helper()
	if !strings.Contains(out, "- [implemented] CHG-2026-010 [project large]") {
		t.Fatalf("expected project umbrella root row, got:\n%s", out)
	}
	if !strings.Contains(out, "| Enhance human-friendly status UX and scaling") {
		t.Fatalf("expected project title continuation line, got:\n%s", out)
	}
	for _, token := range []string{"|- [implemented] CHG-2026-011", "|- [implemented] CHG-2026-012", "\\- [implemented] CHG-2026-013"} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected nested active association token %q, got:\n%s", token, out)
		}
	}
	assertNestedProjectAssociationOrder(t, out)
}

func assertNestedProjectAssociationOrder(t *testing.T, out string) {
	t.Helper()
	idx011 := strings.Index(out, "CHG-2026-011")
	idx012 := strings.Index(out, "CHG-2026-012")
	idx013 := strings.Index(out, "CHG-2026-013")
	if !(idx011 >= 0 && idx012 > idx011 && idx013 > idx012) {
		t.Fatalf("expected dependency-respecting nested order 011->012->013, got:\n%s", out)
	}
}

func TestRenderHumanStatusVerboseUsesFullID(t *testing.T) {
	summary := &contracts.ProjectStatusSummary{
		SelectedConfigPath: "/tmp/runecontext.yaml",
		RuneContextVersion: "0.1.0-alpha.10",
		AssuranceTier:      "verified",
		Active: []contracts.ChangeStatusEntry{{
			ID:     "CHG-2026-010-97a3-enhance-human-friendly-status-ux-and-scaling",
			Title:  "Enhance human-friendly status UX and scaling",
			Type:   "project",
			Size:   "large",
			Status: "implemented",
		}},
	}
	compact := renderHumanStatus("/tmp/project", nil, summary, statusRenderOptions{color: false})
	if !strings.Contains(compact, "CHG-2026-010 [project large]") {
		t.Fatalf("expected compact ID in default output, got:\n%s", compact)
	}
	verbose := renderHumanStatus("/tmp/project", nil, summary, statusRenderOptions{color: false, verbose: true})
	if !strings.Contains(verbose, "CHG-2026-010-97a3-enhance-human-friendly-status-ux-and-scaling [project large]") {
		t.Fatalf("expected full change ID in verbose output, got:\n%s", verbose)
	}
}

func TestRenderHumanStatusWrapsLongTitleOnContinuationLines(t *testing.T) {
	summary := &contracts.ProjectStatusSummary{
		SelectedConfigPath: "/tmp/runecontext.yaml",
		RuneContextVersion: "0.1.0-alpha.10",
		AssuranceTier:      "verified",
		Active: []contracts.ChangeStatusEntry{{
			ID:     "CHG-2026-013-1f97-add-progressive-disclosure-and-history-controls-to-status",
			Type:   "feature",
			Size:   "medium",
			Status: "implemented",
			Title:  "Here is an example of a super long title. Here is an example of a super long title. Here is an example of a super long title.",
		}},
	}
	out := renderHumanStatus("/tmp/project", nil, summary, statusRenderOptions{color: false})
	if !strings.Contains(out, "- [implemented] CHG-2026-013 [feature medium]") {
		t.Fatalf("expected first-line status row, got:\n%s", out)
	}
	if strings.Count(out, "Here is an example of a super long title") < 2 {
		t.Fatalf("expected wrapped continuation title lines, got:\n%s", out)
	}
	if !strings.Contains(out, "| Here is an example of a super long title") {
		t.Fatalf("expected continuation lines with detail prefix, got:\n%s", out)
	}
}

func TestRenderHumanStatusHintLinesUseCompactIDsByDefault(t *testing.T) {
	summary := &contracts.ProjectStatusSummary{
		SelectedConfigPath: "/tmp/runecontext.yaml",
		RuneContextVersion: "0.1.0-alpha.10",
		AssuranceTier:      "verified",
		Active: []contracts.ChangeStatusEntry{{
			ID:             "CHG-2026-013-1f97-add-progressive-disclosure-and-history-controls-to-status",
			Type:           "feature",
			Size:           "medium",
			Status:         "implemented",
			Title:          "History controls",
			DependsOn:      []string{"CHG-2026-011-d50b-extend-status-summaries-with-relationship-and-recency-metadata", "CHG-2026-012-f67a-add-human-friendly-status-rendering-with-ascii-hierarchy-and-color"},
			RelatedChanges: []string{"CHG-2026-010-97a3-enhance-human-friendly-status-ux-and-scaling"},
		}},
	}
	out := renderHumanStatus("/tmp/project", nil, summary, statusRenderOptions{color: false})
	if !strings.Contains(out, "depends on: CHG-2026-011, CHG-2026-012") {
		t.Fatalf("expected compact dependency IDs in default hint lines, got:\n%s", out)
	}
	if strings.Contains(out, "CHG-2026-011-d50b") || strings.Contains(out, "CHG-2026-012-f67a") {
		t.Fatalf("expected default hint lines to avoid full IDs, got:\n%s", out)
	}
}

func TestRenderHumanStatusHintLinesUseFullIDsInVerbose(t *testing.T) {
	summary := &contracts.ProjectStatusSummary{
		SelectedConfigPath: "/tmp/runecontext.yaml",
		RuneContextVersion: "0.1.0-alpha.10",
		AssuranceTier:      "verified",
		Active: []contracts.ChangeStatusEntry{{
			ID:        "CHG-2026-013-1f97-add-progressive-disclosure-and-history-controls-to-status",
			Type:      "feature",
			Size:      "medium",
			Status:    "implemented",
			Title:     "History controls",
			DependsOn: []string{"CHG-2026-011-d50b-extend-status-summaries-with-relationship-and-recency-metadata", "CHG-2026-012-f67a-add-human-friendly-status-rendering-with-ascii-hierarchy-and-color"},
		}},
	}
	out := renderHumanStatus("/tmp/project", nil, summary, statusRenderOptions{color: false, verbose: true})
	if !strings.Contains(out, "depends on: CHG-2026-011-d50b-extend-status-summaries-with-relationship-and-recency-metadata") {
		t.Fatalf("expected full IDs in verbose hint lines, got:\n%s", out)
	}
}

func TestRenderHumanStatusWrapsLongHintLines(t *testing.T) {
	summary := &contracts.ProjectStatusSummary{
		SelectedConfigPath: "/tmp/runecontext.yaml",
		RuneContextVersion: "0.1.0-alpha.10",
		AssuranceTier:      "verified",
		Active: []contracts.ChangeStatusEntry{{
			ID:        "CHG-2026-003-5f38-teach-adapters-to-run-guided-clarification-and-decomposition-flows",
			Type:      "feature",
			Size:      "large",
			Status:    "proposed",
			Title:     "Teach adapters to run guided clarification and decomposition flows",
			DependsOn: []string{"CHG-2026-001-fdc1-add-advisory-intake-and-decomposition-assessment-commands", "CHG-2026-004-5b03-add-first-class-change-decomposition-planning-and-apply-operations", "CHG-2026-005-d9ca-add-structured-change-update-command-for-safe-status-and-relationship-edits"},
		}},
	}
	out := renderHumanStatus("/tmp/project", nil, summary, statusRenderOptions{color: false})
	if strings.Count(out, "depends on:") != 1 {
		t.Fatalf("expected one dependency hint label with wrapped continuation lines, got:\n%s", out)
	}
	if !strings.Contains(out, "CHG-2026-001,") || !strings.Contains(out, "CHG-2026-004,") || !strings.Contains(out, "CHG-2026-005") {
		t.Fatalf("expected all compact dependency IDs in wrapped hint output, got:\n%s", out)
	}
}

func TestRenderHumanStatusFallsBackToFlatRowsWhenAssociationAmbiguous(t *testing.T) {
	summary := &contracts.ProjectStatusSummary{
		SelectedConfigPath: "/tmp/runecontext.yaml",
		RuneContextVersion: "0.1.0-alpha.10",
		AssuranceTier:      "verified",
		Active: []contracts.ChangeStatusEntry{
			{
				ID:             "CHG-PROJECT-A",
				Title:          "Project A",
				Type:           "project",
				Size:           "large",
				RelatedChanges: []string{"CHG-FEATURE-1"},
			},
			{
				ID:             "CHG-PROJECT-B",
				Title:          "Project B",
				Type:           "project",
				Size:           "large",
				RelatedChanges: []string{"CHG-FEATURE-1"},
			},
			{
				ID:             "CHG-FEATURE-1",
				Title:          "Shared child",
				Type:           "feature",
				Size:           "medium",
				RelatedChanges: []string{"CHG-PROJECT-A", "CHG-PROJECT-B"},
			},
		},
	}
	out := renderHumanStatus("/tmp/project", nil, summary, statusRenderOptions{color: false})
	if strings.Contains(out, "|- CHG-FEATURE-1") || strings.Contains(out, "\\- CHG-FEATURE-1") {
		t.Fatalf("expected ambiguous association fallback to avoid forced tree nesting, got:\n%s", out)
	}
	if !strings.Contains(out, "related: CHG-PROJECT-A, CHG-PROJECT-B") {
		t.Fatalf("expected explicit relationship hints during fallback, got:\n%s", out)
	}
}

func TestRenderHumanStatusFallsBackToFlatRowsWhenRelationshipsCycle(t *testing.T) {
	summary := &contracts.ProjectStatusSummary{
		SelectedConfigPath: "/tmp/runecontext.yaml",
		RuneContextVersion: "0.1.0-alpha.10",
		AssuranceTier:      "verified",
		Active: []contracts.ChangeStatusEntry{
			{
				ID:             "CHG-PROJECT-A",
				Title:          "Project A",
				Type:           "project",
				Size:           "large",
				RelatedChanges: []string{"CHG-PROJECT-B"},
			},
			{
				ID:             "CHG-PROJECT-B",
				Title:          "Project B",
				Type:           "project",
				Size:           "large",
				RelatedChanges: []string{"CHG-PROJECT-A"},
			},
		},
	}
	out := renderHumanStatus("/tmp/project", nil, summary, statusRenderOptions{color: false})
	for _, token := range []string{"CHG-PROJECT-A", "CHG-PROJECT-B"} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected cyclic fallback output to include %q, got:\n%s", token, out)
		}
	}
	if strings.Contains(out, "|- [") || strings.Contains(out, "\\- [") {
		t.Fatalf("expected cyclic graph fallback to avoid nested tree connectors, got:\n%s", out)
	}
	if !strings.Contains(out, "related: CHG-PROJECT-A") || !strings.Contains(out, "related: CHG-PROJECT-B") {
		t.Fatalf("expected fallback relationship hints for cyclic graph, got:\n%s", out)
	}
}

func TestRenderHumanStatusStripsControlSequencesFromUserFields(t *testing.T) {
	summary := &contracts.ProjectStatusSummary{
		SelectedConfigPath: "/tmp/runecontext.yaml\x1b[32m",
		RuneContextVersion: "0.1.0-alpha.10\x1b[31m",
		AssuranceTier:      "verified\x1b[1m",
		BundleIDs:          []string{"core\x1b[34m"},
		Active: []contracts.ChangeStatusEntry{{
			ID:        "CHG-2026-013-unsafe\x1b[31m",
			Type:      "feature",
			Size:      "medium",
			Status:    "implemented\x1b[31m",
			Title:     "Unsafe\x1b[31m title",
			Path:      "changes/unsafe\x1b[31m/path.md",
			DependsOn: []string{"CHG-2026-001-safe\x1b[31m"},
		}},
	}
	out := renderHumanStatus("/tmp/project\x1b[35m", nil, summary, statusRenderOptions{color: false, verbose: true})
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("expected control sequences to be stripped from user fields, got:\n%s", out)
	}
	for _, token := range []string{"Unsafe title", "changes/unsafe/path.md", "depends on: CHG-2026-001-safe"} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected sanitized output to retain visible content %q, got:\n%s", token, out)
		}
	}
}
