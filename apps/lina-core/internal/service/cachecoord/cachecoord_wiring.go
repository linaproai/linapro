// This file contains cachecoord construction wiring helpers, including static
// topology adapters and runtime backend replacement helpers.

package cachecoord

import (
	"context"

	"lina-core/internal/service/cluster"
	"lina-core/internal/service/coordination"
)

// staticTopology is a minimal cluster.Service placeholder used before startup
// wiring injects the real cluster topology service.
type staticTopology struct {
	enabled bool
	primary bool
	nodeID  string
}

// NewStaticTopology creates one static topology view for service-level cache
// coordination.
func NewStaticTopology(enabled bool) cluster.Service {
	return staticTopology{
		enabled: enabled,
		primary: !enabled,
		nodeID:  "local-node",
	}
}

// Start records no behavior for static cachecoord topology placeholders.
func (staticTopology) Start(context.Context) {}

// Stop records no behavior for static cachecoord topology placeholders.
func (staticTopology) Stop(context.Context) {}

// IsEnabled reports the configured cluster switch.
func (t staticTopology) IsEnabled() bool {
	return t.enabled
}

// IsPrimary reports the configured primary flag.
func (t staticTopology) IsPrimary() bool {
	return t.primary
}

// NodeID returns the configured node identifier.
func (t staticTopology) NodeID() string {
	if t.nodeID == "" {
		return "local-node"
	}
	return t.nodeID
}

// setTopology replaces the coordinator topology without resetting cache-domain
// observations or diagnostic state.
func (s *serviceImpl) setTopology(topology cluster.Service) {
	if s == nil {
		return
	}
	if topology == nil {
		topology = NewStaticTopology(false)
	}
	s.topologyMu.Lock()
	s.topology = topology
	s.topologyMu.Unlock()
}

// setCoordination replaces the distributed coordination backend used in
// clustered mode without resetting local cache observations.
func (s *serviceImpl) setCoordination(coordinationSvc coordination.Service) {
	if s == nil {
		return
	}
	s.coordMu.Lock()
	s.coord = coordinationSvc
	s.coordMu.Unlock()
}
