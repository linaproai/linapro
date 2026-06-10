// Package hostconfigcap defines host configuration capabilities published to
// plugins. The ordinary Service is read-only host configuration access. The
// AdminService manages governed runtime configuration projections and remains
// available only through source-plugin management surfaces.
package hostconfigcap

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	"lina-core/pkg/plugin/capability/capmodel"
)

// Service defines read-only host config values that source plugins may read.
type Service interface {
	// Get returns the raw host config value for the requested key or root snapshot.
	Get(ctx context.Context, key string) (*gvar.Var, error)
	// Exists reports whether a host config key is available.
	Exists(ctx context.Context, key string) (bool, error)
	// String reads a host config string value or returns defaultValue when
	// the key is absent or blank.
	String(ctx context.Context, key string, defaultValue string) (string, error)
	// Bool reads a host config bool value or returns defaultValue when the key is absent.
	Bool(ctx context.Context, key string, defaultValue bool) (bool, error)
	// Int reads a host config int value or returns defaultValue when the key is absent.
	Int(ctx context.Context, key string, defaultValue int) (int, error)
	// Duration reads a host config duration value or returns defaultValue when
	// the key is absent or blank.
	Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error)
}

// RuntimeConfigKey identifies one governed runtime configuration key.
type RuntimeConfigKey string

// RuntimeConfigProjection describes one runtime configuration value visible to
// a plugin management caller.
type RuntimeConfigProjection struct {
	// Key is the runtime configuration key.
	Key RuntimeConfigKey
	// ValueJSON is the JSON-encoded value projection.
	ValueJSON []byte
	// LabelKey is the optional i18n label key.
	LabelKey string
	// Label is the optional locale-resolved label.
	Label string
}

// AdminService defines governed runtime configuration management commands.
type AdminService interface {
	// BatchGetRuntimeConfig returns visible runtime configuration projections
	// and opaque missing keys.
	BatchGetRuntimeConfig(ctx context.Context, capCtx capmodel.CapabilityContext, keys []RuntimeConfigKey) (*capmodel.BatchResult[*RuntimeConfigProjection, RuntimeConfigKey], error)
	// SetRuntimeConfigJSON writes one governed runtime configuration value.
	SetRuntimeConfigJSON(ctx context.Context, capCtx capmodel.CapabilityContext, key RuntimeConfigKey, valueJSON []byte) error
}

// ScopeService defines host-internal runtime configuration visibility helpers.
type ScopeService interface {
	// EnsureRuntimeConfigKeysVisible rejects when any runtime configuration key
	// is outside caller scope.
	EnsureRuntimeConfigKeysVisible(ctx context.Context, capCtx capmodel.CapabilityContext, keys []RuntimeConfigKey) error
}
