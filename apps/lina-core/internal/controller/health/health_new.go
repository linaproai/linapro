// This file constructs the host health controller and wires its dependencies.

package health

import (
	"lina-core/api/health"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
)

// ControllerV1 handles host runtime health probes.
type ControllerV1 struct {
	clusterSvc cluster.Service // clusterSvc provides current deployment role state.
	configSvc  config.Service  // configSvc provides probe timeout configuration.
}

// NewV1 creates a host health controller instance.
func NewV1(configSvc config.Service, clusterSvc cluster.Service) health.IHealthV1 {
	if configSvc == nil {
		configSvc = config.New()
	}
	return &ControllerV1{
		clusterSvc: clusterSvc,
		configSvc:  configSvc,
	}
}
