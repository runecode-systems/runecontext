package contracts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateProjectRejectsInvalidTerminalMetadata(t *testing.T) {
	for _, tc := range terminalMetadataScenarios() {
		t.Run(tc.name, func(t *testing.T) { tc.run(t) })
	}
}

type workflowTestScenario struct {
	name string
	run  func(*testing.T)
}

func terminalMetadataScenarios() []workflowTestScenario {
	return []workflowTestScenario{
		{name: "terminal requires closed_at", run: func(t *testing.T) {
			assertTerminalMetadataFailure(t, func(text string) string {
				text = strings.Replace(text, "status: proposed", "status: closed", 1)
				return strings.Replace(text, "verification_status: pending", "verification_status: passed", 1)
			}, "requires closed_at")
		}},
		{name: "non-terminal must not set closed_at", run: func(t *testing.T) {
			assertTerminalMetadataFailure(t, func(text string) string {
				return strings.Replace(text, "closed_at: null", "closed_at: \"2026-03-18\"", 1)
			}, "must not set closed_at")
		}},
		{name: "closed must not keep pending verification", run: func(t *testing.T) {
			assertTerminalMetadataFailure(t, func(text string) string {
				text = strings.Replace(text, "status: proposed", "status: closed", 1)
				return strings.Replace(text, "closed_at: null", "closed_at: \"2026-03-18\"", 1)
			}, "must not leave verification_status pending")
		}},
		{name: "superseded must not keep pending verification", run: func(t *testing.T) {
			assertTerminalMetadataFailure(t, func(text string) string {
				text = strings.Replace(text, "status: proposed", "status: superseded", 1)
				text = strings.Replace(text, "closed_at: null", "closed_at: \"2026-03-18\"", 1)
				return strings.Replace(text, "superseded_by: []", "superseded_by:\n  - CHG-2026-002-b4c3-auth-revision", 1)
			}, "superseded changes must not leave verification_status pending")
		}},
	}
}

func assertTerminalMetadataFailure(t *testing.T, rewrite func(string) string, want string) {
	t.Helper()
	root := copyTraceabilityFixtureProject(t, "valid-project")
	statusPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml")
	rewriteFile(t, statusPath, rewrite)
	v := NewValidator(schemaRoot(t))
	_, err := v.ValidateProject(root)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("expected %q failure, got %v", want, err)
	}
}

func TestValidateProjectStandardFrontmatterAndMigrationSemantics(t *testing.T) {
	for _, tc := range standardValidationScenarios() {
		t.Run(tc.name, func(t *testing.T) { tc.run(t) })
	}
}

func standardValidationScenarios() []workflowTestScenario {
	scenarios := standardMetadataScenarios()
	scenarios = append(scenarios, standardProposalScenarios()...)
	scenarios = append(scenarios, standardSpecScenarios()...)
	return scenarios
}

func standardMetadataScenarios() []workflowTestScenario {
	return []workflowTestScenario{
		{name: "valid standard metadata", run: func(t *testing.T) { assertStandardValidationPass(t, nil) }},
		{name: "reject standard id mismatch", run: func(t *testing.T) {
			assertStandardValidationFailure(t, func(root string) {
				rewriteDeterministicStandard(t, root, "id: global/deterministic-check-write", "id: global/not-the-path")
			}, "must match path-relative stem")
		}},
		{name: "reject draft standard in standards md", run: func(t *testing.T) {
			assertStandardValidationFailure(t, func(root string) { rewriteDeterministicStandard(t, root, "status: active", "status: draft") }, "section \"Applicable Standards\"")
		}},
		{name: "reject draft standard in added section with section-specific message", run: assertAddedSectionDraftFailure},
		{name: "allow deprecated standard in applicable standards with warning", run: assertDeprecatedStandardWarning},
		{name: "allow excluded draft standard path references", run: assertExcludedDraftStandardAllowed},
		{name: "reject missing replaced_by target", run: func(t *testing.T) {
			assertStandardValidationFailure(t, func(root string) {
				rewriteDeterministicStandard(t, root, "status: active", "status: deprecated\nreplaced_by: standards/global/missing.md")
			}, "references missing standard")
		}},
		{name: "reject self replaced_by target", run: func(t *testing.T) {
			assertStandardValidationFailure(t, func(root string) {
				rewriteDeterministicStandard(t, root, "status: active", "status: deprecated\nreplaced_by: standards/global/deterministic-check-write.md")
			}, "must not reference the standard itself")
		}},
		{name: "reject alias collisions", run: assertAliasCollisionFailure},
		{name: "reject copied standard body text in standards md", run: func(t *testing.T) {
			assertStandardsDocumentValidationFailure(t, "## Applicable Standards\nTrust Boundary Interfaces\n\n## Resolution Notes\nThis copied body text is invalid.\n", "must list standards as")
		}},
	}
}

func standardProposalScenarios() []workflowTestScenario {
	return []workflowTestScenario{
		{name: "reject copied standard body text in proposal", run: func(t *testing.T) {
			assertProposalDocumentValidationFailure(t, "\n\nGenerated and reviewed artifacts must remain deterministic and easy to audit.\n", "appears to copy standard content")
		}},
		{name: "allow standard text inside fenced code in proposal", run: func(t *testing.T) {
			assertProposalDocumentPasses(t, "\n\n```md\nGenerated and reviewed artifacts must remain deterministic and easy to audit.\n```\n")
		}},
		{name: "validate plain standard path reference in proposal", run: func(t *testing.T) {
			assertProposalDocumentPasses(t, "\n\nSee `standards/global/deterministic-check-write.md` for the durable rule.\n")
		}},
		{name: "reject missing plain standard path reference in proposal", run: assertMissingProposalStandardPath},
	}
}

func standardSpecScenarios() []workflowTestScenario {
	return []workflowTestScenario{
		{name: "reject missing standard deep ref in spec body", run: func(t *testing.T) {
			assertSpecDocumentValidationFailure(t, "\n\nSee standards/global/missing.md#missing for the obsolete rule.\n", "points to missing standard")
		}},
		{name: "reject copied standard body text in spec", run: func(t *testing.T) {
			assertSpecDocumentValidationFailure(t, "\n\nGenerated and reviewed artifacts must remain deterministic and easy to audit.\n", "appears to copy standard content")
		}},
		{name: "allow standard text inside blockquote fenced code in spec", run: func(t *testing.T) {
			assertSpecDocumentPasses(t, "\n\n> ```md\n> Generated and reviewed artifacts must remain deterministic and easy to audit.\n> ```\n")
		}},
		{name: "validate plain standard path reference in spec", run: func(t *testing.T) {
			assertSpecDocumentPasses(t, "\n\nSee `standards/global/deterministic-check-write.md` for review guidance.\n")
		}},
	}
}

func assertStandardValidationPass(t *testing.T, mutate func(root string)) {
	t.Helper()
	root := copyTraceabilityFixtureProject(t, "valid-project")
	if mutate != nil {
		mutate(root)
	}
	v := NewValidator(schemaRoot(t))
	if _, err := v.ValidateProject(root); err != nil {
		t.Fatalf("expected validation to pass, got %v", err)
	}
}

func assertStandardValidationFailure(t *testing.T, mutate func(root string), want string) {
	t.Helper()
	root := copyTraceabilityFixtureProject(t, "valid-project")
	mutate(root)
	v := NewValidator(schemaRoot(t))
	_, err := v.ValidateProject(root)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("expected %q failure, got %v", want, err)
	}
}

func rewriteDeterministicStandard(t *testing.T, root, old, new string) {
	t.Helper()
	standardPath := filepath.Join(root, "runecontext", "standards", "global", "deterministic-check-write.md")
	rewriteFile(t, standardPath, func(text string) string { return strings.Replace(text, old, new, 1) })
}

func assertAddedSectionDraftFailure(t *testing.T) {
	t.Helper()
	assertStandardValidationFailure(t, func(root string) {
		rewriteDeterministicStandard(t, root, "status: active", "status: draft")
		standardsPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "standards.md")
		rewriteFile(t, standardsPath, func(string) string {
			return "## Applicable Standards\n- `standards/global/other-active.md`: Current active selection.\n\n## Standards Added Since Last Refresh\n- `standards/global/deterministic-check-write.md`: Newly added but still draft.\n"
		})
		writeActiveStandardFixture(t, root)
		rewriteOtherChangeToUseActiveStandard(t, root)
	}, "section \"Standards Added Since Last Refresh\"")
}

func assertDeprecatedStandardWarning(t *testing.T) {
	t.Helper()
	root := copyTraceabilityFixtureProject(t, "valid-project")
	rewriteDeterministicStandard(t, root, "status: active", "status: deprecated\nreplaced_by: standards/global/deterministic-check-write-v2.md")
	v2Path := filepath.Join(root, "runecontext", "standards", "global", "deterministic-check-write-v2.md")
	writeStandardFixture(t, v2Path, "---\nschema_version: 1\nid: global/deterministic-check-write-v2\ntitle: Deterministic Check Write v2\nstatus: active\n---\n\n# Deterministic Check Write v2\n\nUse the newer wording.\n")
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(root)
	if err != nil {
		t.Fatalf("expected deprecated applicable standard to warn, got %v", err)
	}
	defer index.Close()
	for _, diagnostic := range index.Diagnostics {
		if diagnostic.Code == "deprecated_standard_referenced" {
			if diagnostic.Path != "changes/CHG-2026-001-a3f2-auth-gateway/standards.md" {
				t.Fatalf("expected relative diagnostic path, got %#v", diagnostic)
			}
			return
		}
	}
	t.Fatalf("expected deprecated standard warning, got %#v", index.Diagnostics)
}

func assertExcludedDraftStandardAllowed(t *testing.T) {
	t.Helper()
	assertStandardValidationPass(t, func(root string) {
		rewriteDeterministicStandard(t, root, "status: active", "status: draft")
		standardsPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "standards.md")
		rewriteFile(t, standardsPath, func(string) string {
			return "## Applicable Standards\n- `standards/global/other-active.md`: Replacement active standard.\n\n## Standards Considered But Excluded\n- `standards/global/deterministic-check-write.md`: Still draft and intentionally excluded.\n"
		})
		writeActiveStandardFixture(t, root)
		rewriteOtherChangeToUseActiveStandard(t, root)
	})
}

func assertAliasCollisionFailure(t *testing.T) {
	t.Helper()
	assertStandardValidationFailure(t, func(root string) {
		otherPath := filepath.Join(root, "runecontext", "standards", "global", "other-active.md")
		writeStandardFixture(t, otherPath, "---\nschema_version: 1\nid: global/other-active\ntitle: Other Active\nstatus: active\naliases:\n  - global/legacy-id\n---\n\n# Other Active\n\nAlias collision test.\n")
		rewriteDeterministicStandard(t, root, "suggested_context_bundles:\n  - go-control-plane", "suggested_context_bundles:\n  - go-control-plane\naliases:\n  - global/legacy-id")
	}, "alias \"global/legacy-id\" is duplicated")
}

func assertStandardsDocumentValidationFailure(t *testing.T, content, want string) {
	t.Helper()
	assertStandardValidationFailure(t, func(root string) {
		standardsPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "standards.md")
		rewriteFile(t, standardsPath, func(string) string { return content })
	}, want)
}

func assertProposalDocumentValidationFailure(t *testing.T, suffix, want string) {
	t.Helper()
	assertStandardValidationFailure(t, func(root string) {
		proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
		rewriteFile(t, proposalPath, func(text string) string { return text + suffix })
	}, want)
}

func assertProposalDocumentPasses(t *testing.T, suffix string) {
	t.Helper()
	assertStandardValidationPass(t, func(root string) {
		proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
		rewriteFile(t, proposalPath, func(text string) string { return text + suffix })
	})
}

func assertMissingProposalStandardPath(t *testing.T) {
	t.Helper()
	root := copyTraceabilityFixtureProject(t, "valid-project")
	proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
	rewriteFile(t, proposalPath, func(text string) string {
		return text + "\n\nSee `standards/global/missing.md` for the durable rule.\n"
	})
	v := NewValidator(schemaRoot(t))
	_, err := v.ValidateProject(root)
	if err == nil || !strings.Contains(err.Error(), "points to missing standard") || strings.Contains(err.Error(), root) {
		t.Fatalf("expected relative missing-standard failure, got %v", err)
	}
}

func assertSpecDocumentValidationFailure(t *testing.T, suffix, want string) {
	t.Helper()
	assertStandardValidationFailure(t, func(root string) {
		specPath := filepath.Join(root, "runecontext", "specs", "auth-gateway.md")
		rewriteFile(t, specPath, func(text string) string { return text + suffix })
	}, want)
}

func assertSpecDocumentPasses(t *testing.T, suffix string) {
	t.Helper()
	assertStandardValidationPass(t, func(root string) {
		specPath := filepath.Join(root, "runecontext", "specs", "auth-gateway.md")
		rewriteFile(t, specPath, func(text string) string { return text + suffix })
	})
}

func writeActiveStandardFixture(t *testing.T, root string) {
	t.Helper()
	otherPath := filepath.Join(root, "runecontext", "standards", "global", "other-active.md")
	writeStandardFixture(t, otherPath, "---\nschema_version: 1\nid: global/other-active\ntitle: Other Active\nstatus: active\n---\n\n# Other Active\n\nUse the active path.\n")
}

func rewriteOtherChangeToUseActiveStandard(t *testing.T, root string) {
	t.Helper()
	otherChangeStandards := filepath.Join(root, "runecontext", "changes", "CHG-2026-002-b4c3-auth-revision", "standards.md")
	rewriteFile(t, otherChangeStandards, func(text string) string {
		return strings.Replace(text, "standards/global/deterministic-check-write.md", "standards/global/other-active.md", 1)
	})
}

func writeStandardFixture(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write standard fixture: %v", err)
	}
}
