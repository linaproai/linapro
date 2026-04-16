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

func defaultElectionConfig() *ElectionConfig {
	return &ElectionConfig{
		Lease:         30 * time.Second,
		RenewInterval: 10 * time.Second,
	}
}

// GetCluster reads cluster config from configuration file.
func (s *serviceImpl) GetCluster(ctx context.Context) *ClusterConfig {
	cfg := &ClusterConfig{
		Enabled:  false,
		Election: *defaultElectionConfig(),
	}
	mustScanConfig(ctx, "cluster", cfg)
	return cfg
}
