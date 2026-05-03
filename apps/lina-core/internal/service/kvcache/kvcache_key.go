// This file exposes scoped cache-key encoding helpers used by kvcache callers.

package kvcache

import (
	"encoding/base64"
	"strings"
)

// BuildCacheKey encodes one owner-scoped logical cache key into a single
// stable string that can be passed through the public kvcache service methods.
func BuildCacheKey(ownerKey string, namespace string, cacheKey string) string {
	parts := []string{ownerKey, namespace, cacheKey}
	encodedParts := make([]string, 0, len(parts))
	for _, part := range parts {
		encodedParts = append(
			encodedParts,
			base64.RawURLEncoding.EncodeToString([]byte(strings.TrimSpace(part))),
		)
	}
	return strings.Join(encodedParts, ".")
}
