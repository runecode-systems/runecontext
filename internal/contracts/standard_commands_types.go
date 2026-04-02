package contracts

type StandardListOptions struct {
	ScopePaths []string
	Focus      string
	Statuses   []StandardStatus
}

type StandardListResult struct {
	ScopePaths []string
	Focus      string
	Statuses   []StandardStatus
	Standards  []StandardRecord
}

type StandardCreateOptions struct {
	Path                    string
	ID                      string
	Title                   string
	Status                  StandardStatus
	ReplacedBy              string
	Aliases                 []string
	SuggestedContextBundles []string
	Body                    string
}

type StandardUpdateOptions struct {
	Path                           string
	Title                          string
	Status                         string
	ReplacedBy                     string
	ClearReplacedBy                bool
	Aliases                        []string
	ReplaceAliases                 bool
	SuggestedContextBundles        []string
	ReplaceSuggestedContextBundles bool
}

type StandardMutationResult struct {
	Path         string
	ChangedFiles []FileMutation
}

type preparedCreateStandardOptions struct {
	path                    string
	id                      string
	title                   string
	status                  StandardStatus
	replacedBy              string
	aliases                 []string
	suggestedContextBundles []string
}

type standardFrontmatter struct {
	SchemaVersion           int            `yaml:"schema_version"`
	ID                      string         `yaml:"id"`
	Title                   string         `yaml:"title"`
	Status                  StandardStatus `yaml:"status,omitempty"`
	SuggestedContextBundles []string       `yaml:"suggested_context_bundles,omitempty"`
	ReplacedBy              string         `yaml:"replaced_by,omitempty"`
	Aliases                 []string       `yaml:"aliases,omitempty"`
}
