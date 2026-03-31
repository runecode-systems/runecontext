package cli

func isComparableVersionLine(current, installed string) bool {
	currentParts, currentOK := parseSemverLike(current)
	installedParts, installedOK := parseSemverLike(installed)
	if !currentOK || !installedOK {
		return false
	}
	return currentParts.major == installedParts.major && currentParts.minor == installedParts.minor
}
