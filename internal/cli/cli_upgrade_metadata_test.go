package cli

import "testing"

func TestUpgradeCommandMetadataMatchesParserRequiredFlags(t *testing.T) {
	metadata := upgradeCommandMetadata()
	apply := metadata.Subcommands[0]
	cli := metadata.Subcommands[1]
	cliApply := cli.Subcommands[0]

	if flag := flagMetadataByName(apply.Flags, "--target-version"); flag == nil || flag.Required {
		t.Fatalf("expected upgrade apply --target-version to be optional in metadata")
	}
	if flag := flagMetadataByName(cliApply.Flags, "--target-version"); flag == nil || !flag.Required {
		t.Fatalf("expected upgrade cli apply --target-version to be required in metadata")
	}
	if apply.Usage != upgradeApplyUsage {
		t.Fatalf("expected upgrade apply usage parity, got %q want %q", apply.Usage, upgradeApplyUsage)
	}
	if cliApply.Usage != upgradeCLIApplyUsage {
		t.Fatalf("expected upgrade cli apply usage parity, got %q want %q", cliApply.Usage, upgradeCLIApplyUsage)
	}
}
