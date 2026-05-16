// This file contains the concrete lifecycle event delegation behavior for
// source plugins. It keeps nil-runner handling and plugin ID normalization out
// of the package entrypoint while preserving host runner semantics.

package pluginlifecycle

import (
	"context"
	"strings"
)

// EnsureTenantPluginDisableAllowed runs tenant-plugin disable preconditions.
func (s *service) EnsureTenantPluginDisableAllowed(ctx context.Context, pluginID string, tenantID int) error {
	if s == nil || s.runner == nil {
		return nil
	}
	return s.runner.EnsureTenantPluginDisableAllowed(ctx, strings.TrimSpace(pluginID), tenantID)
}

// NotifyTenantPluginDisabled runs tenant-plugin disable notifications.
func (s *service) NotifyTenantPluginDisabled(ctx context.Context, pluginID string, tenantID int) {
	if s == nil || s.runner == nil {
		return
	}
	s.runner.NotifyTenantPluginDisabled(ctx, strings.TrimSpace(pluginID), tenantID)
}

// EnsureTenantDeleteAllowed runs tenant-delete preconditions.
func (s *service) EnsureTenantDeleteAllowed(ctx context.Context, tenantID int) error {
	if s == nil || s.runner == nil {
		return nil
	}
	return s.runner.EnsureTenantDeleteAllowed(ctx, tenantID)
}

// NotifyTenantDeleted runs tenant-delete notifications.
func (s *service) NotifyTenantDeleted(ctx context.Context, tenantID int) {
	if s == nil || s.runner == nil {
		return
	}
	s.runner.NotifyTenantDeleted(ctx, tenantID)
}
