// This file owns the source-plugin scheduled-job registrar implementation used
// by integration startup. Keeping it in integration lets runtime guards reuse
// plugin enablement and cluster topology services directly.

package integration

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcron"

	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/pluginhost"
)

// sourceJobRegistrar registers source-plugin scheduled jobs into GoFrame cron.
type sourceJobRegistrar struct {
	pluginID string
	service  *serviceImpl
	services capability.Services
}

// Ensure sourceJobRegistrar satisfies the published registrar contract.
var _ pluginhost.JobsRegistrar = (*sourceJobRegistrar)(nil)

// newSourceJobRegistrar creates one host-owned jobs registrar for a source plugin.
func newSourceJobRegistrar(pluginID string, service *serviceImpl) pluginhost.JobsRegistrar {
	normalizedPluginID := strings.TrimSpace(pluginID)
	var services capability.Services
	if service != nil {
		services = service.sourceServicesForPlugin(normalizedPluginID)
	}
	return &sourceJobRegistrar{
		pluginID: normalizedPluginID,
		service:  service,
		services: services,
	}
}

// Add registers one guarded scheduled job.
func (r *sourceJobRegistrar) Add(
	ctx context.Context,
	pattern string,
	name string,
	handler pluginhost.JobHandler,
) error {
	return r.AddWithMetadata(ctx, pattern, name, name, "", handler)
}

// AddWithMetadata registers one guarded scheduled job with English source display metadata.
func (r *sourceJobRegistrar) AddWithMetadata(
	ctx context.Context,
	pattern string,
	name string,
	displayName string,
	description string,
	handler pluginhost.JobHandler,
) error {
	if handler == nil {
		return gerror.New("pluginhost: job handler is nil")
	}

	_, err := gcron.Add(ctx, pattern, func(jobCtx context.Context) {
		if !r.canRun(jobCtx) {
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
func (r *sourceJobRegistrar) IsPrimaryNode() bool {
	if r == nil || r.service == nil {
		return false
	}
	return r.service.isPrimaryNode()
}

// Services returns the host-published runtime services for source-plugin construction.
func (r *sourceJobRegistrar) Services() capability.Services {
	if r == nil {
		return nil
	}
	return r.services
}

// canRun reports whether the owning plugin may execute background work.
func (r *sourceJobRegistrar) canRun(ctx context.Context) bool {
	if r == nil || r.service == nil {
		return false
	}
	return r.service.canExposePluginBusinessEntries(ctx, r.pluginID)
}
