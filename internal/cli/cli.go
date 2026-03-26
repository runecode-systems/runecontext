package cli

import (
	"fmt"
	"io"
)

const (
	exitOK      = 0
	exitInvalid = 1
	exitUsage   = 2
)

const (
	validateUsage           = "runectx validate [--json] [--non-interactive] [--explain] [--ssh-allowed-signers PATH] [--path PATH] [path]"
	statusUsage             = "runectx status [--json] [--non-interactive] [--explain] [--path PATH] [path]"
	changeUsage             = "runectx change [--json] [--non-interactive] [--dry-run] [--explain] <new|shape|close|reallocate> ..."
	generateUsage           = "runectx generate [--json] [--non-interactive] [--explain] <indexes>"
	generateIndexesUsage    = "runectx generate indexes [--json] [--non-interactive] [--explain] [--path PATH] [path]"
	bundleUsage             = "runectx bundle [--json] [--non-interactive] [--explain] <resolve>"
	bundleResolveUsage      = "runectx bundle resolve [--json] [--non-interactive] [--explain] [--path PATH] <bundle-id>..."
	doctorUsage             = "runectx doctor [--json] [--non-interactive] [--explain] [--path PATH] [path]"
	changeNewUsage          = "runectx change new [--json] [--non-interactive] [--dry-run] [--explain] --title TITLE --type TYPE [--size SIZE] [--bundle ID] [--shape minimum|full] [--description TEXT] [--path PATH]"
	changeShapeUsage        = "runectx change shape [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID [--design TEXT] [--verification TEXT] [--task TEXT] [--reference TEXT] [--path PATH]"
	changeCloseUsage        = "runectx change close [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID [--verification-status STATUS] [--superseded-by ID] [--closed-at YYYY-MM-DD] [--path PATH]"
	changeReallocateUsage   = "runectx change reallocate [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID [--path PATH]"
	initUsage               = "runectx init [--json] [--non-interactive] [--dry-run] [--explain] [--mode embedded|linked] [--seed-bundle NAME] [--path PATH]"
	promoteUsage            = "runectx promote [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID [--accept | --complete] [--target TYPE:PATH (summary auto-filled per target type)] [--path PATH]"
	standardUsage           = "runectx standard [--json] [--non-interactive] [--explain] <discover>"
	standardDiscoverUsage   = "runectx standard discover [--json] [--non-interactive] [--explain] [--path PATH] [--change CHANGE_ID] [--scope-path PATH] [--focus TEXT] [--confirm-handoff] [--target TYPE:PATH]"
	assuranceUsage          = "runectx assurance [--json] [--non-interactive] [--dry-run] [--explain] <enable|backfill|capture> ..."
	adapterUsage            = "runectx adapter [--json] [--non-interactive] [--dry-run] [--explain] <sync|render-host-native> ..."
	adapterSyncUsage        = "runectx adapter sync [--json] [--non-interactive] [--dry-run] [--explain] [--path PATH] <tool>"
	adapterRenderUsage      = "runectx adapter render-host-native [--json] [--non-interactive] [--dry-run] [--explain] [--role flow_asset|discoverability_shim] <tool> <operation>"
	completionUsage         = "runectx completion <bash|zsh|fish|suggest|metadata>"
	completionSuggestUsage  = "runectx completion suggest [--path PATH] [--prefix PREFIX] <change-ids|bundle-ids|promotion-targets|adapter-names|adapter-names-shell-injection>"
	completionMetadataUsage = "runectx completion metadata"
	versionUsage            = "runectx version [--json] [--non-interactive] (aliases: --version, -v)"
)

func Run(args []string, stdout, stderr io.Writer) int {
	command, remaining, ok := parseRootCommand(args)
	if !ok {
		printUsage(stdout)
		return exitOK
	}
	if isRootHelpCommand(command) {
		printUsage(stdout)
		return exitOK
	}
	if handler, found := rootCommandHandler(command); found {
		return handler(remaining, stdout, stderr)
	}
	writeCommandUsageError(stderr, command, "runectx help", fmt.Errorf("unknown command %q", command))
	return exitUsage
}

func printUsage(w io.Writer) {
	printUsageHeader(w)
	printUsageCommands(w)
	printUsageExamples(w)
}

type rootCommandFunc func([]string, io.Writer, io.Writer) int

func parseRootCommand(args []string) (string, []string, bool) {
	if len(args) == 0 {
		return "", nil, false
	}
	return args[0], args[1:], true
}

func isRootHelpCommand(command string) bool {
	return isHelpToken(command)
}

func rootCommandHandler(command string) (rootCommandFunc, bool) {
	handlers := map[string]rootCommandFunc{
		"validate":   runValidate,
		"status":     runStatus,
		"change":     runChange,
		"generate":   runGenerate,
		"bundle":     runBundle,
		"doctor":     runDoctor,
		"init":       runInit,
		"promote":    runPromote,
		"standard":   runStandard,
		"assurance":  runAssurance,
		"adapter":    runAdapter,
		"completion": runCompletion,
		"version":    runVersion,
		"--version":  runVersion,
		"-v":         runVersion,
	}
	handler, ok := handlers[command]
	return handler, ok
}

func printUsageHeader(w io.Writer) {
	fmt.Fprintln(w, "RuneContext CLI")
	fmt.Fprintln(w)
}

func printUsageCommands(w io.Writer) {
	fmt.Fprintln(w, "Commands:")
	for _, entry := range usageCommandDescriptions() {
		fmt.Fprintln(w, entry)
	}
	fmt.Fprintln(w)
}

func printUsageExamples(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	for _, usage := range usageExamples() {
		fmt.Fprintln(w, "  "+usage)
	}
}

func usageCommandDescriptions() []string {
	return []string{
		"  help       Show CLI usage",
		"  status     Report active, closed, and superseded changes",
		"  change     Create, shape, close, and reallocate changes",
		"  generate   Write optional generated indexes and manifest",
		"  bundle     Resolve context bundles",
		"  validate   Validate RuneContext contracts for a project root",
		"  doctor     Run environment and resolution diagnostics",
		"  init       Scaffold a RuneContext project",
		"  promote    Explicitly advance promotion assessment state (summary auto-filled for --target entries)",
		"  standard   Discover advisory standards candidates for promotion handoff",
		"  assurance  Enable, backfill, or capture Verified assurance artifacts",
		"  adapter    Sync and render tool host-native adapter artifacts",
		"  completion Emit shell completion scripts",
		"  version    Show RuneContext CLI version (--version, -v)",
	}
}

func usageExamples() []string {
	return []string{
		"runectx help",
		statusUsage,
		changeNewUsage,
		changeShapeUsage,
		changeCloseUsage,
		changeReallocateUsage,
		generateIndexesUsage,
		validateUsage,
		bundleResolveUsage,
		doctorUsage,
		initUsage,
		promoteUsage,
		standardDiscoverUsage,
		assuranceUsage,
		adapterSyncUsage,
		adapterRenderUsage,
		completionUsage,
		completionSuggestUsage,
		completionMetadataUsage,
		versionUsage,
	}
}

func isHelpToken(arg string) bool {
	switch arg {
	case "help", "--help", "-h":
		return true
	default:
		return false
	}
}
