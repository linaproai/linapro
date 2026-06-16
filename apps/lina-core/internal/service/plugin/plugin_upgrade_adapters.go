// This file adapts root-facade cache publication and freshness helpers into
// the narrow contracts consumed by the unified upgrade component.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
)

// upgradeCachePublisher publishes upgrade cache changes through the root
// facade's single plugin-change path.
type upgradeCachePublisher struct {
	service *serviceImpl
}

// PublishPluginChange publishes a plugin-scoped mutation reason.
func (p upgradeCachePublisher) PublishPluginChange(
	ctx context.Context,
	pluginID string,
	pluginType string,
	reason string,
) error {
	if p.service == nil {
		return gerror.New("plugin upgrade cache publisher is not configured")
	}
	return p.service.PublishPluginChange(ctx, pluginID, pluginType, reason)
}

// SyncEnabledSnapshotAndPublishRuntimeChange refreshes local enablement and
// publishes a scoped mutation through the root facade.
func (p upgradeCachePublisher) SyncEnabledSnapshotAndPublishRuntimeChange(
	ctx context.Context,
	pluginID string,
	reason string,
) error {
	if p.service == nil {
		return gerror.New("plugin upgrade cache publisher is not configured")
	}
	return p.service.syncEnabledSnapshotAndPublishRuntimeChange(ctx, pluginID, reason)
}

// upgradeCacheFreshener refreshes runtime caches before read-only upgrade paths.
type upgradeCacheFreshener struct {
	service *serviceImpl
}

// EnsureRuntimeCacheFresh synchronizes local runtime caches with the shared revision.
func (f upgradeCacheFreshener) EnsureRuntimeCacheFresh(ctx context.Context) error {
	if f.service == nil {
		return gerror.New("plugin upgrade cache freshener is not configured")
	}
	return f.service.ensureRuntimeCacheFresh(ctx)
}
