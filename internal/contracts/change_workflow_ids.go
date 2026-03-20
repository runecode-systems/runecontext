package contracts

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// AllocateChangeID returns the next year-scoped change ID candidate.
//
// Callers must still create the target change directory atomically and retry on
// EEXIST-style collisions; this helper allocates but does not reserve IDs.
func AllocateChangeID(contentRoot string, now time.Time, title string, entropy io.Reader) (string, error) {
	if entropy == nil {
		entropy = cryptorand.Reader
	}
	existing, err := existingChangeIDs(filepath.Join(contentRoot, "changes"))
	if err != nil {
		return "", err
	}
	year := now.Year()
	nextCounter := nextYearlyChangeCounter(existing, year)
	if nextCounter > 999 {
		return "", fmt.Errorf("cannot allocate change ID for %d: yearly counter exceeds 999", year)
	}
	return allocateUniqueChangeID(existing, year, nextCounter, slugifyTitle(title), entropy)
}

func nextYearlyChangeCounter(existing map[string]struct{}, year int) int {
	nextCounter := 1
	for id := range existing {
		idYear, counter, _, _, err := parseChangeID(id)
		if err == nil && idYear == year && counter >= nextCounter {
			nextCounter = counter + 1
		}
	}
	return nextCounter
}

func allocateUniqueChangeID(existing map[string]struct{}, year, counter int, slug string, entropy io.Reader) (string, error) {
	for attempt := 0; attempt < 32; attempt++ {
		suffix, err := randomChangeSuffix(entropy)
		if err != nil {
			return "", err
		}
		candidate := fmt.Sprintf("CHG-%04d-%03d-%s-%s", year, counter, suffix, slug)
		if _, ok := existing[candidate]; !ok {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate a unique change ID after repeated suffix collisions")
}

func existingChangeIDs(changesRoot string) (map[string]struct{}, error) {
	ids := map[string]struct{}{}
	entries, err := os.ReadDir(changesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return ids, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() && changeIDPattern.MatchString(entry.Name()) {
			ids[entry.Name()] = struct{}{}
		}
	}
	return ids, nil
}

func parseChangeID(id string) (int, int, string, string, error) {
	if !changeIDPattern.MatchString(id) {
		return 0, 0, "", "", fmt.Errorf("invalid change ID %q", id)
	}
	parts := strings.SplitN(id, "-", 5)
	if len(parts) != 5 {
		return 0, 0, "", "", fmt.Errorf("invalid change ID %q", id)
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, "", "", err
	}
	counter, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, "", "", err
	}
	return year, counter, parts[3], parts[4], nil
}

func randomChangeSuffix(entropy io.Reader) (string, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(entropy, buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func slugifyTitle(title string) string {
	return slugifyASCII(title, "change")
}

func slugifyASCII(value, fallback string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return fallback
	}
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		lastDash = appendSlugRune(&b, r, lastDash)
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return fallback
	}
	return slug
}

func appendSlugRune(b *strings.Builder, r rune, lastDash bool) bool {
	if isSlugAlphaNumeric(r) {
		b.WriteRune(r)
		return false
	}
	if !isSlugSeparator(r) || b.Len() == 0 || lastDash {
		return lastDash
	}
	b.WriteByte('-')
	return true
}

func isSlugAlphaNumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

func isSlugSeparator(r rune) bool {
	return r == '-' || unicode.IsSpace(r) || r == '_' || r == '/'
}
