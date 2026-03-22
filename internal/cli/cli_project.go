package cli

import (
	"io"
	"path/filepath"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type cliProject struct {
	absRoot      string
	explicitRoot bool
	validator    *contracts.Validator
	loaded       *contracts.LoadedProject
}

func (project *cliProject) close() {
	if project != nil && project.loaded != nil {
		project.loaded.Close()
	}
}

func loadProjectOrReport(root string, explicitRoot bool, stderr io.Writer, command string, machine machineOptions) (*cliProject, int) {
	absRoot, validator, loaded, err := loadProjectForCLI(root, explicitRoot)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(command, absRootOrFallback(root, absRoot), err), machine), exitInvalid, failureClassInvalid)
		return nil, exitInvalid
	}
	return &cliProject{absRoot: absRoot, explicitRoot: explicitRoot, validator: validator, loaded: loaded}, exitOK
}

func loadProjectForCLI(root string, explicitRoot bool) (string, *contracts.Validator, *contracts.LoadedProject, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", nil, nil, err
	}
	schemaRoot, err := locateSchemaRoot()
	if err != nil {
		return absRoot, nil, nil, err
	}
	validator := contracts.NewValidator(schemaRoot)
	options := contracts.ResolveOptions{
		ConfigDiscovery: contracts.ConfigDiscoveryNearestAncestor,
		ExecutionMode:   contracts.ExecutionModeLocal,
	}
	if explicitRoot {
		options.ConfigDiscovery = contracts.ConfigDiscoveryExplicitRoot
	}
	loaded, err := validator.LoadProject(absRoot, options)
	if err != nil {
		return absRoot, nil, nil, err
	}
	return absRoot, validator, loaded, nil
}

func absRootOrFallback(root, absRoot string) string {
	if absRoot != "" {
		return absRoot
	}
	if value, err := filepath.Abs(root); err == nil {
		return value
	}
	return root
}
