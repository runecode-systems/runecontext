package cli

func versionRootCommandsMetadata() []CommandMetadata {
	flags := versionCommandFlags()
	return []CommandMetadata{
		{Name: "version", Path: "version", Usage: versionUsage, Flags: flags},
		{Name: "--version", Path: "--version", Usage: versionUsage, Flags: flags},
		{Name: "-v", Path: "-v", Usage: versionUsage, Flags: flags},
	}
}

func versionCommandFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--json", Value: noValueSpec()},
		{Name: "--non-interactive", Value: noValueSpec()},
	}
}
