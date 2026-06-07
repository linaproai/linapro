// Package configcap defines governed runtime-configuration capability
// contracts for plugins without leaking host configuration internals.
package configcap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
)

// ConfigKey identifies one governed runtime configuration key.
type ConfigKey string

// Projection describes one configuration value visible to a plugin.
type Projection struct {
	// Key is the configuration key.
	Key ConfigKey
	// ValueJSON is the JSON-encoded value projection.
	ValueJSON []byte
	// LabelKey is the optional i18n label key.
	LabelKey string
	// Label is the optional locale-resolved label.
	Label string
}

// Service defines read-oriented configuration capability methods.
type Service interface {
	// BatchGetConfig returns visible configuration projections and opaque missing keys.
	BatchGetConfig(ctx context.Context, capCtx capmodel.CapabilityContext, keys []ConfigKey) (*capmodel.BatchResult[*Projection, ConfigKey], error)
}

// AdminService defines governed configuration management commands.
type AdminService interface {
	// SetConfigJSON writes one governed configuration value.
	SetConfigJSON(ctx context.Context, capCtx capmodel.CapabilityContext, key ConfigKey, valueJSON []byte) error
}

// ScopeService defines host-internal configuration visibility helpers.
type ScopeService interface {
	// EnsureKeysVisible rejects when any configuration key is outside caller scope.
	EnsureKeysVisible(ctx context.Context, capCtx capmodel.CapabilityContext, keys []ConfigKey) error
}
