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
	GetJwt(ctx context.Context) (*JwtConfig, error)
	// GetJwtSecret returns the static JWT signing secret loaded from config.yaml.
	GetJwtSecret(ctx context.Context) string
	// GetJwtExpire returns the runtime-effective JWT expiration duration.
	GetJwtExpire(ctx context.Context) (time.Duration, error)
	// GetPublicFrontend returns the public-safe frontend branding and display
	// settings that can be consumed by login pages and the admin workspace.
	GetPublicFrontend(ctx context.Context) (*PublicFrontendConfig, error)
	// GetI18n reads runtime internationalization settings from configuration file.
	GetI18n(ctx context.Context) *I18nConfig
	// GetLogin reads runtime login parameters from sys_config.
	GetLogin(ctx context.Context) *LoginConfig
	// GetCron reads runtime cron-management parameters from protected sys_config entries.
	GetCron(ctx context.Context) (*CronConfig, error)
	// IsCronShellEnabled reports whether shell-type cron jobs are currently allowed.
	IsCronShellEnabled(ctx context.Context) (bool, error)
	// GetCronLogRetention returns the runtime-effective default cron log retention policy.
	GetCronLogRetention(ctx context.Context) (*CronLogRetentionConfig, error)
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
	// GetPluginAutoEnable returns the plugin IDs that the host should
	// automatically install and enable during startup bootstrap. Used by
	// callers that only need the ID list — e.g., the management UI's
	// "startup-managed" badge map.
	GetPluginAutoEnable(ctx context.Context) []string
	// GetPluginAutoEnableEntries returns the full normalized startup
	// auto-enable entries, including the per-entry WithMockData opt-in flag.
	// Used by the startup bootstrap flow to decide whether each plugin
	// should also load its mock-data SQL during the auto-install.
	GetPluginAutoEnableEntries(ctx context.Context) []PluginAutoEnableEntry
	// GetPluginDynamicStoragePath returns the runtime-resolved dynamic wasm storage directory.
	GetPluginDynamicStoragePath(ctx context.Context) string
	// GetSession reads session config from configuration file.
	GetSession(ctx context.Context) (*SessionConfig, error)
	// GetSessionTimeout returns the runtime-effective online-session timeout.
	GetSessionTimeout(ctx context.Context) (time.Duration, error)
	// GetUpload reads upload config from configuration file.
	GetUpload(ctx context.Context) (*UploadConfig, error)
	// GetUploadPath returns the runtime-resolved static upload directory loaded from config.yaml.
	GetUploadPath(ctx context.Context) string
	// GetUploadMaxSize returns the runtime-effective upload size ceiling in MB.
	GetUploadMaxSize(ctx context.Context) (int64, error)
	// MarkRuntimeParamsChanged bumps the shared runtime-parameter revision and clears
	// the current process snapshot after one protected runtime parameter mutation.
	MarkRuntimeParamsChanged(ctx context.Context) error
	// NotifyRuntimeParamsChanged best-effort refreshes the shared runtime-parameter revision.
	NotifyRuntimeParamsChanged(ctx context.Context)
	// SyncRuntimeParamSnapshot synchronizes the process-local runtime-parameter
	// snapshot cache with the shared revision visible to the current node.
	SyncRuntimeParamSnapshot(ctx context.Context) error
}

// Ensure serviceImpl implements Service.
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
