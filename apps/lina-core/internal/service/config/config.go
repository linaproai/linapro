// Package config implements host configuration access for runtime settings,
// embedded delivery metadata, and related normalization helpers.
package config

import (
	"context"
	"time"

	"lina-core/internal/service/kvcache"
)

// Service defines the config service contract.
type Service interface {
	// GetCluster reads cluster config from configuration file.
	GetCluster(ctx context.Context) *ClusterConfig
	// IsClusterEnabled reports whether multi-node cluster mode is enabled.
	IsClusterEnabled(ctx context.Context) bool
	// GetJwt reads JWT config from configuration file.
	GetJwt(ctx context.Context) *JwtConfig
	// GetJwtSecret returns the static JWT signing secret loaded from config.yaml.
	GetJwtSecret(ctx context.Context) string
	// GetJwtExpire returns the runtime-effective JWT expiration duration.
	GetJwtExpire(ctx context.Context) time.Duration
	// GetPublicFrontend returns the public-safe frontend branding and display
	// settings that can be consumed by login pages and the admin workspace.
	GetPublicFrontend(ctx context.Context) *PublicFrontendConfig
	// GetLogin reads runtime login parameters from sys_config.
	GetLogin(ctx context.Context) *LoginConfig
	// IsLoginIPBlacklisted reports whether the input IP is denied by the
	// runtime-effective blacklist without constructing a full config object.
	IsLoginIPBlacklisted(ctx context.Context, ip string) bool
	// GetServerExtensions reads LinaPro-specific server extension settings from configuration file.
	GetServerExtensions(ctx context.Context) *ServerExtensionsConfig
	// GetLogger reads logger config from configuration file.
	GetLogger(ctx context.Context) *LoggerConfig
	// GetMetadata reads embedded delivery metadata from the packaged resource file.
	GetMetadata(ctx context.Context) *MetadataConfig
	// GetMonitor reads monitor config from configuration file.
	GetMonitor(ctx context.Context) *MonitorConfig
	// GetOpenApi reads OpenAPI config from embedded metadata.
	GetOpenApi(ctx context.Context) *OpenApiConfig
	// GetPlugin reads plugin config from configuration file.
	GetPlugin(ctx context.Context) *PluginConfig
	// GetPluginDynamicStoragePath returns the normalized dynamic wasm storage directory.
	GetPluginDynamicStoragePath(ctx context.Context) string
	// GetSession reads session config from configuration file.
	GetSession(ctx context.Context) *SessionConfig
	// GetSessionTimeout returns the runtime-effective online-session timeout.
	GetSessionTimeout(ctx context.Context) time.Duration
	// GetUpload reads upload config from configuration file.
	GetUpload(ctx context.Context) *UploadConfig
	// GetUploadPath returns the static upload directory loaded from config.yaml.
	GetUploadPath(ctx context.Context) string
	// GetUploadMaxSize returns the runtime-effective upload size ceiling in MB.
	GetUploadMaxSize(ctx context.Context) int64
	// MarkRuntimeParamsChanged bumps the shared runtime-parameter revision and clears
	// the current process snapshot after one protected runtime parameter mutation.
	MarkRuntimeParamsChanged(ctx context.Context) error
	// NotifyRuntimeParamsChanged best-effort refreshes the shared runtime-parameter revision.
	NotifyRuntimeParamsChanged(ctx context.Context)
	// SyncRuntimeParamSnapshot synchronizes the process-local runtime-parameter
	// snapshot cache with the shared revision visible to the current node.
	SyncRuntimeParamSnapshot(ctx context.Context) error
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	kvCacheSvc               kvcache.Service
	runtimeParamRevisionCtrl runtimeParamRevisionController
}

// New creates one config service instance.
func New() Service {
	svc := &serviceImpl{
		kvCacheSvc: kvcache.New(),
	}
	svc.runtimeParamRevisionCtrl = newRuntimeParamRevisionController(
		svc.IsClusterEnabled(context.Background()),
		svc.kvCacheSvc,
	)
	return svc
}
