package contracts

type GeneratedManifest struct {
	SchemaVersion int                      `json:"schema_version" yaml:"schema_version"`
	Indexes       GeneratedManifestIndexes `json:"indexes" yaml:"indexes"`
	Counts        GeneratedManifestCounts  `json:"counts" yaml:"counts"`
	Standards     []string                 `json:"standards" yaml:"standards"`
	Bundles       []string                 `json:"bundles" yaml:"bundles"`
	Changes       []string                 `json:"changes" yaml:"changes"`
	Specs         []string                 `json:"specs" yaml:"specs"`
	Decisions     []string                 `json:"decisions" yaml:"decisions"`
}

type GeneratedManifestIndexes struct {
	ChangesByStatus string `json:"changes_by_status" yaml:"changes_by_status"`
	Bundles         string `json:"bundles" yaml:"bundles"`
}

type GeneratedManifestCounts struct {
	Standards int `json:"standards" yaml:"standards"`
	Bundles   int `json:"bundles" yaml:"bundles"`
	Changes   int `json:"changes" yaml:"changes"`
	Specs     int `json:"specs" yaml:"specs"`
	Decisions int `json:"decisions" yaml:"decisions"`
}
