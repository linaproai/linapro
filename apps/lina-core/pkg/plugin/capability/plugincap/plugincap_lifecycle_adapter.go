// This file contains the plugincap lifecycle service adapter implementation.
// The package entry file keeps the public contract and constructor only; this
// file owns input normalization and nil-backed no-op delegation.

package plugincap

import (
	"context"
	"strings"
)

// lifecycleServiceAdapter delegates lifecycle orchestration to the host-owned service.
type lifecycleServiceAdapter struct {
	lifecycle LifecycleService
}

// NewLifecycle creates a source-plugin visible plugin lifecycle service.
func NewLifecycle(lifecycle LifecycleService) LifecycleService {
	return &lifecycleServiceAdapter{lifecycle: lifecycle}
}

// EnsureTenantPluginDisableAllowed runs tenant-plugin disable preconditions.
func (s *lifecycleServiceAdapter) EnsureTenantPluginDisableAllowed(ctx context.Context, pluginID string, tenantID int) error {
	if s == nil || s.lifecycle == nil {
		return nil
	}
	return s.lifecycle.EnsureTenantPluginDisableAllowed(ctx, strings.TrimSpace(pluginID), tenantID)
}

// NotifyTenantPluginDisabled runs tenant-plugin disable notifications.
func (s *lifecycleServiceAdapter) NotifyTenantPluginDisabled(ctx context.Context, pluginID string, tenantID int) {
	if s == nil || s.lifecycle == nil {
		return
	}
	s.lifecycle.NotifyTenantPluginDisabled(ctx, strings.TrimSpace(pluginID), tenantID)
}

// EnsureTenantDeleteAllowed runs tenant-delete preconditions.
func (s *lifecycleServiceAdapter) EnsureTenantDeleteAllowed(ctx context.Context, tenantID int) error {
	if s == nil || s.lifecycle == nil {
		return nil
	}
	return s.lifecycle.EnsureTenantDeleteAllowed(ctx, tenantID)
}

// NotifyTenantDeleted runs tenant-delete notifications.
func (s *lifecycleServiceAdapter) NotifyTenantDeleted(ctx context.Context, tenantID int) {
	if s == nil || s.lifecycle == nil {
		return
	}
	s.lifecycle.NotifyTenantDeleted(ctx, tenantID)
}
