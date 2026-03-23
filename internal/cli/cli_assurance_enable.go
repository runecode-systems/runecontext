package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type assuranceEnableContext struct {
	configPath    string
	configData    []byte
	rootCfg       map[string]any
	baselinePath  string
	priorBaseline []byte
}

type assuranceEnableResult struct {
	baselinePath string
}

type assuranceEnableError struct {
	err         error
	rollbackErr error
}

func (e *assuranceEnableError) Error() string {
	return e.err.Error()
}

func newAssuranceEnableContext(root string) (*assuranceEnableContext, error) {
	configPath := filepath.Join(root, "runecontext.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var rootCfg map[string]any
	if err := yaml.Unmarshal(configData, &rootCfg); err != nil {
		return nil, err
	}
	baselinePath := filepath.Join(root, "assurance", "baseline.yaml")
	var priorBaseline []byte
	if prev, err := os.ReadFile(baselinePath); err == nil {
		priorBaseline = prev
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	return &assuranceEnableContext{
		configPath:    configPath,
		configData:    configData,
		rootCfg:       rootCfg,
		baselinePath:  baselinePath,
		priorBaseline: priorBaseline,
	}, nil
}

func (ctx *assuranceEnableContext) ensureTierNotVerified() error {
	if fmt.Sprint(ctx.rootCfg["assurance_tier"]) == "verified" {
		return fmt.Errorf("assurance_tier already set to verified")
	}
	return nil
}

func finalizeAssuranceEnable(ctx *assuranceEnableContext) (assuranceEnableResult, error) {
	if err := ctx.ensureTierNotVerified(); err != nil {
		return assuranceEnableResult{}, err
	}
	envelope := buildAssuranceBaselineEnvelope(ctx.rootCfg)
	out, err := yaml.Marshal(envelope)
	if err != nil {
		return assuranceEnableResult{}, err
	}
	if err := ctx.writeBaseline(out); err != nil {
		return assuranceEnableResult{}, err
	}
	updatedConfig, err := renderVerifiedTierConfig(ctx.configData)
	if err != nil {
		return assuranceEnableResult{}, err
	}
	if err := writeAtomicFile(ctx.configPath, updatedConfig, 0o644); err != nil {
		rollbackErr := ctx.rollbackBaseline()
		return assuranceEnableResult{}, &assuranceEnableError{err: err, rollbackErr: rollbackErr}
	}
	return assuranceEnableResult{baselinePath: ctx.baselinePath}, nil
}

func buildAssuranceBaselineEnvelope(rootCfg map[string]any) contracts.AssuranceEnvelope {
	adoptionCommit, sourcePosture := sourceSnapshotFields(rootCfg)
	return contracts.AssuranceEnvelope{
		SchemaVersion:    1,
		Kind:             "baseline",
		SubjectID:        "project-root",
		CreatedAt:        time.Now().Unix(),
		Canonicalization: "runecontext-canonical-json-v1",
		Value: map[string]any{
			"adoption_commit": adoptionCommit,
			"source_posture":  sourcePosture,
		},
	}
}

func sourceSnapshotFields(rootCfg map[string]any) (string, string) {
	sourceRaw, ok := rootCfg["source"]
	if !ok {
		return "", ""
	}
	source, ok := sourceRaw.(map[string]any)
	if !ok {
		return "", ""
	}
	adoptionCommit := readOptionalString(source, "commit")
	if adoptionCommit == "" {
		adoptionCommit = readOptionalString(source, "expect_commit")
	}
	return adoptionCommit, readOptionalString(source, "type")
}

func readOptionalString(values map[string]any, key string) string {
	raw, ok := values[key]
	if !ok || raw == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(raw))
}

func renderVerifiedTierConfig(configData []byte) ([]byte, error) {
	updatedConfig, _ := ensureAssuranceTierConfig(configData)
	var parsed map[string]any
	if err := yaml.Unmarshal(updatedConfig, &parsed); err != nil {
		return nil, fmt.Errorf("render runecontext config with verified tier: %w", err)
	}
	if fmt.Sprint(parsed["assurance_tier"]) != "verified" {
		return nil, fmt.Errorf("render runecontext config with verified tier: assurance_tier not set to verified")
	}
	return updatedConfig, nil
}

func emitAssuranceEnableError(stderr io.Writer, machine machineOptions, root string, err error) {
	commandErr, rollbackErr := splitAssuranceEnableError(err)
	if rollbackErr != nil && !machine.jsonOutput {
		fmt.Fprintf(stderr, "Warning: failed to restore previous baseline: %v\n", rollbackErr)
	}
	errorLines := buildCommandInvalidLines("assurance enable", root, commandErr)
	if rollbackErr != nil {
		errorLines = append(errorLines, line{"rollback_error", rollbackErr.Error()})
	}
	emitOutput(stderr, machine, appendMachineOptionLines(errorLines, machine), exitInvalid, failureClassInvalid)
}

func splitAssuranceEnableError(err error) (error, error) {
	var enableErr *assuranceEnableError
	if errors.As(err, &enableErr) {
		return enableErr.err, enableErr.rollbackErr
	}
	return err, nil
}

func (ctx *assuranceEnableContext) writeBaseline(data []byte) error {
	if err := os.MkdirAll(filepath.Dir(ctx.baselinePath), 0o755); err != nil {
		return err
	}
	return writeAtomicFile(ctx.baselinePath, data, 0o644)
}

func (ctx *assuranceEnableContext) rollbackBaseline() error {
	if ctx.priorBaseline != nil {
		return writeAtomicFile(ctx.baselinePath, ctx.priorBaseline, 0o644)
	}
	return os.Remove(ctx.baselinePath)
}
