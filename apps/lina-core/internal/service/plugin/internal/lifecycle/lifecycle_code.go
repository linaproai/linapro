// This file defines plugin lifecycle business error codes and their i18n
// metadata.

package lifecycle

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
	// messageKey is derived as error.plugin.install.mode.invalid.
	CodePluginInstallModeInvalid = bizerr.MustDefine(
		"PLUGIN_INSTALL_MODE_INVALID",
		"Plugin install mode supports only global or tenant_scoped",
		gcode.CodeInvalidParameter,
	)
	// CodePluginInstallModeInvalidForScopeNature reports an install-mode and scope-nature mismatch.
	// Error code is chosen so the derived key does not nest under
	// error.plugin.install.mode.invalid (which is already a leaf string).
	// messageKey is derived as error.plugin.install.mode.scope.nature.mismatch.
	CodePluginInstallModeInvalidForScopeNature = bizerr.MustDefine(
		"PLUGIN_INSTALL_MODE_SCOPE_NATURE_MISMATCH",
		"Plugin {pluginId} with scope_nature={scopeNature} cannot use install_mode={installMode}",
		gcode.CodeInvalidParameter,
	)
	// CodePluginEnabledSnapshotRefreshFailed reports enabled snapshot refresh failure.
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
	// CodePluginAutoEnableTenantProvisioningFailed reports startup could not
	// reconcile tenant-scoped auto-enabled plugins to existing tenants.
	CodePluginAutoEnableTenantProvisioningFailed = bizerr.MustDefine(
		"PLUGIN_AUTO_ENABLE_TENANT_PROVISIONING_FAILED",
		"Startup auto-enable tenant provisioning failed for plugin {pluginId}",
		gcode.CodeInternalError,
	)
	// CodePluginLifecyclePreconditionVetoed reports that one or more lifecycle
	// precondition callbacks blocked an operation.
	CodePluginLifecyclePreconditionVetoed = bizerr.MustDefine(
		"PLUGIN_LIFECYCLE_PRECONDITION_VETOED",
		"Plugin lifecycle operation {operation} for {pluginId} was blocked by lifecycle preconditions: {reasons}",
		gcode.CodeInvalidOperation,
	)
	// CodePluginDependencyBlocked reports that plugin dependency checks rejected a lifecycle action.
	CodePluginDependencyBlocked = bizerr.MustDefine(
		"PLUGIN_DEPENDENCY_BLOCKED",
		"Plugin {pluginId} dependency check failed: {blockers}",
		gcode.CodeInvalidParameter,
	)
	// CodePluginReverseDependencyBlocked reports that installed downstream plugins depend on the target plugin.
	CodePluginReverseDependencyBlocked = bizerr.MustDefine(
		"PLUGIN_REVERSE_DEPENDENCY_BLOCKED",
		"Plugin {pluginId} cannot be changed because installed plugins depend on it: {dependents}",
		gcode.CodeInvalidOperation,
	)
	// CodePluginReverseEnabledDependencyBlocked reports that enabled downstream
	// plugins still depend on the target plugin during a disable request.
	CodePluginReverseEnabledDependencyBlocked = bizerr.MustDefine(
		"PLUGIN_REVERSE_ENABLED_DEPENDENCY_BLOCKED",
		"Plugin {pluginId} cannot be disabled because enabled plugins depend on it: {dependents}",
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
	// Error code is chosen so the derived key does not nest under
	// error.plugin.dynamic.artifact.missing (which is already a leaf string).
	// messageKey is derived as error.plugin.dynamic.uninstall.artifact.missing.
	CodePluginDynamicArtifactMissingForUninstall = bizerr.MustDefine(
		"PLUGIN_DYNAMIC_UNINSTALL_ARTIFACT_MISSING",
		"Dynamic plugin {pluginId} cannot run full uninstall because its wasm artifact is missing. Use force uninstall to clear host governance only",
		gcode.CodeInvalidOperation,
	)
	// CodePluginUninstallExecutionFailed reports a failed uninstall after the
	// request passed dependency and lifecycle precondition checks.
	CodePluginUninstallExecutionFailed = bizerr.MustDefine(
		"PLUGIN_UNINSTALL_EXECUTION_FAILED",
		"Plugin {pluginId} uninstall failed",
		gcode.CodeInternalError,
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
	// CodeSourcePluginInstallUnsupported reports that source plugins cannot be installed by dynamic lifecycle.
	CodeSourcePluginInstallUnsupported = bizerr.MustDefine(
		"PLUGIN_SOURCE_INSTALL_UNSUPPORTED",
		"Source plugins are compiled into the host and cannot be installed by the dynamic lifecycle",
		gcode.CodeInvalidParameter,
	)
	// CodeDynamicPluginDowngradeUnsupported reports that dynamic release rollback is unsupported here.
	CodeDynamicPluginDowngradeUnsupported = bizerr.MustDefine(
		"PLUGIN_DYNAMIC_VERSION_DOWNGRADE_UNSUPPORTED",
		"Downgrading to an older dynamic plugin version is not supported. Use host rollback output or upload a newer version",
		gcode.CodeInvalidParameter,
	)
	// CodeSourcePluginUninstallUnsupported reports that source plugins cannot be uninstalled by dynamic lifecycle.
	CodeSourcePluginUninstallUnsupported = bizerr.MustDefine(
		"PLUGIN_SOURCE_UNINSTALL_UNSUPPORTED",
		"Source plugins are compiled into the host and cannot be uninstalled by the dynamic lifecycle",
		gcode.CodeInvalidParameter,
	)
)
