// This file defines plugin upgrade business error codes owned by the unified
// upgrade component while preserving the existing runtime code and fallback
// values expected by callers.

package upgrade

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodePluginNotFound reports that a plugin management query could not find the target plugin.
	CodePluginNotFound = bizerr.MustDefine(
		"PLUGIN_NOT_FOUND",
		"Plugin does not exist: {pluginId}",
		gcode.CodeNotFound,
	)
	// CodePluginRuntimeUpgradePreviewUnavailable reports that no upgrade preview can be produced.
	CodePluginRuntimeUpgradePreviewUnavailable = bizerr.MustDefine(
		"PLUGIN_RUNTIME_UPGRADE_PREVIEW_UNAVAILABLE",
		"Plugin {pluginId} runtime upgrade preview is available only when runtimeState is pending_upgrade or upgrade_failed; current runtimeState={runtimeState}",
		gcode.CodeInvalidOperation,
	)
	// CodePluginRuntimeUpgradeConfirmationRequired reports that an upgrade
	// request omitted the explicit operator confirmation.
	CodePluginRuntimeUpgradeConfirmationRequired = bizerr.MustDefine(
		"PLUGIN_RUNTIME_UPGRADE_CONFIRMATION_REQUIRED",
		"Plugin {pluginId} runtime upgrade requires explicit confirmation",
		gcode.CodeInvalidParameter,
	)
	// CodePluginRuntimeUpgradeUnavailable reports that execution is allowed only
	// for plugins still marked as pending upgrade after server-side state re-read.
	CodePluginRuntimeUpgradeUnavailable = bizerr.MustDefine(
		"PLUGIN_RUNTIME_UPGRADE_UNAVAILABLE",
		"Plugin {pluginId} runtime upgrade is available only when runtimeState is pending_upgrade or upgrade_failed; current runtimeState={runtimeState}",
		gcode.CodeInvalidOperation,
	)
	// CodePluginRuntimeUpgradeLockUnavailable reports that clustered runtime
	// upgrade cannot safely acquire a deployment-wide lock backend.
	CodePluginRuntimeUpgradeLockUnavailable = bizerr.MustDefine(
		"PLUGIN_RUNTIME_UPGRADE_LOCK_UNAVAILABLE",
		"Plugin {pluginId} runtime upgrade requires a cluster lock backend when cluster.enabled=true",
		gcode.CodeInternalError,
	)
	// CodePluginRuntimeUpgradeAlreadyRunning reports that another node already
	// owns the runtime-upgrade lock for the target plugin.
	CodePluginRuntimeUpgradeAlreadyRunning = bizerr.MustDefine(
		"PLUGIN_RUNTIME_UPGRADE_ALREADY_RUNNING",
		"Plugin {pluginId} runtime upgrade is already running on another node",
		gcode.CodeInvalidOperation,
	)
	// CodePluginRuntimeUpgradeTypeUnsupported reports that the target plugin type
	// cannot run through the runtime upgrade executor.
	CodePluginRuntimeUpgradeTypeUnsupported = bizerr.MustDefine(
		"PLUGIN_RUNTIME_UPGRADE_TYPE_UNSUPPORTED",
		"Plugin {pluginId} with type {pluginType} does not support runtime upgrade execution",
		gcode.CodeInvalidParameter,
	)
	// CodePluginRuntimeUpgradeExecutionFailed reports a failed explicit upgrade
	// after the request passed confirmation and state validation.
	CodePluginRuntimeUpgradeExecutionFailed = bizerr.MustDefine(
		"PLUGIN_RUNTIME_UPGRADE_EXECUTION_FAILED",
		"Plugin {pluginId} runtime upgrade from {fromVersion} to {toVersion} failed",
		gcode.CodeInternalError,
	)
	// CodePluginNotInstalled reports that a lifecycle operation requires an installed plugin.
	CodePluginNotInstalled = bizerr.MustDefine(
		"PLUGIN_NOT_INSTALLED",
		"Plugin is not installed",
		gcode.CodeInvalidParameter,
	)
	// CodePluginLifecyclePreconditionVetoed reports that one or more lifecycle
	// precondition callbacks blocked an operation.
	CodePluginLifecyclePreconditionVetoed = bizerr.MustDefine(
		"PLUGIN_LIFECYCLE_PRECONDITION_VETOED",
		"Plugin lifecycle operation {operation} for {pluginId} was blocked by lifecycle preconditions: {reasons}",
		gcode.CodeInvalidOperation,
	)
	// CodePluginReleaseNotFound reports that a plugin release row is missing.
	CodePluginReleaseNotFound = bizerr.MustDefine(
		"PLUGIN_RELEASE_NOT_FOUND",
		"Plugin release record does not exist: {pluginId}@{version}",
		gcode.CodeNotFound,
	)
	// CodePluginDependencyBlocked reports that plugin dependency checks rejected an upgrade action.
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
	// CodePluginRuntimeUpgradeSnapshotMissing reports missing effective or target manifest snapshot data.
	CodePluginRuntimeUpgradeSnapshotMissing = bizerr.MustDefine(
		"PLUGIN_RUNTIME_UPGRADE_SNAPSHOT_MISSING",
		"Plugin {pluginId}@{version} runtime upgrade manifest snapshot is missing",
		gcode.CodeInternalError,
	)
	// CodePluginRuntimeUpgradeSnapshotInvalid reports invalid upgrade preview snapshot metadata.
	CodePluginRuntimeUpgradeSnapshotInvalid = bizerr.MustDefine(
		"PLUGIN_RUNTIME_UPGRADE_SNAPSHOT_INVALID",
		"Plugin {pluginId}@{version} runtime upgrade manifest snapshot is invalid",
		gcode.CodeInternalError,
	)
)
