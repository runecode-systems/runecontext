package cli

func toolFlowMappings(tool string) []hostNativeFlow {
	return []hostNativeFlow{
		flowForTool(tool, "change-new", "change new", "Create a new RuneContext change"),
		flowForTool(tool, "change-assess-intake", "change assess-intake", "Assess intake readiness and advisory shaping signals"),
		flowForTool(tool, "change-assess-decomposition", "change assess-decomposition", "Assess decomposition and umbrella/sub-change signals"),
		flowForTool(tool, "change-decomposition-plan", "change decomposition-plan", "Plan umbrella and sub-change graph rewrites"),
		flowForTool(tool, "change-decomposition-apply", "change decomposition-apply", "Apply umbrella and sub-change graph rewrites"),
		flowForTool(tool, "change-shape", "change shape", "Shape an existing RuneContext change"),
		flowForTool(tool, "standard-discover", "standard discover", "Discover standards candidates for promotion"),
		flowForTool(tool, "promote", "promote", "Advance RuneContext promotion state"),
	}
}

func flowForTool(tool, id, name, description string) hostNativeFlow {
	return hostNativeFlow{
		id:          id,
		name:        name,
		description: description,
		source:      "adapters/" + tool + "/flows/" + id + ".md",
	}
}
