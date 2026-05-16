// This file defines source-plugin upgrade business error codes and result
// message keys.

package sourceupgrade

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

const (
	// sourceUpgradeNotInstalledSkippedKey identifies the no-op result for uninstalled source plugins.
	sourceUpgradeNotInstalledSkippedKey = "plugin.sourceUpgrade.notInstalledSkipped"
	// sourceUpgradeAlreadyLatestKey identifies the no-op result for already-current source plugins.
	sourceUpgradeAlreadyLatestKey = "plugin.sourceUpgrade.alreadyLatest"
	// sourceUpgradeSuccessKey identifies the successful source-plugin upgrade result.
	sourceUpgradeSuccessKey = "plugin.sourceUpgrade.success"
)

var (
	// CodePluginSourceUpgradeCandidateNotFound reports an inconsistent missing upgrade candidate.
	CodePluginSourceUpgradeCandidateNotFound = bizerr.MustDefine(
		"PLUGIN_SOURCE_UPGRADE_CANDIDATE_NOT_FOUND",
		"Source plugin upgrade candidate does not exist: {pluginId}",
		gcode.CodeNotFound,
	)
	// CodePluginSourceUpgradeDowngradeUnsupported reports that source downgrade is unsupported.
	CodePluginSourceUpgradeDowngradeUnsupported = bizerr.MustDefine(
		"PLUGIN_SOURCE_UPGRADE_DOWNGRADE_UNSUPPORTED",
		"Source plugin {pluginId} effective version {effectiveVersion} is higher than discovered source version {discoveredVersion}. Downgrade or rollback is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodePluginSourceUpgradeTargetReleaseNotFound reports that the target release row is missing.
	CodePluginSourceUpgradeTargetReleaseNotFound = bizerr.MustDefine(
		"PLUGIN_SOURCE_UPGRADE_TARGET_RELEASE_NOT_FOUND",
		"Source plugin upgrade target release record is missing: {pluginId}@{version}",
		gcode.CodeNotFound,
	)
	// CodePluginSourceUpgradeRegistryAfterUpgradeNotFound reports the registry row disappeared after upgrade.
	CodePluginSourceUpgradeRegistryAfterUpgradeNotFound = bizerr.MustDefine(
		"PLUGIN_SOURCE_UPGRADE_REGISTRY_AFTER_UPGRADE_NOT_FOUND",
		"Source plugin registry does not exist after upgrade: {pluginId}",
		gcode.CodeInternalError,
	)
	// CodePluginSourceUpgradePluginIDRequired reports that an upgrade request omitted plugin ID.
	CodePluginSourceUpgradePluginIDRequired = bizerr.MustDefine(
		"PLUGIN_SOURCE_UPGRADE_PLUGIN_ID_REQUIRED",
		"Source plugin ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodePluginSourceUpgradePluginNotFound reports that the requested source plugin was not discovered.
	CodePluginSourceUpgradePluginNotFound = bizerr.MustDefine(
		"PLUGIN_SOURCE_UPGRADE_PLUGIN_NOT_FOUND",
		"Source plugin was not found: {pluginId}",
		gcode.CodeNotFound,
	)
	// CodePluginSourceUpgradeManifestRequired reports that source upgrade requires a manifest.
	CodePluginSourceUpgradeManifestRequired = bizerr.MustDefine(
		"PLUGIN_SOURCE_UPGRADE_MANIFEST_REQUIRED",
		"Source plugin manifest cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodePluginSourceUpgradeRegistryRequired reports that source upgrade requires a registry row.
	CodePluginSourceUpgradeRegistryRequired = bizerr.MustDefine(
		"PLUGIN_SOURCE_UPGRADE_REGISTRY_REQUIRED",
		"Source plugin registry cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodePluginSourceUpgradeTargetReleaseRequired reports that source upgrade requires a target release.
	CodePluginSourceUpgradeTargetReleaseRequired = bizerr.MustDefine(
		"PLUGIN_SOURCE_UPGRADE_TARGET_RELEASE_REQUIRED",
		"Source plugin target release cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodePluginSourceUpgradeLifecycleVetoed reports that a unified lifecycle
	// before-upgrade callback blocked source-plugin runtime upgrade.
	CodePluginSourceUpgradeLifecycleVetoed = bizerr.MustDefine(
		"PLUGIN_SOURCE_UPGRADE_LIFECYCLE_VETOED",
		"Source plugin lifecycle operation {operation} for {pluginId} was blocked: {reasons}",
		gcode.CodeInvalidOperation,
	)
)
