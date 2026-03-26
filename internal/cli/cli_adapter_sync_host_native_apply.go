package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func applyHostNativeArtifactWrites(absRoot string, artifacts []hostNativeArtifact, plannedWrites map[string]struct{}) error {
	for _, artifact := range artifacts {
		rel := filepath.ToSlash(artifact.relPath)
		if _, ok := plannedWrites[rel]; !ok {
			continue
		}
		if err := writeHostNativeArtifact(absRoot, rel, artifact.content); err != nil {
			return err
		}
	}
	return nil
}

func writeHostNativeArtifact(absRoot, rel string, content []byte) error {
	path := filepath.Join(absRoot, filepath.FromSlash(rel))
	if err := validateExistingHostNativeForWrite(path, rel, content); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return writeAtomicFile(path, content, 0o644)
}

func validateExistingHostNativeForWrite(path, rel string, desired []byte) error {
	current, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if string(current) == string(desired) {
		return nil
	}
	return validateHostNativeOwnershipForWrite(current, rel, hostNativeArtifactFromContent(rel, desired))
}

func applyHostNativeArtifactDeletes(absRoot, tool string, plan []contracts.FileMutation) error {
	for _, mutation := range plan {
		if mutation.Action != "deleted" || !isHostNativePath(mutation.Path) {
			continue
		}
		if err := removeHostNativeArtifact(absRoot, mutation.Path, tool); err != nil {
			return err
		}
	}
	return pruneHostNativeRoots(absRoot)
}

func removeHostNativeArtifact(absRoot, rel, tool string) error {
	path := filepath.Join(absRoot, filepath.FromSlash(rel))
	if err := validateExistingHostNativeForDelete(path, rel, tool); err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func validateExistingHostNativeForDelete(path, rel, tool string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return validateHostNativeOwnershipForDelete(data, rel, tool)
}

func hostNativeArtifactFromContent(rel string, content []byte) hostNativeArtifact {
	header, ok := parseHostNativeOwnershipHeader(content)
	if !ok {
		return hostNativeArtifact{relPath: rel}
	}
	return hostNativeArtifact{
		relPath: rel,
		tool:    header.Tool,
		kind:    header.Kind,
		id:      header.ID,
		content: content,
	}
}

func pruneHostNativeRoots(absRoot string) error {
	for _, root := range []string{".opencode", ".claude", ".agents"} {
		if err := pruneEmptyDirs(filepath.Join(absRoot, root)); err != nil {
			return err
		}
	}
	return nil
}

func isHostNativePath(path string) bool {
	path = filepath.ToSlash(path)
	return strings.HasPrefix(path, ".opencode/") || strings.HasPrefix(path, ".claude/") || strings.HasPrefix(path, ".agents/")
}
