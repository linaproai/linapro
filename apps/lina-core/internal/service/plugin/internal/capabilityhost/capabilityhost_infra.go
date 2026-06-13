// This file adapts host infrastructure status to plugin-visible
// infrastructure capability contracts.
package capabilityhost

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/capmodel"
	capabilityinfracap "lina-core/pkg/plugin/capability/infracap"
)

// Service exposes infrastructure status projections and management commands.
type infraCapabilityService interface {
	capabilityinfracap.Service
	capabilityinfracap.AdminService
}

// adapter exposes host infrastructure status projections.
type infraCapabilityAdapter struct{}

var (
	_ capabilityinfracap.Service      = (*infraCapabilityAdapter)(nil)
	_ capabilityinfracap.AdminService = (*infraCapabilityAdapter)(nil)
)

// New creates the host-owned infrastructure capability adapter.
func newInfraCapabilityAdapter() infraCapabilityService {
	return &infraCapabilityAdapter{}
}

// BatchGetStatus returns stable host infrastructure component status projections.
func (a *infraCapabilityAdapter) BatchGetStatus(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityinfracap.ComponentID) (*capmodel.BatchResult[*capabilityinfracap.StatusProjection, capabilityinfracap.ComponentID], error) {
	result := &capmodel.BatchResult[*capabilityinfracap.StatusProjection, capabilityinfracap.ComponentID]{
		Items:      make(map[capabilityinfracap.ComponentID]*capabilityinfracap.StatusProjection, len(ids)),
		MissingIDs: []capabilityinfracap.ComponentID{},
	}
	for _, id := range ids {
		normalizedID := strings.TrimSpace(string(id))
		if normalizedID == "" {
			result.MissingIDs = append(result.MissingIDs, id)
			continue
		}
		result.Items[id] = &capabilityinfracap.StatusProjection{
			ID:        id,
			Available: true,
			Status:    "available",
			LabelKey:  "infra." + normalizedID + ".status.available",
		}
	}
	return result, nil
}

// RefreshStatus is a no-op for static infrastructure projections.
func (a *infraCapabilityAdapter) RefreshStatus(context.Context, capmodel.CapabilityContext, capabilityinfracap.ComponentID) error {
	return nil
}
