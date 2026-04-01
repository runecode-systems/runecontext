package cli

import "github.com/runecode-systems/runecontext/internal/contracts"

func descriptorDistributionLayouts() []descriptorLayout {
	return []descriptorLayout{
		{Profile: "repo_bundle", SchemaPath: "schemas", AdaptersPath: "build/generated/adapters"},
		{Profile: "installed_share_layout", SchemaPath: "share/runecontext/schemas", AdaptersPath: "share/runecontext/adapters"},
	}
}

func descriptorProjectProfiles() []descriptorProject {
	return []descriptorProject{
		{
			Profile:       "portable_project_root",
			RootConfig:    "runecontext.yaml",
			ContentRoot:   "runecontext",
			AssurancePath: "runecontext/assurance",
			ManifestPath:  "runecontext/manifest.yaml",
			IndexesRoot:   "runecontext/indexes",
		},
	}
}

func descriptorFeatures() []string {
	return []string{
		"signed_tag_verification",
		"mutable_ref_opt_in",
		"monorepo_nearest_root_discovery",
		"context_pack_capture",
		"verified_assurance",
		"completion_registry",
		"completion_dynamic_suggestions",
		"generated_indexes",
		"promotion_workflow",
		"upgrade_planning",
		"staged_upgrade_execution",
		"mixed_or_stale_tree_detection",
	}
}

func buildDescriptorAssurance() descriptorAssurance {
	return descriptorAssurance{
		Tiers:             []string{assuranceTierPlain, contracts.AssuranceTierVerified},
		BaselineSupported: true,
		ReceiptFamilies:   contracts.AssuranceReceiptFamiliesForCLI(),
	}
}

func buildDescriptorCanonicalization() descriptorCanonicalization {
	return descriptorCanonicalization{
		ContextPack: descriptorCanonicalizationProfile{
			Profile:       contracts.ContextPackCanonicalizationTokenForCLI(),
			HashAlgorithm: contracts.ContextPackHashAlgorithmForCLI(),
		},
		AssuranceArtifacts: descriptorCanonicalizationProfile{
			Profile:       contracts.AssuranceCanonicalizationToken,
			HashAlgorithm: contracts.AssuranceHashAlgorithmTokenForCLI(),
		},
	}
}
