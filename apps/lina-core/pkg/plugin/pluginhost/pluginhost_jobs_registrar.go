// This file defines the public jobs registrar contract exposed to source
// plugins and the guarded host-side implementation used at runtime.

package pluginhost

import (
	"context"
	"lina-core/pkg/logger"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcron"
)

// PrimaryNodeChecker defines one host callback that reports whether the current node is the primary node.
type PrimaryNodeChecker func() bool

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
	Services() Services
}

// jobsRegistrar is the host-owned JobsRegistrar implementation for one source
// plugin registration session.
type jobsRegistrar struct {
	pluginID           string
	enabledChecker     PluginEnabledChecker
	primaryNodeChecker PrimaryNodeChecker
	services           Services
}

// NewJobsRegistrar creates and returns a new JobsRegistrar instance.
func NewJobsRegistrar(
	pluginID string,
	enabledChecker PluginEnabledChecker,
	primaryNodeChecker PrimaryNodeChecker,
	services Services,
) JobsRegistrar {
	return &jobsRegistrar{
		pluginID:           pluginID,
		enabledChecker:     enabledChecker,
		primaryNodeChecker: primaryNodeChecker,
		services:           services,
	}
}

// Add registers one guarded scheduled job.
func (r *jobsRegistrar) Add(
	ctx context.Context,
	pattern string,
	name string,
	handler JobHandler,
) error {
	return r.AddWithMetadata(ctx, pattern, name, name, "", handler)
}

// AddWithMetadata registers one guarded scheduled job with English source display metadata.
func (r *jobsRegistrar) AddWithMetadata(
	ctx context.Context,
	pattern string,
	name string,
	displayName string,
	description string,
	handler JobHandler,
) error {
	if handler == nil {
		return gerror.New("pluginhost: job handler is nil")
	}

	_, err := gcron.Add(ctx, pattern, func(jobCtx context.Context) {
		if r.enabledChecker != nil && !r.enabledChecker(jobCtx, r.pluginID) {
			return
		}
		// Protect every scheduled-job callback at runtime so disabling a plugin
		// immediately stops future executions without requiring host restart or
		// plugin re-registration.
		if runErr := handler(jobCtx); runErr != nil {
			logger.Warningf(jobCtx, "plugin job failed plugin=%s name=%s err=%v", r.pluginID, name, runErr)
		}
	}, name)
	return err
}

// IsPrimaryNode reports whether the current host node is the primary node.
func (r *jobsRegistrar) IsPrimaryNode() bool {
	if r == nil || r.primaryNodeChecker == nil {
		return true
	}
	return r.primaryNodeChecker()
}

// Services returns the host-published runtime services for source-plugin construction.
func (r *jobsRegistrar) Services() Services {
	if r == nil {
		return nil
	}
	return r.services
}
