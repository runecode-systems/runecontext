package cli

import (
	"fmt"
	"io"
	"strings"
)

const (
	exitOK      = 0
	exitInvalid = 1
	exitUsage   = 2
)

const (
	validateUsage           = "runectx validate [--json] [--non-interactive] [--explain] [--ssh-allowed-signers PATH] [--path PATH] [path]"
	statusUsage             = "runectx status [--json] [--non-interactive] [--explain] [--path PATH] [path] (human output only: --history recent|all|none --history-limit N --verbose)"
	changeUsage             = "runectx change [--json] [--non-interactive] [--dry-run] [--explain] <new|shape|close|reallocate|update|assess-intake|assess-decomposition> ..."
	generateUsage           = "runectx generate [--json] [--non-interactive] [--explain] <indexes>"
	generateIndexesUsage    = "runectx generate indexes [--json] [--non-interactive] [--explain] [--path PATH] [path]"
	bundleUsage             = "runectx bundle [--json] [--non-interactive] [--explain] <resolve>"
	bundleResolveUsage      = "runectx bundle resolve [--json] [--non-interactive] [--explain] [--path PATH] <bundle-id>..."
	doctorUsage             = "runectx doctor [--json] [--non-interactive] [--explain] [--path PATH] [path]"
	changeNewUsage          = "runectx change new [--json] [--non-interactive] [--dry-run] [--explain] --title TITLE --type TYPE [--size SIZE] [--bundle ID] [--shape minimum|full] [--description TEXT] [--path PATH]"
	changeShapeUsage        = "runectx change shape [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID [--design TEXT] [--verification TEXT] [--task TEXT] [--reference TEXT] [--path PATH]"
	changeCloseUsage        = "runectx change close [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID [--verification-status STATUS] [--superseded-by ID] [--closed-at YYYY-MM-DD] [--recursive] [--path PATH]"
	changeReallocateUsage   = "runectx change reallocate [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID [--path PATH]"
	changeUpdateUsage       = "runectx change update [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID --status planned|implemented|verified [--verification-status passed|failed|skipped] [--add-related-change ID] [--remove-related-change ID] [--recursive] [--path PATH]"
	changeAssessIntakeUsage = "runectx change assess-intake [--json] [--non-interactive] [--explain] --title TITLE --type TYPE [--size SIZE] [--bundle ID] [--description TEXT] [--path PATH]"
	changeAssessDecompUsage = "runectx change assess-decomposition [--json] [--non-interactive] [--explain] CHANGE_ID [--path PATH]"
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
	metadataUsage           = "runectx metadata"
	upgradeUsage            = "runectx upgrade [--json] [--non-interactive] [--explain] [--path PATH] [--target-version VERSION]"
	upgradeApplyUsage       = "runectx upgrade apply [--json] [--non-interactive] [--explain] [--path PATH] [--target-version VERSION]"
	upgradeCLIUsage         = "runectx upgrade cli [--json] [--non-interactive] [--explain] [--target-version VERSION]"
	upgradeCLIApplyUsage    = "runectx upgrade cli apply [--json] [--non-interactive] [--explain] [--target-version VERSION]"
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
	printRootHelpHeader(w)
	printRootHelpCommandGroups(w)
	printRootHelpExamples(w)
	printRootHelpFooter(w)
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
		"metadata":   runMetadata,
		"upgrade":    runUpgrade,
		"version":    runVersion,
		"--version":  runVersion,
		"-v":         runVersion,
	}
	handler, ok := handlers[command]
	return handler, ok
}

func printRootHelpHeader(w io.Writer) {
	fmt.Fprintln(w, "RuneContext CLI")
	fmt.Fprintln(w, "Portable, markdown-first project context operations")
	fmt.Fprintln(w)
}

func printRootHelpCommandGroups(w io.Writer) {
	for _, group := range rootHelpCommandGroups() {
		fmt.Fprintln(w, group.title+":")
		for _, command := range group.commands {
			fmt.Fprintln(w, rootHelpCommandLine(command.name, command.description))
		}
		fmt.Fprintln(w)
	}
}

func printRootHelpExamples(w io.Writer) {
	fmt.Fprintln(w, "Quick Start:")
	for _, example := range rootHelpExamples() {
		fmt.Fprintln(w, "  "+example)
	}
	fmt.Fprintln(w)
}

func printRootHelpFooter(w io.Writer) {
	fmt.Fprintln(w, "Use 'runectx <command> --help' for command-specific usage.")
}

type rootHelpGroup struct {
	title    string
	commands []rootHelpCommand
}

type rootHelpCommand struct {
	name        string
	description string
}

func rootHelpCommandGroups() []rootHelpGroup {
	return []rootHelpGroup{
		{
			title: "Core",
			commands: []rootHelpCommand{
				{name: "help", description: "Show this help screen"},
				{name: "version", description: "Show CLI version (--version, -v)"},
				{name: "init", description: "Scaffold a RuneContext project"},
				{name: "validate", description: "Validate RuneContext contracts"},
				{name: "status", description: "Show active, closed, and superseded changes"},
				{name: "doctor", description: "Run environment and resolution diagnostics"},
			},
		},
		{
			title: "Change Workflow",
			commands: []rootHelpCommand{
				{name: "change", description: "Create, shape, assess, update, close, and reallocate changes"},
				{name: "promote", description: "Advance promotion assessment state"},
				{name: "standard", description: "Discover advisory standards candidates"},
			},
		},
		{
			title: "Project Operations",
			commands: []rootHelpCommand{
				{name: "bundle", description: "Resolve context bundles"},
				{name: "upgrade", description: "Preview or apply version upgrades"},
				{name: "assurance", description: "Enable, backfill, or capture assurance artifacts"},
				{name: "generate", description: "Write generated indexes and manifest"},
			},
		},
		{
			title: "Tooling",
			commands: []rootHelpCommand{
				{name: "adapter", description: "Sync and render host-native adapter artifacts"},
				{name: "completion", description: "Emit completion scripts, suggestions, and metadata"},
				{name: "metadata", description: "Emit canonical machine-readable capability metadata"},
			},
		},
	}
}

func rootHelpCommandLine(name, description string) string {
	padding := 12 - len(name)
	if padding < 2 {
		padding = 2
	}
	return "  " + name + strings.Repeat(" ", padding) + description
}

func rootHelpExamples() []string {
	return []string{
		"runectx help",
		"runectx status --path /path/to/project",
		"runectx change update CHANGE_ID --status verified --verification-status passed",
		"runectx change close CHANGE_ID --verification-status passed --closed-at YYYY-MM-DD",
		"runectx completion metadata",
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
