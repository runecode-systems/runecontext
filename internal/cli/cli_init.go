package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// runecontextVersion is overridden at build time via -ldflags -X.
var runecontextVersion = "0.0.0-dev"
var bundleIDPattern = regexp.MustCompile("^[a-z0-9]([a-z0-9-]*[a-z0-9])?$")

func normalizedRunecontextVersion() string {
	trimmed := strings.TrimPrefix(runecontextVersion, "v")
	if trimmed == "" {
		return "0.0.0-dev"
	}
	return trimmed
}

type initRequest struct {
	root         string
	explicitRoot bool
	seedBundle   string
	mode         string
}

func runInit(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowDryRun: true, allowExplain: true})
	if err != nil {
		return emitInitUsage(stderr, machine, err)
	}
	state, err := parseInitState(remaining)
	if err != nil {
		return emitInitUsage(stderr, machine, err)
	}
	if machine.dryRun {
		return runInitDryRun(state, stdout, stderr, machine)
	}
	return runInitApply(state, machine, stdout, stderr)
}

func parseInitState(args []string) (initState, error) {
	request, err := parseInitArgs(args)
	if err != nil {
		return initState{}, err
	}
	absRoot, err := resolveAbsoluteRoot(request.root)
	if err != nil {
		return initState{}, err
	}
	contentRoot := filepath.Join(absRoot, "runecontext")
	bundlesDir := filepath.Join(contentRoot, "bundles")
	changesDir := filepath.Join(contentRoot, "changes")
	configPath := filepath.Join(absRoot, "runecontext.yaml")
	effectiveMode := request.mode
	if effectiveMode == "" {
		effectiveMode = "embedded"
	}
	sourceType := "embedded"
	if effectiveMode == "linked" {
		sourceType = "path"
	}
	seedBundleName := strings.TrimSpace(request.seedBundle)
	bundlePath, err := resolveSeedBundlePath(seedBundleName, bundlesDir)
	if err != nil {
		return initState{}, err
	}
	return buildInitState(absRoot, contentRoot, bundlesDir, changesDir, configPath, effectiveMode, request.mode != "", sourceType, seedBundleName, bundlePath), nil
}

func resolveSeedBundlePath(seedBundleName, bundlesDir string) (string, error) {
	if seedBundleName == "" {
		return "", nil
	}
	if invalidSeedBundleName(seedBundleName) {
		return "", fmt.Errorf("--seed-bundle name must not contain path separators or '..' segments")
	}
	if !bundleIDPattern.MatchString(seedBundleName) {
		return "", fmt.Errorf("--seed-bundle name %q must match %s", seedBundleName, bundleIDPattern)
	}
	return filepath.Join(bundlesDir, seedBundleName+".yaml"), nil
}

func runInitApply(state initState, machine machineOptions, stdout, stderr io.Writer) int {
	if code := ensureInitDirs(state, machine, stderr); code != exitOK {
		return code
	}
	if code := createInitArtifacts(state, machine, stderr); code != exitOK {
		return code
	}
	return emitInitSuccess(stdout, stderr, machine, state)
}

type initState struct {
	absRoot        string
	bundlesDir     string
	changesDir     string
	configPath     string
	contentRoot    string
	effectiveMode  string
	modeExplicit   bool
	sourceType     string
	seedBundleName string
	bundlePath     string
	plan           []string
}

func buildInitState(absRoot, contentRoot, bundlesDir, changesDir, configPath, effectiveMode string, modeExplicit bool, sourceType, seedBundleName, bundlePath string) initState {
	state := initState{
		absRoot:        absRoot,
		contentRoot:    contentRoot,
		bundlesDir:     bundlesDir,
		changesDir:     changesDir,
		configPath:     configPath,
		effectiveMode:  effectiveMode,
		modeExplicit:   modeExplicit,
		sourceType:     sourceType,
		seedBundleName: seedBundleName,
		bundlePath:     bundlePath,
	}
	state.plan = buildInitPlan(state)
	return state
}

func buildInitPlan(state initState) []string {
	plan := []string{
		fmt.Sprintf("ensure directory %s", state.absRoot),
		fmt.Sprintf("ensure directory %s", state.contentRoot),
		fmt.Sprintf("ensure directory %s", state.bundlesDir),
		fmt.Sprintf("ensure directory %s", state.changesDir),
	}
	if state.bundlePath != "" {
		plan = append(plan, fmt.Sprintf("write bundle %s", state.bundlePath))
	}
	plan = append(plan, fmt.Sprintf("write runecontext.yaml at %s", state.configPath))
	return plan
}

func runInitDryRun(state initState, stdout, stderr io.Writer, machine machineOptions) int {
	output := []line{
		{"result", "ok"},
		{"command", "init"},
		{"root", state.absRoot},
	}
	if state.effectiveMode != "embedded" || state.modeExplicit {
		output = append(output, line{"mode", state.effectiveMode})
	}
	output = appendStringItems(output, "plan_action", state.plan)
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	if !machine.jsonOutput {
		fmt.Fprintf(stderr, "Dry run: would initialize RuneContext at %s\n", state.absRoot)
		for _, action := range state.plan {
			fmt.Fprintf(stderr, "  - %s\n", action)
		}
	}
	return exitOK
}

func ensureInitDirs(state initState, machine machineOptions, stderr io.Writer) int {
	if err := os.MkdirAll(state.absRoot, 0o755); err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("init", state.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	if err := os.MkdirAll(state.contentRoot, 0o755); err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("init", state.contentRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	if err := os.MkdirAll(state.bundlesDir, 0o755); err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("init", state.bundlesDir, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	if err := os.MkdirAll(state.changesDir, 0o755); err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("init", state.changesDir, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	return exitOK
}

func emitInitSuccess(stdout, stderr io.Writer, machine machineOptions, state initState) int {
	output := []line{
		{"result", "ok"},
		{"command", "init"},
		{"root", state.absRoot},
		{"config_path", state.configPath},
		{"bundles_dir", state.bundlesDir},
		{"changes_dir", state.changesDir},
	}
	if state.effectiveMode != "embedded" || state.modeExplicit {
		output = append(output, line{"mode", state.effectiveMode})
	}
	if state.bundlePath != "" {
		output = append(output, line{"seed_bundle_path", state.bundlePath})
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	if !machine.jsonOutput {
		fmt.Fprintf(stderr, "Initialized RuneContext at %s\n", state.absRoot)
	}
	return exitOK
}

func emitInitUsage(stderr io.Writer, machine machineOptions, err error) int {
	emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("init", initUsage, err), machine), exitUsage, failureClassUsage)
	return exitUsage
}

func parseInitArgs(args []string) (initRequest, error) {
	request := initRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--mode":
			return assignStringFlag(args, flag, &request.mode)
		case "--seed-bundle":
			return assignStringFlag(args, flag, &request.seedBundle)
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		default:
			return flag.next, fmt.Errorf("unknown init flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return initRequest{}, err
	}
	if len(positionals) > 1 {
		return initRequest{}, fmt.Errorf("expected at most one path argument")
	}
	if len(positionals) == 1 {
		if request.explicitRoot {
			return initRequest{}, fmt.Errorf("cannot use both --path and a positional path argument")
		}
		request.root = positionals[0]
		request.explicitRoot = true
	}
	if request.mode != "" && request.mode != "embedded" && request.mode != "linked" {
		return initRequest{}, fmt.Errorf("--mode must be %q or %q", "embedded", "linked")
	}
	return request, nil
}

func invalidSeedBundleName(name string) bool {
	if name == "" {
		return false
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return true
	}
	if name == ".." {
		return true
	}
	return false
}
