package contracts

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBuildGeneratedChangesByStatusIndexMatchesGolden(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "traceability", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()

	changeIndex, err := index.BuildGeneratedChangesByStatusIndex()
	if err != nil {
		t.Fatalf("build changes-by-status index: %v", err)
	}
	assertGeneratedArtifactValidAgainstSchema(t, v, "changes-by-status-index.schema.json", "generated-changes-by-status.yaml", changeIndex)
	assertGeneratedArtifactMatchesGolden(t, changeIndex, fixturePath(t, "generated-indexes", "golden", "traceability-changes-by-status.yaml"))
}

func TestBuildGeneratedChangesByStatusDeterministic(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "traceability", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()

	first, err := index.BuildGeneratedChangesByStatusIndex()
	if err != nil {
		t.Fatalf("build first changes-by-status index: %v", err)
	}
	second, err := index.BuildGeneratedChangesByStatusIndex()
	if err != nil {
		t.Fatalf("build second changes-by-status index: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected deterministic changes-by-status output\nfirst: %#v\nsecond: %#v", first, second)
	}
}

func TestBuildGeneratedChangesByStatusIndexRejectsUnknownLifecycleStatus(t *testing.T) {
	root := t.TempDir()
	index := &ProjectIndex{
		ContentRoot: root,
		Changes: map[string]*ChangeRecord{
			"CHG-2026-123-a1b2-unknown": {
				ID:         "CHG-2026-123-a1b2-unknown",
				Title:      "Unknown lifecycle",
				Type:       "feature",
				Status:     LifecycleStatus("unknown"),
				StatusPath: filepath.Join(root, "changes", "CHG-2026-123-a1b2-unknown", "status.yaml"),
			},
		},
	}
	_, err := index.BuildGeneratedChangesByStatusIndex()
	if err == nil || !strings.Contains(err.Error(), "unsupported lifecycle status") {
		t.Fatalf("expected unsupported lifecycle status failure, got %v", err)
	}
}

func TestBuildGeneratedChangesByStatusIndexRejectsPathOutsideContentRoot(t *testing.T) {
	root := t.TempDir()
	outsideRoot := t.TempDir()
	index := &ProjectIndex{
		ContentRoot: root,
		Changes: map[string]*ChangeRecord{
			"CHG-2026-124-a1b2-outside": {
				ID:         "CHG-2026-124-a1b2-outside",
				Title:      "Outside path",
				Type:       "feature",
				Status:     StatusProposed,
				StatusPath: filepath.Join(outsideRoot, "changes", "CHG-2026-124-a1b2-outside", "status.yaml"),
			},
		},
	}
	_, err := index.BuildGeneratedChangesByStatusIndex()
	if err == nil || !strings.Contains(err.Error(), "escapes RuneContext content root") {
		t.Fatalf("expected out-of-root path rejection, got %v", err)
	}
}

func TestChangesByStatusSchemaRejectsUnknownFields(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "traceability", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	changesByStatus, err := index.BuildGeneratedChangesByStatusIndex()
	if err != nil {
		t.Fatalf("build changes index: %v", err)
	}
	changesData, err := yaml.Marshal(changesByStatus)
	if err != nil {
		t.Fatalf("marshal changes index: %v", err)
	}
	changesValue, err := parseYAML(changesData)
	if err != nil {
		t.Fatalf("parse changes index yaml: %v", err)
	}
	changesMap := changesValue.(map[string]any)
	statuses := changesMap["statuses"].(map[string]any)
	proposed := statuses["proposed"].([]any)
	if len(proposed) == 0 {
		t.Fatal("expected fixture proposed status entries")
	}
	proposedEntry := proposed[0].(map[string]any)
	proposedEntry["unexpected"] = true
	if err := v.ValidateValue("changes-by-status-index.schema.json", "generated-changes-by-status.yaml", changesMap); err == nil {
		t.Fatal("expected changes-by-status schema to reject unknown entry fields")
	}
}
