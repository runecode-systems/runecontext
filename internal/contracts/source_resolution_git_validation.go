package contracts

import (
	"fmt"
	"strings"
)

func validateGitURL(url string) error {
	if strings.HasPrefix(url, "-") {
		return fmt.Errorf("git source url must not start with '-'")
	}
	if strings.ContainsAny(url, gitURLControlChars+" ") {
		return fmt.Errorf("git source url contains unsupported whitespace or control characters")
	}
	lower := strings.ToLower(url)
	if strings.HasPrefix(lower, "ext::") {
		return fmt.Errorf("git source url must not use remote-helper forms")
	}
	if !gitURLSchemePattern.MatchString(url) {
		if strings.Contains(url, "::") {
			return fmt.Errorf("git source url must not use remote-helper forms")
		}
		return nil
	}
	scheme := strings.ToLower(url[:strings.Index(url, "://")])
	for _, allowed := range gitAllowedProtocols {
		if scheme == allowed {
			return nil
		}
	}
	return fmt.Errorf("git source url scheme %q is not allowed", scheme)
}

func validateGitCommit(commit string) error {
	if strings.HasPrefix(commit, "-") {
		return fmt.Errorf("git commit must not start with '-'")
	}
	if !gitCommitPattern.MatchString(commit) {
		return fmt.Errorf("git commit must be a 40-character lowercase hex SHA")
	}
	return nil
}

func validateGitRef(ref string) error {
	if err := validateGitRefStructure(ref); err != nil {
		return err
	}
	for _, segment := range strings.Split(ref, "/") {
		if err := validateGitRefSegment(segment); err != nil {
			return err
		}
	}
	return nil
}

func validateGitRefStructure(ref string) error {
	if strings.HasPrefix(ref, "-") {
		return fmt.Errorf("git ref must not start with '-'")
	}
	if !gitRefPattern.MatchString(ref) {
		return fmt.Errorf("git ref contains unsupported characters")
	}
	if strings.Contains(ref, "..") {
		return fmt.Errorf("git ref must not contain '..'")
	}
	if strings.Contains(ref, "//") {
		return fmt.Errorf("git ref must not contain consecutive '/'")
	}
	if strings.HasPrefix(ref, "/") || strings.HasSuffix(ref, "/") {
		return fmt.Errorf("git ref must not start or end with '/'")
	}
	if strings.HasSuffix(ref, ".lock") {
		return fmt.Errorf("git ref must not end with '.lock'")
	}
	return nil
}

func validateGitRefSegment(segment string) error {
	if segment == "" || segment == "." || segment == ".." {
		return fmt.Errorf("git ref contains an invalid path segment")
	}
	if strings.HasPrefix(segment, ".") {
		return fmt.Errorf("git ref segments must not start with '.'")
	}
	return nil
}
