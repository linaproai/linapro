// This file defines exported metadata constants together with lightweight
// snapshot and descriptor models used by plugin governance persistence.

package catalog

import "lina-core/pkg/pluginbridge"

// MigrationDirection defines the install or uninstall phase persisted in migration records.
type MigrationDirection string

// ReleaseStatus defines the normalized release status persisted in sys_plugin_release.
type ReleaseStatus string

// MigrationExecutionStatus defines the migration execution result persisted in sys_plugin_migration.
type MigrationExecutionStatus string

// ResourceKind defines the abstract governance resource category indexed in
// sys_plugin_resource_ref.
type ResourceKind string

// ResourceOwnerType defines the abstract owner category indexed in
// sys_plugin_resource_ref.
type ResourceOwnerType string

// NodeState defines the current node-state projection enum.
type NodeState string

// HostState defines the desired/current host lifecycle state enum.
type HostState string

// LifecycleState defines the lifecycle summary enum exposed by plugin governance.
type LifecycleState string

// MigrationState defines the review-friendly migration state enum.
type MigrationState string

// ResourceSpecType defines the supported plugin backend resource declaration type.
type ResourceSpecType string

// ResourceFilterOperator defines supported resource filter operators.
type ResourceFilterOperator string

// ResourceOrderDirection defines supported ordering directions in resource specs.
type ResourceOrderDirection string

// ResourceOperation defines the supported structured data operations for one resource.
type ResourceOperation string

// ResourceAccessMode defines which execution contexts may invoke one resource.
type ResourceAccessMode string

const (
	// MigrationDirection values.
	MigrationDirectionInstall   MigrationDirection = "install"
	MigrationDirectionUninstall MigrationDirection = "uninstall"
	MigrationDirectionUpgrade   MigrationDirection = "upgrade"
	MigrationDirectionRollback  MigrationDirection = "rollback"

	// Migration execution status sentinel values.
	MigrationStatusFailed    = 0
	MigrationStatusSucceeded = 1

	// ReleaseStatus values.
	ReleaseStatusPrepared    ReleaseStatus = "prepared"
	ReleaseStatusUninstalled ReleaseStatus = "uninstalled"
	ReleaseStatusInstalled   ReleaseStatus = "installed"
	ReleaseStatusActive      ReleaseStatus = "active"
	ReleaseStatusFailed      ReleaseStatus = "failed"

	// MigrationExecutionStatus values.
	MigrationExecutionStatusSucceeded MigrationExecutionStatus = "succeeded"
	MigrationExecutionStatusFailed    MigrationExecutionStatus = "failed"

	// ResourceKind values.
	ResourceKindManifest        ResourceKind = "manifest"
	ResourceKindBackendEntry    ResourceKind = "backend_entry"
	ResourceKindRuntimeWasm     ResourceKind = "runtime_wasm"
	ResourceKindRuntimeFrontend ResourceKind = "runtime_frontend"
	ResourceKindFrontendPage    ResourceKind = "frontend_page"
	ResourceKindFrontendSlot    ResourceKind = "frontend_slot"
	ResourceKindMenu            ResourceKind = "menu"
	ResourceKindInstallSQL      ResourceKind = "install_sql"
	ResourceKindUninstallSQL    ResourceKind = "uninstall_sql"
	ResourceKindHostStorage     ResourceKind = "host_storage"
	ResourceKindHostUpstream    ResourceKind = "host_upstream"
	ResourceKindHostData        ResourceKind = "host_data_table"
	ResourceKindHostCache       ResourceKind = "host_cache"
	ResourceKindHostLock        ResourceKind = "host_lock"
	ResourceKindHostSecret      ResourceKind = "host_secret"
	ResourceKindHostEventTopic  ResourceKind = "host_event_topic"
	ResourceKindHostQueue       ResourceKind = "host_queue"
	ResourceKindHostNotify      ResourceKind = "host_notify_channel"

	// ResourceOwnerType values.
	ResourceOwnerTypeFile                ResourceOwnerType = "file"
	ResourceOwnerTypeBackendRegistration ResourceOwnerType = "backend-registration"
	ResourceOwnerTypeRuntimeArtifact     ResourceOwnerType = "runtime-artifact"
	ResourceOwnerTypeRuntimeFrontend     ResourceOwnerType = "runtime-frontend"
	ResourceOwnerTypeInstallSQL          ResourceOwnerType = "install-sql"
	ResourceOwnerTypeUninstallSQL        ResourceOwnerType = "uninstall-sql"
	ResourceOwnerTypeFrontendPageEntry   ResourceOwnerType = "frontend-page-entry"
	ResourceOwnerTypeFrontendSlotEntry   ResourceOwnerType = "frontend-slot-entry"
	ResourceOwnerTypeMenuEntry           ResourceOwnerType = "menu-entry"
	ResourceOwnerTypeHostServiceResource ResourceOwnerType = "host-service-resource"

	// NodeState values.
	NodeStateReconciling NodeState = "reconciling"
	NodeStateFailed      NodeState = "failed"
	NodeStateEnabled     NodeState = "enabled"
	NodeStateInstalled   NodeState = "installed"
	NodeStateUninstalled NodeState = "uninstalled"

	// HostState values.
	HostStateReconciling HostState = "reconciling"
	HostStateFailed      HostState = "failed"
	HostStateEnabled     HostState = "enabled"
	HostStateInstalled   HostState = "installed"
	HostStateUninstalled HostState = "uninstalled"

	// LifecycleState values.
	LifecycleStateSourceEnabled      LifecycleState = "source_enabled"
	LifecycleStateSourceDisabled     LifecycleState = "source_disabled"
	LifecycleStateRuntimeUninstalled LifecycleState = "runtime_uninstalled"
	LifecycleStateRuntimeInstalled   LifecycleState = "runtime_installed"
	LifecycleStateRuntimeEnabled     LifecycleState = "runtime_enabled"

	// MigrationState values.
	MigrationStateNone      MigrationState = "none"
	MigrationStateSucceeded MigrationState = "succeeded"
	MigrationStateFailed    MigrationState = "failed"

	// ResourceSpecType values.
	ResourceSpecTypeTableList ResourceSpecType = "table-list"

	// ResourceFilterOperator values.
	ResourceFilterOperatorEQ      ResourceFilterOperator = "eq"
	ResourceFilterOperatorLike    ResourceFilterOperator = "like"
	ResourceFilterOperatorGTEDate ResourceFilterOperator = "gte-date"
	ResourceFilterOperatorLTEDate ResourceFilterOperator = "lte-date"

	// ResourceOrderDirection values.
	ResourceOrderDirectionASC  ResourceOrderDirection = "asc"
	ResourceOrderDirectionDESC ResourceOrderDirection = "desc"

	// ResourceOperation values.
	ResourceOperationQuery       ResourceOperation = "query"
	ResourceOperationGet         ResourceOperation = "get"
	ResourceOperationCreate      ResourceOperation = "create"
	ResourceOperationUpdate      ResourceOperation = "update"
	ResourceOperationDelete      ResourceOperation = "delete"
	ResourceOperationTransaction ResourceOperation = "transaction"

	// ResourceAccessMode values.
	ResourceAccessModeRequest ResourceAccessMode = "request"
	ResourceAccessModeSystem  ResourceAccessMode = "system"
	ResourceAccessModeBoth    ResourceAccessMode = "both"
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

// String returns the canonical lifecycle-state value.
func (value LifecycleState) String() string { return string(value) }

// String returns the canonical migration-state value.
func (value MigrationState) String() string { return string(value) }

// String returns the canonical resource spec type value.
func (value ResourceSpecType) String() string { return string(value) }

// String returns the canonical resource filter-operator value.
func (value ResourceFilterOperator) String() string { return string(value) }

// String returns the canonical resource order-direction value.
func (value ResourceOrderDirection) String() string { return string(value) }

// String returns the canonical resource operation value.
func (value ResourceOperation) String() string { return string(value) }

// String returns the canonical resource access-mode value.
func (value ResourceAccessMode) String() string { return string(value) }

// ManifestSnapshot stores the review-friendly manifest snapshot persisted in sys_plugin_release.
type ManifestSnapshot struct {
	ID                        string                          `yaml:"id"`
	Name                      string                          `yaml:"name"`
	Version                   string                          `yaml:"version"`
	Type                      string                          `yaml:"type"`
	Description               string                          `yaml:"description,omitempty"`
	Author                    string                          `yaml:"author,omitempty"`
	Homepage                  string                          `yaml:"homepage,omitempty"`
	License                   string                          `yaml:"license,omitempty"`
	RuntimeKind               string                          `yaml:"runtimeKind,omitempty"`
	RuntimeABIVersion         string                          `yaml:"runtimeAbiVersion,omitempty"`
	ManifestDeclared          bool                            `yaml:"manifestDeclared"`
	InstallSQLCount           int                             `yaml:"installSqlCount,omitempty"`
	UninstallSQLCount         int                             `yaml:"uninstallSqlCount,omitempty"`
	FrontendPageCount         int                             `yaml:"frontendPageCount,omitempty"`
	FrontendSlotCount         int                             `yaml:"frontendSlotCount,omitempty"`
	MenuCount                 int                             `yaml:"menuCount,omitempty"`
	BackendHookCount          int                             `yaml:"backendHookCount,omitempty"`
	ResourceSpecCount         int                             `yaml:"resourceSpecCount,omitempty"`
	RouteCount                int                             `yaml:"routeCount,omitempty"`
	RouteExecutionEnabled     bool                            `yaml:"routeExecutionEnabled,omitempty"`
	RouteRequestCodec         string                          `yaml:"routeRequestCodec,omitempty"`
	RouteResponseCodec        string                          `yaml:"routeResponseCodec,omitempty"`
	RuntimeFrontendAssetCount int                             `yaml:"runtimeFrontendAssetCount,omitempty"`
	RuntimeSQLAssetCount      int                             `yaml:"runtimeSqlAssetCount,omitempty"`
	RequestedHostServices     []*pluginbridge.HostServiceSpec `yaml:"requestedHostServices,omitempty"`
	AuthorizedHostServices    []*pluginbridge.HostServiceSpec `yaml:"authorizedHostServices,omitempty"`
	HostServiceAuthRequired   bool                            `yaml:"hostServiceAuthRequired,omitempty"`
	HostServiceAuthConfirmed  bool                            `yaml:"hostServiceAuthConfirmed,omitempty"`
}

// ResourceRefDescriptor represents one governance resource index entry derived
// from the current plugin release.
type ResourceRefDescriptor struct {
	Kind      ResourceKind
	Key       string
	OwnerType ResourceOwnerType
	OwnerKey  string
	Remark    string
}

// DeriveNodeState converts installation and enablement flags into one
// stable node-state key for the governance projection.
func DeriveNodeState(installed int, enabled int) string {
	if installed != InstalledYes {
		return NodeStateUninstalled.String()
	}
	if enabled == StatusEnabled {
		return NodeStateEnabled.String()
	}
	return NodeStateInstalled.String()
}

// DeriveHostState converts install and enablement flags into the stable
// host lifecycle state stored in sys_plugin desired_state/current_state.
func DeriveHostState(installed int, enabled int) string {
	if installed != InstalledYes {
		return HostStateUninstalled.String()
	}
	if enabled == StatusEnabled {
		return HostStateEnabled.String()
	}
	return HostStateInstalled.String()
}

// DeriveLifecycleState converts the plugin type and runtime flags into the
// lifecycle state exposed by the management API.
func DeriveLifecycleState(pluginType string, installed int, enabled int) string {
	if NormalizeType(pluginType) == TypeSource {
		if enabled == StatusEnabled {
			return LifecycleStateSourceEnabled.String()
		}
		return LifecycleStateSourceDisabled.String()
	}
	if installed != InstalledYes {
		return LifecycleStateRuntimeUninstalled.String()
	}
	if enabled == StatusEnabled {
		return LifecycleStateRuntimeEnabled.String()
	}
	return LifecycleStateRuntimeInstalled.String()
}
