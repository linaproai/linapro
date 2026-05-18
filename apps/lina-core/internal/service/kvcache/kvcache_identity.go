// This file parses and validates public kvcache keys for non-SQL backend
// adapters.

package kvcache

import (
	"encoding/base64"
	"strconv"
	"strings"

	"lina-core/pkg/bizerr"
)

// Cache-size constants bound the persisted identity and payload lengths for KV
// cache entries regardless of backend.
const (
	cacheKeyPartCount = 3
	maxOwnerTypeBytes = 16
	maxOwnerKeyBytes  = 64
	maxNamespaceBytes = 64
	maxCacheKeyBytes  = 128
	maxValueBytes     = 4096
)

// cacheIdentity stores the decoded owner key, namespace, and logical cache
// key parsed from a public cache key string.
type cacheIdentity struct {
	tenantID  int64
	ownerKey  string
	namespace string
	cacheKey  string
}

// String returns the canonical owner type value.
func (value OwnerType) String() string {
	return string(value)
}

// resolveIdentity parses and validates one public cache key under the provided
// owner type.
func resolveIdentity(ownerType OwnerType, cacheKey string) (*cacheIdentity, error) {
	identity, err := parseCacheKey(cacheKey)
	if err != nil {
		return nil, err
	}
	if err = validateIdentity(ownerType, identity.ownerKey, identity.namespace, identity.cacheKey); err != nil {
		return nil, err
	}
	return identity, nil
}

// parseCacheKey decodes one public cache key back into its owner-scoped
// identity parts.
func parseCacheKey(cacheKey string) (*cacheIdentity, error) {
	parts := strings.Split(strings.TrimSpace(cacheKey), ".")
	if len(parts) != cacheKeyPartCount {
		return nil, bizerr.NewCode(CodeKVCacheKeyInvalid)
	}

	decodedParts := make([]string, 0, len(parts))
	for _, part := range parts {
		decoded, err := base64.RawURLEncoding.DecodeString(part)
		if err != nil {
			return nil, bizerr.NewCode(CodeKVCacheKeyInvalid)
		}
		decodedParts = append(decodedParts, string(decoded))
	}
	return &cacheIdentity{
		tenantID:  tenantIDFromOwnerKey(decodedParts[0]),
		ownerKey:  decodedParts[0],
		namespace: decodedParts[1],
		cacheKey:  decodedParts[2],
	}, nil
}

// tenantIDFromOwnerKey extracts the canonical tenant discriminator from
// tenantcap.CacheKey owner keys. Legacy owner keys remain platform scoped.
func tenantIDFromOwnerKey(ownerKey string) int64 {
	trimmedOwnerKey := strings.TrimSpace(ownerKey)
	if !strings.HasPrefix(trimmedOwnerKey, "tenant=") {
		return 0
	}
	tenantPart := strings.TrimPrefix(trimmedOwnerKey, "tenant=")
	separatorIndex := strings.Index(tenantPart, ":")
	if separatorIndex >= 0 {
		tenantPart = tenantPart[:separatorIndex]
	}
	tenantID, err := strconv.ParseInt(tenantPart, 10, 64)
	if err != nil {
		return 0
	}
	return tenantID
}

// validateIdentity validates the byte-length constraints for one decoded cache
// identity.
func validateIdentity(ownerType OwnerType, ownerKey string, namespace string, cacheKey string) error {
	if err := validateByteLength("ownerType", ownerType.String(), maxOwnerTypeBytes); err != nil {
		return err
	}
	if err := validateByteLength("ownerKey", ownerKey, maxOwnerKeyBytes); err != nil {
		return err
	}
	if err := validateByteLength("namespace", namespace, maxNamespaceBytes); err != nil {
		return err
	}
	if err := validateByteLength("cacheKey", cacheKey, maxCacheKeyBytes); err != nil {
		return err
	}
	return nil
}

// validateByteLength enforces a non-empty string field with a maximum byte
// length.
func validateByteLength(field string, value string, maxBytes int) error {
	if strings.TrimSpace(value) == "" {
		return bizerr.NewCode(CodeKVCacheFieldRequired, bizerr.P("field", field))
	}
	if len([]byte(value)) > maxBytes {
		return bizerr.NewCode(
			CodeKVCacheFieldTooLong,
			bizerr.P("field", field),
			bizerr.P("maxBytes", maxBytes),
		)
	}
	return nil
}

// validateMaxByteLength enforces only the maximum byte length for an optional
// string field.
func validateMaxByteLength(field string, value string, maxBytes int) error {
	if len([]byte(value)) > maxBytes {
		return bizerr.NewCode(
			CodeKVCacheValueTooLong,
			bizerr.P("field", field),
			bizerr.P("maxBytes", maxBytes),
		)
	}
	return nil
}
