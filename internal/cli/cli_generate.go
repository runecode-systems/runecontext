package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

const generateIndexesCommand = "generate indexes"

type generateIndexesRequest struct {
	root         string
	explicitRoot bool
}

func runGenerate(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("generate", generateUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if len(remaining) == 0 {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("generate", generateUsage, fmt.Errorf("generate subcommand is required")), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if isHelpToken(remaining[0]) {
		if len(remaining) != 1 {
			emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("generate", generateUsage, fmt.Errorf("help does not accept additional arguments")), machine), exitUsage, failureClassUsage)
			return exitUsage
		}
		emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", "generate"}, {"usage", generateUsage}}, machine), exitOK, failureClassNone)
		return exitOK
	}
	switch remaining[0] {
	case "indexes":
		return runGenerateIndexes(remaining[1:], machine, stdout, stderr)
	default:
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("generate", generateUsage, fmt.Errorf("unknown generate subcommand %q", remaining[0])), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
}

func runGenerateIndexes(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	request, err := parseGenerateIndexesArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines(generateIndexesCommand, generateIndexesUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, generateIndexesCommand, machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	index, err := project.validator.ValidateLoadedProjectAllowMalformedGeneratedIndexes(project.loaded)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(generateIndexesCommand, project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	defer index.Close()
	if err := index.WriteGeneratedIndexes(); err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(generateIndexesCommand, project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	changedFiles, err := generatedIndexesMutations(index.ContentRoot)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(generateIndexesCommand, project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildGenerateIndexesOutput(project.absRoot, project.loaded, changedFiles)
	if machine.explain {
		output = appendGenerateIndexesExplainLines(output)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func parseGenerateIndexesArgs(args []string) (generateIndexesRequest, error) {
	request := generateIndexesRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		if flag.name != "--path" {
			return flag.next, fmt.Errorf("unknown generate indexes flag %q", flag.raw)
		}
		return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return generateIndexesRequest{}, err
	}
	if len(positionals) > 1 {
		return generateIndexesRequest{}, fmt.Errorf("expected at most one path argument")
	}
	if len(positionals) == 1 {
		if request.explicitRoot {
			return generateIndexesRequest{}, fmt.Errorf("cannot use both --path and a positional path argument")
		}
		request.root = positionals[0]
		request.explicitRoot = true
	}
	return request, nil
}

func buildGenerateIndexesOutput(absRoot string, loaded *contracts.LoadedProject, changed []contracts.FileMutation) []line {
	output := []line{
		{"result", "ok"},
		{"command", generateIndexesCommand},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
	}
	if loaded != nil && loaded.Resolution != nil {
		output = append(output,
			line{"project_root", loaded.Resolution.ProjectRoot},
			line{"source_root", loaded.Resolution.SourceRoot},
			line{"source_mode", string(loaded.Resolution.SourceMode)},
		)
	}
	return appendChangedFiles(output, changed)
}

func appendGenerateIndexesExplainLines(lines []line) []line {
	return append(lines,
		line{"explain_scope", "generated-indexes"},
		line{"explain_generated_artifact_kind", "manifest-and-indexes"},
		line{"explain_generated_artifact_count", "3"},
	)
}

func generatedIndexesMutations(contentRoot string) ([]contracts.FileMutation, error) {
	manifestPath, err := generatedIndexPathMutation(contentRoot, "manifest.yaml")
	if err != nil {
		return nil, err
	}
	changesPath, err := generatedIndexPathMutation(contentRoot, filepath.Join("indexes", "changes-by-status.yaml"))
	if err != nil {
		return nil, err
	}
	bundlesPath, err := generatedIndexPathMutation(contentRoot, filepath.Join("indexes", "bundles.yaml"))
	if err != nil {
		return nil, err
	}
	return []contracts.FileMutation{
		{Path: changesPath, Action: "created_or_updated"},
		{Path: bundlesPath, Action: "created_or_updated"},
		{Path: manifestPath, Action: "created_or_updated"},
	}, nil
}

func generatedIndexPathMutation(contentRoot, suffix string) (string, error) {
	if contentRoot == "" {
		return "", fmt.Errorf("project content root is unavailable")
	}
	absolute := filepath.Join(contentRoot, suffix)
	relative, err := filepath.Rel(contentRoot, absolute)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(relative), nil
}
