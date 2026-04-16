package config

import

// serviceImpl implements Service.
"context"

// Service defines the config service contract.
type Service interface {
	// GetCluster reads cluster config from configuration file.
	GetCluster(ctx context.Context) *ClusterConfig
	// GetInit reads initialization config from configuration file.
	GetInit(ctx context.Context) *InitConfig
	// GetJwt reads JWT config from configuration file.
	GetJwt(ctx context.Context) *JwtConfig
	// GetMonitor reads monitor config from configuration file.
	GetMonitor(ctx context.Context) *MonitorConfig
	// GetOpenApi reads OpenAPI config from configuration file.
	GetOpenApi(ctx context.Context) *OpenApiConfig
	// GetPlugin reads plugin config from configuration file.
	GetPlugin(ctx context.Context) *PluginConfig
	// GetPluginDynamicStoragePath returns the normalized dynamic wasm storage directory.
	GetPluginDynamicStoragePath(ctx context.Context) string
	// GetSession reads session config from configuration file.
	GetSession(ctx context.Context) *SessionConfig
	// GetUpload reads upload config from configuration file.
	GetUpload(ctx context.Context) *UploadConfig
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

func New() Service {
	return &serviceImpl{}
}
