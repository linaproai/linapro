// This file contains cachecoord implementation helpers that mutate service
// wiring while preserving observed revision and diagnostic state.

package cachecoord

import "lina-core/internal/service/coordination"

// setTopology replaces the coordinator topology without resetting cache-domain
// observations or diagnostic state.
func (s *serviceImpl) setTopology(topology Topology) {
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
