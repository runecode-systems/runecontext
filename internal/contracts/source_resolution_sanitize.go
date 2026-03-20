package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func parseTrustedSSHVerifyTagOutput(output string) (string, string, error) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		const prefix = `Good "git" signature for `
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		remainder := strings.TrimPrefix(line, prefix)
		fingerprintIndex := strings.LastIndex(remainder, " SHA256:")
		if fingerprintIndex <= 0 {
			continue
		}
		identitySection := strings.TrimSpace(remainder[:fingerprintIndex])
		withIndex := strings.LastIndex(identitySection, " with ")
		if withIndex <= 0 {
			continue
		}
		identity := strings.TrimSpace(identitySection[:withIndex])
		fingerprint := strings.TrimSpace(remainder[fingerprintIndex+1:])
		if identity == "" || fingerprint == "" || !strings.HasPrefix(fingerprint, "SHA256:") {
			continue
		}
		return identity, fingerprint, nil
	}
	return "", "", fmt.Errorf("missing trusted signer identity/fingerprint in verification output")
}

func classifySignedTagFailure(output string) SignedTagFailureReason {
	lower := strings.ToLower(output)
	switch {
	case strings.Contains(lower, "no signature found"):
		return SignedTagFailureUnsignedTag
	case strings.Contains(lower, "could not verify signature") || strings.Contains(lower, "couldn't verify signature") || strings.Contains(lower, "bad signature") || strings.Contains(lower, "invalid format"):
		return SignedTagFailureInvalidSignature
	case strings.Contains(lower, "no principal matched"):
		return SignedTagFailureUntrustedSigner
	default:
		return SignedTagFailureVerificationFailed
	}
}

func signedTagFailureMessage(tagName string, reason SignedTagFailureReason, output string) string {
	sanitizedOutput := sanitizeGitMessage(strings.TrimSpace(output))
	switch reason {
	case SignedTagFailureUnsignedTag:
		return fmt.Sprintf("signed tag %q is unsigned", tagName)
	case SignedTagFailureInvalidSignature:
		return signedTagFailureDetail(tagName, sanitizedOutput, "has an invalid signature")
	case SignedTagFailureUntrustedSigner:
		return signedTagFailureDetail(tagName, sanitizedOutput, "was signed by an untrusted signer")
	default:
		return signedTagFailureDetail(tagName, sanitizedOutput, "verification failed")
	}
}

func signedTagFailureDetail(tagName, detail, label string) string {
	if detail != "" {
		return fmt.Sprintf("signed tag %q %s: %s", tagName, label, detail)
	}
	return fmt.Sprintf("signed tag %q %s", tagName, label)
}

func sanitizedGitEnv() []string {
	env := []string{
		"HOME=" + os.TempDir(),
		"XDG_CONFIG_HOME=" + os.TempDir(),
		"GNUPGHOME=" + os.TempDir(),
		"GIT_CONFIG_GLOBAL=" + os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_ALLOW_PROTOCOL=" + strings.Join(gitAllowedProtocols, ":"),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
		"SSH_ASKPASS=",
		"SSH_AUTH_SOCK=",
		"GIT_SSH=",
		"GIT_SSH_COMMAND=",
		"GCM_INTERACTIVE=Never",
		"LANG=C",
		"LC_ALL=C",
	}
	for _, key := range []string{"PATH", "TMPDIR", "TMP", "TEMP", "SYSTEMROOT"} {
		if value, ok := os.LookupEnv(key); ok && value != "" {
			env = append(env, key+"="+value)
		}
	}
	return env
}

func runeContextRelativePath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

func sanitizeGitArgs(args []string) string {
	sanitized := make([]string, len(args))
	for i, arg := range args {
		sanitized[i] = sanitizeGitMessage(arg)
	}
	return strings.Join(sanitized, " ")
}

func sanitizeGitMessage(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return trimmed
	}
	trimmed = redactGitFingerprints(trimmed)
	trimmed = redactGitURLs(trimmed)
	return redactGitIdentities(trimmed)
}

func redactGitFingerprints(message string) string {
	for _, prefix := range []string{"sha256:", "SHA256:"} {
		message = redactPrefixedToken(message, prefix, "<redacted-fingerprint>")
	}
	return message
}

func redactGitURLs(message string) string {
	for _, prefix := range []string{"https://", "http://", "ssh://", "file://"} {
		message = redactCaseInsensitivePrefixedToken(message, prefix, "<redacted-url>")
	}
	return message
}

func redactGitIdentities(message string) string {
	searchFrom := 0
	for {
		at := nextIdentityAt(message, searchFrom)
		if at < 0 {
			return message
		}
		if shouldSkipIdentityAt(message, at) {
			searchFrom = at + 1
			continue
		}
		start, end := identityBounds(message, at)
		message = message[:start] + "<redacted-identity>" + message[end:]
		searchFrom = start + len("<redacted-identity>")
	}
}

func nextIdentityAt(message string, searchFrom int) int {
	rel := strings.Index(message[searchFrom:], "@")
	if rel < 0 {
		return -1
	}
	return searchFrom + rel
}

func shouldSkipIdentityAt(message string, at int) bool {
	return at <= 0 || (at+1 < len(message) && message[at+1] == '{')
}

func identityBounds(message string, at int) (int, int) {
	start := strings.LastIndexAny(message[:at], " /\n\r\t\"")
	if start < 0 {
		start = 0
	} else {
		start++
	}
	end := at + 1
	for end < len(message) && !strings.ContainsRune(" /\n\r\t\"'", rune(message[end])) {
		end++
	}
	return start, end
}

func redactPrefixedToken(message, prefix, replacement string) string {
	for {
		idx := strings.Index(message, prefix)
		if idx < 0 {
			return message
		}
		end := idx + len(prefix)
		for end < len(message) && !strings.ContainsRune(" \n\r\t'\"", rune(message[end])) {
			end++
		}
		message = message[:idx] + replacement + message[end:]
	}
}

func redactCaseInsensitivePrefixedToken(message, prefix, replacement string) string {
	for {
		idx := strings.Index(strings.ToLower(message), prefix)
		if idx < 0 {
			return message
		}
		end := idx + len(prefix)
		for end < len(message) && !strings.ContainsRune(" \n\r\t'\"", rune(message[end])) {
			end++
		}
		message = message[:idx] + replacement + message[end:]
	}
}
