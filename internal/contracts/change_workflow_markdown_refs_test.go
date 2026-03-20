package contracts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRewriteMarkdownReferenceTargets(t *testing.T) {
	data := []byte("See specs/auth-gateway.md#auth-gateway and decisions/DEC-0001-trust-boundary-model.md#trust-boundary-model.\n")
	rewritten, count, err := RewriteMarkdownReferenceTargets(data, []MarkdownReferenceRewrite{{OldPath: "specs/auth-gateway.md", NewPath: "specs/security/auth-gateway.md", OldFragment: "auth-gateway", NewFragment: "gateway-overview"}})
	if err != nil {
		t.Fatalf("rewrite markdown refs: %v", err)
	}
	if got, want := count, 1; got != want {
		t.Fatalf("expected %d rewrite, got %d", want, got)
	}
	text := string(rewritten)
	if !strings.Contains(text, "specs/security/auth-gateway.md#gateway-overview") || !strings.Contains(text, "decisions/DEC-0001-trust-boundary-model.md#trust-boundary-model") {
		t.Fatalf("unexpected rewrite result: %q", text)
	}
}

func TestRewriteMarkdownReferenceTargetsIgnoresFencedCode(t *testing.T) {
	assertMarkdownRewriteScenario(t, "```md\nspecs/auth-gateway.md#auth-gateway\n```\n\nSee specs/auth-gateway.md#auth-gateway.\n", "```md\nspecs/auth-gateway.md#auth-gateway\n```", "See specs/security/auth-gateway.md#gateway-overview.")
}

func TestRewriteMarkdownReferenceTargetsIgnoresLongerFenceNesting(t *testing.T) {
	data := []byte("````md\n```\nspecs/auth-gateway.md#auth-gateway\n```\n````\n\nSee specs/auth-gateway.md#auth-gateway.\n")
	rewritten, count, err := rewriteAuthGatewayRef(data)
	if err != nil {
		t.Fatalf("rewrite markdown refs: %v", err)
	}
	if got, want := count, 1; got != want {
		t.Fatalf("expected %d rewrite, got %d", want, got)
	}
	if !strings.Contains(string(rewritten), "````md\n```\nspecs/auth-gateway.md#auth-gateway\n```\n````") {
		t.Fatalf("expected nested fenced code ref to stay unchanged, got %q", string(rewritten))
	}
}

func TestRewriteMarkdownReferenceTargetsUsesFirstMatch(t *testing.T) {
	data := []byte("See specs/auth-gateway.md#auth-gateway.\n")
	rewritten, count, err := RewriteMarkdownReferenceTargets(data, []MarkdownReferenceRewrite{{OldPath: "specs/auth-gateway.md", NewPath: "specs/security/auth-gateway.md"}, {OldPath: "specs/auth-gateway.md", OldFragment: "auth-gateway", NewFragment: "gateway-overview"}})
	if err != nil {
		t.Fatalf("rewrite markdown refs: %v", err)
	}
	if got, want := count, 1; got != want {
		t.Fatalf("expected %d rewrite, got %d", want, got)
	}
	if got, want := string(rewritten), "See specs/security/auth-gateway.md#auth-gateway.\n"; got != want {
		t.Fatalf("unexpected rewrite result: %q", got)
	}
}

func TestExtractMarkdownDeepRefsFromTextStopsAtUTF8Punctuation(t *testing.T) {
	assertExtractedMarkdownRefCount(t, "See \"specs/auth-gateway.md#auth-gateway\" and “specs/auth-gateway.md#auth-gateway” for context.", 2)
}

func TestExtractMarkdownDeepRefsFromTextStopsBeforeNonASCIIFragmentSuffix(t *testing.T) {
	assertSingleMarkdownRef(t, "See specs/auth-gateway.md#auth-gatewayの情報 for context.")
}

func TestExtractMarkdownDeepRefsFromTextRequiresIndexedRootBoundary(t *testing.T) {
	refs, err := extractMarkdownDeepRefsFromText("见specs/auth-gateway.md#auth-gateway and docs/auth-gateway.md#auth-gateway should stay plain text.", 0)
	if err != nil {
		t.Fatalf("extract markdown refs: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("expected no refs for non-indexed or boundaryless paths, got %#v", refs)
	}
}

func TestValidateProjectMarkdownDeepRefs(t *testing.T) {
	for _, tc := range markdownDeepRefScenarios() {
		t.Run(tc.name, func(t *testing.T) { tc.run(t) })
	}
}

type markdownDeepRefScenario struct {
	name string
	run  func(*testing.T)
}

func markdownDeepRefScenarios() []markdownDeepRefScenario {
	return []markdownDeepRefScenario{
		{name: "valid heading fragment", run: func(t *testing.T) {
			assertProposalValidation(t, "\nSee specs/auth-gateway.md#auth-gateway for the stable reference.\n", "", true)
		}},
		{name: "missing heading fragment", run: func(t *testing.T) {
			assertProposalValidation(t, "\nSee specs/auth-gateway.md#missing-heading for more context.\n", "missing heading fragment", false)
		}},
		{name: "line-number fragment rejected", run: func(t *testing.T) {
			assertProposalValidation(t, "\nAvoid specs/auth-gateway.md#L10 because line numbers are not durable.\n", "must use a heading fragment", false)
		}},
		{name: "absolute path rejected", run: func(t *testing.T) {
			assertProposalValidation(t, "\nBad ref: /specs/auth-gateway.md#auth-gateway\n", "must not use an absolute path", false)
		}},
		{name: "relative traversal rejected", run: func(t *testing.T) {
			assertProposalValidation(t, "\nBad ref: ../specs/auth-gateway.md#auth-gateway\n", "RuneContext-root-relative path", false)
		}},
		{name: "fenced code refs ignored", run: func(t *testing.T) {
			assertProposalValidation(t, "\n```md\n/specs/auth-gateway.md#L10\n```\n", "", true)
		}},
		{name: "blockquote fenced code refs ignored", run: func(t *testing.T) {
			assertProposalValidation(t, "\n> ```md\n> /specs/auth-gateway.md#L10\n> ```\n", "", true)
		}},
		{name: "numeric line fragment rejected", run: func(t *testing.T) {
			assertProposalValidation(t, "\nAvoid specs/auth-gateway.md#42 because line numbers are not durable.\n", "must use a heading fragment", false)
		}},
		{name: "line range fragment rejected", run: func(t *testing.T) { assertProposalLineRangeFailure(t) }},
		{name: "external markdown URL ignored", run: func(t *testing.T) {
			assertProposalValidation(t, "\nReference https://example.com/docs/auth-gateway.md#overview for external context.\n", "", true)
		}},
		{name: "utf8 punctuation around deep ref ignored", run: func(t *testing.T) {
			assertProposalValidation(t, "\nSee “specs/auth-gateway.md#auth-gateway” for quoted context.\n", "", true)
		}},
		{name: "utf8 prose after fragment ignored", run: func(t *testing.T) {
			assertProposalValidation(t, "\nSee specs/auth-gateway.md#auth-gatewayの情報 for localized context.\n", "", true)
		}},
		{name: "project markdown target allowed", run: func(t *testing.T) { assertChangeProjectMarkdownRefAllowed(t) }},
	}
}

func assertProposalValidation(t *testing.T, suffix, want string, shouldPass bool) {
	t.Helper()
	root := copyTraceabilityFixtureProject(t, "valid-project")
	proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
	rewriteFile(t, proposalPath, func(text string) string { return text + suffix })
	v := NewValidator(schemaRoot(t))
	_, err := v.ValidateProject(root)
	assertValidationOutcome(t, err, want, shouldPass)
}

func assertProposalLineRangeFailure(t *testing.T) {
	t.Helper()
	root := copyTraceabilityFixtureProject(t, "valid-project")
	proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
	rewriteFile(t, proposalPath, func(text string) string {
		return text + "\nAvoid specs/auth-gateway.md#l10-l20 because line ranges are not durable.\n"
	})
	v := NewValidator(schemaRoot(t))
	_, err := v.ValidateProject(root)
	if err == nil || (!strings.Contains(err.Error(), "must use a heading fragment") && !strings.Contains(err.Error(), "missing heading fragment")) {
		t.Fatalf("expected line-range fragment failure, got %v", err)
	}
}

func assertChangeProjectMarkdownRefAllowed(t *testing.T) {
	t.Helper()
	root := copyTraceabilityFixtureProject(t, "valid-project")
	projectDir := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir change dir: %v", err)
	}
	projectPath := filepath.Join(projectDir, "mission.md")
	if err := os.WriteFile(projectPath, []byte("# Mission\n\nShip stable change workflows.\n"), 0o644); err != nil {
		t.Fatalf("write project markdown: %v", err)
	}
	proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
	rewriteFile(t, proposalPath, func(text string) string {
		return text + "\nSee changes/CHG-2026-001-a3f2-auth-gateway/mission.md#mission for the high-level intent.\n"
	})
	v := NewValidator(schemaRoot(t))
	if _, err := v.ValidateProject(root); err != nil {
		t.Fatalf("expected indexed change markdown deep ref to validate, got %v", err)
	}
}

func assertValidationOutcome(t *testing.T, err error, want string, shouldPass bool) {
	t.Helper()
	if shouldPass {
		if err != nil {
			t.Fatalf("expected validation to pass, got %v", err)
		}
		return
	}
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("expected %q failure, got %v", want, err)
	}
}

func rewriteAuthGatewayRef(data []byte) ([]byte, int, error) {
	return RewriteMarkdownReferenceTargets(data, []MarkdownReferenceRewrite{{OldPath: "specs/auth-gateway.md", NewPath: "specs/security/auth-gateway.md", OldFragment: "auth-gateway", NewFragment: "gateway-overview"}})
}

func assertMarkdownRewriteScenario(t *testing.T, body, keep, want string) {
	t.Helper()
	rewritten, count, err := rewriteAuthGatewayRef([]byte(body))
	if err != nil {
		t.Fatalf("rewrite markdown refs: %v", err)
	}
	if got, expected := count, 1; got != expected {
		t.Fatalf("expected %d rewrite, got %d", expected, got)
	}
	text := string(rewritten)
	if !strings.Contains(text, keep) || !strings.Contains(text, want) {
		t.Fatalf("unexpected rewrite result: %q", text)
	}
}

func assertExtractedMarkdownRefCount(t *testing.T, text string, want int) {
	t.Helper()
	refs, err := extractMarkdownDeepRefsFromText(text, 0)
	if err != nil {
		t.Fatalf("extract markdown refs: %v", err)
	}
	if got := len(refs); got != want {
		t.Fatalf("expected %d refs, got %d", want, got)
	}
	for _, ref := range refs {
		if ref.Path != "specs/auth-gateway.md" || ref.Fragment != "auth-gateway" {
			t.Fatalf("unexpected ref: %#v", ref)
		}
	}
}

func assertSingleMarkdownRef(t *testing.T, text string) {
	t.Helper()
	refs, err := extractMarkdownDeepRefsFromText(text, 0)
	if err != nil {
		t.Fatalf("extract markdown refs: %v", err)
	}
	if len(refs) != 1 || refs[0].Path != "specs/auth-gateway.md" || refs[0].Fragment != "auth-gateway" {
		t.Fatalf("unexpected refs: %#v", refs)
	}
}
