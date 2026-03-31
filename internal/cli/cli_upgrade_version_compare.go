package cli

import (
	"strconv"
	"strings"
)

func compareKnownRunecontextVersions(left, right string) (int, bool) {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" || right == "" {
		return 0, false
	}
	if left == right {
		return 0, true
	}
	leftParts, leftOK := parseSemverLike(left)
	rightParts, rightOK := parseSemverLike(right)
	if !leftOK || !rightOK {
		return 0, false
	}
	return compareSemverLikeParts(leftParts, rightParts), true
}

type semverLikeParts struct {
	major int
	minor int
	patch int
	pre   []string
}

func parseSemverLike(version string) (semverLikeParts, bool) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(version), "v")
	if !semverLikePattern.MatchString(trimmed) {
		return semverLikeParts{}, false
	}
	base := trimmed
	if idx := strings.Index(base, "+"); idx >= 0 {
		base = base[:idx]
	}
	parts := strings.SplitN(base, "-", 2)
	core := strings.Split(parts[0], ".")
	if len(core) != 3 {
		return semverLikeParts{}, false
	}
	major, err := strconv.Atoi(core[0])
	if err != nil {
		return semverLikeParts{}, false
	}
	minor, err := strconv.Atoi(core[1])
	if err != nil {
		return semverLikeParts{}, false
	}
	patch, err := strconv.Atoi(core[2])
	if err != nil {
		return semverLikeParts{}, false
	}
	parsed := semverLikeParts{major: major, minor: minor, patch: patch}
	if len(parts) == 2 && parts[1] != "" {
		parsed.pre = strings.Split(parts[1], ".")
	}
	return parsed, true
}

func compareSemverLikeParts(left, right semverLikeParts) int {
	if left.major != right.major {
		if left.major < right.major {
			return -1
		}
		return 1
	}
	if left.minor != right.minor {
		if left.minor < right.minor {
			return -1
		}
		return 1
	}
	if left.patch != right.patch {
		if left.patch < right.patch {
			return -1
		}
		return 1
	}
	return comparePreRelease(left.pre, right.pre)
}

func comparePreRelease(left, right []string) int {
	if len(left) == 0 && len(right) == 0 {
		return 0
	}
	if len(left) == 0 {
		return 1
	}
	if len(right) == 0 {
		return -1
	}
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	for i := 0; i < limit; i++ {
		comparison := comparePreReleaseIdentifier(left[i], right[i])
		if comparison != 0 {
			return comparison
		}
	}
	if len(left) < len(right) {
		return -1
	}
	if len(left) > len(right) {
		return 1
	}
	return 0
}

func comparePreReleaseIdentifier(left, right string) int {
	leftNumeric := isNumericIdentifier(left)
	rightNumeric := isNumericIdentifier(right)
	if leftNumeric && rightNumeric {
		leftValue, _ := strconv.Atoi(left)
		rightValue, _ := strconv.Atoi(right)
		if leftValue < rightValue {
			return -1
		}
		if leftValue > rightValue {
			return 1
		}
		return 0
	}
	if leftNumeric {
		return -1
	}
	if rightNumeric {
		return 1
	}
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func isNumericIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
