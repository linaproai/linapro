// This file defines plugin lifecycle business error codes and their i18n
// metadata.

package plugin

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodePluginStatusInvalid reports that a lifecycle status value is not supported.
	CodePluginStatusInvalid = bizerr.MustDefine(
		"PLUGIN_STATUS_INVALID",
		"Plugin status supports only 0 or 1",
		gcode.CodeInvalidParameter,
	)
	// CodePluginNotInstalled reports that a lifecycle operation requires an installed plugin.
	CodePluginNotInstalled = bizerr.MustDefine(
		"PLUGIN_NOT_INSTALLED",
		"Plugin is not installed",
		gcode.CodeInvalidParameter,
	)
	// CodePluginInstallModeInvalid reports that an install request used an unsupported install mode.
	CodePluginInstallModeInvalid = bizerr.MustDefine(
		"PLUGIN_INSTALL_MODE_INVALID",
		"Plugin install mode supports only global or tenant_scoped",
		gcode.CodeInvalidParameter,
	)
	// CodePluginInstallModeInvalidForScopeNature reports an install-mode and scope-nature mismatch.
	CodePluginInstallModeInvalidForScopeNature = bizerr.MustDefine(
		"PLUGIN_INSTALL_MODE_INVALID_FOR_SCOPE_NATURE",
		"Plugin {pluginId} with scope_nature={scopeNature} cannot use install_mode={installMode}",
		gcode.CodeInvalidParameter,
	)
	// CodePluginTenantProvisioningPolicyInvalid reports that a new-tenant provisioning policy cannot apply to the plugin.
	CodePluginTenantProvisioningPolicyInvalid = bizerr.MustDefine(
		"PLUGIN_TENANT_PROVISIONING_POLICY_INVALID",
		"Plugin {pluginId} must support multi-tenant governance and be installed in tenant_scoped mode before it can be auto-enabled for new tenants",
		gcode.CodeInvalidParameter,
	)
	// CodePluginSourceManifestRequired reports that a source-plugin manifest is required.
	CodePluginSourceManifestRequired = bizerr.MustDefine(
		"PLUGIN_SOURCE_MANIFEST_REQUIRED",
		"Source plugin manifest cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodePluginSourceRegistryRequired reports that a source-plugin registry row is required.
	CodePluginSourceRegistryRequired = bizerr.MustDefine(
		"PLUGIN_SOURCE_REGISTRY_REQUIRED",
		"Source plugin registry cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodePluginSourceRegistryNotFound reports that a synchronized source-plugin registry row is missing.
	CodePluginSourceRegistryNotFound = bizerr.MustDefine(
		"PLUGIN_SOURCE_REGISTRY_NOT_FOUND",
		"Source plugin registry does not exist: {pluginId}",
		gcode.CodeNotFound,
	)
	// CodePluginReleaseNotFound reports that a plugin release row is missing.
	CodePluginReleaseNotFound = bizerr.MustDefine(
		"PLUGIN_RELEASE_NOT_FOUND",
		"Plugin release record does not exist: {pluginId}@{version}",
		gcode.CodeNotFound,
	)
	// CodePluginSourceRegistryAfterInstallNotFound reports install lost the source-plugin registry row.
	CodePluginSourceRegistryAfterInstallNotFound = bizerr.MustDefine(
		"PLUGIN_SOURCE_REGISTRY_AFTER_INSTALL_NOT_FOUND",
		"Source plugin registry does not exist after install: {pluginId}",
		gcode.CodeInternalError,
	)
	// CodePluginSourceRegistryAfterUninstallNotFound reports uninstall lost the source-plugin registry row.
	CodePluginSourceRegistryAfterUninstallNotFound = bizerr.MustDefine(
		"PLUGIN_SOURCE_REGISTRY_AFTER_UNINSTALL_NOT_FOUND",
		"Source plugin registry does not exist after uninstall: {pluginId}",
		gcode.CodeInternalError,
	)
	// CodePluginEnabledSnapshotRefreshFailed reports startup could not refresh enabled plugin state.
	CodePluginEnabledSnapshotRefreshFailed = bizerr.MustDefine(
		"PLUGIN_ENABLED_SNAPSHOT_REFRESH_FAILED",
		"Failed to refresh plugin enabled snapshot",
		gcode.CodeInternalError,
	)
	// CodePluginAutoEnableDiscoveryFailed reports startup auto-enable could not discover one plugin.
	CodePluginAutoEnableDiscoveryFailed = bizerr.MustDefine(
		"PLUGIN_AUTO_ENABLE_DISCOVERY_FAILED",
		"Startup auto-enable failed while discovering plugin {pluginId}",
		gcode.CodeInternalError,
	)
	// CodePluginAutoEnableManifestNotFound reports a configured auto-enable plugin has no manifest.
	CodePluginAutoEnableManifestNotFound = bizerr.MustDefine(
		"PLUGIN_AUTO_ENABLE_MANIFEST_NOT_FOUND",
		"Startup auto-enable plugin manifest does not exist: {pluginId}",
		gcode.CodeNotFound,
	)
	// CodePluginAutoEnableTypeUnsupported reports an unsupported plugin type in startup auto-enable.
	CodePluginAutoEnableTypeUnsupported = bizerr.MustDefine(
		"PLUGIN_AUTO_ENABLE_TYPE_UNSUPPORTED",
		"Startup auto-enable does not support plugin type {pluginType} for plugin {pluginId}",
		gcode.CodeInvalidParameter,
	)
	// CodePluginAutoEnableSourceManifestRequired reports a missing source manifest during startup.
	CodePluginAutoEnableSourceManifestRequired = bizerr.MustDefine(
		"PLUGIN_AUTO_ENABLE_SOURCE_MANIFEST_REQUIRED",
		"Startup auto-enable source plugin manifest cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodePluginAutoEnableDynamicManifestRequired reports a missing dynamic manifest during startup.
	CodePluginAutoEnableDynamicManifestRequired = bizerr.MustDefine(
		"PLUGIN_AUTO_ENABLE_DYNAMIC_MANIFEST_REQUIRED",
		"Startup auto-enable dynamic plugin manifest cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodePluginSourceInstallFailed reports startup source-plugin installation failed.
	CodePluginSourceInstallFailed = bizerr.MustDefine(
		"PLUGIN_SOURCE_INSTALL_FAILED",
		"Failed to install source plugin",
		gcode.CodeInternalError,
	)
	// CodePluginSourceEnableFailed reports startup source-plugin enabling failed.
	CodePluginSourceEnableFailed = bizerr.MustDefine(
		"PLUGIN_SOURCE_ENABLE_FAILED",
		"Failed to enable source plugin",
		gcode.CodeInternalError,
	)
	// CodePluginDynamicInstallFailed reports startup dynamic-plugin installation failed.
	CodePluginDynamicInstallFailed = bizerr.MustDefine(
		"PLUGIN_DYNAMIC_INSTALL_FAILED",
		"Failed to install dynamic plugin",
		gcode.CodeInternalError,
	)
	// CodePluginDynamicEnableFailed reports startup dynamic-plugin enabling failed.
	CodePluginDynamicEnableFailed = bizerr.MustDefine(
		"PLUGIN_DYNAMIC_ENABLE_FAILED",
		"Failed to enable dynamic plugin",
		gcode.CodeInternalError,
	)
	// CodePluginDynamicManifestRequired reports that a dynamic-plugin manifest is required.
	CodePluginDynamicManifestRequired = bizerr.MustDefine(
		"PLUGIN_DYNAMIC_MANIFEST_REQUIRED",
		"Dynamic plugin manifest cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodePluginDynamicAutoEnableReleaseMissing reports startup cannot reuse authorization without a release.
	CodePluginDynamicAutoEnableReleaseMissing = bizerr.MustDefine(
		"PLUGIN_DYNAMIC_AUTO_ENABLE_RELEASE_MISSING",
		"Dynamic plugin {pluginId} has no release record and cannot reuse authorization snapshot",
		gcode.CodeNotFound,
	)
	// CodePluginDynamicAutoEnableAuthSnapshotMissing reports startup requires prior authorization review.
	CodePluginDynamicAutoEnableAuthSnapshotMissing = bizerr.MustDefine(
		"PLUGIN_DYNAMIC_AUTO_ENABLE_AUTH_SNAPSHOT_MISSING",
		"Dynamic plugin {pluginId} has no confirmed host-service authorization snapshot. Complete review through the regular install or enable flow first",
		gcode.CodeInvalidParameter,
	)
	// CodePluginRegistryReadFailed reports startup could not read a plugin registry row.
	CodePluginRegistryReadFailed = bizerr.MustDefine(
		"PLUGIN_REGISTRY_READ_FAILED",
		"Failed to read plugin {pluginId} registry",
		gcode.CodeInternalError,
	)
	// CodePluginAutoEnableSharedExecutorMissing reports startup lacks the shared lifecycle executor.
	CodePluginAutoEnableSharedExecutorMissing = bizerr.MustDefine(
		"PLUGIN_AUTO_ENABLE_SHARED_EXECUTOR_MISSING",
		"Startup auto-enable plugin {pluginId} failed because shared executor is missing",
		gcode.CodeInternalError,
	)
	// CodePluginAutoEnableFailed reports startup auto-enable failed for one plugin.
	CodePluginAutoEnableFailed = bizerr.MustDefine(
		"PLUGIN_AUTO_ENABLE_FAILED",
		"Startup auto-enable plugin {pluginId} failed",
		gcode.CodeInternalError,
	)
	// CodePluginAutoEnableWaitCanceled reports startup waiting was canceled.
	CodePluginAutoEnableWaitCanceled = bizerr.MustDefine(
		"PLUGIN_AUTO_ENABLE_WAIT_CANCELED",
		"Startup wait for plugin {pluginId} auto-enable was canceled",
		gcode.CodeInternalError,
	)
	// CodePluginAutoEnableTimeoutRegistryMissing reports timeout before a registry row appeared.
	CodePluginAutoEnableTimeoutRegistryMissing = bizerr.MustDefine(
		"PLUGIN_AUTO_ENABLE_TIMEOUT_REGISTRY_MISSING",
		"Startup auto-enable plugin {pluginId} timed out because registry does not exist",
		gcode.CodeInternalError,
	)
	// CodePluginAutoEnableTimeoutState reports timeout with the last observed registry state.
	CodePluginAutoEnableTimeoutState = bizerr.MustDefine(
		"PLUGIN_AUTO_ENABLE_TIMEOUT_STATE",
		"Startup auto-enable plugin {pluginId} timed out: installed={installed} status={status} desiredState={desiredState} currentState={currentState}",
		gcode.CodeInternalError,
	)
	// CodePluginInstallMockDataFailed reports that the optional mock-data load
	// phase of an install request failed and was rolled back. The install SQL
	// itself succeeded; only the mock data was discarded. Callers can decide to
	// keep the plugin in its installed-without-mock state or to uninstall and
	// reinstall after fixing the mock SQL.
	CodePluginInstallMockDataFailed = bizerr.MustDefine(
		"PLUGIN_INSTALL_MOCK_DATA_FAILED",
		"Plugin {pluginId} installed successfully, but mock data file {failedFile} failed to load and was rolled back: {cause}",
		gcode.CodeInternalError,
	)
	// CodePluginLifecycleGuardVetoed reports that one or more lifecycle guards blocked an operation.
	CodePluginLifecycleGuardVetoed = bizerr.MustDefine(
		"PLUGIN_LIFECYCLE_GUARD_VETOED",
		"Plugin lifecycle operation {operation} for {pluginId} was blocked by lifecycle guards: {reasons}",
		gcode.CodeInvalidOperation,
	)
	// CodePluginForceUninstallDisabled reports that force uninstall is not enabled in host configuration.
	CodePluginForceUninstallDisabled = bizerr.MustDefine(
		"PLUGIN_FORCE_UNINSTALL_DISABLED",
		"Force uninstall is disabled by plugin.allowForceUninstall",
		gcode.CodeInvalidOperation,
	)
	// CodePluginDynamicArtifactMissingForUninstall reports that a dynamic
	// plugin cannot run a full uninstall because both staged and active release
	// artifacts are missing. Operators may use force uninstall to clear only
	// host governance state.
	CodePluginDynamicArtifactMissingForUninstall = bizerr.MustDefine(
		"PLUGIN_DYNAMIC_ARTIFACT_MISSING_FOR_UNINSTALL",
		"Dynamic plugin {pluginId} cannot run full uninstall because its wasm artifact is missing. Use force uninstall to clear host governance only",
		gcode.CodeInvalidOperation,
	)
	// CodePluginStartupConsistencyFailed reports invalid persisted plugin or tenant-governance startup state.
	CodePluginStartupConsistencyFailed = bizerr.MustDefine(
		"PLUGIN_STARTUP_CONSISTENCY_FAILED",
		"Plugin startup consistency validation failed: {details}",
		gcode.CodeInternalError,
	)
)
