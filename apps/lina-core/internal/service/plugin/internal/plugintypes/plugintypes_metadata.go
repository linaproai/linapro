// This file defines governance, release, runtime-upgrade, migration, and
// resource value objects shared across plugin sub-components.

package plugintypes

// MigrationDirection defines the install or uninstall phase persisted in migration records.
type MigrationDirection string

// ReleaseStatus defines the normalized release status persisted in sys_plugin_release.
type ReleaseStatus string

// MigrationExecutionStatus defines the migration execution result persisted in sys_plugin_migration.
type MigrationExecutionStatus string

// ResourceKind defines the abstract governance resource category indexed in sys_plugin_resource_ref.
type ResourceKind string

// ResourceOwnerType defines the abstract owner category indexed in sys_plugin_resource_ref.
type ResourceOwnerType string

// NodeState defines the current node-state projection enum.
type NodeState string

// HostState defines the desired/current host lifecycle state enum.
type HostState string

// RuntimeUpgradeState identifies whether discovered plugin files match the effective host registry state.
type RuntimeUpgradeState string

// RuntimeUpgradeAbnormalReason identifies why a plugin cannot be treated as normally upgradeable.
type RuntimeUpgradeAbnormalReason string

// RuntimeUpgradeFailurePhase identifies the upgrade phase associated with the latest observable failure.
type RuntimeUpgradeFailurePhase string

// LifecycleState defines the lifecycle summary enum exposed by plugin governance.
type LifecycleState string

// MigrationState defines the review-friendly migration state enum.
type MigrationState string

const (
	// MigrationDirectionInstall identifies an install SQL migration phase.
	MigrationDirectionInstall MigrationDirection = "install"
	// MigrationDirectionUninstall identifies an uninstall SQL migration phase.
	MigrationDirectionUninstall MigrationDirection = "uninstall"
	// MigrationDirectionUpgrade identifies an upgrade SQL migration phase.
	MigrationDirectionUpgrade MigrationDirection = "upgrade"
	// MigrationDirectionRollback identifies a rollback SQL migration phase.
	MigrationDirectionRollback MigrationDirection = "rollback"
	// MigrationDirectionMock identifies the optional install-time mock data load phase.
	MigrationDirectionMock MigrationDirection = "mock"

	// MigrationStatusFailed is the persisted integer failed marker.
	MigrationStatusFailed = 0
	// MigrationStatusSucceeded is the persisted integer succeeded marker.
	MigrationStatusSucceeded = 1

	// ReleaseStatusPrepared marks a staged release.
	ReleaseStatusPrepared ReleaseStatus = "prepared"
	// ReleaseStatusUninstalled marks an uninstalled release.
	ReleaseStatusUninstalled ReleaseStatus = "uninstalled"
	// ReleaseStatusInstalled marks an installed but disabled release.
	ReleaseStatusInstalled ReleaseStatus = "installed"
	// ReleaseStatusActive marks an enabled release.
	ReleaseStatusActive ReleaseStatus = "active"
	// ReleaseStatusFailed marks a release that failed before becoming effective.
	ReleaseStatusFailed ReleaseStatus = "failed"

	// MigrationExecutionStatusSucceeded marks a successful migration row.
	MigrationExecutionStatusSucceeded MigrationExecutionStatus = "succeeded"
	// MigrationExecutionStatusFailed marks a failed migration row.
	MigrationExecutionStatusFailed MigrationExecutionStatus = "failed"

	// ResourceKindManifest identifies the plugin manifest resource.
	ResourceKindManifest ResourceKind = "manifest"
	// ResourceKindBackendEntry identifies source backend registration.
	ResourceKindBackendEntry ResourceKind = "backend_entry"
	// ResourceKindRuntimeWasm identifies a runtime WASM artifact.
	ResourceKindRuntimeWasm ResourceKind = "runtime_wasm"
	// ResourceKindRuntimeFrontend identifies runtime frontend assets.
	ResourceKindRuntimeFrontend ResourceKind = "runtime_frontend"
	// ResourceKindFrontendPage identifies source frontend pages.
	ResourceKindFrontendPage ResourceKind = "frontend_page"
	// ResourceKindFrontendSlot identifies source frontend slots.
	ResourceKindFrontendSlot ResourceKind = "frontend_slot"
	// ResourceKindMenu identifies plugin menus.
	ResourceKindMenu ResourceKind = "menu"
	// ResourceKindInstallSQL identifies install SQL resources.
	ResourceKindInstallSQL ResourceKind = "install_sql"
	// ResourceKindUninstallSQL identifies uninstall SQL resources.
	ResourceKindUninstallSQL ResourceKind = "uninstall_sql"
	// ResourceKindMockSQL identifies mock-data SQL resources.
	ResourceKindMockSQL ResourceKind = "mock_sql"
	// ResourceKindHostStorage identifies host storage service resources.
	ResourceKindHostStorage ResourceKind = "host_storage"
	// ResourceKindHostUpstream identifies host upstream service resources.
	ResourceKindHostUpstream ResourceKind = "host_upstream"
	// ResourceKindHostData identifies host data table resources.
	ResourceKindHostData ResourceKind = "host_data_table"
	// ResourceKindHostCache identifies host cache resources.
	ResourceKindHostCache ResourceKind = "host_cache"
	// ResourceKindHostLock identifies host lock resources.
	ResourceKindHostLock ResourceKind = "host_lock"
	// ResourceKindHostSecret identifies host secret resources.
	ResourceKindHostSecret ResourceKind = "host_secret"
	// ResourceKindHostEventTopic identifies host event topic resources.
	ResourceKindHostEventTopic ResourceKind = "host_event_topic"
	// ResourceKindHostQueue identifies host queue resources.
	ResourceKindHostQueue ResourceKind = "host_queue"
	// ResourceKindHostNotify identifies host notify channel resources.
	ResourceKindHostNotify ResourceKind = "host_notify_channel"

	// ResourceOwnerTypeFile identifies a file-backed resource owner.
	ResourceOwnerTypeFile ResourceOwnerType = "file"
	// ResourceOwnerTypeBackendRegistration identifies backend registration.
	ResourceOwnerTypeBackendRegistration ResourceOwnerType = "backend-registration"
	// ResourceOwnerTypeRuntimeArtifact identifies runtime artifacts.
	ResourceOwnerTypeRuntimeArtifact ResourceOwnerType = "runtime-artifact"
	// ResourceOwnerTypeRuntimeFrontend identifies runtime frontend entries.
	ResourceOwnerTypeRuntimeFrontend ResourceOwnerType = "runtime-frontend"
	// ResourceOwnerTypeInstallSQL identifies install SQL entries.
	ResourceOwnerTypeInstallSQL ResourceOwnerType = "install-sql"
	// ResourceOwnerTypeUninstallSQL identifies uninstall SQL entries.
	ResourceOwnerTypeUninstallSQL ResourceOwnerType = "uninstall-sql"
	// ResourceOwnerTypeMockSQL identifies mock SQL entries.
	ResourceOwnerTypeMockSQL ResourceOwnerType = "mock-sql"
	// ResourceOwnerTypeFrontendPageEntry identifies frontend page entries.
	ResourceOwnerTypeFrontendPageEntry ResourceOwnerType = "frontend-page-entry"
	// ResourceOwnerTypeFrontendSlotEntry identifies frontend slot entries.
	ResourceOwnerTypeFrontendSlotEntry ResourceOwnerType = "frontend-slot-entry"
	// ResourceOwnerTypeMenuEntry identifies menu entries.
	ResourceOwnerTypeMenuEntry ResourceOwnerType = "menu-entry"
	// ResourceOwnerTypeHostServiceResource identifies host-service governed resources.
	ResourceOwnerTypeHostServiceResource ResourceOwnerType = "host-service-resource"

	// NodeStateReconciling marks a node reconciling state.
	NodeStateReconciling NodeState = "reconciling"
	// NodeStateFailed marks a failed node state.
	NodeStateFailed NodeState = "failed"
	// NodeStateEnabled marks an enabled node state.
	NodeStateEnabled NodeState = "enabled"
	// NodeStateInstalled marks an installed node state.
	NodeStateInstalled NodeState = "installed"
	// NodeStateUninstalled marks an uninstalled node state.
	NodeStateUninstalled NodeState = "uninstalled"

	// HostStateReconciling marks a reconciling host state.
	HostStateReconciling HostState = "reconciling"
	// HostStateFailed marks a failed host state.
	HostStateFailed HostState = "failed"
	// HostStateEnabled marks an enabled host state.
	HostStateEnabled HostState = "enabled"
	// HostStateInstalled marks an installed host state.
	HostStateInstalled HostState = "installed"
	// HostStateUninstalled marks an uninstalled host state.
	HostStateUninstalled HostState = "uninstalled"

	// RuntimeUpgradeStateNormal means effective and discovered metadata are aligned.
	RuntimeUpgradeStateNormal RuntimeUpgradeState = "normal"
	// RuntimeUpgradeStatePendingUpgrade means discovered files are newer than the effective version.
	RuntimeUpgradeStatePendingUpgrade RuntimeUpgradeState = "pending_upgrade"
	// RuntimeUpgradeStateAbnormal means discovered files are older or cannot be safely compared.
	RuntimeUpgradeStateAbnormal RuntimeUpgradeState = "abnormal"
	// RuntimeUpgradeStateUpgradeRunning means a runtime upgrade transition is reconciling.
	RuntimeUpgradeStateUpgradeRunning RuntimeUpgradeState = "upgrade_running"
	// RuntimeUpgradeStateUpgradeFailed means the latest target release failed before becoming effective.
	RuntimeUpgradeStateUpgradeFailed RuntimeUpgradeState = "upgrade_failed"

	// RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective means the file version is lower than DB version.
	RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective RuntimeUpgradeAbnormalReason = "discovered_version_lower_than_effective"
	// RuntimeUpgradeAbnormalReasonVersionCompareFailed means at least one version string is invalid.
	RuntimeUpgradeAbnormalReasonVersionCompareFailed RuntimeUpgradeAbnormalReason = "version_compare_failed"

	// RuntimeUpgradeFailurePhaseRelease identifies a release-state failure.
	RuntimeUpgradeFailurePhaseRelease RuntimeUpgradeFailurePhase = "release"
	// RuntimeUpgradeFailurePhaseBeforeUpgrade identifies a before-upgrade callback failure.
	RuntimeUpgradeFailurePhaseBeforeUpgrade RuntimeUpgradeFailurePhase = "before_upgrade"
	// RuntimeUpgradeFailurePhaseUpgradeCallback identifies an upgrade callback failure.
	RuntimeUpgradeFailurePhaseUpgradeCallback RuntimeUpgradeFailurePhase = "upgrade_callback"
	// RuntimeUpgradeFailurePhaseSQL identifies a SQL migration failure.
	RuntimeUpgradeFailurePhaseSQL RuntimeUpgradeFailurePhase = "sql"
	// RuntimeUpgradeFailurePhaseGovernance identifies a governance sync failure.
	RuntimeUpgradeFailurePhaseGovernance RuntimeUpgradeFailurePhase = "governance"
	// RuntimeUpgradeFailurePhaseReleaseSwitch identifies a release switch failure.
	RuntimeUpgradeFailurePhaseReleaseSwitch RuntimeUpgradeFailurePhase = "release_switch"
	// RuntimeUpgradeFailurePhaseCacheInvalidation identifies a cache invalidation failure.
	RuntimeUpgradeFailurePhaseCacheInvalidation RuntimeUpgradeFailurePhase = "cache_invalidation"

	// LifecycleStateSourceEnabled identifies an enabled source plugin.
	LifecycleStateSourceEnabled LifecycleState = "source_enabled"
	// LifecycleStateSourceDisabled identifies a disabled source plugin.
	LifecycleStateSourceDisabled LifecycleState = "source_disabled"
	// LifecycleStateRuntimeUninstalled identifies an uninstalled runtime plugin.
	LifecycleStateRuntimeUninstalled LifecycleState = "runtime_uninstalled"
	// LifecycleStateRuntimeInstalled identifies an installed runtime plugin.
	LifecycleStateRuntimeInstalled LifecycleState = "runtime_installed"
	// LifecycleStateRuntimeEnabled identifies an enabled runtime plugin.
	LifecycleStateRuntimeEnabled LifecycleState = "runtime_enabled"

	// MigrationStateNone indicates no migration row exists.
	MigrationStateNone MigrationState = "none"
	// MigrationStateSucceeded indicates the latest migration succeeded.
	MigrationStateSucceeded MigrationState = "succeeded"
	// MigrationStateFailed indicates the latest migration failed.
	MigrationStateFailed MigrationState = "failed"
)

// String returns the canonical migration direction value.
func (value MigrationDirection) String() string { return string(value) }

// String returns the canonical release status value.
func (value ReleaseStatus) String() string { return string(value) }

// String returns the canonical migration execution status value.
func (value MigrationExecutionStatus) String() string { return string(value) }

// String returns the canonical resource kind value.
func (value ResourceKind) String() string { return string(value) }

// String returns the canonical resource owner-type value.
func (value ResourceOwnerType) String() string { return string(value) }

// String returns the canonical node-state value.
func (value NodeState) String() string { return string(value) }

// String returns the canonical host-state value.
func (value HostState) String() string { return string(value) }

// String returns the canonical runtime-upgrade state value.
func (value RuntimeUpgradeState) String() string { return string(value) }

// String returns the canonical runtime-upgrade abnormal reason value.
func (value RuntimeUpgradeAbnormalReason) String() string { return string(value) }

// String returns the canonical runtime-upgrade failure phase value.
func (value RuntimeUpgradeFailurePhase) String() string { return string(value) }

// String returns the canonical lifecycle-state value.
func (value LifecycleState) String() string { return string(value) }

// String returns the canonical migration-state value.
func (value MigrationState) String() string { return string(value) }

// ResourceRefDescriptor represents one governance resource index entry derived
// from the current plugin release.
type ResourceRefDescriptor struct {
	Kind      ResourceKind
	Key       string
	OwnerType ResourceOwnerType
	OwnerKey  string
	Remark    string
}

// RuntimeUpgradeFailure exposes the latest observable runtime-upgrade failure.
type RuntimeUpgradeFailure struct {
	// Phase is the upgrade phase associated with the failure.
	Phase RuntimeUpgradeFailurePhase
	// ErrorCode is a stable machine-readable failure code.
	ErrorCode string
	// MessageKey is the i18n key that management clients can render.
	MessageKey string
	// ReleaseID identifies the failed target release when known.
	ReleaseID int
	// ReleaseVersion identifies the failed target release version when known.
	ReleaseVersion string
	// Detail carries the latest persisted failure detail for operator diagnosis.
	Detail string
}

// RuntimeUpgradeProjection is the flattened version-drift state for one plugin.
type RuntimeUpgradeProjection struct {
	// State is the current runtime-upgrade state.
	State RuntimeUpgradeState
	// EffectiveVersion is the version currently active in sys_plugin.
	EffectiveVersion string
	// DiscoveredVersion is the version currently discovered from plugin files.
	DiscoveredVersion string
	// UpgradeAvailable reports whether a user can attempt a runtime upgrade.
	UpgradeAvailable bool
	// AbnormalReason stores a stable reason code when State is abnormal.
	AbnormalReason RuntimeUpgradeAbnormalReason
	// LastFailure stores the latest observable failed target release.
	LastFailure *RuntimeUpgradeFailure
}

// DeriveNodeState converts installation and enablement flags into one stable node-state key.
func DeriveNodeState(installed int, enabled int) string {
	if NormalizeInstalledStatus(installed) != PluginInstalledYes {
		return NodeStateUninstalled.String()
	}
	if NormalizeStatus(enabled) == PluginStatusEnabled {
		return NodeStateEnabled.String()
	}
	return NodeStateInstalled.String()
}

// DeriveHostState converts install and enablement flags into the stable host lifecycle state.
func DeriveHostState(installed int, enabled int) string {
	if NormalizeInstalledStatus(installed) != PluginInstalledYes {
		return HostStateUninstalled.String()
	}
	if NormalizeStatus(enabled) == PluginStatusEnabled {
		return HostStateEnabled.String()
	}
	return HostStateInstalled.String()
}

// DeriveLifecycleState converts plugin type and runtime flags into a lifecycle state.
func DeriveLifecycleState(pluginType string, installed int, enabled int) string {
	if NormalizeType(pluginType) == TypeSource {
		if NormalizeStatus(enabled) == PluginStatusEnabled {
			return LifecycleStateSourceEnabled.String()
		}
		return LifecycleStateSourceDisabled.String()
	}
	if NormalizeInstalledStatus(installed) != PluginInstalledYes {
		return LifecycleStateRuntimeUninstalled.String()
	}
	if NormalizeStatus(enabled) == PluginStatusEnabled {
		return LifecycleStateRuntimeEnabled.String()
	}
	return LifecycleStateRuntimeInstalled.String()
}
