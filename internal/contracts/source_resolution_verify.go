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
	unsigned, err := v.tagIsUnsigned(repoRoot, tagName)
	if err != nil {
		message := fmt.Sprintf("signed tag %q verification failed: %s", tagName, sanitizeGitMessage(err.Error()))
		return nil, signedTagVerificationFailure(tagName, SignedTagFailureVerificationFailed, message)
	}
	if unsigned {
		reason := SignedTagFailureUnsignedTag
		message := signedTagFailureMessage(tagName, reason, "")
		return nil, &SignedTagVerificationError{Tag: tagName, Reason: reason, Message: message, Diagnostics: []ResolutionDiagnostic{{Severity: DiagnosticSeverityError, Code: string(reason), Message: message}}}
	}
	output := strings.TrimSpace(result.Output)
	reason := classifySignedTagFailure(output)
	message := signedTagFailureMessage(tagName, reason, output)
	return nil, &SignedTagVerificationError{Tag: tagName, Reason: reason, Message: message, Diagnostics: []ResolutionDiagnostic{{Severity: DiagnosticSeverityError, Code: string(reason), Message: message}}}
}

func (v *SSHAllowedSignersVerifier) tagIsUnsigned(repoRoot, tagName string) (bool, error) {
	exists, err := tagRefExists(v.gitExecutable, repoRoot, tagName)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("tag ref %q was not found", normalizeTagRef(tagName))
	}
	hasTagObject, err := tagRefHasTagObject(v.gitExecutable, repoRoot, tagName)
	if err != nil {
		return false, err
	}
	if !hasTagObject {
		return true, nil
	}
	tagBody, ok := gitTagObjectBody(v.gitExecutable, repoRoot, tagName)
	if !ok {
		return false, nil
	}
	return !strings.Contains(tagBody, "-----BEGIN SSH SIGNATURE-----") && !strings.Contains(tagBody, "-----BEGIN PGP SIGNATURE-----"), nil
}

func tagRefHasTagObject(executable, repoRoot, tagName string) (bool, error) {
	insideWorkTree, err := gitIsInsideWorkTree(executable, repoRoot)
	if err != nil {
		return false, err
	}
	if !insideWorkTree {
		return false, fmt.Errorf("git repository lookup failed: repository is not a git work tree")
	}
	result := runGitCaptured([]string{"-C", repoRoot, "cat-file", "-t", normalizeTagRef(tagName)}, executable)
	if result.TimedOut {
		return false, fmt.Errorf("git %s: command timed out after %s", sanitizeGitArgs(result.Args), gitCommandTimeout)
	}
	if result.Err != nil && result.ExitCode == -1 {
		return false, fmt.Errorf("%s", sanitizedResultMessage(result))
	}
	if result.ExitCode != 0 {
		return false, nil
	}
	return strings.TrimSpace(result.Output) == "tag", nil
}

func tagRefExists(executable, repoRoot, tagName string) (bool, error) {
	result := runGitCaptured([]string{"-C", repoRoot, "rev-parse", "--verify", normalizeTagRef(tagName)}, executable)
	if result.TimedOut {
		return false, fmt.Errorf("git %s: command timed out after %s", sanitizeGitArgs(result.Args), gitCommandTimeout)
	}
	if result.Err != nil && result.ExitCode == -1 {
		return false, fmt.Errorf("%s", sanitizedResultMessage(result))
	}
	if result.ExitCode != 0 {
		return false, nil
	}
	return true, nil
}

func gitIsInsideWorkTree(executable, repoRoot string) (bool, error) {
	result := runGitCaptured([]string{"-C", repoRoot, "rev-parse", "--is-inside-work-tree"}, executable)
	if result.TimedOut {
		return false, fmt.Errorf("git %s: command timed out after %s", sanitizeGitArgs(result.Args), gitCommandTimeout)
	}
	if result.Err != nil && result.ExitCode == -1 {
		return false, fmt.Errorf("%s", sanitizedResultMessage(result))
	}
	if result.ExitCode != 0 {
		return false, nil
	}
	return strings.TrimSpace(result.Output) == "true", nil
}

func gitTagObjectBody(executable, repoRoot, tagName string) (string, bool) {
	result := runGitCaptured([]string{"-C", repoRoot, "cat-file", "-p", normalizeTagRef(tagName)}, executable)
	if result.TimedOut || result.ExitCode != 0 || result.Err != nil {
		return "", false
	}
	return result.Output, true
}

func normalizeTagRef(tagName string) string {
	if strings.HasPrefix(tagName, "refs/tags/") {
		return tagName
	}
	return "refs/tags/" + tagName
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
