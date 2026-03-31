package cli

import (
	"fmt"
	"strings"
)

func executeUpgradeHops(stageCtx upgradeMigrationContext, plan upgradePlan) error {
	registry := upgradeApplyMigrationRegistryFn()
	stagedVersion := stageVersionForHopExecution(stageCtx, plan)
	for _, hop := range plan.UpgradeHops {
		nextVersion, advanced, err := advanceStagedVersionToHopFrom(stageCtx, stagedVersion, hop)
		if err != nil {
			return err
		}
		stagedVersion = nextVersion
		_ = advanced
		if err := applyAndValidateUpgradeHop(stageCtx, registry, hop); err != nil {
			return err
		}
		stagedVersion = strings.TrimSpace(hop.To)
	}
	return nil
}

func stageVersionForHopExecution(stageCtx upgradeMigrationContext, plan upgradePlan) string {
	stagedVersion := strings.TrimSpace(plan.CurrentVersion)
	if stagedVersion != "" {
		return stagedVersion
	}
	return readRunecontextVersionFromConfig(stageCtx.ConfigPath)
}

func applyAndValidateUpgradeHop(stageCtx upgradeMigrationContext, registry upgradeApplyMigrationRegistry, hop upgradeHop) error {
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
	return nil
}

func advanceStagedVersionToHopFrom(stageCtx upgradeMigrationContext, stagedVersion string, hop upgradeHop) (string, bool, error) {
	from := strings.TrimSpace(hop.From)
	if from == "" {
		return stagedVersion, false, fmt.Errorf("upgrade hop missing from version")
	}
	current := strings.TrimSpace(stagedVersion)
	if current == "" {
		current = readRunecontextVersionFromConfig(stageCtx.ConfigPath)
	}
	if current == from {
		return current, false, nil
	}
	comparison, comparable := compareKnownRunecontextVersions(current, from)
	if !comparable {
		return current, false, fmt.Errorf("cannot compare staged runecontext_version %s with hop start %s", current, from)
	}
	if comparison > 0 {
		return current, false, fmt.Errorf("staged runecontext_version %s is newer than hop start %s", current, from)
	}
	if err := rewriteStageRunecontextVersion(stageCtx.ConfigPath, from); err != nil {
		return current, false, fmt.Errorf("advance staged runecontext_version to %s: %w", from, err)
	}
	return from, true, nil
}
