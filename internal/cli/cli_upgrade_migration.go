package cli

import "fmt"

type upgradeMigrationContext struct {
	Root       string
	ConfigPath string
}

type upgradeHopMigration interface {
	Apply(ctx upgradeMigrationContext, hop upgradeHop) error
	Verify(ctx upgradeMigrationContext, hop upgradeHop) error
}

type upgradeApplyMigrationRegistry struct {
	hopSpecific map[upgradeEdgeKey]upgradeHopMigration
	defaultHop  upgradeHopMigration
}

func defaultUpgradeApplyMigrationRegistry() upgradeApplyMigrationRegistry {
	defaultMigration := defaultVersionRewriteUpgradeMigration{}
	hopSpecific := map[upgradeEdgeKey]upgradeHopMigration{}
	hopSpecific[upgradeEdgeKey{From: "0.1.0-alpha.12", To: "0.1.0-alpha.13"}] = assuranceLayoutAlpha13Migration{}
	return upgradeApplyMigrationRegistry{
		hopSpecific: hopSpecific,
		defaultHop:  defaultMigration,
	}
}

func (r upgradeApplyMigrationRegistry) forHop(hop upgradeHop) upgradeHopMigration {
	if migration, ok := r.hopSpecific[upgradeEdgeKey{From: hop.From, To: hop.To}]; ok {
		return migration
	}
	if r.defaultHop == nil {
		return defaultVersionRewriteUpgradeMigration{}
	}
	return r.defaultHop
}

type defaultVersionRewriteUpgradeMigration struct{}

func (m defaultVersionRewriteUpgradeMigration) Apply(ctx upgradeMigrationContext, hop upgradeHop) error {
	return rewriteStageRunecontextVersion(ctx.ConfigPath, hop.To)
}

func (m defaultVersionRewriteUpgradeMigration) Verify(ctx upgradeMigrationContext, hop upgradeHop) error {
	version := readRunecontextVersionFromConfig(ctx.ConfigPath)
	if version != hop.To {
		return fmt.Errorf("expected staged runecontext_version %s after hop, got %s", hop.To, version)
	}
	return nil
}
