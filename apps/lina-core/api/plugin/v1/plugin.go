package v1

// PluginType identifies the host plugin implementation family.
type PluginType string

// Supported plugin types.
const (
	PluginTypeSource  PluginType = "source"
	PluginTypeDynamic PluginType = "dynamic"
)

// String returns the serialized plugin type value.
func (value PluginType) String() string { return string(value) }

// PluginDistribution identifies the host governance model for plugin delivery.
type PluginDistribution string

// Supported plugin distribution governance values.
const (
	PluginDistributionManaged PluginDistribution = "managed"
	PluginDistributionBuiltin PluginDistribution = "builtin"
)

// String returns the serialized plugin distribution value.
func (value PluginDistribution) String() string { return string(value) }

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

// String returns the serialized runtime state value.
func (value RuntimeState) String() string { return string(value) }

// RuntimeAbnormalReason identifies why a plugin cannot be treated as normally upgradeable.
type RuntimeAbnormalReason string

// Supported runtime abnormal reasons.
const (
	RuntimeAbnormalReasonDiscoveredVersionLowerThanEffective RuntimeAbnormalReason = "discovered_version_lower_than_effective"
	RuntimeAbnormalReasonVersionCompareFailed                RuntimeAbnormalReason = "version_compare_failed"
)

// String returns the serialized runtime abnormal reason value.
func (value RuntimeAbnormalReason) String() string { return string(value) }

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

// String returns the serialized runtime failure phase value.
func (value RuntimeFailurePhase) String() string { return string(value) }

// ScopeNature defines how a plugin participates in tenant governance.
type ScopeNature string

// Supported plugin scope natures.
const (
	ScopeNaturePlatformOnly ScopeNature = "platform_only"
	ScopeNatureTenantAware  ScopeNature = "tenant_aware"
)

// String returns the serialized scope nature value.
func (value ScopeNature) String() string { return string(value) }

// InstallMode defines how a tenant-aware plugin is enabled across tenants.
type InstallMode string

// Supported plugin install modes.
const (
	InstallModeGlobal       InstallMode = "global"
	InstallModeTenantScoped InstallMode = "tenant_scoped"
)

// String returns the serialized install mode value.
func (value InstallMode) String() string { return string(value) }

// AuthorizationStatus identifies host-service authorization review state.
type AuthorizationStatus string

// Supported host-service authorization states.
const (
	AuthorizationStatusNotRequired AuthorizationStatus = "not_required"
	AuthorizationStatusPending     AuthorizationStatus = "pending"
	AuthorizationStatusConfirmed   AuthorizationStatus = "confirmed"
)

// String returns the serialized authorization status value.
func (value AuthorizationStatus) String() string { return string(value) }

// DependencyStatus identifies one plugin dependency edge state.
type DependencyStatus string

// Supported dependency edge states.
const (
	DependencyStatusSatisfied          DependencyStatus = "satisfied"
	DependencyStatusMissing            DependencyStatus = "missing"
	DependencyStatusVersionUnsatisfied DependencyStatus = "version_unsatisfied"
	DependencyStatusNotEnabled         DependencyStatus = "not_enabled"
)

// String returns the serialized dependency status value.
func (value DependencyStatus) String() string { return string(value) }

// FrameworkStatus identifies framework-version compatibility.
type FrameworkStatus string

// Supported framework compatibility states.
const (
	FrameworkStatusNotDeclared FrameworkStatus = "not_declared"
	FrameworkStatusSatisfied   FrameworkStatus = "satisfied"
	FrameworkStatusUnsatisfied FrameworkStatus = "unsatisfied"
)

// String returns the serialized framework status value.
func (value FrameworkStatus) String() string { return string(value) }

// BlockerCode identifies one plugin dependency check failure category.
type BlockerCode string

// Supported dependency blocker categories.
const (
	BlockerCodeFrameworkVersionUnsatisfied  BlockerCode = "framework_version_unsatisfied"
	BlockerCodeDependencyMissing            BlockerCode = "dependency_missing"
	BlockerCodeDependencyVersionUnsatisfied BlockerCode = "dependency_version_unsatisfied"
	BlockerCodeDependencyNotEnabled         BlockerCode = "dependency_not_enabled"
	BlockerCodeDependencyCycle              BlockerCode = "dependency_cycle"
	BlockerCodeDependencySnapshotUnknown    BlockerCode = "dependency_snapshot_unknown"
	BlockerCodeReverseDependency            BlockerCode = "reverse_dependency"
	BlockerCodeReverseEnabledDependency     BlockerCode = "reverse_enabled_dependency"
	BlockerCodeReverseDependencyVersion     BlockerCode = "reverse_dependency_version"
)

// String returns the serialized blocker code value.
func (value BlockerCode) String() string { return string(value) }
