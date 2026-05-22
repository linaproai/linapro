// This file centralizes coordination key construction and component scoping.

package core

import (
	"strconv"
	"strings"

	"lina-core/pkg/bizerr"
)

// Key namespace constants keep Redis keys grouped by component.
const (
	defaultApplication = "linapro"
	defaultEnvironment = "default"
	defaultInstance    = "default"
)

// KeyBuilder builds backend keys with stable application and environment
// prefixes so tests and deployments can isolate coordination data.
type KeyBuilder struct {
	application string
	environment string
	instance    string
}

// NewKeyBuilder creates a key builder using normalized namespace parts.
func NewKeyBuilder(application string, environment string, instance string) *KeyBuilder {
	builder := &KeyBuilder{
		application: normalizeNamespacePart(application, defaultApplication),
		environment: normalizeNamespacePart(environment, defaultEnvironment),
		instance:    normalizeNamespacePart(instance, defaultInstance),
	}
	return builder
}

// DefaultKeyBuilder creates the default LinaPro key namespace builder.
func DefaultKeyBuilder() *KeyBuilder {
	return NewKeyBuilder("", "", "")
}

// LockKey builds the key used to store one distributed lock owner token.
func (b *KeyBuilder) LockKey(name string) (string, error) {
	segment, err := requireSegment("lock", name)
	if err != nil {
		return "", err
	}
	return b.join("lock", segment), nil
}

// LockFenceKey builds the key used to generate fencing tokens for a lock.
func (b *KeyBuilder) LockFenceKey(name string) (string, error) {
	segment, err := requireSegment("lock", name)
	if err != nil {
		return "", err
	}
	return b.join("lock-fence", segment), nil
}

// KVKey builds a tenant-aware short-lived key-value cache key.
func (b *KeyBuilder) KVKey(tenantID int64, ownerType string, ownerKey string, namespace string, key string) (string, error) {
	parts := []string{
		strconv.FormatInt(tenantID, 10),
		ownerType,
		ownerKey,
		namespace,
		key,
	}
	segments, err := requireSegments("kv", parts...)
	if err != nil {
		return "", err
	}
	return b.join(append([]string{"kv"}, segments...)...), nil
}

// RawKVKey builds a short-lived key-value key for host-internal auth/session
// state that already has a component-specific logical key.
func (b *KeyBuilder) RawKVKey(component string, parts ...string) (string, error) {
	segments, err := requireSegments(component, parts...)
	if err != nil {
		return "", err
	}
	return b.join(append([]string{component}, segments...)...), nil
}

// RevisionKey builds the key used for one cache-domain revision.
func (b *KeyBuilder) RevisionKey(key RevisionKey) (string, error) {
	parts := []string{
		strconv.FormatInt(key.TenantID, 10),
		key.Domain,
		key.Scope,
	}
	segments, err := requireSegments("revision", parts...)
	if err != nil {
		return "", err
	}
	return b.join(append([]string{"rev"}, segments...)...), nil
}

// EventChannel returns the Redis pub/sub channel for coordination events.
func (b *KeyBuilder) EventChannel() string {
	return b.join("event", "coordination")
}

// normalizeNamespacePart normalizes one namespace segment with a default value.
func normalizeNamespacePart(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

// join creates a namespaced backend key.
func (b *KeyBuilder) join(parts ...string) string {
	prefix := []string{b.application, b.environment, b.instance}
	return strings.Join(append(prefix, parts...), ":")
}

// requireSegments validates required key parts and returns trimmed segments.
func requireSegments(field string, parts ...string) ([]string, error) {
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		value, err := requireSegment(field, part)
		if err != nil {
			return nil, err
		}
		segments = append(segments, value)
	}
	return segments, nil
}

// requireSegment validates and trims one non-empty key component.
func requireSegment(field string, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", bizerr.NewCode(CodeCoordinationKeyInvalid, bizerr.P("field", field))
	}
	return trimmed, nil
}
