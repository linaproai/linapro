// This file defines cluster topology configuration loading and default election
// settings for single-node and multi-node deployments.

package config

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/cluster"
)

// Supported cluster coordination backend names.
const (
	// ClusterCoordinationRedis identifies Redis as the cluster coordination backend.
	ClusterCoordinationRedis = "redis"
)

// ClusterConfig aliases the cluster-owned topology configuration value.
type ClusterConfig = cluster.ClusterConfig

// ElectionConfig aliases the cluster-owned leader election configuration value.
type ElectionConfig = cluster.ElectionConfig

// ClusterRedisConfig aliases the cluster-owned Redis coordination settings.
type ClusterRedisConfig = cluster.ClusterRedisConfig

// defaultElectionConfig returns the host defaults for leader-election timing.
func defaultElectionConfig() *ElectionConfig {
	return &ElectionConfig{
		Lease:         30 * time.Second,
		RenewInterval: 10 * time.Second,
	}
}

// defaultClusterRedisConfig returns safe startup defaults for Redis timeouts.
func defaultClusterRedisConfig() *ClusterRedisConfig {
	return &ClusterRedisConfig{
		ConnectTimeout: 3 * time.Second,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
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
			Redis:    *defaultClusterRedisConfig(),
		}
		mustScanConfig(ctx, "cluster", cfg)
		cfg.Coordination = strings.TrimSpace(cfg.Coordination)
		cfg.Election.Lease = mustLoadDurationConfig(ctx, "cluster.election.lease", cfg.Election.Lease)
		cfg.Election.RenewInterval = mustLoadDurationConfig(ctx, "cluster.election.renewInterval", cfg.Election.RenewInterval)
		cfg.Redis.ConnectTimeout = mustLoadDurationConfig(ctx, "cluster.redis.connectTimeout", cfg.Redis.ConnectTimeout)
		cfg.Redis.ReadTimeout = mustLoadDurationConfig(ctx, "cluster.redis.readTimeout", cfg.Redis.ReadTimeout)
		cfg.Redis.WriteTimeout = mustLoadDurationConfig(ctx, "cluster.redis.writeTimeout", cfg.Redis.WriteTimeout)
		mustValidateClusterConfig(cfg)
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

// GetClusterRedis reads the Redis coordination config from configuration file.
func (s *serviceImpl) GetClusterRedis(ctx context.Context) *ClusterRedisConfig {
	cfg := s.GetCluster(ctx)
	if cfg == nil {
		return nil
	}
	return cloneClusterRedisConfig(&cfg.Redis)
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
	s.runtimeParamRevisionCtrl = newCacheCoordRuntimeParamRevisionController(value)
}

// mustValidateClusterConfig validates deployment-mode coordination settings.
func mustValidateClusterConfig(cfg *ClusterConfig) {
	if cfg == nil || !cfg.Enabled {
		return
	}
	if cfg.Coordination == "" {
		panic(clusterStartupDiagnosticError(
			"cluster.coordination",
			"required when cluster.enabled=true",
			"set cluster.coordination=redis",
		))
	}
	if cfg.Coordination != ClusterCoordinationRedis {
		panic(clusterStartupDiagnosticError(
			"cluster.coordination",
			"unsupported value "+cfg.Coordination,
			"set cluster.coordination=redis",
		))
	}
	if strings.TrimSpace(cfg.Redis.Address) == "" {
		panic(clusterStartupDiagnosticError(
			"cluster.redis.address",
			"required when cluster.coordination=redis",
			"set cluster.redis.address to the Redis host:port endpoint",
		))
	}
}

// clusterStartupDiagnosticError formats static cluster configuration failures
// so startup logs identify the broken field and a concrete remediation.
func clusterStartupDiagnosticError(field string, reason string, fix string) error {
	return gerror.Newf("cluster startup diagnostic field=%s reason=%s fix=%s", field, reason, fix)
}
