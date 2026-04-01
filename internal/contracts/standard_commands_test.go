package contracts

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestListStandardsFiltersByScopeFocusAndStatus(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()

	result, err := ListStandards(v, loaded, StandardListOptions{
		ScopePaths: []string{"security"},
		Focus:      "review",
		Statuses:   []StandardStatus{StandardStatusActive},
	})
	if err != nil {
		t.Fatalf("list standards: %v", err)
	}
	if got, want := len(result.Standards), 1; got != want {
		t.Fatalf("expected %d standards, got %d", want, got)
	}
	if got, want := result.Standards[0].Path, "standards/security/review.md"; got != want {
		t.Fatalf("expected standard path %q, got %q", want, got)
	}

	empty, err := ListStandards(v, loaded, StandardListOptions{Statuses: []StandardStatus{StandardStatusDeprecated}})
	if err != nil {
		t.Fatalf("list deprecated standards: %v", err)
	}
	if got := len(empty.Standards); got != 0 {
		t.Fatalf("expected zero deprecated standards, got %d", got)
	}
}

func TestCreateStandardWritesFileAndReturnsMutation(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()

	result, err := CreateStandard(v, loaded, StandardCreateOptions{
		Path:   "custom/authoring",
		Title:  "Authoring Standard",
		Status: StandardStatusDraft,
		Body:   "# Authoring Standard\n\nKeep standard authoring workflows explicit.",
	})
	if err != nil {
		t.Fatalf("create standard: %v", err)
	}
	if got, want := result.Path, "standards/custom/authoring.md"; got != want {
		t.Fatalf("expected standard path %q, got %q", want, got)
	}
	if len(result.ChangedFiles) != 1 || result.ChangedFiles[0].Action != "created" {
		t.Fatalf("expected one created file mutation, got %#v", result.ChangedFiles)
	}

	assertValidatedWorkflowProject(t, v, root)
	path := filepath.Join(root, "runecontext", filepath.FromSlash(result.Path))
	text := strings.ReplaceAll(string(mustReadBytes(t, path)), "\r\n", "\n")
	if !strings.Contains(text, "id: custom/authoring") {
		t.Fatalf("expected inferred id in created standard, got:\n%s", text)
	}
	if !strings.Contains(text, "status: draft") {
		t.Fatalf("expected draft status in created standard, got:\n%s", text)
	}
}

func TestUpdateStandardRewritesFrontmatterFields(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()

	if _, err := CreateStandard(v, loaded, StandardCreateOptions{Path: "custom/authoring", Title: "Authoring Standard", Status: StandardStatusDraft}); err != nil {
		t.Fatalf("create standard: %v", err)
	}

	updated, err := UpdateStandard(v, loaded, StandardUpdateOptions{
		Path:                           "standards/custom/authoring.md",
		Title:                          "Authoring Standard v2",
		Status:                         string(StandardStatusActive),
		ReplaceAliases:                 true,
		Aliases:                        []string{"custom/authoring-v1"},
		ReplaceSuggestedContextBundles: true,
		SuggestedContextBundles:        []string{"base"},
	})
	if err != nil {
		t.Fatalf("update standard: %v", err)
	}
	if len(updated.ChangedFiles) != 1 || updated.ChangedFiles[0].Action != "updated" {
		t.Fatalf("expected one updated file mutation, got %#v", updated.ChangedFiles)
	}

	assertValidatedWorkflowProject(t, v, root)
	path := filepath.Join(root, "runecontext", filepath.FromSlash(updated.Path))
	text := strings.ReplaceAll(string(mustReadBytes(t, path)), "\r\n", "\n")
	for _, expected := range []string{"title: Authoring Standard v2", "status: active", "aliases:", "- custom/authoring-v1", "suggested_context_bundles:", "- base"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected %q in updated standard, got:\n%s", expected, text)
		}
	}
}

func TestUpdateStandardRejectsReplacedByWithoutDeprecatedStatus(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()

	if _, err := CreateStandard(v, loaded, StandardCreateOptions{Path: "custom/authoring", Title: "Authoring Standard", Status: StandardStatusActive}); err != nil {
		t.Fatalf("create standard: %v", err)
	}

	_, err := UpdateStandard(v, loaded, StandardUpdateOptions{
		Path:       "standards/custom/authoring.md",
		ReplacedBy: "standards/global/base.md",
	})
	if err == nil || !strings.Contains(err.Error(), "--replaced-by requires --status deprecated") {
		t.Fatalf("expected replaced_by status rejection, got %v", err)
	}
}
