// This file defines read-only host configuration access for source plugins. It
// is intentionally separate from plugin config because plugin config is scoped
// to one plugin's own runtime configuration files.
package hostconfigcap

import (
	"context"

	"github.com/gogf/gf/v2/container/gvar"
)

// RawConfigReader is implemented by the host-owned config service. It keeps
// this adapter dependent on the startup-injected config instance without
// coupling this public capability package to host internal service packages.
type RawConfigReader interface {
	// GetRaw returns one raw host configuration value or root snapshot.
	GetRaw(ctx context.Context, key string) (*gvar.Var, error)
}

// serviceAdapter reads individual host config keys from the host config service.
type serviceAdapter struct {
	configSvc RawConfigReader
}

// New creates a host config reader backed by the host config service.
func New(configSvc RawConfigReader) Service {
	return &serviceAdapter{configSvc: configSvc}
}
