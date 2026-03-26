package cli

func upgradeCommandMetadata() CommandMetadata {
	return CommandMetadata{
		Name:  "upgrade",
		Path:  "upgrade",
		Usage: upgradeUsage,
		Flags: readOnlyCommandFlags(upgradeFlags()),
		Subcommands: []CommandMetadata{
			{Name: "apply", Path: "upgrade apply", Usage: upgradeApplyUsage, Flags: readOnlyCommandFlags(upgradeApplyFlags())},
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
