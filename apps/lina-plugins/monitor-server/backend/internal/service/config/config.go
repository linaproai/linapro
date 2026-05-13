// Package config implements monitor-server plugin configuration loading.
package config

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginservice/contract"
)

// Monitor-server configuration keys and defaults.
const (
	// configKeyMonitor identifies the monitor configuration section.
	configKeyMonitor = "monitor"
	// configKeyMonitorInterval identifies the monitor collection interval value.
	configKeyMonitorInterval = "monitor.interval"
	// defaultInterval is the default monitor collection period.
	defaultInterval = time.Minute
	// defaultRetentionMultiplier is the default stale-record retention multiplier.
	defaultRetentionMultiplier = 5
)

// Config holds monitor-server plugin configuration.
type Config struct {
	// Interval is the metrics collection period.
	Interval time.Duration `json:"interval"`
	// RetentionMultiplier multiplies Interval to produce the cleanup threshold.
	RetentionMultiplier int `json:"retentionMultiplier"`
}

// rawConfig captures values that can be scanned directly from the monitor section.
type rawConfig struct {
	// RetentionMultiplier is the optional cleanup retention multiplier.
	RetentionMultiplier int `json:"retentionMultiplier"`
}

// Load reads and validates monitor-server configuration through an explicit host reader.
func Load(ctx context.Context, reader contract.ConfigService) (*Config, error) {
	if reader == nil {
		return nil, gerror.New("monitor server config reader cannot be nil")
	}

	cfg := &Config{
		Interval:            defaultInterval,
		RetentionMultiplier: defaultRetentionMultiplier,
	}
	raw := &rawConfig{}
	if err := reader.Scan(ctx, configKeyMonitor, raw); err != nil {
		return nil, gerror.Wrap(err, "scan monitor server config failed")
	}
	if raw.RetentionMultiplier > 0 {
		cfg.RetentionMultiplier = raw.RetentionMultiplier
	}

	interval, err := reader.Duration(ctx, configKeyMonitorInterval, cfg.Interval)
	if err != nil {
		return nil, gerror.Wrap(err, "read monitor server interval config failed")
	}
	if err := validateInterval(interval); err != nil {
		return nil, err
	}
	cfg.Interval = interval
	return cfg, nil
}

// validateInterval verifies the monitor interval business constraints.
func validateInterval(interval time.Duration) error {
	if interval < time.Second {
		return gerror.Newf("config %s must be at least 1s", configKeyMonitorInterval)
	}
	if interval%time.Second != 0 {
		return gerror.Newf("config %s must align to whole seconds", configKeyMonitorInterval)
	}
	return nil
}
