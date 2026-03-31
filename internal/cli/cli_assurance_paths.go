package cli

import "path/filepath"

func assuranceRootPath(root string) string {
	return filepath.Join(root, "runecontext", "assurance")
}

func assuranceBaselinePath(root string) string {
	return filepath.Join(assuranceRootPath(root), "baseline.yaml")
}

func assuranceBackfillRootPath(root string) string {
	return filepath.Join(assuranceRootPath(root), "backfill")
}

func assuranceBackfillRelativePath(commit string) string {
	return filepath.ToSlash(filepath.Join("runecontext", "assurance", "backfill", "imported-git-history-"+commit+".json"))
}
