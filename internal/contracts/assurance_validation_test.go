package contracts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateProjectRejectsVerifiedTierMissingBaseline(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	configPath := filepath.Join(root, "runecontext.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	rewritten := strings.Replace(string(data), "assurance_tier: plain", "assurance_tier: verified", 1)
	if err := os.WriteFile(configPath, []byte(rewritten), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	v := NewValidator(schemaRoot(t))
	_, err = v.ValidateProject(root)
	if err == nil || !strings.Contains(err.Error(), "assurance baseline is required") {
		t.Fatalf("expected missing baseline error, got %v", err)
	}
}

func TestValidateProjectRejectsInvalidAssuranceReceiptHash(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	writeVerifiedAssuranceTierConfig(t, root)
	writeBaselineForAssuranceValidation(t, root)
	writeInvalidHashAssuranceReceipt(t, root)

	v := NewValidator(schemaRoot(t))
	_, err := v.ValidateProject(root)
	if err == nil || !strings.Contains(err.Error(), "receipt_hash does not match") {
		t.Fatalf("expected receipt hash mismatch error, got %v", err)
	}
}

func writeVerifiedAssuranceTierConfig(t *testing.T, root string) {
	t.Helper()

	configPath := filepath.Join(root, "runecontext.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	rewritten := strings.Replace(string(data), "assurance_tier: plain", "assurance_tier: verified", 1)
	if err := os.WriteFile(configPath, []byte(rewritten), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func writeBaselineForAssuranceValidation(t *testing.T, root string) {
	t.Helper()

	baseline := strings.Join([]string{
		"schema_version: 1",
		"kind: baseline",
		"subject_id: project-root",
		"created_at: 1710000000",
		"canonicalization: runecontext-canonical-json-v1",
		"value:",
		"  adoption_commit: abcdef1234567890abcdef1234567890abcdef12",
		"  source_posture: embedded",
		"",
	}, "\n")
	baselinePath := filepath.Join(root, "assurance", "baseline.yaml")
	if err := os.MkdirAll(filepath.Dir(baselinePath), 0o755); err != nil {
		t.Fatalf("mkdir baseline dir: %v", err)
	}
	if err := os.WriteFile(baselinePath, []byte(baseline), 0o644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}
}

func writeInvalidHashAssuranceReceipt(t *testing.T, root string) {
	t.Helper()

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
}
