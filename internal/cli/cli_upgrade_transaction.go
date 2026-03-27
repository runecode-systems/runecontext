package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

type upgradeFileSnapshot struct {
	path    string
	exists  bool
	mode    os.FileMode
	content []byte
}

func runUpgradeTransaction(paths []string, apply func() error) error {
	snapshots, err := snapshotUpgradeFiles(paths)
	if err != nil {
		return err
	}
	if err := apply(); err != nil {
		if rollbackErr := restoreUpgradeFiles(snapshots); rollbackErr != nil {
			return fmt.Errorf("%v; rollback also failed: %v", err, rollbackErr)
		}
		return err
	}
	return nil
}

func snapshotUpgradeFiles(paths []string) ([]upgradeFileSnapshot, error) {
	unique := map[string]struct{}{}
	snapshots := make([]upgradeFileSnapshot, 0, len(paths))
	for _, path := range paths {
		clean := filepath.Clean(path)
		if _, ok := unique[clean]; ok {
			continue
		}
		unique[clean] = struct{}{}
		snapshot, err := snapshotUpgradeFile(clean)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}

func snapshotUpgradeFile(path string) (upgradeFileSnapshot, error) {
	snapshot := upgradeFileSnapshot{path: path}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return snapshot, nil
		}
		return upgradeFileSnapshot{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return upgradeFileSnapshot{}, err
	}
	snapshot.exists = true
	snapshot.mode = info.Mode().Perm()
	snapshot.content = data
	return snapshot, nil
}

func restoreUpgradeFiles(snapshots []upgradeFileSnapshot) error {
	for _, snapshot := range snapshots {
		if !snapshot.exists {
			if err := os.Remove(snapshot.path); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}
		if err := restoreExistingUpgradeFile(snapshot); err != nil {
			return err
		}
	}
	return nil
}

func restoreExistingUpgradeFile(snapshot upgradeFileSnapshot) error {
	if err := os.MkdirAll(filepath.Dir(snapshot.path), 0o755); err != nil {
		return err
	}
	return writeAtomicFile(snapshot.path, snapshot.content, snapshot.mode)
}
