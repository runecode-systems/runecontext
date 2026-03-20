package contracts

import "time"

type LifecycleStatus string

const (
	StatusProposed    LifecycleStatus = "proposed"
	StatusPlanned     LifecycleStatus = "planned"
	StatusImplemented LifecycleStatus = "implemented"
	StatusVerified    LifecycleStatus = "verified"
	StatusClosed      LifecycleStatus = "closed"
	StatusSuperseded  LifecycleStatus = "superseded"
)

var lifecycleOrder = map[LifecycleStatus]int{
	StatusProposed:    0,
	StatusPlanned:     1,
	StatusImplemented: 2,
	StatusVerified:    3,
	StatusClosed:      4,
	StatusSuperseded:  4,
}

type ChangeRecord struct {
	ID                  string
	DirPath             string
	StatusPath          string
	Title               string
	Status              LifecycleStatus
	Type                string
	Size                string
	VerificationStatus  string
	ClosedAt            string
	HasClosedAt         bool
	ContextBundles      []string
	StandardRefs        []string
	ApplicableStandards []string
	AddedStandards      []string
	ExcludedStandards   []string
	RelatedSpecs        []string
	RelatedDecisions    []string
	RelatedChanges      []string
	DependsOn           []string
	InformedBy          []string
	Supersedes          []string
	SupersededBy        []string
	Data                map[string]any
}

type SpecRecord struct {
	Path               string
	OriginatingChanges []string
	RevisedByChanges   []string
}

type DecisionRecord struct {
	Path               string
	OriginatingChanges []string
	RelatedChanges     []string
}

type CloseChangeOptions struct {
	ClosedAt     time.Time
	SupersededBy []string
}

type SplitChangePlan struct {
	UmbrellaID string
	SubChanges []SplitSubChange
}

type SplitSubChange struct {
	ID        string
	DependsOn []string
}

type ChangeGraphLinks struct {
	RelatedChanges []string
	DependsOn      []string
}

func (p *ProjectIndex) OpenChangeIDs() []string {
	return p.changeIDsMatching(func(record *ChangeRecord) bool {
		return !isTerminalLifecycleStatus(record.Status)
	})
}

func (p *ProjectIndex) ClosedChangeIDs() []string {
	return p.changeIDsMatching(func(record *ChangeRecord) bool {
		return record.Status == StatusClosed
	})
}

func (p *ProjectIndex) SupersededChangeIDs() []string {
	return p.changeIDsMatching(func(record *ChangeRecord) bool {
		return record.Status == StatusSuperseded
	})
}

func (p *ProjectIndex) changeIDsMatching(keep func(*ChangeRecord) bool) []string {
	if p == nil {
		return nil
	}
	ids := make([]string, 0)
	for _, id := range SortedKeys(p.Changes) {
		if keep(p.Changes[id]) {
			ids = append(ids, id)
		}
	}
	return ids
}

func isTerminalLifecycleStatus(status LifecycleStatus) bool {
	return status == StatusClosed || status == StatusSuperseded
}
