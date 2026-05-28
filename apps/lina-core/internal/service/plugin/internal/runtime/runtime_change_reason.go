// This file centralizes internal dynamic-plugin runtime change reason values
// used for cache invalidation, reconciler wake-ups, and runtime revision
// diagnostics. Keep these values stable because external cache coordinators and
// logs may persist them as plain strings even though runtime code uses the
// named type.

package runtime

// runtimeChangeReason identifies why the dynamic-plugin runtime cache or
// reconciler revision changed.
type runtimeChangeReason string

const (
	// runtimeChangeReasonPluginInstalled means a dynamic plugin release became installed.
	runtimeChangeReasonPluginInstalled runtimeChangeReason = "plugin_installed"
	// runtimeChangeReasonPluginUpgraded means an installed dynamic plugin switched release versions.
	runtimeChangeReasonPluginUpgraded runtimeChangeReason = "plugin_upgraded"
	// runtimeChangeReasonPluginDisabled means an installed dynamic plugin became disabled.
	runtimeChangeReasonPluginDisabled runtimeChangeReason = "plugin_disabled"
	// runtimeChangeReasonPluginEnabled means an installed dynamic plugin became enabled.
	runtimeChangeReasonPluginEnabled runtimeChangeReason = "plugin_enabled"
	// runtimeChangeReasonPluginStatusChanged means a dynamic plugin status toggle completed.
	runtimeChangeReasonPluginStatusChanged runtimeChangeReason = "plugin_status_changed"
	// runtimeChangeReasonPluginRefreshed means an installed dynamic plugin release was refreshed.
	runtimeChangeReasonPluginRefreshed runtimeChangeReason = "plugin_refreshed"
	// runtimeChangeReasonPluginUninstalled means a dynamic plugin was uninstalled using its manifest.
	runtimeChangeReasonPluginUninstalled runtimeChangeReason = "plugin_uninstalled"
	// runtimeChangeReasonPluginOrphanUninstalled means host governance was force-cleared for a missing artifact.
	runtimeChangeReasonPluginOrphanUninstalled runtimeChangeReason = "plugin_orphan_uninstalled"
	// runtimeChangeReasonRuntimePackageUploaded means a dynamic runtime package was uploaded.
	runtimeChangeReasonRuntimePackageUploaded runtimeChangeReason = "runtime_package_uploaded"
	// runtimeChangeReasonDynamicPackageUploaded means a dynamic package upload changed desired reconciliation input.
	runtimeChangeReasonDynamicPackageUploaded runtimeChangeReason = "dynamic_package_uploaded"
	// runtimeChangeReasonRuntimeArtifactMissing means an active runtime artifact disappeared and registry state changed.
	runtimeChangeReasonRuntimeArtifactMissing runtimeChangeReason = "runtime_artifact_missing"
	// runtimeChangeReasonStaleReconcilingRestored means an abandoned transient
	// reconcile state was restored to a stable state.
	runtimeChangeReasonStaleReconcilingRestored runtimeChangeReason = "stale_reconciling_restored"
	// runtimeChangeReasonDesiredStateChanged means a management request changed one plugin target state.
	runtimeChangeReasonDesiredStateChanged runtimeChangeReason = "desired_state_changed"
	// runtimeChangeReasonSingleNode means single-node mode should run reconciliation every tick.
	runtimeChangeReasonSingleNode runtimeChangeReason = "single_node"
	// runtimeChangeReasonRevisionChanged means the shared reconciler revision advanced.
	runtimeChangeReasonRevisionChanged runtimeChangeReason = "revision_changed"
	// runtimeChangeReasonSafetySweep means the low-frequency safety scan interval elapsed.
	runtimeChangeReasonSafetySweep runtimeChangeReason = "safety_sweep"
)
