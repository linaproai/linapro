// This file defines the public jobs registrar contract exposed to source
// plugins. Host-owned implementations live in the integration layer where they
// can directly reuse the owning services for enablement and topology decisions.

package pluginhost

import (
	"context"
	"lina-core/pkg/plugin/capability"
)

// JobHandler defines one plugin-owned scheduled job callback.
type JobHandler func(ctx context.Context) error

// JobsRegistrar exposes host job registration and node-role inspection for one plugin.
type JobsRegistrar interface {
	// Add registers one guarded scheduled job.
	Add(ctx context.Context, pattern string, name string, handler JobHandler) error
	// AddWithMetadata registers one guarded scheduled job with English source display
	// metadata used by the unified scheduled-job management view.
	AddWithMetadata(ctx context.Context, pattern string, name string, displayName string, description string, handler JobHandler) error
	// IsPrimaryNode reports whether the current host node is the primary node.
	IsPrimaryNode() bool
	// Services returns the host-published runtime services for source-plugin construction.
	Services() capability.Services
}
