// Package infracap adapts host infrastructure status to plugin-visible
// infrastructure capability contracts.
package infracap

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/capmodel"
	capabilityinfracap "lina-core/pkg/plugin/capability/infracap"
)

// Service exposes infrastructure status projections and management commands.
type Service interface {
	capabilityinfracap.Service
	capabilityinfracap.AdminService
}

// adapter exposes host infrastructure status projections.
type adapter struct{}

var (
	_ capabilityinfracap.Service      = (*adapter)(nil)
	_ capabilityinfracap.AdminService = (*adapter)(nil)
)

// New creates the host-owned infrastructure capability adapter.
func New() Service {
	return &adapter{}
}

// BatchGetStatus returns stable host infrastructure component status projections.
func (a *adapter) BatchGetStatus(_ context.Context, _ capmodel.CapabilityContext, ids []capabilityinfracap.ComponentID) (*capmodel.BatchResult[*capabilityinfracap.StatusProjection, capabilityinfracap.ComponentID], error) {
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
func (a *adapter) RefreshStatus(context.Context, capmodel.CapabilityContext, capabilityinfracap.ComponentID) error {
	return nil
}
