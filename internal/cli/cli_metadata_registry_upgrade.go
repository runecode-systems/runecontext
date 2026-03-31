package cli

func upgradeCommandMetadata() CommandMetadata {
	return CommandMetadata{
		Name:  "upgrade",
		Path:  "upgrade",
		Usage: upgradeUsage,
		Flags: readOnlyCommandFlags(upgradeFlags()),
		Subcommands: []CommandMetadata{
			{Name: "apply", Path: "upgrade apply", Usage: upgradeApplyUsage, Flags: readOnlyCommandFlags(upgradeApplyFlags())},
			{
				Name:  "cli",
				Path:  "upgrade cli",
				Usage: upgradeCLIUsage,
				Flags: readOnlyCommandFlags(upgradeCLIFlags()),
				Subcommands: []CommandMetadata{
					{Name: "apply", Path: "upgrade cli apply", Usage: upgradeCLIApplyUsage, Flags: readOnlyCommandFlags(upgradeCLIApplyFlags())},
				},
			},
		},
	}
}

func upgradeFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--path", Value: textValueSpec()},
		{Name: "--target-version", Value: textValueSpec()},
	}
}

func upgradeApplyFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--path", Value: textValueSpec()},
		{Name: "--target-version", Value: textValueSpec(), Required: true},
	}
}

func upgradeCLIFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--target-version", Value: textValueSpec()},
	}
}

func upgradeCLIApplyFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--target-version", Value: textValueSpec()},
	}
}
