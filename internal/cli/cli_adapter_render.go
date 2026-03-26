package cli

import (
	"fmt"
	"io"
	"strings"
)

const adapterRenderCommand = "adapter_render_host_native"

type adapterRenderRequest struct {
	tool      string
	operation string
	role      string
}

func runAdapterRenderHostNative(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	if len(args) == 1 && isHelpToken(args[0]) {
		emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", adapterRenderCommand}, {"usage", adapterRenderUsage}}, machine), exitOK, failureClassNone)
		return exitOK
	}
	if len(args) > 1 && isHelpToken(args[0]) {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines(adapterRenderCommand, adapterRenderUsage, fmt.Errorf("help does not accept additional arguments")), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	request, err := parseAdapterRenderArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines(adapterRenderCommand, adapterRenderUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if !supportsShellInjection(request.tool) {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(adapterRenderCommand, "", fmt.Errorf("adapter %q does not support shell-output injection render mode", request.tool)), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	rendered, err := renderHostNativeOperationMarkdown(request)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(adapterRenderCommand, "", err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	if machine.jsonOutput {
		emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", adapterRenderCommand}, {"adapter", request.tool}, {"operation", request.operation}, {"role", request.role}, {"body", strings.TrimSpace(rendered)}}, machine), exitOK, failureClassNone)
		return exitOK
	}
	if _, err := io.WriteString(stdout, rendered); err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(adapterRenderCommand, "", err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	return exitOK
}

func parseAdapterRenderArgs(args []string) (adapterRenderRequest, error) {
	request := adapterRenderRequest{role: hostNativeKindFlowAsset}
	positionals := make([]string, 0, 2)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		return parseAdapterRenderFlag(args, flag, &request)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return adapterRenderRequest{}, err
	}
	tool, operation, err := parseAdapterRenderPositionals(positionals)
	if err != nil {
		return adapterRenderRequest{}, err
	}
	request.tool = tool
	request.operation = operation
	if request.role != hostNativeKindFlowAsset && request.role != hostNativeKindDiscoverabilityShim {
		return adapterRenderRequest{}, fmt.Errorf("adapter render-host-native role %q is invalid", request.role)
	}
	return request, nil
}

func parseAdapterRenderFlag(args []string, flag parsedFlag, request *adapterRenderRequest) (int, error) {
	if flag.name != "--role" {
		return flag.next, fmt.Errorf("unknown adapter render-host-native flag %q", flag.raw)
	}
	value, next, err := flag.requireValue(args)
	if err != nil {
		return flag.next, err
	}
	request.role = normalizeHostNativeRole(strings.TrimSpace(value))
	return next, nil
}

func parseAdapterRenderPositionals(positionals []string) (string, string, error) {
	if len(positionals) != 2 {
		return "", "", fmt.Errorf("adapter render-host-native requires <tool> <operation>")
	}
	assignReq := adapterRequest{}
	if err := assignAdapterTool(&assignReq, positionals[0]); err != nil {
		return "", "", err
	}
	operation := strings.TrimSpace(positionals[1])
	if operation == "" {
		return "", "", fmt.Errorf("adapter render-host-native operation must not be empty")
	}
	if operation != sanitizeHostNativeOperation(operation) {
		return "", "", fmt.Errorf("adapter render-host-native operation %q is invalid", operation)
	}
	return assignReq.tool, operation, nil
}

func normalizeHostNativeRole(role string) string {
	role = strings.ToLower(strings.TrimSpace(role))
	role = strings.ReplaceAll(role, "-", "_")
	return role
}
