package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestRunCompletionSuggestUsageAndHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if code := Run([]string{"completion", "suggest"}, &stdout, &stderr); code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "usage="+completionSuggestUsage) {
		t.Fatalf("expected completion suggest usage output, got %q", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"completion", "suggest", "--help"}, &stdout, &stderr); code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage="+completionSuggestUsage) {
		t.Fatalf("expected completion suggest help usage, got %q", stdout.String())
	}
}

func TestRunCompletionSuggestSoftFailsOutsideRuneContextProject(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, provider := range []string{suggestionProviderChangeIDs, suggestionProviderBundleIDs, suggestionProviderPromotionTargets} {
		t.Run(provider, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := Run([]string{"completion", "suggest", provider}, &stdout, &stderr)
			if code != exitOK {
				t.Fatalf("expected success outside project for provider %q, got %d (%s)", provider, code, stderr.String())
			}
			if strings.TrimSpace(stdout.String()) != "" {
				t.Fatalf("expected no suggestions outside project for %q, got %q", provider, stdout.String())
			}
			if strings.TrimSpace(stderr.String()) != "" {
				t.Fatalf("expected empty stderr outside project for %q, got %q", provider, stderr.String())
			}
		})
	}
}

func TestRunCompletionSuggestAdapterNamesShellInjection(t *testing.T) {
	workspaceRoot, _ := stageGeneratedAdapterWorkspaceForTests(t)
	t.Chdir(workspaceRoot)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"completion", "suggest", suggestionProviderAdapterNamesShellInjection}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	items := completionSuggestLines(stdout.String())
	if !slices.Equal(items, []string{"claude-code", "opencode"}) {
		t.Fatalf("expected shell-injection adapter suggestions, got %#v", items)
	}
}

func TestRunCompletionSuggestExplicitPathErrors(t *testing.T) {
	t.Run("explicit path missing project config", func(t *testing.T) {
		root := t.TempDir()
		assertCompletionSuggestInvalid(t, []string{"completion", "suggest", "--path", root, suggestionProviderChangeIDs}, "failed to load project")
	})

	t.Run("explicit path malformed project surfaces validation", func(t *testing.T) {
		root := repoFixtureRoot(t, "traceability", "reject-bundle-invalid")
		assertCompletionSuggestInvalid(t, []string{"completion", "suggest", "--path", root, suggestionProviderBundleIDs}, "bundle")
	})
}

func TestHandleAdapterSuggestionReadErrorIncludesContextForExplicitRoot(t *testing.T) {
	request := completionSuggestRequest{root: "/tmp/project", explicitRoot: true}
	_, err := handleAdapterSuggestionReadError(request, os.ErrNotExist)
	if err == nil {
		t.Fatal("expected explicit-root read failure")
	}
	if !strings.Contains(err.Error(), "failed to load adapter packs for \"/tmp/project\"") {
		t.Fatalf("expected contextual adapter-pack read error, got %q", err.Error())
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected wrapped not-exist cause, got %v", err)
	}
}

func TestHandleAdapterSuggestionRootErrorImplicitRootPropagatesUnexpectedErrors(t *testing.T) {
	request := completionSuggestRequest{root: ".", explicitRoot: false}
	want := errors.New("schema discovery failed")
	_, err := handleAdapterSuggestionRootError(request, want)
	if !errors.Is(err, want) {
		t.Fatalf("expected implicit-root unexpected error propagation, got %v", err)
	}
}

func assertCompletionSuggestInvalid(t *testing.T, args []string, wantSubstring string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "result=invalid") {
		t.Fatalf("expected invalid result output, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), wantSubstring) {
		t.Fatalf("expected error output containing %q, got %q", wantSubstring, stderr.String())
	}
}

func TestRunCompletionSuggestAdapterNames(t *testing.T) {
	workspaceRoot, _ := stageGeneratedAdapterWorkspaceForTests(t)
	t.Chdir(workspaceRoot)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"completion", "suggest", suggestionProviderAdapterNames}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	items := completionSuggestLines(stdout.String())
	for _, want := range []string{"claude-code", "codex", "generic", "opencode"} {
		if !slices.Contains(items, want) {
			t.Fatalf("expected adapter suggestion %q in %#v", want, items)
		}
	}
}

func TestRunCompletionSuggestAdapterNamesInvalidWhenSchemaDiscoveryFails(t *testing.T) {
	t.Chdir(t.TempDir())
	assertCompletionSuggestInvalid(
		t,
		[]string{"completion", "suggest", suggestionProviderAdapterNames},
		"could not locate RuneContext schemas",
	)
}

func TestRunCompletionSuggestRepoModesWithPath(t *testing.T) {
	for _, tc := range completionSuggestRepoModeCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			assertCompletionSuggestContains(t, tc.root, tc.provider, tc.want)
		})
	}
}

type completionSuggestRepoModeCase struct {
	name     string
	root     string
	provider string
	want     []string
}

func completionSuggestRepoModeCases(t *testing.T) []completionSuggestRepoModeCase {
	t.Helper()
	return []completionSuggestRepoModeCase{
		{
			name:     "embedded change ids",
			root:     filepath.Join(repoFixtureRoot(t, "source-resolution"), "embedded-project"),
			provider: suggestionProviderChangeIDs,
			want:     []string{"CHG-2026-001-a3f2-source-resolution"},
		},
		{
			name:     "linked change ids",
			root:     filepath.Join(repoFixtureRoot(t, "source-resolution"), "path-project"),
			provider: suggestionProviderChangeIDs,
			want:     []string{"CHG-2026-001-a3f2-source-resolution"},
		},
		{
			name:     "monorepo change ids",
			root:     filepath.Join(repoFixtureRoot(t, "source-resolution", "monorepo"), "packages", "service"),
			provider: suggestionProviderChangeIDs,
			want:     []string{"CHG-2026-001-a3f2-source-resolution"},
		},
		{
			name:     "monorepo promotion targets",
			root:     filepath.Join(repoFixtureRoot(t, "source-resolution", "monorepo"), "packages", "service"),
			provider: suggestionProviderPromotionTargets,
			want:     []string{"standard:standards/global/source-integrity.md"},
		},
	}
}

func assertCompletionSuggestContains(t *testing.T, root, provider string, want []string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"completion", "suggest", "--path", root, provider}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	items := completionSuggestLines(stdout.String())
	for _, expected := range want {
		if !slices.Contains(items, expected) {
			t.Fatalf("expected suggestion %q in %#v", expected, items)
		}
	}
}

func TestRunCompletionSuggestHonorsPrefixAndPath(t *testing.T) {
	projectRoot := filepath.Join(repoFixtureRoot(t, "traceability"), "valid-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"completion", "suggest", "--path", projectRoot, "--prefix", "auth", suggestionProviderBundleIDs}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	items := completionSuggestLines(stdout.String())
	if !slices.Equal(items, []string{"auth-review"}) {
		t.Fatalf("expected prefix-filtered bundle suggestion, got %#v", items)
	}
}

func TestRunCompletionSuggestReadOnly(t *testing.T) {
	projectRoot := filepath.Join(repoFixtureRoot(t, "traceability"), "valid-project")
	statusPath := filepath.Join(projectRoot, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml")
	before, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status file before suggestions: %v", err)
	}

	for _, provider := range []string{suggestionProviderChangeIDs, suggestionProviderBundleIDs, suggestionProviderPromotionTargets} {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := Run([]string{"completion", "suggest", "--path", projectRoot, provider}, &stdout, &stderr)
		if code != exitOK {
			t.Fatalf("expected success exit code for provider %q, got %d (%s)", provider, code, stderr.String())
		}
	}

	after, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status file after suggestions: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("expected completion suggest to be read-only")
	}
}

func TestCompletionMetadataIncludesSuggestionProviders(t *testing.T) {
	metadata := CompletionMetadataRegistry()

	flagProviders := collectFlagSuggestionProviders(metadata)
	assertSuggestionProviders(t, flagProviders, []providerExpectation{
		{key: "change close|--superseded-by", provider: suggestionProviderChangeIDs},
		{key: "promote|--target", provider: suggestionProviderPromotionTargets},
		{key: "change new|--bundle", provider: suggestionProviderBundleIDs},
		{key: "change assess-intake|--bundle", provider: suggestionProviderBundleIDs},
		{key: "init|--seed-bundle", provider: suggestionProviderBundleIDs},
		{key: "standard discover|--change", provider: suggestionProviderChangeIDs},
	})
	positionalProviders := collectPositionalSuggestionProviders(metadata)
	assertSuggestionProviders(t, positionalProviders, []providerExpectation{
		{key: "promote|1", provider: suggestionProviderChangeIDs},
		{key: "change shape|1", provider: suggestionProviderChangeIDs},
		{key: "change decomposition-plan|1", provider: suggestionProviderChangeIDs},
		{key: "change decomposition-apply|1", provider: suggestionProviderChangeIDs},
		{key: "bundle resolve|1", provider: suggestionProviderBundleIDs},
		{key: "adapter render-host-native|1", provider: suggestionProviderAdapterNamesShellInjection},
	})
}

type providerExpectation struct {
	key      string
	provider string
}

func assertSuggestionProviders(t *testing.T, actual map[string]string, want []providerExpectation) {
	t.Helper()
	for _, expected := range want {
		if got := actual[expected.key]; got != expected.provider {
			t.Fatalf("expected %q suggestion provider %q, got %q", expected.key, expected.provider, got)
		}
	}
}

func collectFlagSuggestionProviders(metadata CompletionMetadata) map[string]string {
	providers := map[string]string{}
	for _, flag := range metadata.Flags {
		if flag.SuggestionProvider == "" {
			continue
		}
		providers[flag.CommandPath+"|"+flag.Name] = flag.SuggestionProvider
	}
	return providers
}

func collectPositionalSuggestionProviders(metadata CompletionMetadata) map[string]string {
	providers := map[string]string{}
	for _, positional := range metadata.PositionalSuggestions {
		providers[fmt.Sprintf("%s|%d", positional.CommandPath, positional.Position)] = positional.SuggestionProvider
	}
	return providers
}

func completionSuggestLines(output string) []string {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil
	}
	lines := strings.Split(trimmed, "\n")
	slices.Sort(lines)
	return lines
}
