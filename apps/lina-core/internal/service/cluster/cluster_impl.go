// This file implements cluster runtime state access, including primary-election
// start/stop delegation, primary-node decisions, node identity, and topology
// snapshots. It preserves single-node defaults when clustering is disabled and
// only delegates to the injected election service when cluster mode is active.

package cluster

import "context"

// Start starts clustered primary-election infrastructure when cluster mode is enabled.
func (s *serviceImpl) Start(ctx context.Context) {
	if s == nil || !s.IsEnabled() || s.electionSvc == nil {
		return
	}
	s.electionSvc.Start(ctx)
}

// Stop stops clustered primary-election infrastructure when it is running.
func (s *serviceImpl) Stop(ctx context.Context) {
	if s == nil || !s.IsEnabled() || s.electionSvc == nil {
		return
	}
	s.electionSvc.Stop(ctx)
}

// IsEnabled reports whether clustered deployment mode is enabled.
func (s *serviceImpl) IsEnabled() bool {
	return s != nil && s.cfg != nil && s.cfg.Enabled
}

// IsPrimary reports whether the current node should behave as the primary node.
func (s *serviceImpl) IsPrimary() bool {
	if s == nil || !s.IsEnabled() {
		return true
	}
	if s.electionSvc == nil {
		return false
	}
	return s.electionSvc.IsLeader()
}

// NodeID returns the stable identifier of the current host node.
func (s *serviceImpl) NodeID() string {
	if s == nil || s.nodeID == "" {
		return "local-node"
	}
	return s.nodeID
}
