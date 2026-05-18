// This file implements shared string normalization helpers for host-service contracts.

package hostservice

import (
	"sort"
	"strings"
)

// normalizeHostServiceName trims and lowercases one host service identifier.
func normalizeHostServiceName(value string) string {
	return normalizeLower(value, "")
}

// normalizeHostServiceMethod trims and lowercases one host service method name.
func normalizeHostServiceMethod(value string) string {
	return normalizeLower(value, "")
}

// normalizeLower trims and lowercases one string, applying the default when
// the normalized result is empty.
func normalizeLower(value string, defaultValue string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return defaultValue
	}
	return normalized
}

// normalizeStoragePathSlice normalizes declared storage paths and drops invalid
// entries for clone-style normalization flows.
func normalizeStoragePathSlice(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	items := make([]string, 0, len(paths))
	for _, rawPath := range paths {
		normalizedPath, err := normalizeStorageDeclaredPath(rawPath)
		if err != nil {
			continue
		}
		items = append(items, normalizedPath)
	}
	sort.Strings(items)
	return items
}

// normalizeLowerStringSlice trims, lowercases, de-duplicates, and sorts one
// string slice.
func normalizeLowerStringSlice(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := normalizeLower(item, "")
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

// normalizeUpperStringSlice trims, uppercases, de-duplicates, and sorts one
// string slice.
func normalizeUpperStringSlice(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.ToUpper(strings.TrimSpace(item))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

// normalizeTableSlice trims, de-duplicates, and sorts declared data table
// names.
func normalizeTableSlice(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

// normalizeStringMap trims keys and values while discarding empty keys from a
// metadata map.
func normalizeStringMap(items map[string]string) map[string]string {
	if len(items) == 0 {
		return nil
	}
	result := make(map[string]string, len(items))
	for key, value := range items {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		result[trimmedKey] = strings.TrimSpace(value)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
