package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func TestRunBundleUsageMissingSubcommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"bundle"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "result=usage_error") {
		t.Fatalf("expected usage error output, got %q", stderr.String())
	}
}

func TestRunBundleResolveRequiresBundleID(t *testing.T) {
	projectRoot := repoFixtureRoot(t, "bundle-resolution", "valid-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"bundle", "resolve", "--path", projectRoot}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "bundle resolve requires at least one bundle ID") {
		t.Fatalf("expected bundle resolve usage error, got %q", stderr.String())
	}
}

func TestRunBundleResolveSuccess(t *testing.T) {
	projectRoot := repoFixtureRoot(t, "bundle-resolution", "valid-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"bundle", "resolve", "child-reinclude", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], bundleResolveCommand; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := fields["requested_bundle_1"], "child-reinclude"; got != want {
		t.Fatalf("expected requested bundle child-reinclude, got %q", got)
	}
	if got, want := fields["resolved_bundle_1"], "base"; got != want {
		t.Fatalf("expected resolved bundle base, got %q", got)
	}
	if got, want := fields["resolved_bundle_2"], "child-reinclude"; got != want {
		t.Fatalf("expected resolved bundle child-reinclude, got %q", got)
	}
	if fields["context_pack_hash"] == "" {
		t.Fatalf("expected context_pack_hash in output, got %#v", fields)
	}
	if got, want := fields["context_pack_report_schema_version"], "1"; got != want {
		t.Fatalf("expected report schema version %q, got %q", want, got)
	}
	if got, want := fields["context_pack_id"], "child-reinclude"; got != want {
		t.Fatalf("expected context_pack_id %q, got %q", want, got)
	}
}

func TestRunBundleResolveExplainIncludesContextPackFields(t *testing.T) {
	projectRoot := repoFixtureRoot(t, "bundle-resolution", "valid-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"bundle", "resolve", "--explain", "child-reinclude", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if !strings.Contains(fields["explain_scope"], "context-pack-report") {
		t.Fatalf("expected context-pack explain scope, got %#v", fields)
	}
	if fields["explain_context_pack_warning_count"] == "" {
		t.Fatalf("expected explain_context_pack_warning_count, got %#v", fields)
	}
}

func TestRunDoctorSuccess(t *testing.T) {
	projectRoot := repoFixtureRoot(t, "bundle-resolution", "valid-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"doctor", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], doctorCommand; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if fields["warning_count"] == "" {
		t.Fatalf("expected warning count output, got %#v", fields)
	}
	if fields["project_root"] == "" {
		t.Fatalf("expected project_root in doctor output, got %#v", fields)
	}
}

func TestRunDoctorPositionalPath(t *testing.T) {
	projectRoot := repoFixtureRoot(t, "bundle-resolution", "valid-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"doctor", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], doctorCommand; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
}

func TestRunDoctorPathConflict(t *testing.T) {
	projectRoot := repoFixtureRoot(t, "bundle-resolution", "valid-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"doctor", "--path", projectRoot, "other"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "cannot use both --path and a positional path argument") {
		t.Fatalf("expected --path conflict error, got %q", stderr.String())
	}
}

func TestBuildBundleResolveContextPackReportUsesCurrentTime(t *testing.T) {
	v := contracts.NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(repoFixtureRoot(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()

	before := time.Now().UTC().Add(-1 * time.Second)
	report, err := buildBundleResolveContextPackReport(index, []string{"child-reinclude"}, false)
	if err != nil {
		t.Fatalf("build context-pack report: %v", err)
	}
	after := time.Now().UTC().Add(1 * time.Second)
	if report == nil || report.Pack == nil {
		t.Fatalf("expected report pack, got %#v", report)
	}
	generatedAt, err := time.Parse(time.RFC3339, report.Pack.GeneratedAt)
	if err != nil {
		t.Fatalf("parse generated_at: %v", err)
	}
	if generatedAt.Before(before) || generatedAt.After(after) {
		t.Fatalf("expected generated_at to be near now, got %s", generatedAt.Format(time.RFC3339))
	}
}
