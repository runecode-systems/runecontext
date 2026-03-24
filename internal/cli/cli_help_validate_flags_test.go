package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunValidateRejectsUnknownFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "--bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown validate flag") {
		t.Fatalf("expected unknown-flag output, got %q", stderr.String())
	}
}

func TestRunValidateRejectsMissingAllowedSignersPath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "--ssh-allowed-signers"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "requires a path") {
		t.Fatalf("expected missing-path output, got %q", stderr.String())
	}
}

func TestRunValidateRejectsEmptyAllowedSignersEqualsValue(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "--ssh-allowed-signers="}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "requires a path") {
		t.Fatalf("expected empty-value usage output, got %q", stderr.String())
	}
}

func TestRunValidateRejectsEmptyAllowedSignersSeparateValue(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "--ssh-allowed-signers", ""}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "requires a path") {
		t.Fatalf("expected empty separate-value usage output, got %q", stderr.String())
	}
}

func TestRunValidateRejectsBlankAllowedSignersEqualsValue(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "--ssh-allowed-signers=   "}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "requires a path") {
		t.Fatalf("expected blank equals-value usage output, got %q", stderr.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "result=usage_error") {
		t.Fatalf("expected usage result output, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "error_message=unknown command") {
		t.Fatalf("expected unknown command output, got %q", stderr.String())
	}
}

func TestRunNoCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected help exit code, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected help output, got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected help exit code, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected help output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "runectx help") {
		t.Fatalf("expected help subcommand in usage output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "help       Show CLI usage") {
		t.Fatalf("expected help command description, got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunSubcommandHelpTokens(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "change", args: []string{"change", "--help"}},
		{name: "bundle", args: []string{"bundle", "--help"}},
		{name: "standard", args: []string{"standard", "--help"}},
		{name: "init", args: []string{"init", "--help"}},
		{name: "promote", args: []string{"promote", "--help"}},
		{name: "assurance", args: []string{"assurance", "--help"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := Run(tc.args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("expected help exit code, got %d (%s)", code, stderr.String())
			}
			if !strings.Contains(stdout.String(), "usage=") {
				t.Fatalf("expected usage output, got %q", stdout.String())
			}
			if stderr.String() != "" {
				t.Fatalf("expected empty stderr, got %q", stderr.String())
			}
		})
	}
}

func TestRunSubcommandHelpRejectsExtraArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "change", args: []string{"change", "--help", "extra"}},
		{name: "bundle", args: []string{"bundle", "--help", "extra"}},
		{name: "standard", args: []string{"standard", "--help", "extra"}},
		{name: "init", args: []string{"init", "--help", "extra"}},
		{name: "promote", args: []string{"promote", "--help", "extra"}},
		{name: "assurance", args: []string{"assurance", "--help", "extra"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := Run(tc.args, &stdout, &stderr)
			if code != 2 {
				t.Fatalf("expected usage exit code, got %d", code)
			}
			if !strings.Contains(stderr.String(), "help does not accept additional arguments") {
				t.Fatalf("expected help extra-arg error, got %q", stderr.String())
			}
			if stdout.String() != "" {
				t.Fatalf("expected empty stdout, got %q", stdout.String())
			}
		})
	}
}

func TestRunStatusHelpTokens(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"status", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected help exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage=runectx status") {
		t.Fatalf("expected status usage output, got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunStatusHelpRejectsExtraArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"status", "--help", "extra"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "help does not accept additional arguments") {
		t.Fatalf("expected help extra-arg error, got %q", stderr.String())
	}
}

func TestRunValidateHelpTokens(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected help exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage=runectx validate") {
		t.Fatalf("expected validate usage output, got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunValidateHelpRejectsExtraArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", "--help", "extra"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "help does not accept additional arguments") {
		t.Fatalf("expected help extra-arg error, got %q", stderr.String())
	}
}
