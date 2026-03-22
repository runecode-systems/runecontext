package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func createExclusiveFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func writeSeedBundle(path, seedBundleName string) error {
	// bundle.schema.json requires `includes` to be an object with at least
	// one property. Provide a valid empty aspect map to satisfy the schema.
	data := fmt.Sprintf("schema_version: 1\nid: %q\nincludes:\n  project: []\n", seedBundleName)
	return createExclusiveFile(path, []byte(data))
}

func createInitArtifacts(state initState, machine machineOptions, stderr io.Writer) int {
	bundleCreated, code := createInitSeedBundle(state, machine, stderr)
	if code != exitOK {
		return code
	}
	return createInitConfig(state, bundleCreated, machine, stderr)
}

func createInitSeedBundle(state initState, machine machineOptions, stderr io.Writer) (bool, int) {
	if state.bundlePath == "" {
		return false, exitOK
	}
	if err := writeSeedBundle(state.bundlePath, state.seedBundleName); err != nil {
		return false, reportInitCreateError("init", state.bundlePath, err, machine, stderr)
	}
	return true, exitOK
}

func createInitConfig(state initState, bundleCreated bool, machine machineOptions, stderr io.Writer) int {
	configData := fmt.Sprintf("schema_version: 1\nrunecontext_version: %q\nassurance_tier: plain\nsource:\n  type: %s\n  path: runecontext\n", normalizedRunecontextVersion(), state.sourceType)
	if err := createExclusiveFile(state.configPath, []byte(configData)); err != nil {
		if bundleCreated {
			_ = os.Remove(state.bundlePath)
		}
		return reportInitCreateError("init", state.configPath, err, machine, stderr)
	}
	return exitOK
}

func reportInitCreateError(command, path string, err error, machine machineOptions, stderr io.Writer) int {
	if errors.Is(err, os.ErrExist) {
		if path == "" {
			err = fmt.Errorf("resource already exists")
		} else if strings.HasSuffix(path, "runecontext.yaml") {
			err = fmt.Errorf("runecontext.yaml already exists")
		} else {
			err = fmt.Errorf("bundle %s already exists", path)
		}
	}
	emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(command, path, err), machine), exitInvalid, failureClassInvalid)
	return exitInvalid
}
