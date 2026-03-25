package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

const (
	suggestionProviderChangeIDs        = "change-ids"
	suggestionProviderBundleIDs        = "bundle-ids"
	suggestionProviderPromotionTargets = "promotion-targets"
	suggestionProviderAdapterNames     = "adapter-names"
)

type completionSuggestRequest struct {
	root         string
	explicitRoot bool
	provider     string
	prefix       string
}

func runCompletionSuggest(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && isHelpToken(args[0]) {
		writeLines(stdout,
			line{"result", "ok"},
			line{"command", "completion suggest"},
			line{"usage", completionSuggestUsage},
		)
		return exitOK
	}
	if len(args) > 1 && isHelpToken(args[0]) {
		writeCommandUsageError(stderr, "completion suggest", completionSuggestUsage, fmt.Errorf("help does not accept additional arguments"))
		return exitUsage
	}
	request, err := parseCompletionSuggestArgs(args)
	if err != nil {
		writeCommandUsageError(stderr, "completion suggest", completionSuggestUsage, err)
		return exitUsage
	}
	suggestions, err := collectCompletionSuggestions(request)
	if err != nil {
		writeCommandInvalid(stderr, "completion suggest", "", err)
		return exitInvalid
	}
	if len(suggestions) == 0 {
		return exitOK
	}
	if _, err := io.WriteString(stdout, strings.Join(suggestions, "\n")+"\n"); err != nil {
		writeCommandInvalid(stderr, "completion suggest", "", err)
		return exitInvalid
	}
	return exitOK
}

func parseCompletionSuggestArgs(args []string) (completionSuggestRequest, error) {
	request := completionSuggestRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		case "--prefix":
			return assignStringFlag(args, flag, &request.prefix)
		default:
			return flag.next, fmt.Errorf("unknown completion suggest flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return completionSuggestRequest{}, err
	}
	provider, err := requireExactPositional(positionals, "completion suggest requires exactly one provider")
	if err != nil {
		return completionSuggestRequest{}, err
	}
	request.provider = provider
	if !isKnownSuggestionProvider(provider) {
		return completionSuggestRequest{}, fmt.Errorf("unknown suggestion provider %q", provider)
	}
	return request, nil
}

func isKnownSuggestionProvider(provider string) bool {
	switch provider {
	case suggestionProviderChangeIDs, suggestionProviderBundleIDs, suggestionProviderPromotionTargets, suggestionProviderAdapterNames:
		return true
	default:
		return false
	}
}

func collectCompletionSuggestions(request completionSuggestRequest) ([]string, error) {
	suggestions, err := rawCompletionSuggestions(request)
	if err != nil {
		return nil, err
	}
	return filterSuggestionPrefix(suggestions, request.prefix), nil
}

func rawCompletionSuggestions(request completionSuggestRequest) ([]string, error) {
	switch request.provider {
	case suggestionProviderAdapterNames:
		return adapterNameSuggestions(request)
	case suggestionProviderChangeIDs:
		index, ok, err := loadSuggestionProjectIndex(request)
		if err != nil || !ok {
			return nil, err
		}
		defer index.Close()
		return contracts.SortedKeys(index.ChangeIDs), nil
	case suggestionProviderBundleIDs:
		index, ok, err := loadSuggestionProjectIndex(request)
		if err != nil || !ok {
			return nil, err
		}
		defer index.Close()
		return index.BundleIDs(), nil
	case suggestionProviderPromotionTargets:
		index, ok, err := loadSuggestionProjectIndex(request)
		if err != nil || !ok {
			return nil, err
		}
		defer index.Close()
		return promotionTargetSuggestions(index), nil
	default:
		return nil, fmt.Errorf("unknown suggestion provider %q", request.provider)
	}
}

func loadSuggestionProjectIndex(request completionSuggestRequest) (*contracts.ProjectIndex, bool, error) {
	project, code := loadProjectOrReport(request.root, request.explicitRoot, io.Discard, "completion suggest", machineOptions{})
	if project == nil {
		if code != exitOK && request.explicitRoot {
			return nil, false, fmt.Errorf("failed to load project at %q", request.root)
		}
		return nil, false, nil
	}
	defer project.close()
	index, err := project.validator.ValidateLoadedProject(project.loaded)
	if err != nil {
		return nil, false, err
	}
	return index, true, nil
}

func filterSuggestionPrefix(items []string, prefix string) []string {
	if prefix == "" {
		return items
	}
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		if strings.HasPrefix(item, prefix) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func adapterNameSuggestions(request completionSuggestRequest) ([]string, error) {
	adaptersRoot, err := locateAdaptersRoot()
	if err != nil {
		return handleAdapterSuggestionRootError(request, err)
	}
	entries, err := os.ReadDir(adaptersRoot)
	if err != nil {
		return handleAdapterSuggestionReadError(request, err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

func handleAdapterSuggestionRootError(request completionSuggestRequest, err error) ([]string, error) {
	if request.explicitRoot {
		return nil, err
	}
	if os.IsNotExist(err) {
		return nil, nil
	}
	return nil, nil
}

func handleAdapterSuggestionReadError(request completionSuggestRequest, err error) ([]string, error) {
	if !os.IsNotExist(err) {
		return nil, err
	}
	if request.explicitRoot {
		return nil, fmt.Errorf("failed to load adapter packs for %q: %w", request.root, err)
	}
	return nil, nil
}

func promotionTargetSuggestions(index *contracts.ProjectIndex) []string {
	if index == nil {
		return nil
	}
	items := make([]string, 0, len(index.Specs)+len(index.Standards)+len(index.Decisions))
	for _, path := range contracts.SortedKeys(index.Specs) {
		items = append(items, "spec:"+path)
	}
	for _, path := range contracts.SortedKeys(index.Standards) {
		items = append(items, "standard:"+path)
	}
	for _, path := range contracts.SortedKeys(index.Decisions) {
		items = append(items, "decision:"+path)
	}
	return items
}
