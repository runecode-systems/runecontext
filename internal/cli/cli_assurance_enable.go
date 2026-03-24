package cli

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type assuranceEnableContext struct {
	root          string
	configPath    string
	configData    []byte
	rootCfg       map[string]any
	loaded        *contracts.LoadedProject
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

func newAssuranceEnableContext(root string, loaded *contracts.LoadedProject) (*assuranceEnableContext, error) {
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
		root:          root,
		configPath:    configPath,
		configData:    configData,
		rootCfg:       rootCfg,
		loaded:        loaded,
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
	envelope, err := ctx.buildAssuranceBaselineEnvelope()
	if err != nil {
		return assuranceEnableResult{}, err
	}
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

func (ctx *assuranceEnableContext) buildAssuranceBaselineEnvelope() (contracts.AssuranceEnvelope, error) {
	adoptionCommit, sourcePosture := sourceSnapshotFields(ctx.root, ctx.rootCfg, ctx.loaded)
	if adoptionCommit == "" || !isCanonicalLowerHex40(adoptionCommit) {
		return contracts.AssuranceEnvelope{}, fmt.Errorf("verified assurance enablement requires a canonical lowercase 40-char adoption_commit from resolved source metadata")
	}
	return contracts.AssuranceEnvelope{
		SchemaVersion:    1,
		Kind:             "baseline",
		SubjectID:        "project-root",
		CreatedAt:        time.Now().Unix(),
		Canonicalization: contracts.AssuranceCanonicalizationToken,
		Value: map[string]any{
			"adoption_commit": adoptionCommit,
			"source_posture":  sourcePosture,
		},
	}, nil
}

func sourceSnapshotFields(root string, rootCfg map[string]any, loaded *contracts.LoadedProject) (string, string) {
	if commit, posture, ok := sourceSnapshotFromResolvedProject(loaded); ok {
		return commit, posture
	}
	source, ok := rootSourceMap(rootCfg)
	if !ok {
		return "", ""
	}
	adoptionCommit, posture := sourceSnapshotFromConfigSource(source)
	if adoptionCommit != "" {
		return adoptionCommit, posture
	}
	if headCommit := resolveGitHeadCommit(root); headCommit != "" {
		return headCommit, posture
	}
	if posture == "embedded" || posture == "path" {
		return syntheticAdoptionCommit(source), posture
	}
	return adoptionCommit, posture
}

func sourceSnapshotFromResolvedProject(loaded *contracts.LoadedProject) (string, string, bool) {
	if loaded == nil || loaded.Resolution == nil {
		return "", "", false
	}
	commit := strings.TrimSpace(loaded.Resolution.ResolvedCommit)
	posture := strings.TrimSpace(string(loaded.Resolution.SourceMode))
	if commit == "" || posture == "" {
		return "", "", false
	}
	return commit, posture, true
}

func rootSourceMap(rootCfg map[string]any) (map[string]any, bool) {
	sourceRaw, ok := rootCfg["source"]
	if !ok {
		return nil, false
	}
	source, ok := sourceRaw.(map[string]any)
	if !ok {
		return nil, false
	}
	return source, true
}

func sourceSnapshotFromConfigSource(source map[string]any) (string, string) {
	adoptionCommit := readOptionalString(source, "commit")
	if adoptionCommit == "" {
		adoptionCommit = readOptionalString(source, "expect_commit")
	}
	posture := readOptionalString(source, "type")
	return adoptionCommit, posture
}

func resolveGitHeadCommit(root string) string {
	root = strings.TrimSpace(root)
	if root == "" {
		return ""
	}
	output, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	commit := strings.TrimSpace(string(output))
	if !isCanonicalLowerHex40(commit) {
		return ""
	}
	return commit
}

func syntheticAdoptionCommit(source map[string]any) string {
	canonical, err := contracts.ComputeArtifactCanonicalJSON(source)
	if err != nil {
		return ""
	}
	sum := sha1.Sum([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

func isCanonicalLowerHex40(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
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
