package contracts

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func assertResolutionMatchesGolden(t *testing.T, resolution *SourceResolution, goldenPath string, replacements map[string]string) {
	t.Helper()
	if resolution == nil {
		t.Fatal("expected resolution metadata")
	}
	expected := normalizeResolutionValue(t, mustParseYAML(t, replacePlaceholders(string(readFixture(t, goldenPath)), replacements)))
	actual := normalizeResolutionValue(t, comparableResolution(resolution))
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("resolution metadata mismatch\nexpected: %#v\nactual:   %#v", expected, actual)
	}
}

func comparableResolution(resolution *SourceResolution) map[string]any {
	result := map[string]any{
		"selected_config_path": filepath.ToSlash(resolution.SelectedConfigPath),
		"project_root":         filepath.ToSlash(resolution.ProjectRoot),
		"source_root":          filepath.ToSlash(resolution.SourceRoot),
		"source_mode":          string(resolution.SourceMode),
		"source_ref":           filepath.ToSlash(resolution.SourceRef),
		"verification_posture": string(resolution.VerificationPosture),
	}
	if resolution.ResolvedCommit != "" {
		result["resolved_commit"] = resolution.ResolvedCommit
	}
	if resolution.VerifiedSignerIdentity != "" {
		result["verified_signer_identity"] = resolution.VerifiedSignerIdentity
	}
	if resolution.VerifiedSignerFingerprint != "" {
		result["verified_signer_fingerprint"] = resolution.VerifiedSignerFingerprint
	}
	diagnostics := make([]any, 0, len(resolution.Diagnostics))
	for _, diagnostic := range resolution.Diagnostics {
		diagnostics = append(diagnostics, map[string]any{
			"severity": string(diagnostic.Severity),
			"code":     diagnostic.Code,
			"message":  diagnostic.Message,
		})
	}
	result["diagnostics"] = diagnostics
	return result
}

func normalizeResolutionValue(t *testing.T, value any) any {
	t.Helper()
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = normalizeResolutionValue(t, item)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for i, item := range typed {
			result[i] = normalizeResolutionValue(t, item)
		}
		return result
	case []string:
		result := make([]any, len(typed))
		for i, item := range typed {
			result[i] = normalizeResolutionValue(t, item)
		}
		return result
	case string:
		return filepath.ToSlash(typed)
	default:
		return typed
	}
}

func mustParseYAML(t *testing.T, text string) any {
	t.Helper()
	value, err := parseYAML([]byte(text))
	if err != nil {
		t.Fatalf("parse YAML: %v", err)
	}
	return value
}

func replacePlaceholders(text string, replacements map[string]string) string {
	for oldValue, newValue := range replacements {
		text = strings.ReplaceAll(text, oldValue, newValue)
	}
	return text
}
