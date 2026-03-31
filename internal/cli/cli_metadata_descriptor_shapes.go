package cli

type descriptorLayout struct {
	Profile      string `json:"profile" yaml:"profile"`
	SchemaPath   string `json:"schema_path" yaml:"schema_path"`
	AdaptersPath string `json:"adapters_path" yaml:"adapters_path"`
}

type descriptorProject struct {
	Profile       string `json:"profile" yaml:"profile"`
	RootConfig    string `json:"root_config" yaml:"root_config"`
	ContentRoot   string `json:"content_root" yaml:"content_root"`
	AssurancePath string `json:"assurance_path" yaml:"assurance_path"`
	ManifestPath  string `json:"manifest_path" yaml:"manifest_path"`
	IndexesRoot   string `json:"indexes_root" yaml:"indexes_root"`
}

type descriptorAssurance struct {
	Tiers             []string `json:"tiers" yaml:"tiers"`
	BaselineSupported bool     `json:"baseline_supported" yaml:"baseline_supported"`
	ReceiptFamilies   []string `json:"receipt_families" yaml:"receipt_families"`
}

type descriptorCanonicalization struct {
	ContextPack        descriptorCanonicalizationProfile `json:"context_pack" yaml:"context_pack"`
	AssuranceArtifacts descriptorCanonicalizationProfile `json:"assurance_artifacts" yaml:"assurance_artifacts"`
}

type descriptorCanonicalizationProfile struct {
	Profile       string `json:"profile" yaml:"profile"`
	HashAlgorithm string `json:"hash_algorithm" yaml:"hash_algorithm"`
}
