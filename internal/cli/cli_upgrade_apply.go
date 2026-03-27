package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

var upgradeApplyAdapterSyncFn = applyAdapterSync

func applyUpgradePlan(project *cliProject, plan upgradePlan) error {
	if plan.ConfigPath == "" {
		return fmt.Errorf("selected config path is required")
	}
	root := mutationRootForPlan(project, plan)
	paths := mutationPathsForPlan(root, plan)
	return runUpgradeTransaction(paths, func() error {
		if err := applyUpgradeStagedAndValidated(root, plan); err != nil {
			return err
		}
		return applyUpgradeAdapterPlans(plan)
	})
}

func applyUpgradeAdapterPlans(plan upgradePlan) error {
	for _, tool := range sortedMapKeys(plan.AdapterPlans) {
		state := plan.AdapterPlans[tool]
		if len(state.plan) == 0 {
			continue
		}
		if err := upgradeApplyAdapterSyncFn(state); err != nil {
			return err
		}
	}
	return nil
}

func mutationRootForPlan(project *cliProject, plan upgradePlan) string {
	if plan.ProjectRoot != "" {
		return plan.ProjectRoot
	}
	return project.absRoot
}

func mutationPathsForPlan(root string, plan upgradePlan) []string {
	paths := []string{plan.ConfigPath}
	for _, state := range plan.AdapterPlans {
		for _, mutation := range state.plan {
			paths = append(paths, filepath.Join(root, filepath.FromSlash(mutation.Path)))
		}
	}
	return paths
}

func applyUpgradeStagedAndValidated(root string, plan upgradePlan) error {
	tempRoot, err := os.MkdirTemp("", "runectx-upgrade-stage-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tempRoot) }()
	stageRoot := filepath.Join(tempRoot, "project")
	if err := copyUpgradeTree(root, stageRoot); err != nil {
		return err
	}
	configRel, err := filepath.Rel(root, plan.ConfigPath)
	if err != nil {
		return err
	}
	stagedConfig := filepath.Join(stageRoot, configRel)
	data, err := os.ReadFile(stagedConfig)
	if err != nil {
		return err
	}
	rewritten, err := rewriteRunecontextVersion(data, plan.TargetVersion)
	if err != nil {
		return err
	}
	if err := writeAtomicUpgradeConfig(stagedConfig, rewritten, configFileMode(stagedConfig)); err != nil {
		return err
	}
	if err := validateUpgradeStage(stageRoot); err != nil {
		return fmt.Errorf("validate staged upgrade tree: %w", err)
	}
	return writeAtomicUpgradeConfig(plan.ConfigPath, rewritten, configFileMode(plan.ConfigPath))
}

func validateUpgradeStage(stageRoot string) error {
	schemaRoot, err := locateSchemaRoot()
	if err != nil {
		return err
	}
	validator := contracts.NewValidator(schemaRoot)
	index, err := validator.ValidateProject(stageRoot)
	if err != nil {
		return err
	}
	defer index.Close()
	return nil
}

func copyUpgradeTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("upgrade staging rejects symlinked path %s", filepath.ToSlash(rel))
		}
		return copyUpgradeFile(path, target)
	})
}

func copyUpgradeFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode().Perm())
}
