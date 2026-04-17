// This file defines host logger configuration loading.

package config

import "context"

const defaultLoggerFilePattern = "{Y-m-d}.log"

// LoggerConfig holds logger behavior switches loaded from configuration.
type LoggerConfig struct {
	Path       string `json:"path"`       // Path is the shared log directory for business and server logs.
	File       string `json:"file"`       // File is the shared log file pattern.
	Stdout     bool   `json:"stdout"`     // Stdout controls whether logs are also written to stdout.
	Structured bool   `json:"structured"` // Structured enables GoFrame JSON structured logging.
}

// GetLogger reads logger config from configuration file.
func (s *serviceImpl) GetLogger(ctx context.Context) *LoggerConfig {
	return cloneLoggerConfig(processStaticConfigCaches.logger.load(func() *LoggerConfig {
		cfg := &LoggerConfig{
			Path:       "",
			File:       defaultLoggerFilePattern,
			Stdout:     true,
			Structured: false,
		}
		mustScanConfig(ctx, "logger", cfg)
		return cfg
	}))
}
