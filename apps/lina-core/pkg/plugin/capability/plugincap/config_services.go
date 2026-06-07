// This file defines the source-plugin visible configuration contract.

package plugincap

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
)

// ConfigService defines the configuration operations published to source plugins.
type ConfigService interface {
	// Get returns the raw configuration value for the given key.
	Get(ctx context.Context, key string) (*gvar.Var, error)
	// Exists reports whether the given configuration key exists.
	Exists(ctx context.Context, key string) (bool, error)
	// Scan scans the configuration section into target.
	Scan(ctx context.Context, key string, target any) error
	// String reads a string value or returns defaultValue when the key is absent or blank.
	String(ctx context.Context, key string, defaultValue string) (string, error)
	// Bool reads a bool value or returns defaultValue when the key is absent.
	Bool(ctx context.Context, key string, defaultValue bool) (bool, error)
	// Int reads an int value or returns defaultValue when the key is absent.
	Int(ctx context.Context, key string, defaultValue int) (int, error)
	// Duration reads a time.Duration value or returns defaultValue when the key is absent or blank.
	Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error)
}

// ConfigServiceFactory creates plugin-scoped configuration service views.
type ConfigServiceFactory interface {
	// ForPlugin returns a configuration service scoped to pluginID. A blank
	// pluginID returns a service that rejects reads rather than falling back to
	// host-wide configuration.
	ForPlugin(pluginID string) ConfigService
	// WithArtifactConfig returns a new factory view that can use artifactContent
	// as the release-bound default config for pluginID when no external or
	// development config file exists.
	WithArtifactConfig(pluginID string, artifactContent []byte) ConfigServiceFactory
}
