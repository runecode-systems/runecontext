package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type validateRequest struct {
	root         string
	explicitRoot bool
	gitTrust     contracts.GitTrustInputs
}

func runValidate(args []string, stdout, stderr io.Writer) int {
	request, err := parseValidateArgs(args)
	if err != nil {
		writeCommandUsageError(stderr, "validate", validateUsage, err)
		return exitUsage
	}
	absRoot, err := resolveAbsoluteRoot(request.root)
	if err != nil {
		writeCommandUsageError(stderr, "validate", validateUsage, err)
		return exitUsage
	}
	index, err := validateProject(request, absRoot)
	if err != nil {
		writeValidateInvalid(stderr, absRoot, err)
		return exitInvalid
	}
	defer index.Close()
	writeLines(stdout, buildValidateOutput(absRoot, index)...)
	return exitOK
}

func parseValidateArgs(args []string) (validateRequest, error) {
	request := validateRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		if flag.name != "--ssh-allowed-signers" {
			return flag.next, fmt.Errorf("unknown validate flag %q", flag.raw)
		}
		value, next, err := requireAllowedSignersPath(args, flag)
		if err != nil {
			return flag.next, err
		}
		verifier, err := loadSignedTagVerifier(value)
		if err != nil {
			return flag.next, err
		}
		request.gitTrust.SignedTagVerifier = verifier
		return next, nil
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return validateRequest{}, err
	}
	return finalizeValidateRequest(request, positionals)
}

func requireAllowedSignersPath(args []string, flag parsedFlag) (string, int, error) {
	value, next, err := flag.requireValue(args)
	if err != nil {
		return "", flag.next, fmt.Errorf("--ssh-allowed-signers requires a path")
	}
	return value, next, nil
}

func finalizeValidateRequest(request validateRequest, positionals []string) (validateRequest, error) {
	if len(positionals) > 1 {
		return validateRequest{}, fmt.Errorf("expected at most one path argument")
	}
	if len(positionals) == 1 {
		request.root = positionals[0]
		request.explicitRoot = true
	}
	return request, nil
}

func loadSignedTagVerifier(path string) (contracts.SignedTagVerifier, error) {
	allowedSignersData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read ssh allowed signers file %q: %w", path, err)
	}
	verifier, err := contracts.NewSSHAllowedSignersVerifier(allowedSignersData)
	if err != nil {
		return nil, fmt.Errorf("load ssh allowed signers file %q: %w", path, err)
	}
	return verifier, nil
}

func resolveAbsoluteRoot(root string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path %q: %v", root, err)
	}
	return absRoot, nil
}

func validateProject(request validateRequest, absRoot string) (*contracts.ProjectIndex, error) {
	schemaRoot, err := locateSchemaRoot()
	if err != nil {
		return nil, err
	}
	validator := contracts.NewValidator(schemaRoot)
	return validator.ValidateProjectWithOptions(absRoot, buildValidateResolveOptions(request))
}

func buildValidateResolveOptions(request validateRequest) contracts.ResolveOptions {
	options := contracts.ResolveOptions{
		ConfigDiscovery: contracts.ConfigDiscoveryNearestAncestor,
		ExecutionMode:   contracts.ExecutionModeLocal,
		GitTrust:        request.gitTrust,
	}
	if request.explicitRoot {
		options.ConfigDiscovery = contracts.ConfigDiscoveryExplicitRoot
	}
	return options
}

func buildValidateOutput(absRoot string, index *contracts.ProjectIndex) []line {
	output := []line{{"result", "ok"}, {"command", "validate"}, {"root", absRoot}}
	diagnostics := collectDiagnostics(index)
	output = appendValidateResolutionLines(output, index, diagnostics)
	return appendValidateDiagnosticLines(output, diagnostics)
}

func appendValidateResolutionLines(lines []line, index *contracts.ProjectIndex, diagnostics []emittedDiagnostic) []line {
	if index == nil || index.Resolution == nil {
		return lines
	}

	lines = append(lines,
		line{"selected_config_path", index.Resolution.SelectedConfigPath},
		line{"project_root", index.Resolution.ProjectRoot},
		line{"source_root", index.Resolution.SourceRoot},
		line{"source_mode", string(index.Resolution.SourceMode)},
		line{"source_ref", index.Resolution.SourceRef},
		line{"verification_posture", string(index.Resolution.VerificationPosture)},
		line{"diagnostic_count", fmt.Sprintf("%d", len(diagnostics))},
	)
	if index.Resolution.ResolvedCommit != "" {
		lines = append(lines, line{"resolved_commit", index.Resolution.ResolvedCommit})
	}
	if index.Resolution.VerifiedSignerIdentity != "" {
		lines = append(lines, line{"verified_signer_identity", index.Resolution.VerifiedSignerIdentity})
	}
	if index.Resolution.VerifiedSignerFingerprint != "" {
		lines = append(lines, line{"verified_signer_fingerprint", index.Resolution.VerifiedSignerFingerprint})
	}
	return lines
}

func appendValidateDiagnosticLines(lines []line, diagnostics []emittedDiagnostic) []line {
	for i, diagnostic := range diagnostics {
		lines = append(lines, validateDiagnosticLines(fmt.Sprintf("diagnostic_%d", i+1), diagnostic)...)
	}
	return lines
}

func validateDiagnosticLines(prefix string, diagnostic emittedDiagnostic) []line {
	lines := []line{
		{prefix + "_severity", string(diagnostic.Severity)},
		{prefix + "_code", diagnostic.Code},
		{prefix + "_message", diagnostic.Message},
	}
	lines = appendOptionalDiagnosticLine(lines, prefix+"_path", diagnostic.Path)
	lines = appendOptionalDiagnosticLine(lines, prefix+"_bundle", diagnostic.Bundle)
	lines = appendOptionalDiagnosticLine(lines, prefix+"_aspect", diagnostic.Aspect)
	lines = appendOptionalDiagnosticLine(lines, prefix+"_rule", diagnostic.Rule)
	lines = appendOptionalDiagnosticLine(lines, prefix+"_pattern", diagnostic.Pattern)
	if len(diagnostic.Matches) > 0 {
		lines = append(lines, line{prefix + "_matches", strings.Join(diagnostic.Matches, ",")})
	}
	return lines
}

func appendOptionalDiagnosticLine(lines []line, key, value string) []line {
	if value == "" {
		return lines
	}
	return append(lines, line{key, value})
}

func writeValidateInvalid(w io.Writer, absRoot string, err error) {
	writeLines(w, buildValidateErrorLines(absRoot, err)...)
}

func buildValidateErrorLines(absRoot string, err error) []line {
	lines := []line{{"result", "invalid"}, {"command", "validate"}, {"root", absRoot}}
	var signedTagErr *contracts.SignedTagVerificationError
	if errors.As(err, &signedTagErr) {
		return appendSignedTagError(lines, signedTagErr)
	}
	var validationErr *contracts.ValidationError
	if errors.As(err, &validationErr) {
		return appendValidationError(lines, validationErr)
	}
	return append(lines, line{"error_message", err.Error()})
}

func appendSignedTagError(lines []line, err *contracts.SignedTagVerificationError) []line {
	if err.Path != "" {
		lines = append(lines, line{"error_path", err.Path})
	}
	if err.Tag != "" {
		lines = append(lines, line{"error_tag", err.Tag})
	}
	lines = append(lines,
		line{"error_reason", string(err.Reason)},
		line{"error_message", err.Message},
		line{"diagnostic_count", fmt.Sprintf("%d", len(err.Diagnostics))},
	)
	if err.ResolvedCommit != "" {
		lines = append(lines, line{"resolved_commit", err.ResolvedCommit})
	}
	if err.SignerIdentity != "" {
		lines = append(lines, line{"verified_signer_identity", err.SignerIdentity})
	}
	if err.SignerFingerprint != "" {
		lines = append(lines, line{"verified_signer_fingerprint", err.SignerFingerprint})
	}
	for i, diagnostic := range err.Diagnostics {
		prefix := fmt.Sprintf("diagnostic_%d", i+1)
		lines = append(lines,
			line{prefix + "_severity", string(diagnostic.Severity)},
			line{prefix + "_code", diagnostic.Code},
			line{prefix + "_message", diagnostic.Message},
		)
	}
	return lines
}

func appendValidationError(lines []line, err *contracts.ValidationError) []line {
	if err.Path != "" {
		lines = append(lines, line{"error_path", err.Path})
	}
	return append(lines, line{"error_message", err.Message})
}
