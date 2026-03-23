package contracts

import (
	"os"
	"path/filepath"
)

func buildProjectIndex(v *Validator, contentRoot string) (*ProjectIndex, error) {
	index := newProjectIndex()
	for _, loader := range []func(*Validator, string, *ProjectIndex) error{
		loadProjectChanges,
		loadProjectSpecs,
		loadProjectDecisions,
		loadProjectStandards,
	} {
		if err := loader(v, contentRoot, index); err != nil {
			return nil, err
		}
	}
	return index, nil
}

func newProjectIndex() *ProjectIndex {
	return &ProjectIndex{
		AssuranceReceipts: map[string]AssuranceReceiptRecord{},
		ChangeIDs:         map[string]struct{}{},
		Changes:           map[string]*ChangeRecord{},
		MarkdownFiles:     map[string]*MarkdownArtifact{},
		StandardPaths:     map[string]struct{}{},
		Standards:         map[string]*StandardRecord{},
		SpecPaths:         map[string]struct{}{},
		Specs:             map[string]*SpecRecord{},
		DecisionPaths:     map[string]struct{}{},
		Decisions:         map[string]*DecisionRecord{},
		StatusFiles:       map[string]StatusFileRecord{},
	}
}

func loadProjectChanges(v *Validator, contentRoot string, index *ProjectIndex) error {
	return walkChangeDirectories(filepath.Join(contentRoot, "changes"), func(changeDir string) error {
		return loadChangeDirectory(v, contentRoot, index, changeDir)
	})
}

func loadChangeDirectory(v *Validator, contentRoot string, index *ProjectIndex, changeDir string) error {
	record, err := loadChangeStatusRecord(v, index, changeDir)
	if err != nil {
		return err
	}
	if err := loadChangeProposal(v, contentRoot, index, changeDir); err != nil {
		return err
	}
	if err := loadChangeStandards(v, contentRoot, index, changeDir, record); err != nil {
		return err
	}
	return loadSupplementalChangeMarkdown(contentRoot, index, changeDir)
}

func loadChangeStatusRecord(v *Validator, index *ProjectIndex, changeDir string) (*ChangeRecord, error) {
	statusPath := filepath.Join(changeDir, "status.yaml")
	statusData, err := requireProjectFile(changeDir, statusPath)
	if err != nil {
		return nil, err
	}
	if err := v.ValidateYAMLFile("change-status.schema.json", statusPath, statusData); err != nil {
		return nil, err
	}
	parsed, err := parseYAML(statusData)
	if err != nil {
		return nil, err
	}
	obj, err := expectObject(statusPath, parsed, "status file")
	if err != nil {
		return nil, err
	}
	record, err := buildChangeRecord(changeDir, statusPath, obj)
	if err != nil {
		return nil, err
	}
	index.ChangeIDs[record.ID] = struct{}{}
	index.Changes[record.ID] = record
	index.StatusFiles[statusPath] = StatusFileRecord{Data: obj, Raw: append([]byte(nil), statusData...)}
	return record, nil
}

func loadChangeProposal(v *Validator, contentRoot string, index *ProjectIndex, changeDir string) error {
	proposalPath := filepath.Join(changeDir, "proposal.md")
	proposalData, err := requireProjectFile(changeDir, proposalPath)
	if err != nil {
		return err
	}
	if err := v.ValidateProposalMarkdown(proposalPath, proposalData); err != nil {
		return err
	}
	return indexMarkdownArtifact(index, contentRoot, proposalPath, proposalData, false)
}

func loadChangeStandards(v *Validator, contentRoot string, index *ProjectIndex, changeDir string, record *ChangeRecord) error {
	standardsPath := filepath.Join(changeDir, "standards.md")
	standardsData, err := requireProjectFile(changeDir, standardsPath)
	if err != nil {
		return err
	}
	if err := v.ValidateStandardsMarkdown(standardsPath, standardsData); err != nil {
		return err
	}
	standardsDoc, err := parseStandardsMarkdown(standardsPath, standardsData)
	if err != nil {
		return err
	}
	record.StandardRefs = append([]string(nil), standardsDoc.Refs...)
	record.ApplicableStandards = append([]string(nil), standardsDoc.RefsBySection["Applicable Standards"]...)
	record.AddedStandards = append([]string(nil), standardsDoc.RefsBySection["Standards Added Since Last Refresh"]...)
	record.ExcludedStandards = append([]string(nil), standardsDoc.RefsBySection["Standards Considered But Excluded"]...)
	return indexMarkdownArtifact(index, contentRoot, standardsPath, standardsData, false)
}

func loadSupplementalChangeMarkdown(contentRoot string, index *ProjectIndex, changeDir string) error {
	entries, err := os.ReadDir(changeDir)
	if err != nil {
		return &ValidationError{Path: changeDir, Message: err.Error()}
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" || entry.Name() == "proposal.md" || entry.Name() == "standards.md" {
			continue
		}
		path := filepath.Join(changeDir, entry.Name())
		data, err := readProjectFile(changeDir, path)
		if err != nil {
			return err
		}
		if err := indexMarkdownArtifact(index, contentRoot, path, data, false); err != nil {
			return err
		}
	}
	return nil
}

func loadProjectSpecs(v *Validator, contentRoot string, index *ProjectIndex) error {
	return walkProjectFiles(filepath.Join(contentRoot, "specs"), func(path string) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}
		data, err := readProjectFile(filepath.Join(contentRoot, "specs"), path)
		if err != nil {
			return err
		}
		doc, err := v.ParseSpec(path, data)
		if err != nil {
			return err
		}
		record, err := buildSpecRecord(path, doc)
		if err != nil {
			return err
		}
		if err := validateArtifactChangeRefs(path, doc.Frontmatter, index.ChangeIDs, []string{"originating_changes", "revised_by_changes"}); err != nil {
			return err
		}
		return registerSpecRecord(index, contentRoot, path, data, record)
	})
}

func registerSpecRecord(index *ProjectIndex, contentRoot, path string, data []byte, record *SpecRecord) error {
	rel, err := filepath.Rel(contentRoot, path)
	if err != nil {
		return err
	}
	record.Path = filepath.ToSlash(rel)
	index.SpecPaths[record.Path] = struct{}{}
	index.Specs[record.Path] = record
	return indexMarkdownArtifact(index, contentRoot, path, data, true)
}

func loadProjectDecisions(v *Validator, contentRoot string, index *ProjectIndex) error {
	return walkProjectFiles(filepath.Join(contentRoot, "decisions"), func(path string) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}
		data, err := readProjectFile(filepath.Join(contentRoot, "decisions"), path)
		if err != nil {
			return err
		}
		doc, err := v.ParseDecision(path, data)
		if err != nil {
			return err
		}
		record, err := buildDecisionRecord(path, doc)
		if err != nil {
			return err
		}
		if err := validateArtifactChangeRefs(path, doc.Frontmatter, index.ChangeIDs, []string{"originating_changes", "related_changes"}); err != nil {
			return err
		}
		return registerDecisionRecord(index, contentRoot, path, data, record)
	})
}

func registerDecisionRecord(index *ProjectIndex, contentRoot, path string, data []byte, record *DecisionRecord) error {
	rel, err := filepath.Rel(contentRoot, path)
	if err != nil {
		return err
	}
	record.Path = filepath.ToSlash(rel)
	index.DecisionPaths[record.Path] = struct{}{}
	index.Decisions[record.Path] = record
	return indexMarkdownArtifact(index, contentRoot, path, data, true)
}

func loadProjectStandards(v *Validator, contentRoot string, index *ProjectIndex) error {
	return walkProjectFiles(filepath.Join(contentRoot, "standards"), func(path string) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}
		data, err := readProjectFile(filepath.Join(contentRoot, "standards"), path)
		if err != nil {
			return err
		}
		doc, err := v.ParseStandard(path, data)
		if err != nil {
			return err
		}
		record, err := buildStandardRecord(path, doc)
		if err != nil {
			return err
		}
		return registerStandardRecord(index, contentRoot, path, data, record)
	})
}

func registerStandardRecord(index *ProjectIndex, contentRoot, path string, data []byte, record *StandardRecord) error {
	rel, err := filepath.Rel(contentRoot, path)
	if err != nil {
		return err
	}
	record.Path = filepath.ToSlash(rel)
	index.StandardPaths[record.Path] = struct{}{}
	index.Standards[record.Path] = record
	return indexMarkdownArtifact(index, contentRoot, path, data, true)
}

func validateArtifactChangeRefs(path string, frontmatter map[string]any, known map[string]struct{}, keys []string) error {
	for _, key := range keys {
		if err := validateChangeIDRefs(path, key, frontmatter[key], known); err != nil {
			return err
		}
	}
	return nil
}

func requireProjectFile(boundaryPath, path string) ([]byte, error) {
	data, err := readProjectFile(boundaryPath, path)
	if err == nil {
		return data, nil
	}
	if os.IsNotExist(err) {
		return nil, &ValidationError{Path: path, Message: "missing required file"}
	}
	return nil, err
}
