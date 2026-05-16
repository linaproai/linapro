// Package cluster provides one topology abstraction for single-node and
// clustered deployments.
package cluster

import (
	"context"
	"time"

	"lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
)

// Default election timing constants keep standalone construction deterministic
// when config values are absent.
const (
	defaultElectionLease         = 30 * time.Second
	defaultElectionRenewInterval = 10 * time.Second
)

// Service defines the cluster service contract.
type Service interface {
	// Start starts clustered primary-election infrastructure when cluster mode is enabled.
	// The call is a no-op for nil services, standalone deployments, or services
	// constructed without a coordination lock backend. It does not return
	// startup errors; callers observe election state through IsPrimary.
	Start(ctx context.Context)
	// Stop stops clustered primary-election infrastructure when it is running.
	// The call is idempotent and only affects the local election worker.
	Stop(ctx context.Context)
	// IsEnabled reports whether clustered deployment mode is enabled from the
	// normalized host configuration. Disabled mode keeps all coordination local.
	IsEnabled() bool
	// IsPrimary reports whether the current node should behave as the primary
	// node. Standalone deployments are always primary; clustered deployments
	// without an election backend return false to avoid split-primary work.
	IsPrimary() bool
	// NodeID returns the stable identifier of the current host node. A fallback
	// local identifier is returned when the service is nil or uninitialized.
	NodeID() string
}

// Interface compliance assertion for the default cluster service
// implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	cfg         *config.ClusterConfig // cfg stores the normalized cluster settings.
	nodeID      string                // nodeID is the stable identifier of the current node.
	electionSvc *electionService      // electionSvc participates in primary election for clustered mode.
}

// New creates and returns a new cluster Service instance.
func New(cfg *config.ClusterConfig) Service {
	return NewWithCoordination(cfg, nil)
}

// NewWithCoordination creates a cluster Service using the provided
// coordination service for distributed leader election in cluster mode.
func NewWithCoordination(cfg *config.ClusterConfig, coordinationSvc coordination.Service) Service {
	normalizedCfg := normalizeClusterConfig(cfg)
	service := &serviceImpl{
		cfg:    normalizedCfg,
		nodeID: generateNodeIdentifier(),
	}
	if !normalizedCfg.Enabled {
		return service
	}

	if coordinationSvc == nil || coordinationSvc.Lock() == nil {
		return service
	}
	service.electionSvc = newElectionService(coordinationSvc.Lock(), &normalizedCfg.Election, service.nodeID)
	return service
}

// normalizeClusterConfig applies default election settings while preserving the
// caller-provided enablement flag and positive timing values.
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
