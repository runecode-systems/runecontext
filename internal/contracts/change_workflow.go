package contracts

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type LifecycleStatus string

const (
	StatusProposed    LifecycleStatus = "proposed"
	StatusPlanned     LifecycleStatus = "planned"
	StatusImplemented LifecycleStatus = "implemented"
	StatusVerified    LifecycleStatus = "verified"
	StatusClosed      LifecycleStatus = "closed"
	StatusSuperseded  LifecycleStatus = "superseded"
)

var lifecycleOrder = map[LifecycleStatus]int{
	StatusProposed:    0,
	StatusPlanned:     1,
	StatusImplemented: 2,
	StatusVerified:    3,
	StatusClosed:      4,
	StatusSuperseded:  4,
}

type ChangeRecord struct {
	ID                  string
	DirPath             string
	StatusPath          string
	Title               string
	Status              LifecycleStatus
	Type                string
	Size                string
	VerificationStatus  string
	ClosedAt            string
	HasClosedAt         bool
	ContextBundles      []string
	StandardRefs        []string
	ApplicableStandards []string
	AddedStandards      []string
	ExcludedStandards   []string
	RelatedSpecs        []string
	RelatedDecisions    []string
	RelatedChanges      []string
	DependsOn           []string
	InformedBy          []string
	Supersedes          []string
	SupersededBy        []string
	Data                map[string]any
}

type SpecRecord struct {
	Path               string
	OriginatingChanges []string
	RevisedByChanges   []string
}

type DecisionRecord struct {
	Path               string
	OriginatingChanges []string
	RelatedChanges     []string
}

type CloseChangeOptions struct {
	ClosedAt     time.Time
	SupersededBy []string
}

type SplitChangePlan struct {
	UmbrellaID string
	SubChanges []SplitSubChange
}

type SplitSubChange struct {
	ID        string
	DependsOn []string
}

type ChangeGraphLinks struct {
	RelatedChanges []string
	DependsOn      []string
}

// AllocateChangeID returns the next year-scoped change ID candidate.
//
// Callers must still create the target change directory atomically and retry on
// EEXIST-style collisions; this helper allocates but does not reserve IDs.
func AllocateChangeID(contentRoot string, now time.Time, title string, entropy io.Reader) (string, error) {
	if entropy == nil {
		entropy = cryptorand.Reader
	}
	existing, err := existingChangeIDs(filepath.Join(contentRoot, "changes"))
	if err != nil {
		return "", err
	}
	year := now.Year()
	nextCounter := 1
	for id := range existing {
		idYear, counter, _, _, err := parseChangeID(id)
		if err != nil || idYear != year {
			continue
		}
		if counter >= nextCounter {
			nextCounter = counter + 1
		}
	}
	if nextCounter > 999 {
		return "", fmt.Errorf("cannot allocate change ID for %d: yearly counter exceeds 999", year)
	}
	slug := slugifyTitle(title)
	for attempt := 0; attempt < 32; attempt++ {
		suffix, err := randomChangeSuffix(entropy)
		if err != nil {
			return "", err
		}
		candidate := fmt.Sprintf("CHG-%04d-%03d-%s-%s", year, nextCounter, suffix, slug)
		if _, ok := existing[candidate]; !ok {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate a unique change ID after repeated suffix collisions")
}

// ValidateLifecycleTransition enforces monotonic lifecycle progress without
// requiring every intermediate state to be visited.
func ValidateLifecycleTransition(from, to string) error {
	fromStatus := LifecycleStatus(from)
	toStatus := LifecycleStatus(to)
	if _, ok := lifecycleOrder[fromStatus]; !ok {
		return fmt.Errorf("unknown lifecycle state %q", from)
	}
	if _, ok := lifecycleOrder[toStatus]; !ok {
		return fmt.Errorf("unknown lifecycle state %q", to)
	}
	if fromStatus == toStatus {
		return nil
	}
	if isTerminalLifecycleStatus(fromStatus) {
		return fmt.Errorf("cannot transition from terminal lifecycle state %q", from)
	}
	if lifecycleOrder[toStatus] < lifecycleOrder[fromStatus] {
		return fmt.Errorf("cannot transition backward from %q to %q", from, to)
	}
	return nil
}

func CloseChangeStatus(raw map[string]any, options CloseChangeOptions) (map[string]any, error) {
	if raw == nil {
		return nil, fmt.Errorf("status data is required")
	}
	statusValue, ok := raw["status"].(string)
	if !ok || strings.TrimSpace(statusValue) == "" {
		return nil, fmt.Errorf("status data must include a valid string status")
	}
	currentID, _ := raw["id"].(string)
	if err := validateSuccessorChangeIDs(options.SupersededBy, currentID); err != nil {
		return nil, err
	}
	nextStatus := string(StatusClosed)
	if len(options.SupersededBy) > 0 {
		nextStatus = string(StatusSuperseded)
	}
	if err := ValidateLifecycleTransition(statusValue, nextStatus); err != nil {
		return nil, err
	}
	closedAt := options.ClosedAt
	if closedAt.IsZero() {
		closedAt = time.Now().UTC()
	}
	updated := make(map[string]any, len(raw))
	for key, value := range raw {
		updated[key] = cloneTopLevelValue(value)
	}
	updated["closed_at"] = closedAt.Format("2006-01-02")
	if len(options.SupersededBy) > 0 {
		updated["status"] = string(StatusSuperseded)
		updated["superseded_by"] = stringSliceToAny(options.SupersededBy)
		return updated, nil
	}
	updated["status"] = string(StatusClosed)
	updated["superseded_by"] = []any{}
	return updated, nil
}

func BuildSplitChangeGraph(plan SplitChangePlan) (map[string]ChangeGraphLinks, error) {
	if plan.UmbrellaID == "" {
		return nil, fmt.Errorf("umbrella change ID is required")
	}
	if err := validateChangeIDValue(plan.UmbrellaID, "umbrella change ID"); err != nil {
		return nil, err
	}
	links := map[string]ChangeGraphLinks{
		plan.UmbrellaID: {},
	}
	seen := map[string]struct{}{plan.UmbrellaID: {}}
	subIDs := make([]string, 0, len(plan.SubChanges))
	for _, sub := range plan.SubChanges {
		if sub.ID == "" {
			return nil, fmt.Errorf("sub-change ID is required")
		}
		if err := validateChangeIDValue(sub.ID, fmt.Sprintf("sub-change ID %q", sub.ID)); err != nil {
			return nil, err
		}
		if sub.ID == plan.UmbrellaID {
			return nil, fmt.Errorf("sub-change ID %q must differ from umbrella change ID", sub.ID)
		}
		if _, ok := seen[sub.ID]; ok {
			return nil, fmt.Errorf("duplicate split-change ID %q", sub.ID)
		}
		seen[sub.ID] = struct{}{}
		subIDs = append(subIDs, sub.ID)
	}
	for _, sub := range plan.SubChanges {
		related := make([]string, 0, len(subIDs))
		related = append(related, plan.UmbrellaID)
		for _, otherID := range subIDs {
			if otherID == sub.ID {
				continue
			}
			related = append(related, otherID)
		}
		dependsOn := append([]string(nil), sub.DependsOn...)
		for _, depID := range dependsOn {
			if depID == sub.ID {
				return nil, fmt.Errorf("sub-change %q must not depend on itself", sub.ID)
			}
			if err := validateChangeIDValue(depID, fmt.Sprintf("depends_on entry %q", depID)); err != nil {
				return nil, err
			}
		}
		links[sub.ID] = ChangeGraphLinks{
			RelatedChanges: uniqueSortedStrings(related),
			DependsOn:      uniqueSortedStrings(dependsOn),
		}
	}
	umbrellaRelated := make([]string, 0, len(subIDs))
	umbrellaRelated = append(umbrellaRelated, subIDs...)
	links[plan.UmbrellaID] = ChangeGraphLinks{RelatedChanges: uniqueSortedStrings(umbrellaRelated)}
	if err := validateSplitChangeCycles(plan.UmbrellaID, plan.SubChanges); err != nil {
		return nil, err
	}
	return links, nil
}

func (p *ProjectIndex) OpenChangeIDs() []string {
	if p == nil {
		return nil
	}
	ids := make([]string, 0)
	for _, id := range SortedKeys(p.Changes) {
		if !isTerminalLifecycleStatus(p.Changes[id].Status) {
			ids = append(ids, id)
		}
	}
	return ids
}

func (p *ProjectIndex) ClosedChangeIDs() []string {
	if p == nil {
		return nil
	}
	ids := make([]string, 0)
	for _, id := range SortedKeys(p.Changes) {
		if p.Changes[id].Status == StatusClosed {
			ids = append(ids, id)
		}
	}
	return ids
}

func (p *ProjectIndex) SupersededChangeIDs() []string {
	if p == nil {
		return nil
	}
	ids := make([]string, 0)
	for _, id := range SortedKeys(p.Changes) {
		if p.Changes[id].Status == StatusSuperseded {
			ids = append(ids, id)
		}
	}
	return ids
}

func buildChangeRecord(changeDir, statusPath string, data map[string]any) (*ChangeRecord, error) {
	id := fmt.Sprint(data["id"])
	if filepath.Base(changeDir) != id {
		return nil, &ValidationError{Path: statusPath, Message: fmt.Sprintf("change folder %q must match status id %q", filepath.Base(changeDir), id)}
	}
	relatedSpecs, err := stringSliceField(statusPath, "related_specs", data["related_specs"])
	if err != nil {
		return nil, err
	}
	relatedDecisions, err := stringSliceField(statusPath, "related_decisions", data["related_decisions"])
	if err != nil {
		return nil, err
	}
	relatedChanges, err := stringSliceField(statusPath, "related_changes", data["related_changes"])
	if err != nil {
		return nil, err
	}
	dependsOn, err := stringSliceField(statusPath, "depends_on", data["depends_on"])
	if err != nil {
		return nil, err
	}
	informedBy, err := stringSliceField(statusPath, "informed_by", data["informed_by"])
	if err != nil {
		return nil, err
	}
	supersedes, err := stringSliceField(statusPath, "supersedes", data["supersedes"])
	if err != nil {
		return nil, err
	}
	supersededBy, err := stringSliceField(statusPath, "superseded_by", data["superseded_by"])
	if err != nil {
		return nil, err
	}
	for _, field := range []struct {
		name  string
		items []string
	}{
		{name: "related_specs", items: relatedSpecs},
		{name: "related_decisions", items: relatedDecisions},
		{name: "related_changes", items: relatedChanges},
		{name: "depends_on", items: dependsOn},
		{name: "informed_by", items: informedBy},
		{name: "supersedes", items: supersedes},
		{name: "superseded_by", items: supersededBy},
	} {
		if duplicate, ok := duplicateString(field.items); ok {
			return nil, &ValidationError{Path: statusPath, Message: fmt.Sprintf("%s contains duplicate value %q", field.name, duplicate)}
		}
	}
	for _, field := range []struct {
		name  string
		items []string
	}{
		{name: "related_changes", items: relatedChanges},
		{name: "depends_on", items: dependsOn},
		{name: "informed_by", items: informedBy},
		{name: "supersedes", items: supersedes},
		{name: "superseded_by", items: supersededBy},
	} {
		if containsString(field.items, id) {
			return nil, &ValidationError{Path: statusPath, Message: fmt.Sprintf("%s must not reference the change itself", field.name)}
		}
	}
	record := &ChangeRecord{
		ID:                 id,
		DirPath:            changeDir,
		StatusPath:         statusPath,
		Title:              requiredStringValue(data["title"]),
		Status:             LifecycleStatus(requiredStringValue(data["status"])),
		Type:               requiredStringValue(data["type"]),
		Size:               optionalStringValue(data["size"]),
		VerificationStatus: requiredStringValue(data["verification_status"]),
		ContextBundles:     extractStringList(data["context_bundles"]),
		RelatedSpecs:       relatedSpecs,
		RelatedDecisions:   relatedDecisions,
		RelatedChanges:     relatedChanges,
		DependsOn:          dependsOn,
		InformedBy:         informedBy,
		Supersedes:         supersedes,
		SupersededBy:       supersededBy,
		Data:               data,
	}
	if closedAt, ok := data["closed_at"]; ok && closedAt != nil {
		record.ClosedAt = optionalStringValue(closedAt)
		record.HasClosedAt = true
	}
	return record, nil
}

func buildSpecRecord(path string, doc *FrontmatterDocument) (*SpecRecord, error) {
	originating, err := stringSliceField(path, "originating_changes", doc.Frontmatter["originating_changes"])
	if err != nil {
		return nil, err
	}
	revisedBy, err := stringSliceField(path, "revised_by_changes", doc.Frontmatter["revised_by_changes"])
	if err != nil {
		return nil, err
	}
	return &SpecRecord{OriginatingChanges: originating, RevisedByChanges: revisedBy}, nil
}

func buildDecisionRecord(path string, doc *FrontmatterDocument) (*DecisionRecord, error) {
	originating, err := stringSliceField(path, "originating_changes", doc.Frontmatter["originating_changes"])
	if err != nil {
		return nil, err
	}
	related, err := stringSliceField(path, "related_changes", doc.Frontmatter["related_changes"])
	if err != nil {
		return nil, err
	}
	return &DecisionRecord{OriginatingChanges: originating, RelatedChanges: related}, nil
}

func validateChangeLifecycleConsistency(index *ProjectIndex) error {
	for _, id := range SortedKeys(index.Changes) {
		record := index.Changes[id]
		if _, ok := lifecycleOrder[record.Status]; !ok {
			return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("unknown lifecycle state %q", record.Status)}
		}
		if isTerminalLifecycleStatus(record.Status) {
			if !record.HasClosedAt {
				return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("terminal change status %q requires closed_at", record.Status)}
			}
		} else if record.HasClosedAt {
			return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("non-terminal change status %q must not set closed_at", record.Status)}
		}
		if record.Status == StatusSuperseded {
			if len(record.SupersededBy) == 0 {
				return &ValidationError{Path: record.StatusPath, Message: "superseded changes must list at least one successor in superseded_by"}
			}
		} else if len(record.SupersededBy) > 0 {
			return &ValidationError{Path: record.StatusPath, Message: "only superseded changes may set superseded_by"}
		}
		if record.Status == StatusVerified && record.VerificationStatus == "pending" {
			return &ValidationError{Path: record.StatusPath, Message: "verified changes must record a completed verification_status"}
		}
		if record.Status == StatusClosed && record.VerificationStatus == "pending" {
			return &ValidationError{Path: record.StatusPath, Message: "closed changes must not leave verification_status pending"}
		}
	}
	return nil
}

func validateRelatedChangeReciprocity(index *ProjectIndex) error {
	for _, id := range SortedKeys(index.Changes) {
		record := index.Changes[id]
		for _, relatedID := range record.RelatedChanges {
			related := index.Changes[relatedID]
			if related == nil {
				continue
			}
			if !containsString(related.RelatedChanges, record.ID) {
				return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("related_changes must be reciprocal: %q links to %q but the reverse link is missing", record.ID, relatedID)}
			}
		}
	}
	return nil
}

func validateSupersessionConsistency(index *ProjectIndex) error {
	for _, id := range SortedKeys(index.Changes) {
		record := index.Changes[id]
		for _, successorID := range record.SupersededBy {
			successor := index.Changes[successorID]
			if successor == nil {
				continue
			}
			if !containsString(successor.Supersedes, record.ID) {
				return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("superseded_by must be bidirectionally consistent: %q lists %q but the successor does not list %q in supersedes", record.ID, successorID, record.ID)}
			}
		}
		for _, supersededID := range record.Supersedes {
			superseded := index.Changes[supersededID]
			if superseded == nil {
				continue
			}
			if superseded.Status != StatusSuperseded {
				return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("supersedes references change %q, but that change is not marked superseded", supersededID)}
			}
			if !containsString(superseded.SupersededBy, record.ID) {
				return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("supersedes must be bidirectionally consistent: %q lists %q but the superseded change does not list %q in superseded_by", record.ID, supersededID, record.ID)}
			}
		}
	}
	return nil
}

func validateArtifactTraceabilityConsistency(index *ProjectIndex) error {
	for _, id := range SortedKeys(index.Changes) {
		record := index.Changes[id]
		for _, specPath := range record.RelatedSpecs {
			spec := index.Specs[specPath]
			if spec == nil {
				continue
			}
			if !containsString(spec.OriginatingChanges, record.ID) && !containsString(spec.RevisedByChanges, record.ID) {
				return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("related_specs entry %q must point to a spec that references change %q in originating_changes or revised_by_changes", specPath, record.ID)}
			}
		}
		for _, decisionPath := range record.RelatedDecisions {
			decision := index.Decisions[decisionPath]
			if decision == nil {
				continue
			}
			if !containsString(decision.OriginatingChanges, record.ID) && !containsString(decision.RelatedChanges, record.ID) {
				return &ValidationError{Path: record.StatusPath, Message: fmt.Sprintf("related_decisions entry %q must point to a decision that references change %q in originating_changes or related_changes", decisionPath, record.ID)}
			}
		}
	}
	for _, specPath := range SortedKeys(index.Specs) {
		spec := index.Specs[specPath]
		for _, changeID := range append(append([]string{}, spec.OriginatingChanges...), spec.RevisedByChanges...) {
			change := index.Changes[changeID]
			if change == nil {
				continue
			}
			if !containsString(change.RelatedSpecs, spec.Path) {
				return &ValidationError{Path: spec.Path, Message: fmt.Sprintf("spec change reference %q must be mirrored by a related_specs entry on %q", changeID, changeID)}
			}
		}
	}
	for _, decisionPath := range SortedKeys(index.Decisions) {
		decision := index.Decisions[decisionPath]
		for _, changeID := range append(append([]string{}, decision.OriginatingChanges...), decision.RelatedChanges...) {
			change := index.Changes[changeID]
			if change == nil {
				continue
			}
			if !containsString(change.RelatedDecisions, decision.Path) {
				return &ValidationError{Path: decision.Path, Message: fmt.Sprintf("decision change reference %q must be mirrored by a related_decisions entry on %q", changeID, changeID)}
			}
		}
	}
	return nil
}

func stringSliceField(path, field string, raw any) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, &ValidationError{Path: path, Message: fmt.Sprintf("%s must be an array", field)}
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, filepath.ToSlash(fmt.Sprint(item)))
	}
	return result, nil
}

func isTerminalLifecycleStatus(status LifecycleStatus) bool {
	return status == StatusClosed || status == StatusSuperseded
}

func existingChangeIDs(changesRoot string) (map[string]struct{}, error) {
	ids := map[string]struct{}{}
	entries, err := os.ReadDir(changesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return ids, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !changeIDPattern.MatchString(name) {
			continue
		}
		ids[name] = struct{}{}
	}
	return ids, nil
}

func parseChangeID(id string) (int, int, string, string, error) {
	if !changeIDPattern.MatchString(id) {
		return 0, 0, "", "", fmt.Errorf("invalid change ID %q", id)
	}
	parts := strings.SplitN(id, "-", 5)
	if len(parts) != 5 {
		return 0, 0, "", "", fmt.Errorf("invalid change ID %q", id)
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, "", "", err
	}
	counter, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, "", "", err
	}
	return year, counter, parts[3], parts[4], nil
}

func randomChangeSuffix(entropy io.Reader) (string, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(entropy, buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func slugifyTitle(title string) string {
	return slugifyASCII(title, "change")
}

func slugifyASCII(value, fallback string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return fallback
	}
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == '-' || unicode.IsSpace(r) || r == '_' || r == '/':
			if b.Len() == 0 || lastDash {
				continue
			}
			b.WriteByte('-')
			lastDash = true
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return fallback
	}
	return slug
}

func duplicateString(items []string) (string, bool) {
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			return item, true
		}
		seen[item] = struct{}{}
	}
	return "", false
}

func uniqueSortedStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	slices.Sort(result)
	return result
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func stringSliceToAny(items []string) []any {
	result := make([]any, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	return result
}

func requiredStringValue(raw any) string {
	value, _ := raw.(string)
	return value
}

func optionalStringValue(raw any) string {
	value, _ := raw.(string)
	return strings.TrimSpace(value)
}

func cloneTopLevelValue(value any) any {
	if value == nil {
		return nil
	}
	cloned := cloneReflectValue(reflect.ValueOf(value))
	if !cloned.IsValid() {
		return nil
	}
	return cloned.Interface()
}

func cloneReflectValue(value reflect.Value) reflect.Value {
	if !value.IsValid() {
		return value
	}
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		cloned := cloneReflectValue(value.Elem())
		wrapped := reflect.New(cloned.Type()).Elem()
		wrapped.Set(cloned)
		return wrapped
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := 0; i < value.Len(); i++ {
			result.Index(i).Set(cloneReflectValue(value.Index(i)))
		}
		return result
	case reflect.Array:
		result := reflect.New(value.Type()).Elem()
		for i := 0; i < value.Len(); i++ {
			result.Index(i).Set(cloneReflectValue(value.Index(i)))
		}
		return result
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.MakeMapWithSize(value.Type(), value.Len())
		iter := value.MapRange()
		for iter.Next() {
			result.SetMapIndex(iter.Key(), cloneReflectValue(iter.Value()))
		}
		return result
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.New(value.Elem().Type())
		result.Elem().Set(cloneReflectValue(value.Elem()))
		return result
	default:
		return value
	}
}

func validateSuccessorChangeIDs(successorIDs []string, currentID string) error {
	if len(successorIDs) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(successorIDs))
	for _, successorID := range successorIDs {
		if err := validateChangeIDValue(successorID, fmt.Sprintf("superseded_by entry %q", successorID)); err != nil {
			return err
		}
		if currentID != "" && successorID == currentID {
			return fmt.Errorf("superseded_by must not reference the change itself")
		}
		if _, ok := seen[successorID]; ok {
			return fmt.Errorf("superseded_by contains duplicate value %q", successorID)
		}
		seen[successorID] = struct{}{}
	}
	return nil
}

func validateChangeIDValue(id, label string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%s must not be empty", label)
	}
	if !changeIDPattern.MatchString(id) {
		return fmt.Errorf("%s must match the canonical change ID format", label)
	}
	return nil
}

func validateSplitChangeCycles(umbrellaID string, subChanges []SplitSubChange) error {
	state := map[string]int{}
	adjacency := map[string][]string{}
	for _, sub := range subChanges {
		for _, depID := range uniqueSortedStrings(sub.DependsOn) {
			if depID == umbrellaID {
				continue
			}
			adjacency[sub.ID] = append(adjacency[sub.ID], depID)
		}
	}
	var visit func(string) error
	visit = func(id string) error {
		switch state[id] {
		case 1:
			return fmt.Errorf("split-change dependencies contain a cycle involving %q", id)
		case 2:
			return nil
		}
		state[id] = 1
		for _, depID := range adjacency[id] {
			if _, ok := adjacency[depID]; !ok {
				continue
			}
			if err := visit(depID); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	for _, sub := range subChanges {
		if err := visit(sub.ID); err != nil {
			return err
		}
	}
	return nil
}
