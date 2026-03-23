package contracts

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func prepareStatusRewrite(v *Validator, path string, raw map[string]any) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("validator is required")
	}
	data, err := renderStatusYAML(raw)
	if err != nil {
		return nil, err
	}
	if err := v.ValidateYAMLFile("change-status.schema.json", path, data); err != nil {
		return nil, err
	}
	return data, nil
}

func applyFileRewritesTransaction(rewrites []fileRewrite, postWriteValidate func() error) error {
	if len(rewrites) == 0 {
		return runPostRewriteValidation(postWriteValidate)
	}
	backups, err := createFileRewriteBackups(rewrites)
	if err != nil {
		return err
	}
	if err := writeFileRewrites(rewrites, backups); err != nil {
		return err
	}
	if err := runPostRewriteValidation(postWriteValidate); err != nil {
		rollbackErr := restoreFileBackups(backups)
		return combineFileRewriteRollbackError(err, rollbackErr)
	}
	return nil
}

func runPostRewriteValidation(postWriteValidate func() error) error {
	if postWriteValidate == nil {
		return nil
	}
	return postWriteValidate()
}

func createFileRewriteBackups(rewrites []fileRewrite) ([]fileBackup, error) {
	backups := make([]fileBackup, 0, len(rewrites))
	for _, rewrite := range rewrites {
		backup, err := createFileBackup(rewrite.Path)
		if err != nil {
			return nil, err
		}
		backups = append(backups, backup)
	}
	return backups, nil
}

func createFileBackup(path string) (fileBackup, error) {
	if err := ensurePathAndParentAreNotSymlinks(path); err != nil {
		return fileBackup{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fileBackup{Path: path, Exists: false}, nil
		}
		return fileBackup{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fileBackup{}, err
	}
	return fileBackup{Path: path, Data: data, Perm: info.Mode().Perm(), Exists: true}, nil
}

func writeFileRewrites(rewrites []fileRewrite, backups []fileBackup) error {
	written := 0
	for _, rewrite := range rewrites {
		if err := os.MkdirAll(filepath.Dir(rewrite.Path), 0o755); err != nil {
			rollbackErr := restoreFileBackups(backups[:written])
			return combineFileRewriteRollbackError(err, rollbackErr)
		}
		perm := rewrite.Perm
		if perm == 0 {
			perm = backups[written].Perm
		}
		if perm == 0 {
			perm = 0o644
		}
		if err := writeFileAtomically(rewrite.Path, rewrite.Data, perm); err != nil {
			rollbackErr := restoreFileBackups(backups[:written])
			return combineFileRewriteRollbackError(err, rollbackErr)
		}
		written++
	}
	return nil
}

func restoreFileBackups(backups []fileBackup) error {
	errMessages := make([]string, 0)
	for i := len(backups) - 1; i >= 0; i-- {
		backup := backups[i]
		if !backup.Exists {
			if err := removeAllPath(backup.Path); err != nil && !os.IsNotExist(err) {
				errMessages = append(errMessages, fmt.Sprintf("restore %q: %v", filepath.ToSlash(backup.Path), err))
			}
			continue
		}
		if err := writeFileAtomically(backup.Path, backup.Data, backup.Perm); err != nil {
			errMessages = append(errMessages, fmt.Sprintf("restore %q: %v", filepath.ToSlash(backup.Path), err))
		}
	}
	if len(errMessages) == 0 {
		return nil
	}
	return errors.New(strings.Join(errMessages, "; "))
}

func combineFileRewriteRollbackError(operationErr, rollbackErr error) error {
	if rollbackErr == nil {
		return operationErr
	}
	return fmt.Errorf("%v; rollback also failed and manual recovery may be required: %v", operationErr, rollbackErr)
}

func ensurePathAndParentAreNotSymlinks(path string) error {
	cleanPath := filepath.Clean(path)
	for _, candidate := range []string{filepath.Dir(cleanPath), cleanPath} {
		info, err := lstatPath(candidate)
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

func writeFileAtomically(path string, data []byte, perm fs.FileMode) error {
	if err := ensurePathAndParentAreNotSymlinks(path); err != nil {
		return err
	}
	tempPath, err := prepareAtomicWriteTempPath(path)
	if err != nil {
		return err
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = removeAllPath(tempPath)
		}
	}()
	if err := writeTempMutationFile(tempPath, data, perm); err != nil {
		return err
	}
	if err := replacePathAtomically(tempPath, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func prepareAtomicWriteTempPath(path string) (string, error) {
	tempFile, err := createTempFilePath(filepath.Dir(path), ".mutation-*")
	if err != nil {
		return "", err
	}
	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		_ = removeAllPath(tempPath)
		return "", err
	}
	return tempPath, nil
}

func writeTempMutationFile(path string, data []byte, perm fs.FileMode) error {
	if err := writeFilePath(path, data, perm); err != nil {
		return err
	}
	return chmodPath(path, perm)
}

func replacePathAtomically(tempPath, targetPath string) error {
	err := renamePath(tempPath, targetPath)
	if err == nil || !atomicReplaceNeedsFallback {
		return err
	}
	return replacePathAtomicallyWithFallback(tempPath, targetPath, err)
}

func replacePathAtomicallyWithFallback(tempPath, targetPath string, renameErr error) error {
	if err := ensureAtomicReplaceTarget(targetPath, renameErr); err != nil {
		return err
	}
	backupPath, err := createAtomicReplaceBackup(targetPath)
	if err != nil {
		return err
	}
	return swapAtomicReplacePaths(tempPath, targetPath, backupPath, renameErr)
}

func ensureAtomicReplaceTarget(targetPath string, renameErr error) error {
	if err := ensurePathAndParentAreNotSymlinks(targetPath); err != nil {
		return err
	}
	if _, err := os.Stat(targetPath); err == nil {
		return nil
	}
	return renameErr
}

func createAtomicReplaceBackup(targetPath string) (string, error) {
	backupFile, err := createTempFilePath(filepath.Dir(targetPath), ".replace-backup-*")
	if err != nil {
		return "", err
	}
	backupPath := backupFile.Name()
	if err := backupFile.Close(); err != nil {
		_ = removeAllPath(backupPath)
		return "", err
	}
	if err := removeAllPath(backupPath); err != nil {
		return "", err
	}
	return backupPath, nil
}

func swapAtomicReplacePaths(tempPath, targetPath, backupPath string, renameErr error) error {
	cleanupBackup := true
	defer func() {
		if cleanupBackup {
			_ = removeAllPath(backupPath)
		}
	}()
	if err := renamePath(targetPath, backupPath); err != nil {
		return renameErr
	}
	if err := renamePath(tempPath, targetPath); err != nil {
		rollbackErr := renamePath(backupPath, targetPath)
		return combineFileRewriteRollbackError(err, rollbackErr)
	}
	cleanupBackup = false
	_ = removeAllPath(backupPath)
	return nil
}
