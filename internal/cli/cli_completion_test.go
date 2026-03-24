package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestRunCompletionUsageAndErrors(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if code := Run([]string{"completion"}, &stdout, &stderr); code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "usage="+completionUsage) {
		t.Fatalf("expected completion usage output, got %q", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"completion", "pwsh"}, &stdout, &stderr); code != exitUsage {
		t.Fatalf("expected usage exit code for unsupported shell, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unsupported shell") {
		t.Fatalf("expected unsupported-shell output, got %q", stderr.String())
	}
}

func TestRunCompletionHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if code := Run([]string{"completion", "--help"}, &stdout, &stderr); code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage="+completionUsage) {
		t.Fatalf("expected completion help usage, got %q", stdout.String())
	}
}

func TestCompletionScriptsGolden(t *testing.T) {
	tests := []string{"bash", "zsh", "fish"}
	for _, shell := range tests {
		t.Run(shell, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			if code := Run([]string{"completion", shell}, &stdout, &stderr); code != exitOK {
				t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
			}
			goldenPath := repoFixtureRoot(t, "cli-completion", shell+".golden")
			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden completion script: %v", err)
			}
			if normalizeNewlines(string(expected)) != normalizeNewlines(stdout.String()) {
				t.Fatalf("unexpected completion script for %s\nexpected:\n%s\nactual:\n%s", shell, string(expected), stdout.String())
			}
		})
	}
}

type errWriter struct{ err error }

func (w errWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestRunCompletionWriteFailureOmitsRoot(t *testing.T) {
	var stderr bytes.Buffer
	wantErr := errors.New("write failed")

	code := Run([]string{"completion", "bash"}, errWriter{err: wantErr}, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "result=invalid") {
		t.Fatalf("expected invalid result output, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "command=completion") {
		t.Fatalf("expected completion command output, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "error_message=write failed") {
		t.Fatalf("expected write failure output, got %q", stderr.String())
	}
	if strings.Contains(stderr.String(), "root=") {
		t.Fatalf("expected completion invalid output to omit root, got %q", stderr.String())
	}
}

func TestCompletionIndexIncludesRootSubcommands(t *testing.T) {
	index := buildCompletionIndex(CommandMetadataRegistry())
	if len(index.subcommandsByPath[""]) == 0 {
		t.Fatalf("expected root completion subcommands to be populated")
	}
	if !slices.Equal(index.subcommandsByPath[""], index.topCommands) {
		t.Fatalf("expected root subcommands to match top commands, got %#v want %#v", index.subcommandsByPath[""], index.topCommands)
	}

	rootCase := "    '') echo '" + strings.Join(index.topCommands, " ") + "' ;;"
	bashScript := buildBashCompletionScript(index)
	if !strings.Contains(bashScript, rootCase) {
		t.Fatalf("expected bash completion script root case %q", rootCase)
	}
}

func TestCompletionMetadataSurfaceParity(t *testing.T) {
	registry := CommandMetadataRegistry()
	assertUsageParity(t, flattenCommandPaths(registry.Commands))
	enums := collectEnumFlagValues()
	assertEnumValues(t, enums, "init|--mode", []string{"embedded", "linked"})
	assertEnumValues(t, enums, "change new|--type", []string{"bug", "chore", "feature", "project", "standard"})
	assertEnumValues(t, enums, "change new|--size", []string{"large", "medium", "small"})
	assertEnumValues(t, enums, "change close|--verification-status", []string{"failed", "passed", "pending", "skipped"})
}

func assertUsageParity(t *testing.T, paths []string) {
	t.Helper()
	for _, path := range paths {
		if path == "help" {
			continue
		}
		usage := usageByPath(path)
		if usage == "" {
			t.Fatalf("missing usage metadata for path %q", path)
		}
		stderr := runUnknownFlagProbe(t, path)
		if !strings.Contains(stderr, "usage="+usage) {
			t.Fatalf("expected usage parity for %q, got %q", path, stderr)
		}
	}
}

func runUnknownFlagProbe(t *testing.T, path string) string {
	t.Helper()
	args := append(strings.Fields(path), "--unknown-registry-parity-flag")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code for %q parity probe, got %d (%s)", path, code, stderr.String())
	}
	return stderr.String()
}

func collectEnumFlagValues() map[string][]string {
	enums := map[string][]string{}
	for _, flag := range CompletionMetadataRegistry().Flags {
		if flag.ValueKind != ValueKindEnum {
			continue
		}
		enums[flag.CommandPath+"|"+flag.Name] = append([]string(nil), flag.EnumValues...)
	}
	return enums
}

func TestCompletionMetadataIncludesPositionalEnums(t *testing.T) {
	metadata := CompletionMetadataRegistry()
	checks := map[string][]string{}
	for _, positional := range metadata.PositionalEnums {
		checks[fmt.Sprintf("%s|%d", positional.CommandPath, positional.Position)] = positional.EnumValues
	}
	if !slices.Equal(checks["completion|1"], []string{"bash", "fish", "zsh"}) {
		t.Fatalf("expected completion shell positional enums, got %#v", checks["completion|1"])
	}
	if !slices.Equal(checks["assurance enable|1"], []string{"verified"}) {
		t.Fatalf("expected assurance enable mode enum, got %#v", checks["assurance enable|1"])
	}
}

func TestCompletionScriptUsesRegistryBinary(t *testing.T) {
	registry := CommandMetadataRegistry()
	registry.Binary = "runectx custom"
	quotedBinary := shellSingleQuote(registry.Binary)

	bashScript, err := generateCompletionScript(registry, "bash")
	if err != nil {
		t.Fatalf("generate bash completion script: %v", err)
	}
	if !strings.Contains(bashScript, "complete -F _runectx_complete "+quotedBinary) {
		t.Fatalf("expected bash completion registration to use custom binary, got %q", bashScript)
	}

	zshScript, err := generateCompletionScript(registry, "zsh")
	if err != nil {
		t.Fatalf("generate zsh completion script: %v", err)
	}
	if !strings.Contains(zshScript, "complete -F _runectx_complete "+quotedBinary) {
		t.Fatalf("expected zsh completion registration to use custom binary")
	}

	fishScript, err := generateCompletionScript(registry, "fish")
	if err != nil {
		t.Fatalf("generate fish completion script: %v", err)
	}
	if !strings.Contains(fishScript, "complete -c "+quotedBinary+" -f -n '__fish_use_subcommand'") {
		t.Fatalf("expected fish completion script to use custom binary")
	}
}

func TestBashCompletionSupportsFlagEqualsForm(t *testing.T) {
	registry := MetadataRegistry{
		Binary: "runectx",
		Commands: []CommandMetadata{
			{
				Name:  "test",
				Path:  "test",
				Usage: "runectx test",
				Flags: []FlagMetadata{
					{Name: "--mode", Value: enumValueSpec("fast", "safe")},
				},
			},
		},
	}

	script, err := generateCompletionScript(registry, "bash")
	if err != nil {
		t.Fatalf("generate bash completion script: %v", err)
	}

	checks := []string{
		"token_flag=\"${token%%=*}\"",
		"if _runectx_flag_takes_value \"$cmd\" \"$token_flag\"; then",
		"if [[ -n \"$token_inline\" ]]; then",
		"if [[ \"$cur\" == *=* ]]; then",
		"enums=$(_runectx_enum_for_flag \"$cmd|$cur_flag\")",
		"COMPREPLY[$idx]=\"$cur_flag=${COMPREPLY[$idx]}\"",
	}
	for _, check := range checks {
		if !strings.Contains(script, check) {
			t.Fatalf("expected bash completion script to contain %q", check)
		}
	}
}

func assertEnumValues(t *testing.T, values map[string][]string, key string, want []string) {
	t.Helper()
	got, ok := values[key]
	if !ok {
		t.Fatalf("expected enum metadata key %q", key)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected enum metadata for %s: got %#v want %#v", key, got, want)
	}
}

func flattenCommandPaths(commands []CommandMetadata) []string {
	paths := make([]string, 0, len(commands))
	for _, command := range commands {
		paths = append(paths, command.Path)
		paths = append(paths, flattenCommandPaths(command.Subcommands)...)
	}
	return paths
}

func usageByPath(path string) string {
	for _, command := range flattenCommands(CommandMetadataRegistry().Commands) {
		if command.Path == path {
			return command.Usage
		}
	}
	return ""
}

func flattenCommands(commands []CommandMetadata) []CommandMetadata {
	items := make([]CommandMetadata, 0, len(commands))
	for _, command := range commands {
		items = append(items, command)
		items = append(items, flattenCommands(command.Subcommands)...)
	}
	return items
}

func normalizeNewlines(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	return strings.ReplaceAll(value, "\r", "\n")
}
