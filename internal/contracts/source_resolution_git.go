package contracts

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func resolveGitSourceMaterialization(resolver gitResolver, subdir string, sourceMap map[string]any, gitTrust GitTrustInputs) (gitSourceResult, error) {
	if commit, ok := gitCommitSource(sourceMap); ok {
		return resolver.resolvePinnedCommit(subdir, commit)
	}
	if ref, ok := gitMutableRefSource(sourceMap); ok {
		return resolver.resolveMutableRef(subdir, ref, sourceMap)
	}
	if tagName, expectCommit, ok, err := gitSignedTagSource(sourceMap); err != nil {
		return gitSourceResult{}, err
	} else if ok {
		return resolver.resolveSignedTag(subdir, tagName, expectCommit, gitTrust)
	}
	return gitSourceResult{}, &ValidationError{Path: resolver.configPath, Message: "git source must declare commit, signed_tag, or ref"}
}

func gitCommitSource(sourceMap map[string]any) (string, bool) {
	if rawCommit, ok := sourceMap["commit"]; ok {
		commit := strings.TrimSpace(fmt.Sprint(rawCommit))
		return commit, commit != ""
	}
	return "", false
}

func gitMutableRefSource(sourceMap map[string]any) (string, bool) {
	if rawRef, ok := sourceMap["ref"]; ok {
		ref := strings.TrimSpace(fmt.Sprint(rawRef))
		return ref, ref != ""
	}
	return "", false
}

func gitSignedTagSource(sourceMap map[string]any) (string, string, bool, error) {
	if _, ok := sourceMap["signed_tag"]; !ok {
		return "", "", false, nil
	}
	tagName := strings.TrimSpace(fmt.Sprint(sourceMap["signed_tag"]))
	if tagName == "" {
		return "", "", false, fmt.Errorf("git signed_tag must not be empty")
	}
	expectCommit, ok := sourceMap["expect_commit"]
	if !ok || strings.TrimSpace(fmt.Sprint(expectCommit)) == "" {
		return "", "", false, fmt.Errorf("git expect_commit must not be empty")
	}
	return tagName, strings.TrimSpace(fmt.Sprint(expectCommit)), true, nil
}

func (r gitResolver) resolvePinnedCommit(subdir, commit string) (gitSourceResult, error) {
	if err := validateGitCommit(commit); err != nil {
		return gitSourceResult{}, &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	tree, err := r.materialize(commit, subdir)
	if err != nil {
		return gitSourceResult{}, err
	}
	return gitSourceResult{tree: tree, commit: commit, ref: commit, posture: VerificationPosturePinnedCommit}, nil
}

func (r gitResolver) resolveMutableRef(subdir, ref string, sourceMap map[string]any) (gitSourceResult, error) {
	if err := validateGitRef(ref); err != nil {
		return gitSourceResult{}, &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	if allow, _ := sourceMap["allow_mutable_ref"].(bool); !allow {
		return gitSourceResult{}, &ValidationError{Path: r.configPath, Message: "mutable git refs require allow_mutable_ref: true"}
	}
	tree, commit, err := r.materializeRef(ref, subdir)
	if err != nil {
		return gitSourceResult{}, err
	}
	return gitSourceResult{tree: tree, commit: commit, ref: ref, posture: VerificationPostureUnverifiedMutableRef, diagnostics: []ResolutionDiagnostic{{Severity: DiagnosticSeverityWarning, Code: "mutable_ref", Message: "mutable git refs are unverified and may resolve differently over time"}}}, nil
}

func (r gitResolver) resolveSignedTag(subdir, tagName, expectCommit string, gitTrust GitTrustInputs) (gitSourceResult, error) {
	if err := validateSignedTagInputs(r.configPath, tagName, expectCommit, gitTrust); err != nil {
		return gitSourceResult{}, err
	}
	tree, commit, verification, err := r.materializeSignedTag(tagName, expectCommit, subdir, gitTrust.SignedTagVerifier)
	if err != nil {
		return gitSourceResult{}, err
	}
	return gitSourceResult{tree: tree, commit: commit, ref: tagName, posture: VerificationPostureVerifiedSignedTag, signedTagVerification: verification}, nil
}

func validateSignedTagInputs(configPath, tagName, expectCommit string, gitTrust GitTrustInputs) error {
	if err := validateGitRef(tagName); err != nil {
		return &ValidationError{Path: configPath, Message: strings.Replace(err.Error(), "git ref", "git signed_tag", 1)}
	}
	if err := validateGitCommit(expectCommit); err != nil {
		return &ValidationError{Path: configPath, Message: strings.Replace(err.Error(), "git commit", "git expect_commit", 1)}
	}
	if gitTrust.SignedTagVerifier != nil {
		return nil
	}
	return &SignedTagVerificationError{Path: configPath, Tag: tagName, Reason: SignedTagFailureMissingTrust, Message: "signed tag resolution requires explicit trusted signer inputs", Diagnostics: []ResolutionDiagnostic{{Severity: DiagnosticSeverityError, Code: string(SignedTagFailureMissingTrust), Message: "signed tag resolution requires explicit trusted signer inputs"}}}
}

func applyGitResolution(base *SourceResolution, subdir string, result gitSourceResult) {
	base.SourceRoot = subdir
	base.SourceMode = SourceModeGit
	base.SourceRef = result.ref
	base.ResolvedCommit = result.commit
	base.VerificationPosture = result.posture
	if result.signedTagVerification != nil {
		base.VerifiedSignerIdentity = result.signedTagVerification.SignerIdentity
		base.VerifiedSignerFingerprint = result.signedTagVerification.SignerFingerprint
		result.diagnostics = append(result.diagnostics, result.signedTagVerification.Diagnostics...)
	}
	base.Diagnostics = result.diagnostics
	base.Tree = result.tree
}

func (r gitResolver) materialize(commit, subdir string) (*LocalSourceTree, error) {
	tree, resolvedCommit, err := r.materializeCommitToTree(commit, subdir)
	if err != nil {
		return nil, err
	}
	if resolvedCommit != commit {
		_ = tree.Close()
		return nil, &ValidationError{Path: r.configPath, Message: fmt.Sprintf("resolved git commit %q did not match pinned commit %q", resolvedCommit, commit)}
	}
	return tree, nil
}

func (r gitResolver) materializeRef(ref, subdir string) (*LocalSourceTree, string, error) {
	tempRoot, repoRoot, err := r.initializeRepository()
	if err != nil {
		return nil, "", err
	}
	if err := runGit("-C", repoRoot, "fetch", "--quiet", "--no-tags", "--depth", "1", "origin", ref); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	if err := runGit("-C", repoRoot, "checkout", "--quiet", "--detach", "FETCH_HEAD"); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	return r.finalizeMaterializedTree(tempRoot, repoRoot, subdir)
}

func (r gitResolver) materializeSignedTag(tagName, expectCommit, subdir string, verifier SignedTagVerifier) (*LocalSourceTree, string, *SignedTagVerification, error) {
	tempRoot, repoRoot, err := r.initializeRepository()
	if err != nil {
		return nil, "", nil, err
	}
	if err := runGit("-C", repoRoot, "fetch", "--quiet", "--no-tags", "origin", "+refs/heads/*:refs/remotes/origin/*", "+refs/tags/*:refs/tags/*"); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", nil, &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	verification, err := verifyMaterializedSignedTag(r.configPath, repoRoot, tagName, verifier)
	if err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", nil, err
	}
	resolvedCommit, err := verifySignedTagCommit(r.configPath, repoRoot, tagName, expectCommit, verification)
	if err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", nil, err
	}
	if err := runGit("-C", repoRoot, "checkout", "--quiet", "--detach", resolvedCommit); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", nil, &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	tree, finalizedCommit, err := r.finalizeMaterializedTree(tempRoot, repoRoot, subdir)
	if err != nil {
		return nil, "", nil, err
	}
	if finalizedCommit != resolvedCommit {
		_ = tree.Close()
		return nil, "", nil, &ValidationError{Path: r.configPath, Message: fmt.Sprintf("resolved git commit %q did not match verified signed-tag commit %q", finalizedCommit, resolvedCommit)}
	}
	return tree, resolvedCommit, verification, nil
}

func verifyMaterializedSignedTag(configPath, repoRoot, tagName string, verifier SignedTagVerifier) (*SignedTagVerification, error) {
	verification, err := verifier.VerifySignedTag(repoRoot, tagName)
	if err != nil {
		return nil, normalizeSignedTagVerificationError(configPath, tagName, err)
	}
	if err := validateSignedTagVerification(verification, tagName); err != nil {
		return nil, normalizeSignedTagVerificationError(configPath, tagName, err)
	}
	return verification, nil
}

func normalizeSignedTagVerificationError(configPath, tagName string, err error) error {
	var verificationErr *SignedTagVerificationError
	if errors.As(err, &verificationErr) {
		if verificationErr.Path == "" {
			verificationErr.Path = configPath
		}
		if verificationErr.Tag == "" {
			verificationErr.Tag = tagName
		}
		return verificationErr
	}
	return &ValidationError{Path: configPath, Message: err.Error()}
}

func verifySignedTagCommit(configPath, repoRoot, tagName, expectCommit string, verification *SignedTagVerification) (string, error) {
	commitOutput, err := gitOutput("-C", repoRoot, "rev-parse", tagName+"^{commit}")
	if err != nil {
		return "", &ValidationError{Path: configPath, Message: err.Error()}
	}
	resolvedCommit := strings.TrimSpace(commitOutput)
	if resolvedCommit == expectCommit {
		return resolvedCommit, nil
	}
	message := fmt.Sprintf("signed tag %q resolved commit %q did not match expect_commit %q", tagName, resolvedCommit, expectCommit)
	return "", &SignedTagVerificationError{Path: configPath, Tag: tagName, Reason: SignedTagFailureExpectCommitMismatch, Message: message, ResolvedCommit: resolvedCommit, SignerIdentity: verification.SignerIdentity, SignerFingerprint: verification.SignerFingerprint, Diagnostics: []ResolutionDiagnostic{{Severity: DiagnosticSeverityError, Code: string(SignedTagFailureExpectCommitMismatch), Message: message}}}
}
