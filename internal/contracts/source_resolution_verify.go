package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (v *SSHAllowedSignersVerifier) VerifySignedTag(repoRoot, tagName string) (*SignedTagVerification, error) {
	if v == nil {
		return nil, fmt.Errorf("signed tag verifier is required")
	}
	allowedSignersPath, cleanup, err := writeAllowedSignersFile(v.allowedSigners)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	result := runGitCaptured(v.gitCommandArgs(repoRoot, allowedSignersPath, tagName), v.gitExecutable)
	if verification, err := trustedTagVerificationFromResult(tagName, result); verification != nil || err != nil {
		return verification, err
	}
	output := strings.TrimSpace(result.Output)
	reason := classifySignedTagFailure(output)
	message := signedTagFailureMessage(tagName, reason, output)
	return nil, &SignedTagVerificationError{Tag: tagName, Reason: reason, Message: message, Diagnostics: []ResolutionDiagnostic{{Severity: DiagnosticSeverityError, Code: string(reason), Message: message}}}
}

func writeAllowedSignersFile(allowedSigners []byte) (string, func(), error) {
	tempRoot, err := os.MkdirTemp("", "runectx-allowed-signers-")
	if err != nil {
		return "", nil, err
	}
	allowedSignersPath := filepath.Join(tempRoot, "allowed_signers")
	if err := os.WriteFile(allowedSignersPath, allowedSigners, 0o600); err != nil {
		_ = os.RemoveAll(tempRoot)
		return "", nil, err
	}
	return allowedSignersPath, func() { _ = os.RemoveAll(tempRoot) }, nil
}

func trustedTagVerificationFromResult(tagName string, result gitCommandResult) (*SignedTagVerification, error) {
	if result.TimedOut {
		message := fmt.Sprintf("signed tag %q verification failed: git %s: command timed out after %s", tagName, sanitizeGitArgs(result.Args), gitCommandTimeout)
		return nil, signedTagVerificationFailure(tagName, SignedTagFailureVerificationFailed, message)
	}
	if result.Err != nil && result.ExitCode == -1 {
		message := sanitizedResultMessage(result)
		message = fmt.Sprintf("signed tag %q verification failed: %s", tagName, message)
		return nil, signedTagVerificationFailure(tagName, SignedTagFailureVerificationFailed, message)
	}
	if result.ExitCode != 0 {
		return nil, nil
	}
	identity, fingerprint, err := parseTrustedSSHVerifyTagOutput(strings.TrimSpace(result.Output))
	if err != nil {
		return nil, fmt.Errorf("parse trusted signed-tag verification output: %w", err)
	}
	return &SignedTagVerification{SignerIdentity: identity, SignerFingerprint: fingerprint}, nil
}

func signedTagVerificationFailure(tagName string, reason SignedTagFailureReason, message string) error {
	return &SignedTagVerificationError{Tag: tagName, Reason: reason, Message: message, Diagnostics: []ResolutionDiagnostic{{Severity: DiagnosticSeverityError, Code: string(reason), Message: message}}}
}

func sanitizedResultMessage(result gitCommandResult) string {
	message := sanitizeGitMessage(strings.TrimSpace(result.Output))
	if message == "" && result.Err != nil {
		message = sanitizeGitMessage(result.Err.Error())
	}
	return message
}

func (v *SSHAllowedSignersVerifier) gitCommandArgs(repoRoot, allowedSignersPath, tagName string) []string {
	return []string{"-C", repoRoot, "-c", "gpg.format=ssh", "-c", "gpg.ssh.allowedSignersFile=" + allowedSignersPath, "verify-tag", "--raw", tagName}
}

func validateSignedTagVerification(verification *SignedTagVerification, tagName string) error {
	if verification == nil {
		message := fmt.Sprintf("signed tag %q verification failed: verifier returned no verification details", tagName)
		return signedTagVerificationFailure(tagName, SignedTagFailureVerificationFailed, message)
	}
	if strings.TrimSpace(verification.SignerIdentity) == "" || strings.TrimSpace(verification.SignerFingerprint) == "" {
		message := fmt.Sprintf("signed tag %q verification failed: verifier returned incomplete signer details", tagName)
		return signedTagVerificationFailure(tagName, SignedTagFailureVerificationFailed, message)
	}
	return nil
}
