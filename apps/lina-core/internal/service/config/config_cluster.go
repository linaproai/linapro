// This file defines cluster topology configuration loading and default election
// settings for single-node and multi-node deployments.

package config

import (
	"context"
	"time"
)

// ClusterConfig holds cluster topology configuration.
type ClusterConfig struct {
	Enabled  bool           `json:"enabled"`  // Enabled reports whether clustered deployment is enabled.
	Election ElectionConfig `json:"election"` // Election contains primary-election settings for clustered mode.
}

// ElectionConfig holds leader election configuration.
type ElectionConfig struct {
	Lease         time.Duration `json:"lease"`         // Lock lease duration
	RenewInterval time.Duration `json:"renewInterval"` // Lease renewal interval
}

// defaultElectionConfig returns the host defaults for leader-election timing.
func defaultElectionConfig() *ElectionConfig {
	return &ElectionConfig{
		Lease:         30 * time.Second,
		RenewInterval: 10 * time.Second,
	}
}

// getStaticClusterConfig lazily loads the cluster deployment mode from
// config.yaml so callers can branch on single-node vs multi-node behavior
// without reparsing the config section on hot paths.
func (s *serviceImpl) getStaticClusterConfig(ctx context.Context) *ClusterConfig {
	return processStaticConfigCaches.cluster.load(func() *ClusterConfig {
		cfg := &ClusterConfig{
			Enabled:  false,
			Election: *defaultElectionConfig(),
		}
		mustScanConfig(ctx, "cluster", cfg)
		return cfg
	})
}

// GetCluster reads cluster config from configuration file.
func (s *serviceImpl) GetCluster(ctx context.Context) *ClusterConfig {
	cfg := cloneClusterConfig(s.getStaticClusterConfig(ctx))
	if s != nil && s.clusterOverride != nil {
		cfg.Enabled = *s.clusterOverride
	}
	return cfg
}

// IsClusterEnabled reports whether multi-node cluster mode is enabled.
func (s *serviceImpl) IsClusterEnabled(ctx context.Context) bool {
	if s != nil && s.clusterOverride != nil {
		return *s.clusterOverride
	}
	cfg := s.getStaticClusterConfig(ctx)
	return cfg != nil && cfg.Enabled
}

// OverrideClusterEnabledForDialect locks cluster.enabled in memory for dialects
// that cannot safely back multi-node coordination.
func (s *serviceImpl) OverrideClusterEnabledForDialect(value bool) {
	if s == nil {
		return
	}
	s.clusterOverride = &value
}
