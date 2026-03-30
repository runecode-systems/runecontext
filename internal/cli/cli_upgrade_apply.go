package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

var upgradeApplyAdapterSyncFn = applyAdapterSync
var upgradeApplyMigrationRegistryFn = defaultUpgradeApplyMigrationRegistry

func applyUpgradePlan(project *cliProject, plan upgradePlan) error {
	if plan.ConfigPath == "" {
		return fmt.Errorf("selected config path is required")
	}
	root := mutationRootForPlan(project, plan)
	tempRoot, err := os.MkdirTemp("", "runectx-upgrade-stage-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tempRoot) }()
	stageRoot := filepath.Join(tempRoot, "project")
	stage, err := applyUpgradeStagedAndValidated(root, stageRoot, plan)
	if err != nil {
		return err
	}
	return runUpgradeTransaction(stage.transactionPaths(), func() error {
		return applyStageCommit(root, stage)
	})
}

func mutationRootForPlan(project *cliProject, plan upgradePlan) string {
	if plan.ProjectRoot != "" {
		return plan.ProjectRoot
	}
	return project.absRoot
}

type stagedUpgradeTree struct {
	stageRoot    string
	changedFiles []string
	deletedFiles []string
}

func (s stagedUpgradeTree) transactionPaths() []string {
	paths := make([]string, 0, len(s.changedFiles)+len(s.deletedFiles))
	paths = append(paths, s.changedFiles...)
	paths = append(paths, s.deletedFiles...)
	return paths
}

func applyUpgradeStagedAndValidated(root, stageRoot string, plan upgradePlan) (stagedUpgradeTree, error) {
	if err := copyUpgradeTree(root, stageRoot); err != nil {
		return stagedUpgradeTree{}, err
	}
	stageConfig, err := stagedConfigPath(root, stageRoot, plan.ConfigPath)
	if err != nil {
		return stagedUpgradeTree{}, err
	}
	stageCtx := upgradeMigrationContext{Root: stageRoot, ConfigPath: stageConfig}
	if err := executeUpgradeHops(stageCtx, plan); err != nil {
		return stagedUpgradeTree{}, err
	}
	if err := applyUpgradeAdapterPlansInStage(stageRoot, plan); err != nil {
		return stagedUpgradeTree{}, err
	}
	if err := validateUpgradeStage(stageRoot); err != nil {
		return stagedUpgradeTree{}, fmt.Errorf("validate staged upgrade tree: %w", err)
	}
	changedRel, deletedRel, err := diffUpgradeTrees(root, stageRoot)
	if err != nil {
		return stagedUpgradeTree{}, err
	}
	return stagedUpgradeTree{
		stageRoot:    stageRoot,
		changedFiles: toAbsoluteUpgradePaths(root, changedRel),
		deletedFiles: toAbsoluteUpgradePaths(root, deletedRel),
	}, nil
}

func executeUpgradeHops(stageCtx upgradeMigrationContext, plan upgradePlan) error {
	registry := upgradeApplyMigrationRegistryFn()
	for _, hop := range plan.UpgradeHops {
		migration := registry.forHop(hop)
		if err := migration.Apply(stageCtx, hop); err != nil {
			return fmt.Errorf("apply upgrade hop %s -> %s: %w", hop.From, hop.To, err)
		}
		if err := validateUpgradeStage(stageCtx.Root); err != nil {
			return fmt.Errorf("validate staged upgrade tree after hop %s -> %s: %w", hop.From, hop.To, err)
		}
		if err := migration.Verify(stageCtx, hop); err != nil {
			return fmt.Errorf("verify upgrade hop %s -> %s: %w", hop.From, hop.To, err)
		}
	}
	return nil
}

func applyUpgradeAdapterPlansInStage(stageRoot string, plan upgradePlan) error {
	includeCreate := plan.TargetVersion != plan.CurrentVersion
	states, conflicts, _, err := collectUpgradeAdapterPlansFn(stageRoot, includeCreate)
	if err != nil {
		return err
	}
	if len(conflicts) > 0 {
		return fmt.Errorf("managed artifact conflicts detected in staged tree: %s", conflicts[0])
	}
	for _, tool := range sortedMapKeys(states) {
		state := states[tool]
		if len(state.plan) == 0 {
			continue
		}
		if err := upgradeApplyAdapterSyncFn(state); err != nil {
			return err
		}
	}
	return nil
}

func applyStageCommit(root string, stage stagedUpgradeTree) error {
	if err := applyStageDeletes(root, stage.deletedFiles); err != nil {
		return err
	}
	if err := applyStageChanges(root, stage); err != nil {
		return err
	}
	return nil
}

func toAbsoluteUpgradePaths(root string, relPaths []string) []string {
	paths := make([]string, 0, len(relPaths))
	for _, rel := range relPaths {
		paths = append(paths, filepath.Join(root, rel))
	}
	return paths
}

func rewriteStageRunecontextVersion(configPath, target string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	rewritten, err := rewriteRunecontextVersion(data, target)
	if err != nil {
		return err
	}
	if err := writeAtomicUpgradeConfig(configPath, rewritten, configFileMode(configPath)); err != nil {
		return err
	}
	return nil
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
