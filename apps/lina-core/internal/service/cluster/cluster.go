// Package cluster provides one topology abstraction for single-node and
// clustered deployments.
package cluster

import (
	"context"
	"time"

	"lina-core/internal/service/config"
	"lina-core/internal/service/locker"
)

const (
	defaultElectionLease         = 30 * time.Second
	defaultElectionRenewInterval = 10 * time.Second
)

// Service defines the cluster service contract.
type Service interface {
	// Start starts clustered primary-election infrastructure when cluster mode is enabled.
	Start(ctx context.Context)
	// Stop stops clustered primary-election infrastructure when it is running.
	Stop(ctx context.Context)
	// IsEnabled reports whether clustered deployment mode is enabled.
	IsEnabled() bool
	// IsPrimary reports whether the current node should behave as the primary node.
	IsPrimary() bool
	// NodeID returns the stable identifier of the current host node.
	NodeID() string
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	cfg         *config.ClusterConfig // cfg stores the normalized cluster settings.
	nodeID      string                // nodeID is the stable identifier of the current node.
	electionSvc *electionService      // electionSvc participates in primary election for clustered mode.
}

// New creates and returns a new cluster Service instance.
func New(cfg *config.ClusterConfig) Service {
	normalizedCfg := normalizeClusterConfig(cfg)
	service := &serviceImpl{
		cfg:    normalizedCfg,
		nodeID: generateNodeIdentifier(),
	}
	if !normalizedCfg.Enabled {
		return service
	}

	service.electionSvc = newElectionService(locker.New(), &normalizedCfg.Election, service.nodeID)
	return service
}

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

func normalizeClusterConfig(cfg *config.ClusterConfig) *config.ClusterConfig {
	normalizedCfg := &config.ClusterConfig{
		Enabled: false,
		Election: config.ElectionConfig{
			Lease:         defaultElectionLease,
			RenewInterval: defaultElectionRenewInterval,
		},
	}
	if cfg == nil {
		return normalizedCfg
	}

	normalizedCfg.Enabled = cfg.Enabled
	if cfg.Election.Lease > 0 {
		normalizedCfg.Election.Lease = cfg.Election.Lease
	}
	if cfg.Election.RenewInterval > 0 {
		normalizedCfg.Election.RenewInterval = cfg.Election.RenewInterval
	}
	return normalizedCfg
}
