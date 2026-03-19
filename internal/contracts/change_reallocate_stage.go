package contracts

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func stageReallocatedChange(oldChangeDir, stagedDir, oldID, newID string) ([]FileMutation, int, error) {
	info, err := os.Stat(oldChangeDir)
	if err != nil {
		return nil, 0, err
	}
	if err := ensureNoSymlinksInTree(oldChangeDir, filepath.ToSlash(filepath.Join("changes", oldID))); err != nil {
		return nil, 0, err
	}
	oldRoot := filepath.ToSlash(filepath.Join("changes", oldID))
	newRoot := filepath.ToSlash(filepath.Join("changes", newID))
	changedFiles, totalRewritten, err := copyReallocatedChangeTree(oldChangeDir, stagedDir, oldRoot, newRoot, newID)
	if err != nil {
		return nil, 0, err
	}
	if err := chmodPath(stagedDir, info.Mode().Perm()); err != nil {
		return nil, 0, err
	}
	return changedFiles, totalRewritten, nil
}

func copyReallocatedChangeTree(oldChangeDir, stagedDir, oldRoot, newRoot, newID string) ([]FileMutation, int, error) {
	changedFiles := make([]FileMutation, 0)
	totalRewritten := 0
	err := filepath.WalkDir(oldChangeDir, func(path string, entry fs.DirEntry, walkErr error) error {
		count, mutation, err := stageReallocatedEntry(oldChangeDir, stagedDir, path, entry, walkErr, oldRoot, newRoot, newID)
		if err != nil {
			return err
		}
		totalRewritten += count
		if mutation.Path != "" {
			changedFiles = append(changedFiles, mutation)
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	return changedFiles, totalRewritten, nil
}

func stageReallocatedEntry(oldChangeDir, stagedDir, path string, entry fs.DirEntry, walkErr error, oldRoot, newRoot, newID string) (int, FileMutation, error) {
	if walkErr != nil {
		return 0, FileMutation{}, walkErr
	}
	info, err := entry.Info()
	if err != nil {
		return 0, FileMutation{}, err
	}
	rel, err := filepath.Rel(oldChangeDir, path)
	if err != nil {
		return 0, FileMutation{}, err
	}
	if rel == "." {
		return 0, FileMutation{}, nil
	}
	targetPath := filepath.Join(stagedDir, rel)
	if info.IsDir() {
		return 0, FileMutation{}, createStagedDirectory(targetPath, info.Mode().Perm())
	}
	data, action, rewritten, err := stagedChangeFileData(path, rel, oldRoot, newRoot, newID)
	if err != nil {
		return 0, FileMutation{}, err
	}
	if err := writeStagedChangeFile(targetPath, data, info.Mode().Perm()); err != nil {
		return 0, FileMutation{}, err
	}
	mutation := FileMutation{Path: filepath.ToSlash(filepath.Join("changes", newID, rel)), Action: action}
	return rewritten, mutation, nil
}

func createStagedDirectory(path string, perm fs.FileMode) error {
	if err := os.MkdirAll(path, perm); err != nil {
		return err
	}
	return chmodPath(path, perm)
}

func stagedChangeFileData(path, rel, oldRoot, newRoot, newID string) ([]byte, string, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", 0, err
	}
	if rel == "status.yaml" {
		updated, err := reallocatedStatusData(path, data, newID)
		return updated, "updated", 0, err
	}
	if filepath.Ext(rel) != ".md" {
		return data, "moved", 0, nil
	}
	rewritten, count, err := rewriteMarkdownChangePathMentions(data, oldRoot, newRoot)
	if err != nil {
		return nil, "", 0, err
	}
	if count > 0 {
		return rewritten, "updated", count, nil
	}
	return rewritten, "moved", count, nil
}

func reallocatedStatusData(path string, data []byte, newID string) ([]byte, error) {
	parsed, err := parseYAML(data)
	if err != nil {
		return nil, err
	}
	statusMap, err := expectObject(path, parsed, "status file")
	if err != nil {
		return nil, err
	}
	updated := cloneMap(statusMap)
	updated["id"] = newID
	return renderStatusYAML(updated)
}

func writeStagedChangeFile(path string, data []byte, perm fs.FileMode) error {
	if err := os.WriteFile(path, data, perm); err != nil {
		return err
	}
	return chmodPath(path, perm)
}

func changeMarkdownPathPrefix(changeID string) string {
	return filepath.ToSlash(filepath.Join("changes", changeID)) + "/"
}

func ensureNoSymlinksInTree(rootPath, relativeRoot string) error {
	return filepath.WalkDir(rootPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == rootPath || entry.Type()&os.ModeSymlink == 0 {
			return nil
		}
		rel, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		return fmt.Errorf("reallocation does not support symlinks in change directories: %s", filepath.ToSlash(filepath.Join(relativeRoot, rel)))
	})
}

func rewriteMarkdownChangePathMentions(data []byte, oldRoot, newRoot string) ([]byte, int, error) {
	newLine := detectPreferredNewline(data)
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	segments := markdownTextSegments(text)
	var out strings.Builder
	total := 0
	for _, segment := range segments {
		if segment.fenced {
			out.WriteString(segment.text)
			continue
		}
		rewritten, count := rewriteLiteralPathRootInText(segment.text, oldRoot, newRoot)
		out.WriteString(rewritten)
		total += count
	}
	if total == 0 {
		return append([]byte(nil), data...), 0, nil
	}
	result := out.String()
	if newLine == "\r\n" {
		result = strings.ReplaceAll(result, "\n", "\r\n")
	}
	return []byte(result), total, nil
}

func rewriteLiteralPathRootInText(text, oldRoot, newRoot string) (string, int) {
	if oldRoot == "" || oldRoot == newRoot {
		return text, 0
	}
	var out strings.Builder
	i, count := 0, 0
	for i < len(text) {
		idx := strings.Index(text[i:], oldRoot)
		if idx < 0 {
			out.WriteString(text[i:])
			break
		}
		idx += i
		nextPos := idx + len(oldRoot)
		if !pathRootBoundaryMatches(text, idx, nextPos) {
			out.WriteString(text[i:nextPos])
			i = nextPos
			continue
		}
		out.WriteString(text[i:idx])
		out.WriteString(newRoot)
		count++
		i = nextPos
	}
	return out.String(), count
}

func pathRootBoundaryMatches(text string, idx, nextPos int) bool {
	prevOK := idx == 0 || !isMarkdownPathChar(previousRune(text, idx))
	nextOK := nextPos == len(text) || text[nextPos] == '/' || !isMarkdownPathChar(nextRune(text, nextPos))
	return prevOK && nextOK
}

func previousRune(text string, index int) rune {
	if index <= 0 || index > len(text) {
		return utf8.RuneError
	}
	r, _ := utf8.DecodeLastRuneInString(text[:index])
	return r
}

func nextRune(text string, index int) rune {
	if index < 0 || index >= len(text) {
		return utf8.RuneError
	}
	r, _ := utf8.DecodeRuneInString(text[index:])
	return r
}

func detectPreferredNewline(data []byte) string {
	text := string(data)
	if strings.Contains(text, "\r\n") && !strings.Contains(strings.ReplaceAll(text, "\r\n", ""), "\n") {
		return "\r\n"
	}
	return "\n"
}
