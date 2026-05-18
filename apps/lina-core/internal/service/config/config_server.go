// This file defines LinaPro-specific server extension configuration loading.

package config

import "context"

// defaultServerApiDocPath is the fallback route where the host publishes the
// managed OpenAPI JSON document.
const defaultServerApiDocPath = "/api.json"

// ServerExtensionsConfig holds LinaPro server settings that extend the GoFrame
// `server` configuration section.
type ServerExtensionsConfig struct {
	// ApiDocPath is the route where the host-managed OpenAPI JSON document is exposed.
	ApiDocPath string `json:"apiDocPath"`
}

// GetServerExtensions reads LinaPro-specific server extension settings from configuration file.
func (s *serviceImpl) GetServerExtensions(ctx context.Context) *ServerExtensionsConfig {
	return cloneServerExtensionsConfig(processStaticConfigCaches.serverExtensions.load(func() *ServerExtensionsConfig {
		cfg := &struct {
			Extensions ServerExtensionsConfig `json:"extensions"`
		}{
			Extensions: ServerExtensionsConfig{
				ApiDocPath: defaultServerApiDocPath,
			},
		}
		mustScanConfig(ctx, "server", cfg)
		return &cfg.Extensions
	}))
}
