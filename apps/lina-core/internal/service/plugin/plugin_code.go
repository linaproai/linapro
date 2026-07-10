// This file defines plugin lifecycle business error codes and their i18n
// metadata.

package plugin

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/upgrade"
	"lina-core/pkg/bizerr"
)

var (
	// CodePluginStatusInvalid reports that a lifecycle status value is not supported.
	CodePluginStatusInvalid = lifecycle.CodePluginStatusInvalid
	// CodePluginNotInstalled reports that a lifecycle operation requires an installed plugin.
	CodePluginNotInstalled = lifecycle.CodePluginNotInstalled
	// CodePluginNotFound reports that a plugin management query could not find the target plugin.
	CodePluginNotFound = bizerr.MustDefine(
		"PLUGIN_NOT_FOUND",
		"Plugin does not exist: {pluginId}",
		gcode.CodeNotFound,
	)
	// CodePluginRuntimeUpgradePreviewUnavailable reports that no upgrade preview can be produced.
	CodePluginRuntimeUpgradePreviewUnavailable = upgrade.CodePluginRuntimeUpgradePreviewUnavailable
	// CodePluginRuntimeUpgradeConfirmationRequired reports that an upgrade
	// request omitted the explicit operator confirmation.
	CodePluginRuntimeUpgradeConfirmationRequired = upgrade.CodePluginRuntimeUpgradeConfirmationRequired
	// CodePluginRuntimeUpgradeUnavailable reports that execution is allowed only
	// for plugins still marked as pending upgrade after server-side state re-read.
	CodePluginRuntimeUpgradeUnavailable = upgrade.CodePluginRuntimeUpgradeUnavailable
	// CodePluginRuntimeUpgradeLockUnavailable reports that clustered runtime
	// upgrade cannot safely acquire a deployment-wide lock backend.
	CodePluginRuntimeUpgradeLockUnavailable = upgrade.CodePluginRuntimeUpgradeLockUnavailable
	// CodePluginRuntimeUpgradeAlreadyRunning reports that another node already
	// owns the runtime-upgrade lock for the target plugin.
	CodePluginRuntimeUpgradeAlreadyRunning = upgrade.CodePluginRuntimeUpgradeAlreadyRunning
	// CodePluginRuntimeUpgradeTypeUnsupported reports that the target plugin type
	// cannot run through the runtime upgrade executor.
	CodePluginRuntimeUpgradeTypeUnsupported = upgrade.CodePluginRuntimeUpgradeTypeUnsupported
	// CodePluginRuntimeUpgradeExecutionFailed reports a failed explicit upgrade
	// after the request passed confirmation and state validation.
	CodePluginRuntimeUpgradeExecutionFailed = upgrade.CodePluginRuntimeUpgradeExecutionFailed
	// CodePluginSourceUpgradeLifecycleVetoed reports that a source upgrade callback blocked execution.
	CodePluginSourceUpgradeLifecycleVetoed = upgrade.CodePluginSourceUpgradeLifecycleVetoed
	// CodePluginUninstallExecutionFailed reports a failed uninstall after the
	// request passed dependency and lifecycle precondition checks.
	CodePluginUninstallExecutionFailed = lifecycle.CodePluginUninstallExecutionFailed
	// CodePluginRuntimeUpgradeSnapshotMissing reports missing effective or target manifest snapshot data.
	CodePluginRuntimeUpgradeSnapshotMissing = upgrade.CodePluginRuntimeUpgradeSnapshotMissing
	// CodePluginRuntimeUpgradeSnapshotInvalid reports invalid upgrade preview snapshot metadata.
	CodePluginRuntimeUpgradeSnapshotInvalid = upgrade.CodePluginRuntimeUpgradeSnapshotInvalid
	// CodePluginInstallModeInvalid reports that an install request used an unsupported install mode.
	CodePluginInstallModeInvalid = lifecycle.CodePluginInstallModeInvalid
	// CodePluginInstallModeInvalidForScopeNature reports an install-mode and scope-nature mismatch.
	CodePluginInstallModeInvalidForScopeNature = lifecycle.CodePluginInstallModeInvalidForScopeNature
	// CodePluginTenantProvisioningPolicyInvalid reports that a new-tenant provisioning policy cannot apply to the plugin.
	CodePluginTenantProvisioningPolicyInvalid = bizerr.MustDefine(
		"PLUGIN_TENANT_PROVISIONING_POLICY_INVALID",
		"Plugin {pluginId} must support multi-tenant governance and be installed in tenant_scoped mode before it can be auto-enabled for new tenants",
		gcode.CodeInvalidParameter,
	)
	// CodePluginBuiltinManagementActionDenied reports that ordinary plugin
	// management attempted to mutate a project built-in plugin.
	CodePluginBuiltinManagementActionDenied = bizerr.MustDefine(
		"PLUGIN_BUILTIN_MANAGEMENT_ACTION_DENIED",
		"Built-in plugin {pluginId} is managed by host startup and cannot be changed through ordinary plugin management",
		gcode.CodeInvalidOperation,
	)
	// CodePluginSourceManifestRequired reports that a source-plugin manifest is required.
	CodePluginSourceManifestRequired = lifecycle.CodePluginSourceManifestRequired
	// CodePluginSourceRegistryRequired reports that a source-plugin registry row is required.
	CodePluginSourceRegistryRequired = lifecycle.CodePluginSourceRegistryRequired
	// CodePluginSourceRegistryNotFound reports that a synchronized source-plugin registry row is missing.
	CodePluginSourceRegistryNotFound = lifecycle.CodePluginSourceRegistryNotFound
	// CodePluginReleaseNotFound reports that a plugin release row is missing.
	CodePluginReleaseNotFound = lifecycle.CodePluginReleaseNotFound
	// CodePluginSourceRegistryAfterInstallNotFound reports install lost the source-plugin registry row.
	CodePluginSourceRegistryAfterInstallNotFound = lifecycle.CodePluginSourceRegistryAfterInstallNotFound
	// CodePluginSourceRegistryAfterUninstallNotFound reports uninstall lost the source-plugin registry row.
	CodePluginSourceRegistryAfterUninstallNotFound = lifecycle.CodePluginSourceRegistryAfterUninstallNotFound
	// CodePluginEnabledSnapshotRefreshFailed reports startup could not refresh enabled plugin state.
	CodePluginEnabledSnapshotRefreshFailed = lifecycle.CodePluginEnabledSnapshotRefreshFailed
	// CodePluginAutoEnableDiscoveryFailed reports startup auto-enable could not discover one plugin.
	CodePluginAutoEnableDiscoveryFailed = lifecycle.CodePluginAutoEnableDiscoveryFailed
	// CodePluginAutoEnableManifestNotFound reports a configured auto-enable plugin has no manifest.
	CodePluginAutoEnableManifestNotFound = lifecycle.CodePluginAutoEnableManifestNotFound
	// CodePluginAutoEnableTypeUnsupported reports an unsupported plugin type in startup auto-enable.
	CodePluginAutoEnableTypeUnsupported = lifecycle.CodePluginAutoEnableTypeUnsupported
	// CodePluginAutoEnableSourceManifestRequired reports a missing source manifest during startup.
	CodePluginAutoEnableSourceManifestRequired = lifecycle.CodePluginAutoEnableSourceManifestRequired
	// CodePluginAutoEnableDynamicManifestRequired reports a missing dynamic manifest during startup.
	CodePluginAutoEnableDynamicManifestRequired = lifecycle.CodePluginAutoEnableDynamicManifestRequired
	// CodePluginSourceInstallFailed reports startup source-plugin installation failed.
	CodePluginSourceInstallFailed = lifecycle.CodePluginSourceInstallFailed
	// CodePluginSourceEnableFailed reports startup source-plugin enabling failed.
	CodePluginSourceEnableFailed = lifecycle.CodePluginSourceEnableFailed
	// CodePluginDynamicInstallFailed reports startup dynamic-plugin installation failed.
	CodePluginDynamicInstallFailed = lifecycle.CodePluginDynamicInstallFailed
	// CodePluginDynamicEnableFailed reports startup dynamic-plugin enabling failed.
	CodePluginDynamicEnableFailed = lifecycle.CodePluginDynamicEnableFailed
	// CodePluginDynamicManifestRequired reports that a dynamic-plugin manifest is required.
	CodePluginDynamicManifestRequired = lifecycle.CodePluginDynamicManifestRequired
	// CodePluginDynamicAutoEnableReleaseMissing reports startup cannot reuse authorization without a release.
	CodePluginDynamicAutoEnableReleaseMissing = lifecycle.CodePluginDynamicAutoEnableReleaseMissing
	// CodePluginDynamicAutoEnableAuthSnapshotMissing reports startup requires prior authorization review.
	CodePluginDynamicAutoEnableAuthSnapshotMissing = lifecycle.CodePluginDynamicAutoEnableAuthSnapshotMissing
	// CodePluginRegistryReadFailed reports startup could not read a plugin registry row.
	CodePluginRegistryReadFailed = lifecycle.CodePluginRegistryReadFailed
	// CodePluginAutoEnableSharedExecutorMissing reports startup lacks the shared lifecycle executor.
	CodePluginAutoEnableSharedExecutorMissing = lifecycle.CodePluginAutoEnableSharedExecutorMissing
	// CodePluginAutoEnableFailed reports startup auto-enable failed for one plugin.
	CodePluginAutoEnableFailed = lifecycle.CodePluginAutoEnableFailed
	// CodePluginAutoEnableWaitCanceled reports startup waiting was canceled.
	CodePluginAutoEnableWaitCanceled = lifecycle.CodePluginAutoEnableWaitCanceled
	// CodePluginAutoEnableTimeoutRegistryMissing reports timeout before a registry row appeared.
	CodePluginAutoEnableTimeoutRegistryMissing = lifecycle.CodePluginAutoEnableTimeoutRegistryMissing
	// CodePluginAutoEnableTimeoutState reports timeout with the last observed registry state.
	CodePluginAutoEnableTimeoutState = lifecycle.CodePluginAutoEnableTimeoutState
	// CodePluginAutoEnableTenantProvisioningFailed reports startup could not
	// reconcile tenant-scoped auto-enabled plugins to existing tenants.
	CodePluginAutoEnableTenantProvisioningFailed = lifecycle.CodePluginAutoEnableTenantProvisioningFailed
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
	// CodePluginLifecyclePreconditionVetoed reports that one or more lifecycle
	// precondition callbacks blocked an operation.
	CodePluginLifecyclePreconditionVetoed = lifecycle.CodePluginLifecyclePreconditionVetoed
	// CodePluginDependencyBlocked reports that plugin dependency checks rejected a lifecycle action.
	CodePluginDependencyBlocked = lifecycle.CodePluginDependencyBlocked
	// CodePluginReverseDependencyBlocked reports that installed downstream plugins depend on the target plugin.
	CodePluginReverseDependencyBlocked = lifecycle.CodePluginReverseDependencyBlocked
	// CodePluginForceUninstallDisabled reports that force uninstall is not enabled in host configuration.
	CodePluginForceUninstallDisabled = lifecycle.CodePluginForceUninstallDisabled
	// CodePluginDynamicArtifactMissingForUninstall reports that a dynamic
	// plugin cannot run a full uninstall because both staged and active release
	// artifacts are missing. Operators may use force uninstall to clear only
	// host governance state.
	CodePluginDynamicArtifactMissingForUninstall = lifecycle.CodePluginDynamicArtifactMissingForUninstall
	// CodePluginStartupConsistencyFailed reports invalid persisted plugin or tenant-governance startup state.
	CodePluginStartupConsistencyFailed = bizerr.MustDefine(
		"PLUGIN_STARTUP_CONSISTENCY_FAILED",
		"Plugin startup consistency validation failed: {details}",
		gcode.CodeInternalError,
	)
)
