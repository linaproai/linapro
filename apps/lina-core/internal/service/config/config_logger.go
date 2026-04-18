// This file defines host logger configuration loading.

package config

import "context"

// defaultLoggerFilePattern is the fallback daily log file pattern used when
// the logger config omits the file name template.
const defaultLoggerFilePattern = "{Y-m-d}.log"

// LoggerExtensionsConfig holds LinaPro logger settings that extend the GoFrame
// `logger` configuration section.
type LoggerExtensionsConfig struct {
	// Structured enables GoFrame JSON structured logging.
	Structured bool `json:"structured"`
}

// LoggerConfig holds logger behavior switches loaded from configuration.
type LoggerConfig struct {
	Path       string                 `json:"path"`       // Path is the shared log directory for business and server logs.
	File       string                 `json:"file"`       // File is the shared log file pattern.
	Stdout     bool                   `json:"stdout"`     // Stdout controls whether logs are also written to stdout.
	Extensions LoggerExtensionsConfig `json:"extensions"` // Extensions holds LinaPro-specific logger behavior switches.
}

// GetLogger reads logger config from configuration file.
func (s *serviceImpl) GetLogger(ctx context.Context) *LoggerConfig {
	return cloneLoggerConfig(processStaticConfigCaches.logger.load(func() *LoggerConfig {
		cfg := &LoggerConfig{
			Path:   "",
			File:   defaultLoggerFilePattern,
			Stdout: true,
			Extensions: LoggerExtensionsConfig{
				Structured: false,
			},
		}
		mustScanConfig(ctx, "logger", cfg)
		return cfg
	}))
}
