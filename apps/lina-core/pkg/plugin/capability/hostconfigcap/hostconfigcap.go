// Package hostconfigcap defines the read-only host configuration capability
// published to plugins. It is separate from plugincap.ConfigService, which only
// reads the current plugin's own static configuration.
package hostconfigcap

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
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
