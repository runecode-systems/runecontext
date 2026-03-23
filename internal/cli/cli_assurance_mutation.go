package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

var assuranceAtomicReplaceNeedsFallback = runtime.GOOS == "windows"

func writeAtomicFile(path string, data []byte, perm os.FileMode) error {
	if err := ensurePathAndParentAreNotSymlinks(path); err != nil {
		return err
	}
	tempPath, err := prepareAssuranceAtomicTempPath(path)
	if err != nil {
		return err
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(tempPath)
		}
	}()
	if err := writeAssuranceTempFile(tempPath, data, fs.FileMode(perm)); err != nil {
		return err
	}
	if err := replaceAssurancePathAtomically(tempPath, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func ensurePathAndParentAreNotSymlinks(path string) error {
	cleanPath := filepath.Clean(path)
	for _, candidate := range []string{filepath.Dir(cleanPath), cleanPath} {
		info, err := os.Lstat(candidate)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("mutation does not support symlinked targets: %s", filepath.ToSlash(candidate))
		}
	}
	return nil
}

func prepareAssuranceAtomicTempPath(path string) (string, error) {
	tempFile, err := os.CreateTemp(filepath.Dir(path), ".assurance-mutation-*")
	if err != nil {
		return "", err
	}
	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		_ = os.RemoveAll(tempPath)
		return "", err
	}
	return tempPath, nil
}

func writeAssuranceTempFile(path string, data []byte, perm fs.FileMode) error {
	if err := os.WriteFile(path, data, perm); err != nil {
		return err
	}
	return os.Chmod(path, perm)
}

func replaceAssurancePathAtomically(tempPath, targetPath string) error {
	err := os.Rename(tempPath, targetPath)
	if err == nil || !assuranceAtomicReplaceNeedsFallback {
		return err
	}
	return replaceAssurancePathWithFallback(tempPath, targetPath, err)
}

func replaceAssurancePathWithFallback(tempPath, targetPath string, renameErr error) error {
	if err := ensurePathAndParentAreNotSymlinks(targetPath); err != nil {
		return err
	}
	if _, err := os.Stat(targetPath); err != nil {
		return renameErr
	}
	backupPath, err := createAssuranceAtomicReplaceBackup(targetPath)
	if err != nil {
		return err
	}
	cleanupBackup := true
	defer func() {
		if cleanupBackup {
			_ = os.RemoveAll(backupPath)
		}
	}()
	if err := os.Rename(targetPath, backupPath); err != nil {
		return renameErr
	}
	if err := os.Rename(tempPath, targetPath); err != nil {
		rollbackErr := os.Rename(backupPath, targetPath)
		return combineAssuranceRollbackError(err, rollbackErr)
	}
	cleanupBackup = false
	_ = os.RemoveAll(backupPath)
	return nil
}

func createAssuranceAtomicReplaceBackup(targetPath string) (string, error) {
	backupFile, err := os.CreateTemp(filepath.Dir(targetPath), ".assurance-replace-backup-*")
	if err != nil {
		return "", err
	}
	backupPath := backupFile.Name()
	if err := backupFile.Close(); err != nil {
		_ = os.RemoveAll(backupPath)
		return "", err
	}
	if err := os.RemoveAll(backupPath); err != nil {
		return "", err
	}
	return backupPath, nil
}

func combineAssuranceRollbackError(operationErr, rollbackErr error) error {
	if rollbackErr == nil {
		return operationErr
	}
	return fmt.Errorf("%v; rollback also failed and manual recovery may be required: %v", operationErr, rollbackErr)
}
