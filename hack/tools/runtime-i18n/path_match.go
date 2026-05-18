// This file provides the small glob matcher used by the runtime i18n scanner.

package main

import (
	"regexp"
	"sort"
	"strings"
)

// pathMatchesAny reports whether relPath matches any repository-relative glob pattern.
func pathMatchesAny(relPath string, patterns []string) bool {
	for _, pattern := range patterns {
		if pathMatches(pattern, relPath) {
			return true
		}
	}
	return false
}

// pathMatches reports whether relPath matches the supported glob pattern.
func pathMatches(pattern string, relPath string) bool {
	regex := globToRegexp(pattern)
	return regex.MatchString(relPath)
}

// globToRegexp converts the small supported glob syntax into a regular expression.
func globToRegexp(pattern string) *regexp.Regexp {
	var builder strings.Builder
	builder.WriteString("^")
	for index := 0; index < len(pattern); {
		if strings.HasPrefix(pattern[index:], "**/") {
			builder.WriteString("(?:.*/)?")
			index += 3
			continue
		}
		if strings.HasPrefix(pattern[index:], "**") {
			builder.WriteString(".*")
			index += 2
			continue
		}
		character := pattern[index]
		switch character {
		case '*':
			builder.WriteString("[^/]*")
		case '?':
			builder.WriteString("[^/]")
		case '.', '+', '(', ')', '|', '[', ']', '{', '}', '^', '$', '\\':
			builder.WriteByte('\\')
			builder.WriteByte(character)
		default:
			builder.WriteByte(character)
		}
		index++
	}
	builder.WriteString("$")
	return regexp.MustCompile(builder.String())
}

// mapKeys returns the keys of a string map.
func mapKeys(values map[string]string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for key := range values {
		result[key] = struct{}{}
	}
	return result
}

// sortedMapKeys returns sorted keys from a map keyed by string.
func sortedMapKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// sortedDifference returns sorted keys present in left and missing from right.
func sortedDifference(left map[string]struct{}, right map[string]struct{}) []string {
	result := make([]string, 0)
	for key := range left {
		if _, ok := right[key]; !ok {
			result = append(result, key)
		}
	}
	sort.Strings(result)
	return result
}

// writeLine writes one line to the provided output stream.
func writeLine(out interface{ Write([]byte) (int, error) }, line string) error {
	_, err := out.Write([]byte(line + "\n"))
	return err
}
