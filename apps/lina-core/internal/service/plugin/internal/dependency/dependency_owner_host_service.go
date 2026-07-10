// This file extracts owner-aware host service summaries for reverse dependency
// diagnostics without making the pure resolver depend on catalog internals.

package dependency

import (
	"sort"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// OwnerHostServiceSummariesFromManifest returns owner-aware host service
// summaries from the discovered manifest state.
func OwnerHostServiceSummariesFromManifest(manifest *catalog.Manifest) []*OwnerHostServiceSummary {
	if manifest == nil {
		return nil
	}
	return ownerHostServiceSummariesFromSpecs(manifest.HostServices)
}

// ownerHostServiceSummariesFromSpecs returns deterministic summaries for
// plugin-owned host service declarations.
func ownerHostServiceSummariesFromSpecs(specs []*protocol.HostServiceSpec) []*OwnerHostServiceSummary {
	if len(specs) == 0 {
		return nil
	}
	out := make([]*OwnerHostServiceSummary, 0, len(specs))
	for _, spec := range specs {
		if spec == nil || strings.TrimSpace(spec.Owner) == "" {
			continue
		}
		methods := cloneSortedStrings(spec.Methods)
		out = append(out, &OwnerHostServiceSummary{
			Owner:   strings.TrimSpace(spec.Owner),
			Service: strings.TrimSpace(spec.Service),
			Version: strings.TrimSpace(spec.Version),
			Methods: methods,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return ownerHostServiceSummarySortKey(out[i]) < ownerHostServiceSummarySortKey(out[j])
	})
	if len(out) == 0 {
		return nil
	}
	return out
}

// ownerHostServiceSummariesForOwner filters summaries to the owner plugin whose
// lifecycle transition is being reverse-checked.
func ownerHostServiceSummariesForOwner(
	summaries []*OwnerHostServiceSummary,
	owner string,
) []*OwnerHostServiceSummary {
	owner = strings.TrimSpace(owner)
	if owner == "" || len(summaries) == 0 {
		return nil
	}
	out := make([]*OwnerHostServiceSummary, 0, len(summaries))
	for _, summary := range summaries {
		if summary == nil || strings.TrimSpace(summary.Owner) != owner {
			continue
		}
		clone := *summary
		clone.Methods = cloneSortedStrings(summary.Methods)
		out = append(out, &clone)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// cloneOwnerHostServiceSummaries returns detached summary copies.
func cloneOwnerHostServiceSummaries(items []*OwnerHostServiceSummary) []*OwnerHostServiceSummary {
	if len(items) == 0 {
		return nil
	}
	out := make([]*OwnerHostServiceSummary, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		clone := *item
		clone.Methods = cloneSortedStrings(item.Methods)
		out = append(out, &clone)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func ownerHostServiceSummarySortKey(summary *OwnerHostServiceSummary) string {
	if summary == nil {
		return ""
	}
	return strings.Join([]string{
		strings.TrimSpace(summary.Owner),
		strings.TrimSpace(summary.Service),
		strings.TrimSpace(summary.Version),
		strings.Join(cloneSortedStrings(summary.Methods), ","),
	}, "|")
}

func cloneSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	sort.Strings(out)
	if len(out) == 0 {
		return nil
	}
	return out
}
