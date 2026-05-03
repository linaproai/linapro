// Package config implements host configuration access for runtime settings,
// embedded delivery metadata, and related normalization helpers.
package config

import (
	"context"
	"time"
)

// Service defines the complete config service contract by composing narrower
// categorized interfaces for callers that need a smaller dependency surface.
type Service interface {
	ClusterConfigReader
	AuthConfigReader
	LoginConfigReader
	FrontendConfigReader
	I18nConfigReader
	CronConfigReader
	HostRuntimeConfigReader
	DeliveryMetadataConfigReader
	PluginConfigReader
	UploadConfigReader
	RuntimeParamSyncer
}

// ClusterConfigReader reads cluster topology settings.
type ClusterConfigReader interface {
	// GetCluster reads cluster config from configuration file.
	GetCluster(ctx context.Context) *ClusterConfig
	// IsClusterEnabled reports whether multi-node cluster mode is enabled.
	IsClusterEnabled(ctx context.Context) bool
}

// AuthConfigReader reads authentication and session settings.
type AuthConfigReader interface {
	// GetJwt reads JWT config from configuration file.
	GetJwt(ctx context.Context) (*JwtConfig, error)
	// GetJwtSecret returns the static JWT signing secret loaded from config.yaml.
	GetJwtSecret(ctx context.Context) string
	// GetJwtExpire returns the runtime-effective JWT expiration duration.
	GetJwtExpire(ctx context.Context) (time.Duration, error)
	// GetSession reads session config from configuration file.
	GetSession(ctx context.Context) (*SessionConfig, error)
	// GetSessionTimeout returns the runtime-effective online-session timeout.
	GetSessionTimeout(ctx context.Context) (time.Duration, error)
}

// LoginConfigReader reads login policy settings.
type LoginConfigReader interface {
	// GetLogin reads runtime login parameters from sys_config.
	GetLogin(ctx context.Context) (*LoginConfig, error)
	// IsLoginIPBlacklisted reports whether the input IP is denied by the
	// runtime-effective blacklist without constructing a full config object.
	IsLoginIPBlacklisted(ctx context.Context, ip string) (bool, error)
}

// FrontendConfigReader reads public frontend shell settings.
type FrontendConfigReader interface {
	// GetPublicFrontend returns the public-safe frontend branding and display
	// settings that can be consumed by login pages and the admin workspace.
	GetPublicFrontend(ctx context.Context) (*PublicFrontendConfig, error)
}

// I18nConfigReader reads host internationalization settings.
type I18nConfigReader interface {
	// GetI18n reads runtime internationalization settings from configuration file.
	GetI18n(ctx context.Context) *I18nConfig
}

// CronConfigReader reads scheduled-job governance settings.
type CronConfigReader interface {
	// GetCron reads runtime cron-management parameters from protected sys_config entries.
	GetCron(ctx context.Context) (*CronConfig, error)
	// IsCronShellEnabled reports whether shell-type cron jobs are currently allowed.
	IsCronShellEnabled(ctx context.Context) (bool, error)
	// GetCronLogRetention returns the runtime-effective default cron log retention policy.
	GetCronLogRetention(ctx context.Context) (*CronLogRetentionConfig, error)
}

// HostRuntimeConfigReader reads process-level host runtime settings.
type HostRuntimeConfigReader interface {
	// GetServerExtensions reads LinaPro-specific server extension settings from configuration file.
	GetServerExtensions(ctx context.Context) *ServerExtensionsConfig
	// GetLogger reads logger config from configuration file.
	GetLogger(ctx context.Context) *LoggerConfig
	// GetMonitor reads monitor config from configuration file.
	GetMonitor(ctx context.Context) *MonitorConfig
	// GetHealth reads health probe config from configuration file.
	GetHealth(ctx context.Context) *HealthConfig
	// GetShutdown reads graceful-shutdown config from configuration file.
	GetShutdown(ctx context.Context) *ShutdownConfig
	// GetScheduler reads scheduler config from configuration file.
	GetScheduler(ctx context.Context) *SchedulerConfig
	// GetSchedulerDefaultTimezone returns the configured default timezone for managed scheduled jobs.
	GetSchedulerDefaultTimezone(ctx context.Context) string
}

// DeliveryMetadataConfigReader reads embedded delivery metadata and OpenAPI settings.
type DeliveryMetadataConfigReader interface {
	// GetMetadata reads embedded delivery metadata from the packaged resource file.
	GetMetadata(ctx context.Context) *MetadataConfig
	// GetOpenApi reads OpenAPI config from embedded metadata.
	GetOpenApi(ctx context.Context) *OpenApiConfig
}

// PluginConfigReader reads plugin runtime and bootstrap settings.
type PluginConfigReader interface {
	// GetPlugin reads plugin config from configuration file.
	GetPlugin(ctx context.Context) *PluginConfig
	// GetPluginAutoEnable returns the plugin IDs that the host should
	// automatically install and enable during startup bootstrap. Used by
	// callers that only need the ID list, such as the management UI's
	// "startup-managed" badge map.
	GetPluginAutoEnable(ctx context.Context) []string
	// GetPluginAutoEnableEntries returns the full normalized startup
	// auto-enable entries, including the per-entry WithMockData opt-in flag.
	// Used by the startup bootstrap flow to decide whether each plugin
	// should also load its mock-data SQL during the auto-install.
	GetPluginAutoEnableEntries(ctx context.Context) []PluginAutoEnableEntry
	// GetPluginDynamicStoragePath returns the runtime-resolved dynamic wasm storage directory.
	GetPluginDynamicStoragePath(ctx context.Context) string
}

// UploadConfigReader reads upload storage and size-limit settings.
type UploadConfigReader interface {
	// GetUpload reads upload config from configuration file.
	GetUpload(ctx context.Context) (*UploadConfig, error)
	// GetUploadPath returns the runtime-resolved static upload directory loaded from config.yaml.
	GetUploadPath(ctx context.Context) string
	// GetUploadMaxSize returns the runtime-effective upload size ceiling in MB.
	GetUploadMaxSize(ctx context.Context) (int64, error)
}

// RuntimeParamSyncer synchronizes protected runtime-parameter snapshots.
type RuntimeParamSyncer interface {
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
	runtimeParamRevisionCtrl runtimeParamRevisionController
}

// New creates one config service instance.
func New() Service {
	svc := &serviceImpl{}
	svc.runtimeParamRevisionCtrl = newCacheCoordRuntimeParamRevisionController(
		svc.IsClusterEnabled(context.Background()),
	)
	return svc
}
