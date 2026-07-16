// Package config implements host configuration access for runtime settings,
// embedded delivery metadata, and related normalization helpers.
package config

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
)

// Service defines the complete host configuration contract.
type Service interface {
	// GetRaw returns one raw host configuration value or root snapshot for key.
	GetRaw(ctx context.Context, key string) (*gvar.Var, error)

	// GetWorkspace reads the admin workspace routing config from configuration file.
	GetWorkspace(ctx context.Context) *WorkspaceConfig
	// GetWorkspaceBasePath returns the normalized admin workspace entry path.
	GetWorkspaceBasePath(ctx context.Context) string

	// GetCluster reads cluster config from configuration file.
	GetCluster(ctx context.Context) *ClusterConfig
	// GetClusterRedis reads the Redis coordination config from configuration file.
	GetClusterRedis(ctx context.Context) *ClusterRedisConfig
	// IsClusterEnabled reports whether multi-node cluster mode is enabled.
	IsClusterEnabled(ctx context.Context) bool
	// OverrideClusterEnabledForDialect locks cluster.enabled in memory when the
	// active database dialect cannot support multi-node coordination.
	OverrideClusterEnabledForDialect(value bool)

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

	// GetLogin reads runtime login parameters from sys_config.
	GetLogin(ctx context.Context) (*LoginConfig, error)
	// IsLoginIPBlacklisted reports whether the input IP is denied by the
	// runtime-effective blacklist without constructing a full config object.
	IsLoginIPBlacklisted(ctx context.Context, ip string) (bool, error)

	// GetLogRetentionDays returns the runtime-effective maximum log retention period in days.
	GetLogRetentionDays(ctx context.Context) (int64, error)

	// GetPublicFrontend returns the public-safe frontend branding and display
	// settings that can be consumed by login pages and the admin workspace.
	GetPublicFrontend(ctx context.Context) (*PublicFrontendConfig, error)

	// GetI18n reads runtime internationalization settings from configuration file.
	GetI18n(ctx context.Context) *I18nConfig

	// GetCron reads runtime cron-management parameters from protected sys_config entries.
	GetCron(ctx context.Context) (*CronConfig, error)
	// IsCronShellEnabled reports whether shell-type cron jobs are currently allowed.
	IsCronShellEnabled(ctx context.Context) (bool, error)
	// GetCronLogRetention returns the runtime-effective default cron log retention policy.
	GetCronLogRetention(ctx context.Context) (*CronLogRetentionConfig, error)

	// GetServerExtensions reads LinaPro-specific server extension settings from configuration file.
	GetServerExtensions(ctx context.Context) *ServerExtensionsConfig
	// GetLogger reads logger config from configuration file.
	GetLogger(ctx context.Context) *LoggerConfig
	// GetScheduler reads scheduler config from configuration file.
	GetScheduler(ctx context.Context) *SchedulerConfig
	// GetSchedulerDefaultTimezone returns the configured default timezone for managed scheduled jobs.
	GetSchedulerDefaultTimezone(ctx context.Context) string

	// GetMetadata reads embedded delivery metadata from the packaged resource file.
	GetMetadata(ctx context.Context) *MetadataConfig
	// GetOpenApi reads OpenAPI config from embedded metadata.
	GetOpenApi(ctx context.Context) *OpenApiConfig

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

	// GetUpload reads upload config from configuration file.
	GetUpload(ctx context.Context) (*UploadConfig, error)
	// GetUploadPath returns the runtime-resolved static upload directory loaded from config.yaml.
	GetUploadPath(ctx context.Context) string
	// GetUploadMaxSize returns the runtime-effective upload size ceiling in MB.
	GetUploadMaxSize(ctx context.Context) (int64, error)
	// GetUploadDirectUrlTTL returns the runtime-effective lifetime for client
	// direct object-storage access (presigned put/get and direct-upload sessions).
	// Default is 1h via sys.upload.directUrlTTL.
	GetUploadDirectUrlTTL(ctx context.Context) (time.Duration, error)
	// GetUploadMultipartEnabled reports whether automatic multipart planning is enabled.
	GetUploadMultipartEnabled(ctx context.Context) (bool, error)
	// GetUploadMultipartThresholdMB returns the auto-multipart threshold in MB.
	GetUploadMultipartThresholdMB(ctx context.Context) (int64, error)
	// GetUploadMultipartPartSizeMB returns the multipart part size in MB.
	GetUploadMultipartPartSizeMB(ctx context.Context) (int64, error)
	// GetUploadMultipartMaxConcurrency returns the suggested client part concurrency.
	GetUploadMultipartMaxConcurrency(ctx context.Context) (int64, error)

	// MarkRuntimeParamsChanged bumps the shared sys_config revision and clears
	// the current process snapshot after one system-configuration mutation.
	MarkRuntimeParamsChanged(ctx context.Context) error
	// NotifyRuntimeParamsChanged best-effort refreshes the shared sys_config revision.
	NotifyRuntimeParamsChanged(ctx context.Context)
	// SyncRuntimeParamSnapshot synchronizes the process-local sys_config
	// snapshot cache with the shared revision visible to the current node.
	SyncRuntimeParamSnapshot(ctx context.Context) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	runtimeParamRevisionCtrl runtimeParamRevisionController
	clusterOverride          *bool
}

// New creates one config service instance.
func New() Service {
	svc := &serviceImpl{}
	svc.runtimeParamRevisionCtrl = newCacheCoordRuntimeParamRevisionController(
		svc.IsClusterEnabled(context.Background()),
	)
	return svc
}
