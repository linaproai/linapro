// This file implements scoped cache-key encoding and decoding helpers used by
// the kvcache service contract.

package kvcache

import (
	"encoding/base64"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// cacheKeyPartCount defines the number of encoded segments carried by one
// public cache key.
const cacheKeyPartCount = 3

// cacheIdentity stores the decoded owner key, namespace, and logical cache
// key parsed from a public cache key string.
type cacheIdentity struct {
	ownerKey  string
	namespace string
	cacheKey  string
}

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

// parseCacheKey decodes one public cache key back into its owner-scoped
// identity parts.
func parseCacheKey(cacheKey string) (*cacheIdentity, error) {
	parts := strings.Split(strings.TrimSpace(cacheKey), ".")
	if len(parts) != cacheKeyPartCount {
		return nil, gerror.New("缓存键格式无效，请使用 kvcache.BuildCacheKey 构造")
	}

	decodedParts := make([]string, 0, len(parts))
	for _, part := range parts {
		decoded, err := base64.RawURLEncoding.DecodeString(part)
		if err != nil {
			return nil, gerror.New("缓存键格式无效，请使用 kvcache.BuildCacheKey 构造")
		}
		decodedParts = append(decodedParts, string(decoded))
	}
	return &cacheIdentity{
		ownerKey:  decodedParts[0],
		namespace: decodedParts[1],
		cacheKey:  decodedParts[2],
	}, nil
}
