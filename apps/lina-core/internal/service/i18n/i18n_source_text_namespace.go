// This file maintains the registry of runtime i18n key namespaces whose
// target-language fallback is owned by source code instead of JSON resources.

package i18n

import (
	"strings"
	"sync"
)

// sourceTextNamespaceRegistry stores registered source-text-backed key prefixes.
var sourceTextNamespaceRegistry = struct {
	sync.RWMutex
	prefixes map[string]string
}{
	prefixes: make(map[string]string),
}

// RegisterSourceTextNamespace registers one runtime i18n key prefix whose
// fallback text is supplied by source-owned metadata.
func RegisterSourceTextNamespace(prefix string, reason string) {
	trimmedPrefix := strings.TrimSpace(prefix)
	if trimmedPrefix == "" {
		return
	}

	sourceTextNamespaceRegistry.Lock()
	defer sourceTextNamespaceRegistry.Unlock()
	sourceTextNamespaceRegistry.prefixes[trimmedPrefix] = strings.TrimSpace(reason)
}

// SourceTextNamespaceReason returns the registration reason for the key's
// matching source-text-backed namespace.
func SourceTextNamespaceReason(key string) (string, bool) {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return "", false
	}

	sourceTextNamespaceRegistry.RLock()
	defer sourceTextNamespaceRegistry.RUnlock()
	for prefix, reason := range sourceTextNamespaceRegistry.prefixes {
		if strings.HasPrefix(trimmedKey, prefix) {
			return reason, true
		}
	}
	return "", false
}

// RegisteredSourceTextNamespaces returns a defensive copy of the namespace registry.
func RegisteredSourceTextNamespaces() map[string]string {
	sourceTextNamespaceRegistry.RLock()
	defer sourceTextNamespaceRegistry.RUnlock()

	result := make(map[string]string, len(sourceTextNamespaceRegistry.prefixes))
	for prefix, reason := range sourceTextNamespaceRegistry.prefixes {
		result[prefix] = reason
	}
	return result
}

// resetSourceTextNamespacesForTest clears the namespace registry for isolated unit tests.
func resetSourceTextNamespacesForTest() {
	sourceTextNamespaceRegistry.Lock()
	defer sourceTextNamespaceRegistry.Unlock()
	sourceTextNamespaceRegistry.prefixes = make(map[string]string)
}
