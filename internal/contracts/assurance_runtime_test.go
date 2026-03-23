package contracts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestCaptureContextPackAssuranceCreatesReceipt(t *testing.T) {
	root, v, loaded := setupVerifiedAssuranceProject(t)
	defer loaded.Close()

	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		t.Fatalf("validate project: %v", err)
	}
	defer index.Close()

	result, err := CaptureContextPackAssurance(v, loaded, index, ContextPackAssuranceCaptureOptions{
		BundleIDs:   []string{"base"},
		GeneratedAt: time.Date(2026, time.March, 23, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("capture assurance context-pack: %v", err)
	}
	if result.ReceiptPath == "" {
		t.Fatalf("expected receipt path, got %#v", result)
	}
	if result.PackHash == "" {
		t.Fatalf("expected pack hash, got %#v", result)
	}
	if len(result.ChangedFiles) != 1 {
		t.Fatalf("expected one changed file, got %#v", result.ChangedFiles)
	}
	receiptAbsolute := filepath.Join(root, filepath.FromSlash(result.ReceiptPath))
	if _, err := os.Stat(receiptAbsolute); err != nil {
		t.Fatalf("expected receipt file on disk: %v", err)
	}
}

func setupVerifiedAssuranceProject(t *testing.T) (string, *Validator, *LoadedProject) {
	t.Helper()

	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	setAssuranceTierVerified(t, root, loaded)
	writeAssuranceBaselineFixture(t, root)
	return root, v, loaded
}

func setAssuranceTierVerified(t *testing.T, root string, loaded *LoadedProject) {
	t.Helper()

	configPath := filepath.Join(root, "runecontext.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	rewritten := strings.Replace(string(configData), "assurance_tier: plain", "assurance_tier: verified", 1)
	if err := os.WriteFile(configPath, []byte(rewritten), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	loaded.RootConfig["assurance_tier"] = AssuranceTierVerified
}

func writeAssuranceBaselineFixture(t *testing.T, root string) {
	t.Helper()

	baselinePath := filepath.Join(root, "assurance", "baseline.yaml")
	if err := os.MkdirAll(filepath.Dir(baselinePath), 0o755); err != nil {
		t.Fatalf("mkdir baseline dir: %v", err)
	}
	baseline := AssuranceEnvelope{
		SchemaVersion:    1,
		Kind:             "baseline",
		SubjectID:        "project-root",
		CreatedAt:        1710000000,
		Canonicalization: AssuranceCanonicalizationToken,
		Value: map[string]any{
			"adoption_commit": "abcdef1234567890abcdef1234567890abcdef12",
			"source_posture":  "embedded",
		},
	}
	baselineData, err := yaml.Marshal(baseline)
	if err != nil {
		t.Fatalf("marshal baseline: %v", err)
	}
	if err := os.WriteFile(baselinePath, baselineData, 0o644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}
}

func TestCaptureContextPackAssuranceRequiresVerifiedTier(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v := NewValidator(schemaRoot(t))
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	defer loaded.Close()
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		t.Fatalf("validate project: %v", err)
	}
	defer index.Close()
	_, err = CaptureContextPackAssurance(v, loaded, index, ContextPackAssuranceCaptureOptions{
		BundleIDs:   []string{"base"},
		GeneratedAt: time.Date(2026, time.March, 23, 12, 0, 0, 0, time.UTC),
	})
	if err == nil || !strings.Contains(err.Error(), "assurance_tier must be verified") {
		t.Fatalf("expected verified-tier error, got %v", err)
	}
}

func TestValidateProjectRejectsReceiptsWhenTierPlain(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	receiptPath := filepath.Join(root, "assurance", "receipts", "changes", "changes--rid-abcdef12-1710000000-0123456789ab-badbadbadbad.json")
	if err := os.MkdirAll(filepath.Dir(receiptPath), 0o755); err != nil {
		t.Fatalf("mkdir receipt dir: %v", err)
	}
	receipt := strings.Join([]string{
		"{",
		"  \"schema_version\": 1,",
		"  \"kind\": \"receipt\",",
		"  \"subject_id\": \"changes/CHG-2026-001-a3f2-auth-gateway\",",
		"  \"created_at\": 1710000000,",
		"  \"canonicalization\": \"runecontext-canonical-json-v1\",",
		"  \"value\": {",
		"    \"receipt_family\": \"changes\",",
		"    \"change_id\": \"CHG-2026-001-a3f2-auth-gateway\",",
		"    \"change_status\": \"closed\",",
		"    \"verification_status\": \"passed\"",
		"  },",
		"  \"receipt_id\": \"rid-abcdef12-1710000000-0123456789ab\",",
		"  \"receipt_hash\": \"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\",",
		"  \"provenance\": \"captured_verified\"",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(receiptPath, []byte(receipt), 0o644); err != nil {
		t.Fatalf("write receipt: %v", err)
	}
	v := NewValidator(schemaRoot(t))
	_, err := v.ValidateProject(root)
	if err == nil || !strings.Contains(err.Error(), "assurance receipts exist but assurance_tier is not verified") {
		t.Fatalf("expected plain-tier receipt error, got %v", err)
	}
}
