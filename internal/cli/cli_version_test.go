package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunVersionCommandOutputsNormalizedVersion(t *testing.T) {
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v1.2.3-test"

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"version"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["result"], "ok"; got != want {
		t.Fatalf("expected result %q, got %q", want, got)
	}
	if got, want := fields["command"], "version"; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := fields["version"], "1.2.3-test"; got != want {
		t.Fatalf("expected version %q, got %q", want, got)
	}
	if got, want := fields["runecontext_version"], "1.2.3-test"; got != want {
		t.Fatalf("expected runecontext_version %q, got %q", want, got)
	}
}

func TestRunVersionAliasesMatchVersionCommand(t *testing.T) {
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v9.9.9-alias"

	assertVersionAliasOutput(t, []string{"version"}, "9.9.9-alias")
	assertVersionAliasOutput(t, []string{"--version"}, "9.9.9-alias")
	assertVersionAliasOutput(t, []string{"-v"}, "9.9.9-alias")
}

func assertVersionAliasOutput(t *testing.T, args []string, wantVersion string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], "version"; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got := fields["version"]; got != wantVersion {
		t.Fatalf("expected normalized version %q, got %q", wantVersion, got)
	}
}

func TestRunVersionJSONEnvelope(t *testing.T) {
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v0.1.0-alpha.8"

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"version", "--json", "--non-interactive"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
	var envelope machineEnvelope
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &envelope); err != nil {
		t.Fatalf("unmarshal json output: %v", err)
	}
	if got, want := envelope.Command, "version"; got != want {
		t.Fatalf("expected envelope command %q, got %q", want, got)
	}
	if got, want := envelope.Data["result"], "ok"; got != want {
		t.Fatalf("expected result %q, got %q", want, got)
	}
	if got, want := envelope.Data["version"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected normalized version %q, got %q", want, got)
	}
	if got, want := envelope.Data["runecontext_version"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected runecontext_version %q, got %q", want, got)
	}
	if got, want := envelope.Data["non_interactive"], "true"; got != want {
		t.Fatalf("expected non_interactive %q, got %q", want, got)
	}
}

func TestRunVersionRejectsUnsupportedMachineFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "explain", args: []string{"version", "--explain"}},
		{name: "dry-run", args: []string{"version", "--dry-run"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := Run(tc.args, &stdout, &stderr)
			if code != exitUsage {
				t.Fatalf("expected usage exit code, got %d", code)
			}
			if stdout.String() != "" {
				t.Fatalf("expected empty stdout, got %q", stdout.String())
			}
			if !strings.Contains(stderr.String(), "usage="+versionUsage) {
				t.Fatalf("expected version usage output, got %q", stderr.String())
			}
		})
	}
}
