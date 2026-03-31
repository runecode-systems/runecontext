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
	policy := newUpgradeWalkPolicy(root, plan)
	if err := copyUpgradeTree(root, stageRoot, policy); err != nil {
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
	if err := finalizeUpgradeTargetVersion(stageCtx, plan); err != nil {
		return stagedUpgradeTree{}, err
	}
	if err := applyUpgradeAdapterPlansInStage(stageRoot, plan); err != nil {
		return stagedUpgradeTree{}, err
	}
	if err := validateUpgradeStage(stageRoot); err != nil {
		return stagedUpgradeTree{}, fmt.Errorf("validate staged upgrade tree: %w", err)
	}
	changedRel, deletedRel, err := diffUpgradeTrees(root, stageRoot, policy)
	if err != nil {
		return stagedUpgradeTree{}, err
	}
	return stagedUpgradeTree{
		stageRoot:    stageRoot,
		changedFiles: toAbsoluteUpgradePaths(root, changedRel),
		deletedFiles: toAbsoluteUpgradePaths(root, deletedRel),
	}, nil
}

func finalizeUpgradeTargetVersion(stageCtx upgradeMigrationContext, plan upgradePlan) error {
	if plan.TargetVersion == plan.CurrentVersion {
		return nil
	}
	return rewriteStageRunecontextVersion(stageCtx.ConfigPath, plan.TargetVersion)
}

func applyUpgradeAdapterPlansInStage(stageRoot string, plan upgradePlan) error {
	states, err := rebuildUpgradeAdapterStatesInStage(stageRoot, plan)
	if err != nil {
		return err
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

func rebuildUpgradeAdapterStatesInStage(stageRoot string, plan upgradePlan) (map[string]adapterSyncState, error) {
	tools, err := collectStageAdapterTools(stageRoot, plan)
	if err != nil {
		return nil, err
	}
	staged := make(map[string]adapterSyncState, len(tools))
	for _, tool := range tools {
		includeCreate := includeCreateForTool(plan.AdapterPlans[tool])
		nextState, conflicts, _, skip, err := collectSingleUpgradeAdapterPlan(stageRoot, tool, includeCreate, nil, nil)
		if err != nil {
			return nil, err
		}
		if len(conflicts) > 0 {
			return nil, fmt.Errorf("staged upgrade adapter conflicts detected for %s", tool)
		}
		if skip || len(nextState.plan) == 0 {
			continue
		}
		staged[tool] = nextState
	}
	return staged, nil
}

func includeCreateForTool(state adapterSyncState) bool {
	for _, mutation := range state.plan {
		if mutation.Action == "created" {
			return true
		}
	}
	return false
}

func collectStageAdapterTools(stageRoot string, plan upgradePlan) ([]string, error) {
	tools := make([]string, 0, len(plan.AdapterPlans))
	seen := map[string]struct{}{}
	for _, tool := range []string{"opencode", "claude-code", "codex"} {
		if _, ok := plan.AdapterPlans[tool]; ok {
			seen[tool] = struct{}{}
			tools = append(tools, tool)
			continue
		}
		managed, err := hasManagedHostNativeArtifactsForTool(stageRoot, tool)
		if err != nil {
			return nil, err
		}
		if managed {
			seen[tool] = struct{}{}
			tools = append(tools, tool)
		}
	}
	return tools, nil
}

func applyStageCommit(root string, stage stagedUpgradeTree) error {
	if err := applyStageDeletes(root, stage.deletedFiles, stage.changedFiles); err != nil {
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

func copyUpgradeTree(src, dst string, policy upgradeWalkPolicy) error {
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, walkErr error) error {
		rel, decision, err := classifyUpgradeWalkEntry(src, path, entry, walkErr, policy)
		if err != nil {
			return err
		}
		switch decision {
		case upgradeWalkSkip:
			return nil
		case upgradeWalkSkipDir:
			return filepath.SkipDir
		case upgradeWalkDir:
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		target := filepath.Join(dst, rel)
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
