package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func renderProposalMarkdown(title, description string, mode ChangeMode, reasons, assumptions []string) []byte {
	summary := strings.TrimSpace(title)
	problem := "The repository needs a reviewable RuneContext change record for this work."
	proposed := fmt.Sprintf("Track %s through the %s RuneContext change artifacts.", strings.TrimSpace(title), mode)
	if trimmed := strings.TrimSpace(description); trimmed != "" {
		problem = trimmed
		proposed = fmt.Sprintf("Track and deliver %s while keeping the intent and standards linkage reviewable.", strings.TrimSpace(title))
	}
	return []byte(strings.Join(proposalSections(summary, problem, proposed, assumptions), "\n"))
}

func proposalSections(summary, problem, proposed string, assumptions []string) []string {
	assumptionsBody := "N/A"
	if len(assumptions) > 0 {
		assumptionsBody = renderBulletList(assumptions)
	}
	return []string{
		"## Summary",
		summary,
		"",
		"## Problem",
		problem,
		"",
		"## Proposed Change",
		proposed,
		"",
		"## Why Now",
		"The work needs stable intent, standards linkage, and verification planning before it moves further.",
		"",
		"## Assumptions",
		assumptionsBody,
		"",
		"## Out of Scope",
		"Work outside the scoped change tracked here.",
		"",
		"## Impact",
		"The change keeps intent, assumptions, and standards linkage reviewable.",
		"",
	}
}

func renderStandardsMarkdown(existing []byte, applicable, added []string, preserved []markdownSection, creating bool) []byte {
	sections := []string{
		"## Applicable Standards",
		renderStandardBullets(applicable, "Selected from the current context bundles."),
	}
	sections = append(sections, standardsRefreshSections(added)...)
	sections = append(sections, preservedStandardsContent(preserved)...)
	if creating && !hasResolutionNotesSection(preserved) {
		sections = append(sections, "", "## Resolution Notes", "Generated from the current context bundle selection; review any automatic refresh before committing.")
	}
	return []byte(strings.Join(sections, "\n") + "\n")
}

func standardsRefreshSections(added []string) []string {
	if len(added) == 0 {
		return nil
	}
	return []string{"", "## Standards Added Since Last Refresh", renderStandardBullets(added, "Newly selected during standards refresh.")}
}

func preservedStandardsContent(preserved []markdownSection) []string {
	sections := make([]string, 0)
	for _, section := range preserved {
		if skipPreservedStandardsSection(section.Heading) {
			continue
		}
		sections = append(sections, "", "## "+section.Heading, section.Body)
	}
	return sections
}

func hasResolutionNotesSection(preserved []markdownSection) bool {
	for _, section := range preserved {
		if section.Heading == "Resolution Notes" {
			return true
		}
	}
	return false
}

func skipPreservedStandardsSection(heading string) bool {
	return heading == "Applicable Standards" || heading == "Standards Added Since Last Refresh"
}

func preservedStandardsSections(data []byte) ([]markdownSection, error) {
	sections, err := parseLevel2Sections("standards.md", data)
	if err != nil {
		return nil, err
	}
	preserved := make([]markdownSection, 0, len(sections))
	for _, section := range sections {
		if skipPreservedStandardsSection(section.Heading) {
			continue
		}
		preserved = append(preserved, section)
	}
	return preserved, nil
}

func renderStandardBullets(paths []string, description string) string {
	if len(paths) == 0 {
		return "- `standards/placeholder.md`: Replace this placeholder once the project defines a selectable standard."
	}
	lines := make([]string, 0, len(paths))
	for _, path := range paths {
		lines = append(lines, fmt.Sprintf("- `%s`: %s", path, description))
	}
	return strings.Join(lines, "\n")
}

func materializeShapeFiles(changeDir, writableRoot, projectRoot, title string, assessment changeIntakeAssessment, options ChangeShapeOptions) ([]FileMutation, error) {
	verificationCommands := assessment.VerificationCmds
	if len(verificationCommands) == 0 {
		verificationCommands, _ = inferVerificationCommands(projectRoot)
	}
	files := shapeFilesToCreate(title, assessment, options, verificationCommands)
	changed := make([]FileMutation, 0, len(files))
	for _, file := range files {
		if !file.ok {
			continue
		}
		created, err := createShapeFile(changeDir, writableRoot, file)
		if err != nil {
			return nil, err
		}
		if created.Path != "" {
			changed = append(changed, created)
		}
	}
	return changed, nil
}

type shapeFileSpec struct {
	name string
	data []byte
	ok   bool
}

func shapeFilesToCreate(title string, assessment changeIntakeAssessment, options ChangeShapeOptions, verificationCommands []string) []shapeFileSpec {
	return []shapeFileSpec{
		{name: "design.md", data: renderDesignMarkdown(title, assessment, options.Design), ok: true},
		{name: "verification.md", data: renderVerificationMarkdown(verificationCommands, assessment.VerificationNote, options.Verification), ok: true},
		{name: "tasks.md", data: renderSupplementalMarkdown("Tasks", options.Tasks), ok: len(options.Tasks) > 0},
		{name: "references.md", data: renderSupplementalMarkdown("References", options.References), ok: len(options.References) > 0},
	}
}

func createShapeFile(changeDir, writableRoot string, file shapeFileSpec) (FileMutation, error) {
	path := filepath.Join(changeDir, file.name)
	if exists, err := fileAlreadyExists(path); err != nil || exists {
		return FileMutation{}, err
	}
	if err := writeFileAtomically(path, file.data, 0o644); err != nil {
		return FileMutation{}, err
	}
	return FileMutation{Path: runeContextRelativePath(writableRoot, path), Action: "created"}, nil
}

func fileAlreadyExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func renderDesignMarkdown(title string, assessment changeIntakeAssessment, design string) []byte {
	overview := strings.TrimSpace(design)
	if overview == "" {
		overview = fmt.Sprintf("Shape %s before implementation so scope, standards linkage, and verification stay reviewable.", strings.TrimSpace(title))
	}
	lines := []string{"# Design", "", "## Overview", overview}
	lines = appendDesignSections(lines, assessment)
	return []byte(strings.Join(append(lines, ""), "\n"))
}

func appendDesignSections(lines []string, assessment changeIntakeAssessment) []string {
	if len(assessment.Reasons) > 0 {
		lines = append(lines, "", "## Shape Rationale", renderBulletList(assessment.Reasons))
	}
	if len(assessment.ChecklistItems) > 0 {
		lines = append(lines, "", "## "+assessment.ChecklistTitle, renderBulletList(assessment.ChecklistItems))
	}
	if len(assessment.FollowUpPrompts) > 0 {
		lines = append(lines, "", "## Ask More When", renderBulletList(assessment.FollowUpPrompts))
	}
	return lines
}

func renderVerificationMarkdown(commands []string, note, verification string) []byte {
	lines := []string{"# Verification", ""}
	if strings.TrimSpace(verification) != "" {
		lines = append(lines, "## Notes", strings.TrimSpace(verification), "")
	}
	lines = append(lines, "## Planned Checks", renderBulletList(plannedVerificationItems(commands)))
	if strings.TrimSpace(note) != "" {
		lines = append(lines, "", "## Close Gate", note)
	}
	return []byte(strings.Join(append(lines, ""), "\n"))
}

func plannedVerificationItems(commands []string) []string {
	if len(commands) == 0 {
		return []string{"Define the repository verification commands before closing this change."}
	}
	planned := make([]string, 0, len(commands))
	for _, command := range commands {
		planned = append(planned, fmt.Sprintf("`%s`", command))
	}
	return planned
}

func renderSupplementalMarkdown(title string, items []string) []byte {
	lines := []string{"# " + title, "", renderBulletList(items), ""}
	return []byte(strings.Join(lines, "\n"))
}

func renderBulletList(items []string) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, "- "+strings.TrimSpace(item))
	}
	return strings.Join(lines, "\n")
}
