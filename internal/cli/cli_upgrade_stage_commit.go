package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func stagedConfigPath(root, stageRoot, configPath string) (string, error) {
	configRel, err := filepath.Rel(root, configPath)
	if err != nil {
		return "", err
	}
	if !isPathWithinUpgradeRoot(configRel) {
		return "", fmt.Errorf("upgrade config path escapes project root: %s", filepath.ToSlash(configPath))
	}
	return filepath.Join(stageRoot, configRel), nil
}

func applyStageDeletes(root string, deletedFiles []string) error {
	for _, path := range deletedFiles {
		if err := deleteOneUpgradePath(path); err != nil {
			return err
		}
	}
	return pruneEmptyUpgradeParentDirs(root, deletedFiles)
}

func deleteOneUpgradePath(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func applyStageChanges(root string, stage stagedUpgradeTree) error {
	for _, path := range stage.changedFiles {
		if err := prepareUpgradeDestination(root, path); err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		stagePath := filepath.Join(stage.stageRoot, rel)
		if err := copyUpgradeFile(stagePath, path); err != nil {
			return err
		}
	}
	return nil
}

func prepareUpgradeDestination(root, path string) error {
	if err := removeUpgradeDirectoryAtPath(path); err != nil {
		return err
	}
	return ensureUpgradeParentDirs(root, path)
}

func removeUpgradeDirectoryAtPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}
	return os.RemoveAll(path)
}

func pruneEmptyUpgradeParentDirs(root string, deletedFiles []string) error {
	for _, path := range deletedFiles {
		if err := pruneOneUpgradeParentChain(root, filepath.Dir(path)); err != nil {
			return err
		}
	}
	return nil
}

func pruneOneUpgradeParentChain(root, startDir string) error {
	for dir := startDir; ; dir = filepath.Dir(dir) {
		inRoot, err := isUpgradePathUnderRoot(root, dir)
		if err != nil {
			return err
		}
		if !inRoot {
			return nil
		}
		stop, err := removeUpgradeEmptyDir(dir)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
	}
}

func removeUpgradeEmptyDir(dir string) (bool, error) {
	if err := os.Remove(dir); err != nil {
		if os.IsNotExist(err) || isUpgradeDirNotEmpty(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func isUpgradePathUnderRoot(root, path string) (bool, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false, err
	}
	if rel == "." {
		return false, nil
	}
	return isPathWithinUpgradeRoot(rel), nil
}

func ensureUpgradeParentDirs(root, path string) error {
	parent := filepath.Dir(path)
	rel, err := filepath.Rel(root, parent)
	if err != nil {
		return err
	}
	if rel == "." {
		return nil
	}
	if !isPathWithinUpgradeRoot(rel) {
		return fmt.Errorf("upgrade destination path escapes project root: %s", filepath.ToSlash(path))
	}
	current := root
	for _, segment := range strings.Split(rel, string(filepath.Separator)) {
		if segment == "" || segment == "." {
			continue
		}
		current = filepath.Join(current, segment)
		if err := ensureUpgradePathIsDirectory(current); err != nil {
			return err
		}
	}
	return nil
}

func ensureUpgradePathIsDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(path, 0o755)
		}
		return err
	}
	if info.IsDir() {
		return nil
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	return os.Mkdir(path, 0o755)
}

func isUpgradeDirNotEmpty(err error) bool {
	pathErr, ok := err.(*os.PathError)
	if !ok {
		return false
	}
	return errors.Is(pathErr.Err, syscall.ENOTEMPTY)
}

func isPathWithinUpgradeRoot(rel string) bool {
	clean := filepath.Clean(rel)
	if clean == "." {
		return true
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return false
	}
	return !filepath.IsAbs(clean)
}
