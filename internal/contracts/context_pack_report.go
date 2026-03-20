package contracts

import (
	"errors"
	"fmt"
	"os"
)

const defaultContextPackBuildAttempts = 2

const contextPackReportSchemaVersion = 1

type ContextPackReportOptions struct {
	ContextPackOptions
	Explain            bool
	AdvisoryThresholds ContextPackAdvisoryThresholds
}

type ContextPackAdvisoryThresholds struct {
	SelectedFiles          int   `json:"selected_files" yaml:"selected_files"`
	ReferencedContentBytes int64 `json:"referenced_content_bytes" yaml:"referenced_content_bytes"`
	ProvenanceBytes        int64 `json:"provenance_bytes" yaml:"provenance_bytes"`
}

var defaultContextPackAdvisoryThresholds = ContextPackAdvisoryThresholds{
	SelectedFiles:          256,
	ReferencedContentBytes: 1 << 20,
	ProvenanceBytes:        256 << 10,
}

func DefaultContextPackAdvisoryThresholds() ContextPackAdvisoryThresholds {
	return defaultContextPackAdvisoryThresholds
}

type ContextPackReport struct {
	ReportSchemaVersion int                       `json:"report_schema_version"`
	Pack                *ContextPack              `json:"pack"`
	Summary             ContextPackSummary        `json:"summary"`
	Warnings            []ContextPackAdvisory     `json:"warnings"`
	Explain             *ContextPackExplainReport `json:"explain,omitempty"`
}

type contextPackBuildArtifacts struct {
	inputs          contextPackInputs
	selectedDigests []contextPackFileDigest
}

type contextPackFileDigest struct {
	Path            string
	SHA256          string
	ReferencedBytes int64
}

func (p *ProjectIndex) BuildContextPackReport(options ContextPackReportOptions) (*ContextPackReport, error) {
	thresholds := normalizeContextPackAdvisoryThresholds(options.AdvisoryThresholds)
	pack, artifacts, err := p.buildStableContextPack(options.ContextPackOptions)
	if err != nil {
		return nil, err
	}
	return buildContextPackReport(pack, artifacts.selectedDigests, options.Explain, thresholds)
}

func (p *ProjectIndex) buildStableContextPack(options ContextPackOptions) (*ContextPack, contextPackBuildArtifacts, error) {
	for attempt := 0; attempt < defaultContextPackBuildAttempts; attempt++ {
		pack, artifacts, stable, err := p.buildContextPackAttempt(options)
		if err != nil {
			return nil, contextPackBuildArtifacts{}, err
		}
		if stable {
			return pack, artifacts, nil
		}
	}
	return nil, contextPackBuildArtifacts{}, fmt.Errorf("context pack inputs changed during build; rerun after files stop changing")
}

func (p *ProjectIndex) buildContextPackAttempt(options ContextPackOptions) (*ContextPack, contextPackBuildArtifacts, bool, error) {
	artifacts, err := p.buildContextPackArtifacts(options)
	if err != nil {
		return nil, contextPackBuildArtifacts{}, false, err
	}
	pack := newContextPack(artifacts.inputs)
	if err := validateContextPackIdentity(pack); err != nil {
		return nil, contextPackBuildArtifacts{}, false, err
	}
	packHash, err := pack.computePackHash()
	if err != nil {
		return nil, contextPackBuildArtifacts{}, false, err
	}
	pack.PackHash = packHash
	stable, err := p.contextPackBuildStable(artifacts, pack)
	if err != nil {
		return nil, contextPackBuildArtifacts{}, false, err
	}
	return pack, artifacts, stable, nil
}

func (p *ProjectIndex) buildContextPackArtifacts(options ContextPackOptions) (contextPackBuildArtifacts, error) {
	if err := validateContextPackProjectIndex(p); err != nil {
		return contextPackBuildArtifacts{}, err
	}
	requested, err := normalizeContextPackBundleIDs(options.BundleIDs)
	if err != nil {
		return contextPackBuildArtifacts{}, err
	}
	resolution, err := p.Bundles.ResolveRequest(requested)
	if err != nil {
		return contextPackBuildArtifacts{}, err
	}
	resolved, err := buildContextPackResolvedFrom(p.Resolution, resolution.Linearization)
	if err != nil {
		return contextPackBuildArtifacts{}, err
	}
	generatedAt, err := formatContextPackGeneratedAt(options.GeneratedAt)
	if err != nil {
		return contextPackBuildArtifacts{}, err
	}
	selected, excluded, selectedDigests, err := buildContextPackInventories(p.ContentRoot, resolution)
	if err != nil {
		return contextPackBuildArtifacts{}, err
	}
	return contextPackBuildArtifacts{
		inputs: contextPackInputs{
			requested:   requested,
			resolved:    resolved,
			selected:    selected,
			excluded:    excluded,
			generatedAt: generatedAt,
		},
		selectedDigests: selectedDigests,
	}, nil
}

func normalizeContextPackAdvisoryThresholds(thresholds ContextPackAdvisoryThresholds) ContextPackAdvisoryThresholds {
	// A fully zero-valued thresholds struct means "use defaults". Once any field
	// is set explicitly, zero is treated as an intentional threshold value and
	// negative values opt back into the default for that field.
	if thresholds == (ContextPackAdvisoryThresholds{}) {
		return DefaultContextPackAdvisoryThresholds()
	}
	if thresholds.SelectedFiles < 0 {
		thresholds.SelectedFiles = DefaultContextPackAdvisoryThresholds().SelectedFiles
	}
	if thresholds.ReferencedContentBytes < 0 {
		thresholds.ReferencedContentBytes = DefaultContextPackAdvisoryThresholds().ReferencedContentBytes
	}
	if thresholds.ProvenanceBytes < 0 {
		thresholds.ProvenanceBytes = DefaultContextPackAdvisoryThresholds().ProvenanceBytes
	}
	return thresholds
}

func (p *ProjectIndex) contextPackBuildStable(artifacts contextPackBuildArtifacts, pack *ContextPack) (bool, error) {
	resolution, err := p.Bundles.ResolveRequest(artifacts.inputs.requested)
	if err != nil {
		return false, err
	}
	equal, err := contextPackExplainReportsEqual(
		contextPackExplainReportFromResolution(artifacts.inputs.requested, resolution),
		contextPackExplainReportFromPack(pack),
	)
	if err != nil {
		return false, err
	}
	if !equal {
		return false, nil
	}
	return p.contextPackFileDigestsStable(artifacts.selectedDigests)
}

func (p *ProjectIndex) contextPackFileDigestsStable(digests []contextPackFileDigest) (bool, error) {
	for _, digest := range digests {
		current, err := digestContextPackFile(p.ContentRoot, digest.Path)
		if err != nil {
			if isContextPackRetriableDigestError(err) {
				return false, nil
			}
			return false, err
		}
		if current.SHA256 != digest.SHA256 || current.ReferencedBytes != digest.ReferencedBytes {
			return false, nil
		}
	}
	return true, nil
}

func isContextPackRetriableDigestError(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
