package cli

import (
	"os"
	"path/filepath"
	"sort"
)

func diffUpgradeTrees(root, stageRoot string, policy upgradeWalkPolicy) ([]string, []string, error) {
	liveFiles, err := collectRegularFilesByRel(root, policy)
	if err != nil {
		return nil, nil, err
	}
	stageFiles, err := collectRegularFilesByRel(stageRoot, upgradeWalkPolicy{})
	if err != nil {
		return nil, nil, err
	}
	changed, err := collectChangedStageRelPaths(liveFiles, stageFiles)
	if err != nil {
		return nil, nil, err
	}
	deleted := collectDeletedStageRelPaths(liveFiles, stageFiles)
	sort.Strings(changed)
	sort.Strings(deleted)
	return changed, deleted, nil
}

func collectChangedStageRelPaths(liveFiles, stageFiles map[string]string) ([]string, error) {
	changed := make([]string, 0)
	for rel, stagePath := range stageFiles {
		livePath, ok := liveFiles[rel]
		if !ok {
			changed = append(changed, rel)
			continue
		}
		equal, err := areUpgradeFilesEqual(livePath, stagePath)
		if err != nil {
			return nil, err
		}
		if !equal {
			changed = append(changed, rel)
		}
	}
	return changed, nil
}

func collectDeletedStageRelPaths(liveFiles, stageFiles map[string]string) []string {
	deleted := make([]string, 0)
	for rel := range liveFiles {
		if _, ok := stageFiles[rel]; ok {
			continue
		}
		deleted = append(deleted, rel)
	}
	return deleted
}

func collectRegularFilesByRel(root string, policy upgradeWalkPolicy) (map[string]string, error) {
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		rel, decision, err := classifyUpgradeWalkEntry(root, path, entry, walkErr, policy)
		if err != nil {
			return err
		}
		switch decision {
		case upgradeWalkSkip:
			return nil
		case upgradeWalkSkipDir:
			return filepath.SkipDir
		case upgradeWalkDir:
			return nil
		}
		files[rel] = path
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func areUpgradeFilesEqual(a, b string) (bool, error) {
	aData, err := os.ReadFile(a)
	if err != nil {
		return false, err
	}
	bData, err := os.ReadFile(b)
	if err != nil {
		return false, err
	}
	if string(aData) != string(bData) {
		return false, nil
	}
	aInfo, err := os.Stat(a)
	if err != nil {
		return false, err
	}
	bInfo, err := os.Stat(b)
	if err != nil {
		return false, err
	}
	return aInfo.Mode().Perm() == bInfo.Mode().Perm(), nil
}
