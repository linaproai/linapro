package v1

// PluginType identifies the host plugin implementation family.
type PluginType string

// Supported plugin types.
const (
	PluginTypeSource  PluginType = "source"
	PluginTypeDynamic PluginType = "dynamic"
)

// RuntimeState identifies whether discovered plugin files match effective state.
type RuntimeState string

// Supported plugin runtime-upgrade states.
const (
	RuntimeStateNormal         RuntimeState = "normal"
	RuntimeStatePendingUpgrade RuntimeState = "pending_upgrade"
	RuntimeStateAbnormal       RuntimeState = "abnormal"
	RuntimeStateUpgradeRunning RuntimeState = "upgrade_running"
	RuntimeStateUpgradeFailed  RuntimeState = "upgrade_failed"
)

// RuntimeAbnormalReason identifies why a plugin cannot be treated as normally upgradeable.
type RuntimeAbnormalReason string

// Supported runtime abnormal reasons.
const (
	RuntimeAbnormalReasonDiscoveredVersionLowerThanEffective RuntimeAbnormalReason = "discovered_version_lower_than_effective"
	RuntimeAbnormalReasonVersionCompareFailed                RuntimeAbnormalReason = "version_compare_failed"
)

// RuntimeFailurePhase identifies the phase associated with the latest failure.
type RuntimeFailurePhase string

// Supported runtime upgrade failure phases.
const (
	RuntimeFailurePhaseRelease           RuntimeFailurePhase = "release"
	RuntimeFailurePhaseBeforeUpgrade     RuntimeFailurePhase = "before_upgrade"
	RuntimeFailurePhaseUpgradeCallback   RuntimeFailurePhase = "upgrade_callback"
	RuntimeFailurePhaseSQL               RuntimeFailurePhase = "sql"
	RuntimeFailurePhaseGovernance        RuntimeFailurePhase = "governance"
	RuntimeFailurePhaseReleaseSwitch     RuntimeFailurePhase = "release_switch"
	RuntimeFailurePhaseCacheInvalidation RuntimeFailurePhase = "cache_invalidation"
)

// ScopeNature defines how a plugin participates in tenant governance.
type ScopeNature string

// Supported plugin scope natures.
const (
	ScopeNaturePlatformOnly ScopeNature = "platform_only"
	ScopeNatureTenantAware  ScopeNature = "tenant_aware"
)

// InstallMode defines how a tenant-aware plugin is enabled across tenants.
type InstallMode string

// Supported plugin install modes.
const (
	InstallModeGlobal       InstallMode = "global"
	InstallModeTenantScoped InstallMode = "tenant_scoped"
)

// DependencyInstallMode defines how a missing dependency may be installed.
type DependencyInstallMode string

// Supported dependency install modes.
const (
	DependencyInstallModeManual DependencyInstallMode = "manual"
	DependencyInstallModeAuto   DependencyInstallMode = "auto"
)

// AuthorizationStatus identifies host-service authorization review state.
type AuthorizationStatus string

// Supported host-service authorization states.
const (
	AuthorizationStatusNotRequired AuthorizationStatus = "not_required"
	AuthorizationStatusPending     AuthorizationStatus = "pending"
	AuthorizationStatusConfirmed   AuthorizationStatus = "confirmed"
)

// DependencyStatus identifies one plugin dependency edge state.
type DependencyStatus string

// Supported dependency edge states.
const (
	DependencyStatusSatisfied             DependencyStatus = "satisfied"
	DependencyStatusAutoInstallPlanned    DependencyStatus = "auto_install_planned"
	DependencyStatusManualInstallRequired DependencyStatus = "manual_install_required"
	DependencyStatusMissing               DependencyStatus = "missing"
	DependencyStatusVersionUnsatisfied    DependencyStatus = "version_unsatisfied"
	DependencyStatusSoftUnsatisfied       DependencyStatus = "soft_unsatisfied"
)

// FrameworkStatus identifies framework-version compatibility.
type FrameworkStatus string

// Supported framework compatibility states.
const (
	FrameworkStatusNotDeclared FrameworkStatus = "not_declared"
	FrameworkStatusSatisfied   FrameworkStatus = "satisfied"
	FrameworkStatusUnsatisfied FrameworkStatus = "unsatisfied"
)

// BlockerCode identifies one plugin dependency check failure category.
type BlockerCode string

// Supported dependency blocker categories.
const (
	BlockerCodeFrameworkVersionUnsatisfied  BlockerCode = "framework_version_unsatisfied"
	BlockerCodeDependencyMissing            BlockerCode = "dependency_missing"
	BlockerCodeDependencyVersionUnsatisfied BlockerCode = "dependency_version_unsatisfied"
	BlockerCodeDependencyManualRequired     BlockerCode = "dependency_manual_install_required"
	BlockerCodeDependencyCycle              BlockerCode = "dependency_cycle"
	BlockerCodeDependencySnapshotUnknown    BlockerCode = "dependency_snapshot_unknown"
	BlockerCodeReverseDependency            BlockerCode = "reverse_dependency"
	BlockerCodeReverseDependencyVersion     BlockerCode = "reverse_dependency_version"
)
