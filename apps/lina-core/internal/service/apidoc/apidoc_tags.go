// This file dynamically collects OpenAPI operation tags from the projected
// document and writes a stable top-level tags array for Stoplight sidebar order.
// Ordering is framework-generic: no host/plugin business tag names are hardcoded.
// Hierarchical tags (using " / " or "/") sort by parent group then leaf so related
// families stay contiguous; remaining tags sort case-insensitively by display name.

package apidoc

import (
	"sort"
	"strings"

	"github.com/gogf/gf/v2/net/goai"
)

// assignOpenAPIDocumentTags collects every unique tag currently present on
// operations and writes document.Tags in deterministic display order. Call this
// after localization so sidebar titles match the request locale. Stoplight uses
// the top-level tags array for sidebar group order when present.
func assignOpenAPIDocumentTags(document *goai.OpenApiV3) {
	if document == nil {
		return
	}
	orderedNames := orderOpenAPIDocumentTags(collectOpenAPIOperationTags(document.Paths))
	if len(orderedNames) == 0 {
		document.Tags = nil
		return
	}
	ordered := make(goai.Tags, 0, len(orderedNames))
	for _, name := range orderedNames {
		ordered = append(ordered, goai.Tag{Name: name})
	}
	document.Tags = &ordered
}

// collectOpenAPIOperationTags returns the unique, non-empty tag names found on
// all projected operations in the OpenAPI paths map.
func collectOpenAPIOperationTags(paths goai.Paths) []string {
	if len(paths) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	for _, pathItem := range paths {
		collectOpenAPIOperationTagsFromOperation(seen, pathItem.Connect)
		collectOpenAPIOperationTagsFromOperation(seen, pathItem.Delete)
		collectOpenAPIOperationTagsFromOperation(seen, pathItem.Get)
		collectOpenAPIOperationTagsFromOperation(seen, pathItem.Head)
		collectOpenAPIOperationTagsFromOperation(seen, pathItem.Options)
		collectOpenAPIOperationTagsFromOperation(seen, pathItem.Patch)
		collectOpenAPIOperationTagsFromOperation(seen, pathItem.Post)
		collectOpenAPIOperationTagsFromOperation(seen, pathItem.Put)
		collectOpenAPIOperationTagsFromOperation(seen, pathItem.Trace)
	}
	if len(seen) == 0 {
		return nil
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	return names
}

// collectOpenAPIOperationTagsFromOperation records tags from one operation.
func collectOpenAPIOperationTagsFromOperation(seen map[string]struct{}, operation *goai.Operation) {
	if seen == nil || operation == nil {
		return
	}
	for _, tag := range operation.Tags {
		name := strings.TrimSpace(tag)
		if name == "" {
			continue
		}
		seen[name] = struct{}{}
	}
}

// orderOpenAPIDocumentTags sorts tag display names with a generic hierarchical
// rule so same-prefix families stay together without naming specific modules.
func orderOpenAPIDocumentTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	ordered := append([]string(nil), tags...)
	sort.SliceStable(ordered, func(i, j int) bool {
		leftGroup, leftLeaf := openAPITagHierarchy(ordered[i])
		rightGroup, rightLeaf := openAPITagHierarchy(ordered[j])
		leftGroupKey := strings.ToLower(leftGroup)
		rightGroupKey := strings.ToLower(rightGroup)
		if leftGroupKey != rightGroupKey {
			return leftGroupKey < rightGroupKey
		}
		leftLeafKey := strings.ToLower(leftLeaf)
		rightLeafKey := strings.ToLower(rightLeaf)
		if leftLeafKey != rightLeafKey {
			return leftLeafKey < rightLeafKey
		}
		return strings.ToLower(ordered[i]) < strings.ToLower(ordered[j])
	})
	return ordered
}

// openAPITagHierarchy splits one tag into parent group and leaf using project
// conventions ("Parent / Child" or "Parent/Child"). Flat tags return themselves
// as the group with an empty leaf so they sort among other top-level groups.
func openAPITagHierarchy(name string) (group string, leaf string) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", ""
	}
	if group, leaf, ok := splitOpenAPITagHierarchy(trimmed, " / "); ok {
		return group, leaf
	}
	if group, leaf, ok := splitOpenAPITagHierarchy(trimmed, "/"); ok {
		return group, leaf
	}
	return trimmed, ""
}

// splitOpenAPITagHierarchy splits name on the first occurrence of sep when both
// sides are non-empty after trim.
func splitOpenAPITagHierarchy(name string, sep string) (group string, leaf string, ok bool) {
	index := strings.Index(name, sep)
	if index <= 0 {
		return "", "", false
	}
	group = strings.TrimSpace(name[:index])
	leaf = strings.TrimSpace(name[index+len(sep):])
	if group == "" || leaf == "" {
		return "", "", false
	}
	return group, leaf, true
}
