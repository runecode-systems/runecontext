package contracts

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	gitCommitPattern    = regexp.MustCompile(`^[a-f0-9]{40}$`)
	gitRefPattern       = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)
	gitCommandTimeout   = 30 * time.Second
	gitURLControlChars  = "\x00\r\n\t"
	localSnapshotLimits = snapshotLimits{
		MaxFiles: 10000,
		MaxBytes: 128 << 20,
		MaxDepth: 64,
		Excludes: map[string]struct{}{
			".git": {},
		},
	}
)

type snapshotLimits struct {
	MaxFiles int
	MaxBytes int64
	MaxDepth int
	Excludes map[string]struct{}
}

type snapshotState struct {
	files int
	bytes int64
}

type ExecutionMode string

const (
	ExecutionModeLocal    ExecutionMode = "local"
	ExecutionModeRemoteCI ExecutionMode = "remote_ci"
)

type ConfigDiscoveryMode string

const (
	ConfigDiscoveryExplicitRoot    ConfigDiscoveryMode = "explicit_root"
	ConfigDiscoveryNearestAncestor ConfigDiscoveryMode = "nearest_ancestor"
)

type SourceMode string

const (
	SourceModeEmbedded SourceMode = "embedded"
	SourceModeGit      SourceMode = "git"
	SourceModePath     SourceMode = "path"
)

type VerificationPosture string

const (
	VerificationPosturePinnedCommit         VerificationPosture = "pinned_commit"
	VerificationPostureVerifiedSignedTag    VerificationPosture = "verified_signed_tag"
	VerificationPostureUnverifiedMutableRef VerificationPosture = "unverified_mutable_ref"
	VerificationPostureUnverifiedLocal      VerificationPosture = "unverified_local_source"
	VerificationPostureEmbedded             VerificationPosture = "embedded"
)

type DiagnosticSeverity string

const (
	DiagnosticSeverityInfo    DiagnosticSeverity = "info"
	DiagnosticSeverityWarning DiagnosticSeverity = "warning"
	DiagnosticSeverityError   DiagnosticSeverity = "error"
)

type ResolveOptions struct {
	ConfigDiscovery ConfigDiscoveryMode
	ExecutionMode   ExecutionMode
}

type ResolutionDiagnostic struct {
	Severity DiagnosticSeverity `json:"severity" yaml:"severity"`
	Code     string             `json:"code" yaml:"code"`
	Message  string             `json:"message" yaml:"message"`
}

type LocalSourceTree struct {
	Root         string `json:"-" yaml:"-"`
	SnapshotKind string `json:"snapshot_kind,omitempty" yaml:"snapshot_kind,omitempty"`
	cleanupRoot  string
}

func (t *LocalSourceTree) Close() error {
	if t == nil || t.cleanupRoot == "" {
		return nil
	}
	err := os.RemoveAll(t.cleanupRoot)
	t.cleanupRoot = ""
	return err
}

type SourceResolution struct {
	SelectedConfigPath  string                 `json:"selected_config_path" yaml:"selected_config_path"`
	ProjectRoot         string                 `json:"project_root" yaml:"project_root"`
	SourceRoot          string                 `json:"source_root" yaml:"source_root"`
	SourceMode          SourceMode             `json:"source_mode" yaml:"source_mode"`
	SourceRef           string                 `json:"source_ref" yaml:"source_ref"`
	ResolvedCommit      string                 `json:"resolved_commit,omitempty" yaml:"resolved_commit,omitempty"`
	VerificationPosture VerificationPosture    `json:"verification_posture" yaml:"verification_posture"`
	Diagnostics         []ResolutionDiagnostic `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`
	Tree                *LocalSourceTree       `json:"-" yaml:"-"`
}

func (r *SourceResolution) Close() error {
	if r == nil || r.Tree == nil {
		return nil
	}
	return r.Tree.Close()
}

func (r *SourceResolution) MaterializedRoot() string {
	if r == nil || r.Tree == nil {
		return ""
	}
	return r.Tree.Root
}

type LoadedProject struct {
	RootConfigData []byte
	RootConfig     map[string]any
	Resolution     *SourceResolution
}

func (p *LoadedProject) Close() error {
	if p == nil || p.Resolution == nil {
		return nil
	}
	return p.Resolution.Close()
}

func (v *Validator) LoadProject(path string, options ResolveOptions) (*LoadedProject, error) {
	options = normalizeResolveOptions(options)
	configPath, projectRoot, err := discoverConfig(path, options.ConfigDiscovery)
	if err != nil {
		return nil, err
	}
	rootData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, &ValidationError{Path: configPath, Message: err.Error()}
	}
	if err := v.ValidateYAMLFile("runecontext.schema.json", configPath, rootData); err != nil {
		return nil, err
	}
	parsed, err := parseYAML(rootData)
	if err != nil {
		return nil, &ValidationError{Path: configPath, Message: err.Error()}
	}
	rootConfig, err := expectObject(configPath, parsed, "root config")
	if err != nil {
		return nil, err
	}
	resolution, err := resolveSourceFromConfig(configPath, projectRoot, rootData, options)
	if err != nil {
		return nil, err
	}
	return &LoadedProject{
		RootConfigData: append([]byte(nil), rootData...),
		RootConfig:     rootConfig,
		Resolution:     resolution,
	}, nil
}

func (v *Validator) ValidateLoadedProject(loaded *LoadedProject) (*ProjectIndex, error) {
	if loaded == nil || loaded.Resolution == nil {
		return nil, fmt.Errorf("loaded project is required")
	}
	contentRoot := loaded.Resolution.MaterializedRoot()
	if contentRoot == "" {
		return nil, fmt.Errorf("resolved source root is unavailable")
	}
	rootConfigPath := loaded.Resolution.SelectedConfigPath
	rootData := loaded.RootConfigData
	index, err := buildProjectIndex(v, contentRoot)
	if err != nil {
		return nil, err
	}
	index.RootConfigPath = rootConfigPath
	index.ContentRoot = contentRoot
	index.Resolution = loaded.Resolution
	if err := v.validateBundles(rootConfigPath, rootData, filepath.Join(contentRoot, "bundles")); err != nil {
		return nil, err
	}
	for path, record := range index.StatusFiles {
		if err := v.ValidateExtensionOptIn(rootConfigPath, rootData, path, record.Raw); err != nil {
			return nil, err
		}
		for _, key := range []string{"depends_on", "informed_by", "related_changes"} {
			if err := validateChangeIDRefs(path, key, record.Data[key], index.ChangeIDs); err != nil {
				return nil, err
			}
		}
		if err := validatePathRefs(path, "related_specs", record.Data["related_specs"], index.SpecPaths); err != nil {
			return nil, err
		}
		if err := validatePathRefs(path, "related_decisions", record.Data["related_decisions"], index.DecisionPaths); err != nil {
			return nil, err
		}
	}
	return index, nil
}

func (v *Validator) ValidateProjectWithOptions(path string, options ResolveOptions) (*ProjectIndex, error) {
	loaded, err := v.LoadProject(path, options)
	if err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		_ = loaded.Close()
		return nil, err
	}
	return index, nil
}

func normalizeResolveOptions(options ResolveOptions) ResolveOptions {
	if options.ConfigDiscovery == "" {
		options.ConfigDiscovery = ConfigDiscoveryExplicitRoot
	}
	if options.ExecutionMode == "" {
		options.ExecutionMode = ExecutionModeLocal
	}
	return options
}

func discoverConfig(path string, mode ConfigDiscoveryMode) (string, string, error) {
	start, err := filepath.Abs(path)
	if err != nil {
		return "", "", &ValidationError{Path: path, Message: err.Error()}
	}
	info, err := os.Stat(start)
	if err == nil && !info.IsDir() {
		if filepath.Base(start) == "runecontext.yaml" {
			start = filepath.Dir(start)
		} else {
			start = filepath.Dir(start)
		}
	}
	switch mode {
	case ConfigDiscoveryNearestAncestor:
		current := start
		for {
			candidate := filepath.Join(current, "runecontext.yaml")
			if _, err := os.Stat(candidate); err == nil {
				return filepath.Clean(candidate), filepath.Clean(current), nil
			}
			next := filepath.Dir(current)
			if next == current {
				return "", "", &ValidationError{Path: start, Message: "no runecontext.yaml found in current directory or ancestors"}
			}
			current = next
		}
	case ConfigDiscoveryExplicitRoot:
		candidate := filepath.Join(start, "runecontext.yaml")
		if _, err := os.Stat(candidate); err != nil {
			return "", "", &ValidationError{Path: candidate, Message: err.Error()}
		}
		return filepath.Clean(candidate), filepath.Clean(start), nil
	default:
		return "", "", &ValidationError{Path: start, Message: fmt.Sprintf("unsupported config discovery mode %q", mode)}
	}
}

func resolveSourceFromConfig(configPath, projectRoot string, rootData []byte, options ResolveOptions) (*SourceResolution, error) {
	parsed, err := parseYAML(rootData)
	if err != nil {
		return nil, &ValidationError{Path: configPath, Message: err.Error()}
	}
	rootMap, err := expectObject(configPath, parsed, "root config")
	if err != nil {
		return nil, err
	}
	sourceRaw, ok := rootMap["source"]
	if !ok {
		return nil, &ValidationError{Path: configPath, Message: "root config is missing source"}
	}
	sourceMap, ok := sourceRaw.(map[string]any)
	if !ok {
		return nil, &ValidationError{Path: configPath, Message: "source must decode to an object"}
	}
	sourceType := fmt.Sprint(sourceMap["type"])
	base := &SourceResolution{
		SelectedConfigPath: filepath.Clean(configPath),
		ProjectRoot:        filepath.Clean(projectRoot),
	}
	switch sourceType {
	case string(SourceModeEmbedded):
		return resolveEmbeddedSource(base, configPath, projectRoot, sourceMap)
	case string(SourceModeGit):
		return resolveGitSource(base, configPath, sourceMap)
	case string(SourceModePath):
		return resolvePathSource(base, configPath, projectRoot, sourceMap, options.ExecutionMode)
	default:
		return nil, &ValidationError{Path: configPath, Message: fmt.Sprintf("unsupported source type %q", sourceType)}
	}
}

func resolveEmbeddedSource(base *SourceResolution, configPath, projectRoot string, sourceMap map[string]any) (*SourceResolution, error) {
	declaredRoot, absRoot, err := resolveEmbeddedSourceRoot(configPath, projectRoot, sourceMap)
	if err != nil {
		return nil, err
	}
	base.SourceRoot = declaredRoot
	base.SourceMode = SourceModeEmbedded
	base.SourceRef = "embedded"
	base.VerificationPosture = VerificationPostureEmbedded
	base.Tree = &LocalSourceTree{Root: absRoot, SnapshotKind: "live"}
	return base, nil
}

func resolvePathSource(base *SourceResolution, configPath, projectRoot string, sourceMap map[string]any, executionMode ExecutionMode) (*SourceResolution, error) {
	if executionMode == ExecutionModeRemoteCI {
		return nil, &ValidationError{Path: configPath, Message: "source.type=path is invalid in execution mode remote_ci"}
	}
	declaredRoot, absRoot, err := resolveDeclaredLocalSourceRoot(configPath, projectRoot, sourceMap)
	if err != nil {
		return nil, err
	}
	tree, err := snapshotLocalTree(absRoot)
	if err != nil {
		return nil, &ValidationError{Path: configPath, Message: err.Error()}
	}
	base.SourceRoot = declaredRoot
	base.SourceMode = SourceModePath
	base.SourceRef = declaredRoot
	base.VerificationPosture = VerificationPostureUnverifiedLocal
	base.Diagnostics = []ResolutionDiagnostic{
		{
			Severity: DiagnosticSeverityWarning,
			Code:     "unverified_local_source",
			Message:  "local path sources are unverified and non-auditable",
		},
	}
	base.Tree = tree
	return base, nil
}

func resolveGitSource(base *SourceResolution, configPath string, sourceMap map[string]any) (*SourceResolution, error) {
	url := strings.TrimSpace(fmt.Sprint(sourceMap["url"]))
	if url == "" {
		return nil, &ValidationError{Path: configPath, Message: "git source url must not be empty"}
	}
	if err := validateGitURL(url); err != nil {
		return nil, &ValidationError{Path: configPath, Message: err.Error()}
	}
	subdir := "runecontext"
	if rawSubdir, ok := sourceMap["subdir"]; ok && strings.TrimSpace(fmt.Sprint(rawSubdir)) != "" {
		normalizedSubdir, err := normalizeContainedRelativePath(fmt.Sprint(rawSubdir))
		if err != nil {
			return nil, &ValidationError{Path: configPath, Message: fmt.Sprintf("git subdir %v", err)}
		}
		subdir = normalizedSubdir
	}
	resolver := gitResolver{configPath: configPath, url: url}
	var (
		tree        *LocalSourceTree
		commit      string
		ref         string
		posture     VerificationPosture
		diagnostics []ResolutionDiagnostic
		err         error
	)
	if rawCommit, ok := sourceMap["commit"]; ok && strings.TrimSpace(fmt.Sprint(rawCommit)) != "" {
		ref = strings.TrimSpace(fmt.Sprint(rawCommit))
		if err := validateGitCommit(ref); err != nil {
			return nil, &ValidationError{Path: configPath, Message: err.Error()}
		}
		commit = ref
		posture = VerificationPosturePinnedCommit
		tree, err = resolver.materialize(ref, subdir)
	} else if rawRef, ok := sourceMap["ref"]; ok && strings.TrimSpace(fmt.Sprint(rawRef)) != "" {
		ref = strings.TrimSpace(fmt.Sprint(rawRef))
		if err := validateGitRef(ref); err != nil {
			return nil, &ValidationError{Path: configPath, Message: err.Error()}
		}
		if allow, _ := sourceMap["allow_mutable_ref"].(bool); !allow {
			return nil, &ValidationError{Path: configPath, Message: "mutable git refs require allow_mutable_ref: true"}
		}
		posture = VerificationPostureUnverifiedMutableRef
		diagnostics = append(diagnostics, ResolutionDiagnostic{
			Severity: DiagnosticSeverityWarning,
			Code:     "mutable_ref",
			Message:  "mutable git refs are unverified and may resolve differently over time",
		})
		tree, commit, err = resolver.materializeRef(ref, subdir)
	} else if _, ok := sourceMap["signed_tag"]; ok {
		return nil, &ValidationError{Path: configPath, Message: "signed tag resolution is not implemented in alpha.2"}
	} else {
		return nil, &ValidationError{Path: configPath, Message: "git source must declare commit, signed_tag, or ref"}
	}
	if err != nil {
		return nil, err
	}
	base.SourceRoot = subdir
	base.SourceMode = SourceModeGit
	base.SourceRef = ref
	base.ResolvedCommit = commit
	base.VerificationPosture = posture
	base.Diagnostics = diagnostics
	base.Tree = tree
	return base, nil
}

func resolveEmbeddedSourceRoot(configPath, projectRoot string, sourceMap map[string]any) (string, string, error) {
	rawPath := strings.TrimSpace(fmt.Sprint(sourceMap["path"]))
	if rawPath == "" {
		return "", "", &ValidationError{Path: configPath, Message: "content root path must not be empty"}
	}
	declared, err := normalizeContainedRelativePath(rawPath)
	if err != nil {
		return "", "", &ValidationError{Path: configPath, Message: fmt.Sprintf("embedded source path %v", err)}
	}
	absRoot := filepath.Clean(filepath.Join(projectRoot, filepath.FromSlash(declared)))
	if !isWithinRoot(projectRoot, absRoot) {
		return "", "", &ValidationError{Path: configPath, Message: fmt.Sprintf("embedded source path %q escapes the selected project root", rawPath)}
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return "", "", &ValidationError{Path: configPath, Message: err.Error()}
	}
	if !info.IsDir() {
		return "", "", &ValidationError{Path: configPath, Message: fmt.Sprintf("resolved source root %q is not a directory", absRoot)}
	}
	return declared, absRoot, nil
}

func resolveDeclaredLocalSourceRoot(configPath, projectRoot string, sourceMap map[string]any) (string, string, error) {
	rawPath := strings.TrimSpace(fmt.Sprint(sourceMap["path"]))
	if rawPath == "" {
		return "", "", &ValidationError{Path: configPath, Message: "content root path must not be empty"}
	}
	declared := cleanSourceRootValue(rawPath)
	absRoot := declared
	if !filepath.IsAbs(rawPath) {
		absRoot = filepath.Join(projectRoot, rawPath)
	}
	absRoot = filepath.Clean(absRoot)
	info, err := os.Stat(absRoot)
	if err != nil {
		return "", "", &ValidationError{Path: configPath, Message: err.Error()}
	}
	if !info.IsDir() {
		return "", "", &ValidationError{Path: configPath, Message: fmt.Sprintf("resolved source root %q is not a directory", absRoot)}
	}
	return declared, absRoot, nil
}

func normalizeContainedRelativePath(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("must not be empty")
	}
	if filepath.IsAbs(value) {
		return "", fmt.Errorf("must not be absolute")
	}
	cleaned := filepath.Clean(value)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("must not escape its containing root")
	}
	return filepath.ToSlash(cleaned), nil
}

func validateGitURL(url string) error {
	if strings.HasPrefix(url, "-") {
		return fmt.Errorf("git source url must not start with '-'")
	}
	if strings.ContainsAny(url, gitURLControlChars+" ") {
		return fmt.Errorf("git source url contains unsupported whitespace or control characters")
	}
	return nil
}

func validateGitCommit(commit string) error {
	if strings.HasPrefix(commit, "-") {
		return fmt.Errorf("git commit must not start with '-'")
	}
	if !gitCommitPattern.MatchString(commit) {
		return fmt.Errorf("git commit must be a 40-character lowercase hex SHA")
	}
	return nil
}

func validateGitRef(ref string) error {
	if strings.HasPrefix(ref, "-") {
		return fmt.Errorf("git ref must not start with '-'")
	}
	if !gitRefPattern.MatchString(ref) {
		return fmt.Errorf("git ref contains unsupported characters")
	}
	if strings.Contains(ref, "..") {
		return fmt.Errorf("git ref must not contain '..'")
	}
	if strings.Contains(ref, "//") {
		return fmt.Errorf("git ref must not contain consecutive '/'")
	}
	if strings.HasPrefix(ref, "/") || strings.HasSuffix(ref, "/") {
		return fmt.Errorf("git ref must not start or end with '/'")
	}
	if strings.HasSuffix(ref, ".lock") {
		return fmt.Errorf("git ref must not end with '.lock'")
	}
	for _, segment := range strings.Split(ref, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("git ref contains an invalid path segment")
		}
		if strings.HasPrefix(segment, ".") {
			return fmt.Errorf("git ref segments must not start with '.'")
		}
	}
	return nil
}

func cleanSourceRootValue(value string) string {
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.ToSlash(filepath.Clean(value))
}

func snapshotLocalTree(sourceRoot string) (*LocalSourceTree, error) {
	realRoot, err := filepath.EvalSymlinks(sourceRoot)
	if err != nil {
		return nil, err
	}
	tempRoot, err := os.MkdirTemp("", "runectx-local-source-")
	if err != nil {
		return nil, err
	}
	snapshotRoot := filepath.Join(tempRoot, "snapshot")
	if err := copyResolvedTree(realRoot, snapshotRoot, realRoot, map[string]struct{}{}, localSnapshotLimits, &snapshotState{}, 0); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, err
	}
	return &LocalSourceTree{Root: snapshotRoot, SnapshotKind: "snapshot_copy", cleanupRoot: tempRoot}, nil
}

func copyResolvedTree(sourcePath, destPath, root string, active map[string]struct{}, limits snapshotLimits, state *snapshotState, depth int) error {
	if depth > limits.MaxDepth {
		return fmt.Errorf("local source tree exceeds maximum depth of %d", limits.MaxDepth)
	}
	resolved, err := filepath.EvalSymlinks(sourcePath)
	if err != nil {
		return err
	}
	if !isWithinRoot(root, resolved) {
		return fmt.Errorf("resolved path %q escapes declared local source tree", resolved)
	}
	if _, ok := active[resolved]; ok {
		return fmt.Errorf("symlink cycle detected at %q", resolved)
	}
	active[resolved] = struct{}{}
	defer delete(active, resolved)

	info, err := os.Stat(resolved)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if depth > 0 {
			if _, excluded := limits.Excludes[filepath.Base(resolved)]; excluded {
				return nil
			}
		}
		if err := os.MkdirAll(destPath, 0o755); err != nil {
			return err
		}
		entries, err := os.ReadDir(resolved)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			childSource := filepath.Join(resolved, entry.Name())
			childDest := filepath.Join(destPath, entry.Name())
			if err := copyResolvedTree(childSource, childDest, root, active, limits, state, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("unsupported non-regular file %q in local source tree", resolved)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	state.files++
	if state.files > limits.MaxFiles {
		return fmt.Errorf("local source tree exceeds maximum file count of %d", limits.MaxFiles)
	}
	state.bytes += info.Size()
	if state.bytes > limits.MaxBytes {
		return fmt.Errorf("local source tree exceeds maximum snapshot size of %d bytes", limits.MaxBytes)
	}
	src, err := os.Open(resolved)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	closeErr := dst.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func isWithinRoot(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

type gitResolver struct {
	configPath string
	url        string
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

func (r gitResolver) materializeCommitToTree(commit, subdir string) (*LocalSourceTree, string, error) {
	tempRoot, repoRoot, err := r.initializeRepository()
	if err != nil {
		return nil, "", err
	}
	if err := runGit("-C", repoRoot, "fetch", "--quiet", "--no-tags", "origin", "+refs/heads/*:refs/remotes/origin/*", "+refs/tags/*:refs/tags/*"); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	if err := runGit("-C", repoRoot, "cat-file", "-e", commit+"^{commit}"); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: fmt.Sprintf("pinned git commit %q was not found after fetching advertised refs: %v", commit, err)}
	}
	if err := runGit("-C", repoRoot, "checkout", "--quiet", "--detach", commit); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	return r.finalizeMaterializedTree(tempRoot, repoRoot, subdir)
}

func (r gitResolver) initializeRepository() (string, string, error) {
	tempRoot, err := os.MkdirTemp("", "runectx-git-source-")
	if err != nil {
		return "", "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	repoRoot := filepath.Join(tempRoot, "repo")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		_ = os.RemoveAll(tempRoot)
		return "", "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	for _, args := range [][]string{{"init", "--quiet", repoRoot}, {"-C", repoRoot, "remote", "add", "origin", r.url}} {
		if err := runGit(args...); err != nil {
			_ = os.RemoveAll(tempRoot)
			return "", "", &ValidationError{Path: r.configPath, Message: err.Error()}
		}
	}
	return tempRoot, repoRoot, nil
}

func (r gitResolver) finalizeMaterializedTree(tempRoot, repoRoot, subdir string) (*LocalSourceTree, string, error) {
	commitOutput, err := gitOutput("-C", repoRoot, "rev-parse", "HEAD")
	if err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	commit := strings.TrimSpace(commitOutput)
	materializedRoot := filepath.Clean(filepath.Join(repoRoot, filepath.FromSlash(subdir)))
	if !isWithinRoot(repoRoot, materializedRoot) {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: fmt.Sprintf("git source subdir %q escapes the fetched repository root", subdir)}
	}
	info, err := os.Stat(materializedRoot)
	if err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	if !info.IsDir() {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: fmt.Sprintf("resolved git source root %q is not a directory", materializedRoot)}
	}
	return &LocalSourceTree{Root: materializedRoot, SnapshotKind: "git_checkout", cleanupRoot: tempRoot}, commit, nil
}

func runGit(args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = sanitizedGitEnv()
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("git %s: command timed out after %s", strings.Join(args, " "), gitCommandTimeout)
	}
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
	}
	return nil
}

func gitOutput(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = sanitizedGitEnv()
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("git %s: command timed out after %s", strings.Join(args, " "), gitCommandTimeout)
	}
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
	}
	return string(output), nil
}

func sanitizedGitEnv() []string {
	env := []string{
		"GIT_CONFIG_GLOBAL=" + os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
		"SSH_ASKPASS=",
		"SSH_AUTH_SOCK=",
		"GIT_SSH=",
		"GIT_SSH_COMMAND=",
		"GCM_INTERACTIVE=Never",
		"LANG=C",
		"LC_ALL=C",
	}
	for _, key := range []string{"PATH", "TMPDIR", "TMP", "TEMP", "SYSTEMROOT"} {
		if value, ok := os.LookupEnv(key); ok && value != "" {
			env = append(env, key+"="+value)
		}
	}
	return env
}

func runeContextRelativePath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

func walkRuneContextFiles(root string, visit func(path string, d fs.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return visit(path, d)
	})
}
