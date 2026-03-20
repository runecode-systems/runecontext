package contracts

type GeneratedChangesByStatusIndex struct {
	SchemaVersion int                         `json:"schema_version" yaml:"schema_version"`
	Statuses      GeneratedChangeStatusGroups `json:"statuses" yaml:"statuses"`
}

type GeneratedChangeStatusGroups struct {
	Proposed    []GeneratedChangeStatusEntry `json:"proposed" yaml:"proposed"`
	Planned     []GeneratedChangeStatusEntry `json:"planned" yaml:"planned"`
	Implemented []GeneratedChangeStatusEntry `json:"implemented" yaml:"implemented"`
	Verified    []GeneratedChangeStatusEntry `json:"verified" yaml:"verified"`
	Closed      []GeneratedChangeStatusEntry `json:"closed" yaml:"closed"`
	Superseded  []GeneratedChangeStatusEntry `json:"superseded" yaml:"superseded"`
}

type GeneratedChangeStatusEntry struct {
	ID    string `json:"id" yaml:"id"`
	Title string `json:"title" yaml:"title"`
	Type  string `json:"type" yaml:"type"`
	Size  string `json:"size,omitempty" yaml:"size,omitempty"`
	Path  string `json:"path" yaml:"path"`
}
