package contracts

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func sliceDifference(items, base []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	for _, item := range base {
		seen[item] = struct{}{}
	}
	result := make([]string, 0)
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		result = append(result, item)
	}
	return uniqueSortedStrings(result)
}

func uniqueStringsInOrder(items []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func sortFileMutations(items []FileMutation) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Path == items[j].Path {
			return items[i].Action < items[j].Action
		}
		return items[i].Path < items[j].Path
	})
}

func createUniqueChangeDir(contentRoot string, now time.Time, title string, entropy io.Reader) (string, string, error) {
	changesRoot := filepath.Join(contentRoot, "changes")
	for attempt := 0; attempt < maxCreateChangeDirAttempts; attempt++ {
		id, err := AllocateChangeID(contentRoot, now, title, entropy)
		if err != nil {
			return "", "", err
		}
		changeDir := filepath.Join(changesRoot, id)
		if err := os.Mkdir(changeDir, 0o755); err != nil {
			if os.IsExist(err) {
				continue
			}
			return "", "", err
		}
		return id, changeDir, nil
	}
	return "", "", fmt.Errorf("could not allocate a unique change directory after %d attempts", maxCreateChangeDirAttempts)
}
