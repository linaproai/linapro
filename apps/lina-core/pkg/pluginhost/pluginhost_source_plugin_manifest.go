// This file defines manifest snapshot wrappers published to source-plugin
// upgrade callbacks.

package pluginhost

// ManifestSnapshot exposes the review-oriented manifest snapshot fields needed
// by source-plugin upgrade callbacks without leaking host catalog internals.
type ManifestSnapshot interface {
	// ID returns the plugin identifier recorded in the manifest snapshot.
	ID() string
	// Name returns the plugin display name recorded in the manifest snapshot.
	Name() string
	// Version returns the plugin version recorded in the manifest snapshot.
	Version() string
	// Type returns the plugin type recorded in the manifest snapshot.
	Type() string
	// Values returns a copy of all published snapshot fields.
	Values() map[string]interface{}
}

// ManifestSnapshotValues contains the stable manifest snapshot fields published
// to source-plugin upgrade callbacks.
type ManifestSnapshotValues struct {
	// ID is the plugin identifier recorded in the manifest snapshot.
	ID string
	// Name is the plugin display name recorded in the manifest snapshot.
	Name string
	// Version is the plugin version recorded in the manifest snapshot.
	Version string
	// Type is the plugin type recorded in the manifest snapshot.
	Type string
	// ScopeNature is the plugin tenant-scope nature recorded in the manifest snapshot.
	ScopeNature string
	// SupportsMultiTenant reports whether the plugin declares multi-tenant support.
	SupportsMultiTenant bool
	// DefaultInstallMode is the plugin default installation mode.
	DefaultInstallMode string
	// Description is the plugin description recorded in the manifest snapshot.
	Description string
	// InstallSQLCount is the number of install SQL assets recorded in the snapshot.
	InstallSQLCount int
	// UninstallSQLCount is the number of uninstall SQL assets recorded in the snapshot.
	UninstallSQLCount int
	// MockSQLCount is the number of mock SQL assets recorded in the snapshot.
	MockSQLCount int
	// MenuCount is the number of menu definitions recorded in the snapshot.
	MenuCount int
	// BackendHookCount is the number of backend hook registrations recorded in the snapshot.
	BackendHookCount int
	// ResourceSpecCount is the number of resource specs recorded in the snapshot.
	ResourceSpecCount int
	// HostServiceAuthNeeded reports whether host-service authorization is required.
	HostServiceAuthNeeded bool
}

const (
	// manifestSnapshotFieldID is the published manifest snapshot value key for the plugin identifier.
	manifestSnapshotFieldID = "id"
	// manifestSnapshotFieldName is the published manifest snapshot value key for the display name.
	manifestSnapshotFieldName = "name"
	// manifestSnapshotFieldVersion is the published manifest snapshot value key for the plugin version.
	manifestSnapshotFieldVersion = "version"
	// manifestSnapshotFieldType is the published manifest snapshot value key for the plugin type.
	manifestSnapshotFieldType = "type"
	// manifestSnapshotFieldScopeNature is the published manifest snapshot value key for tenant-scope nature.
	manifestSnapshotFieldScopeNature = "scopeNature"
	// manifestSnapshotFieldSupportsMultiTenant is the published manifest snapshot value key for multi-tenant support.
	manifestSnapshotFieldSupportsMultiTenant = "supportsMultiTenant"
	// manifestSnapshotFieldDefaultInstallMode is the published manifest snapshot value key for default install mode.
	manifestSnapshotFieldDefaultInstallMode = "defaultInstallMode"
	// manifestSnapshotFieldDescription is the published manifest snapshot value key for the plugin description.
	manifestSnapshotFieldDescription = "description"
	// manifestSnapshotFieldInstallSQLCount is the published manifest snapshot value key for install SQL count.
	manifestSnapshotFieldInstallSQLCount = "installSqlCount"
	// manifestSnapshotFieldUninstallSQLCount is the published manifest snapshot value key for uninstall SQL count.
	manifestSnapshotFieldUninstallSQLCount = "uninstallSqlCount"
	// manifestSnapshotFieldMockSQLCount is the published manifest snapshot value key for mock SQL count.
	manifestSnapshotFieldMockSQLCount = "mockSqlCount"
	// manifestSnapshotFieldMenuCount is the published manifest snapshot value key for menu count.
	manifestSnapshotFieldMenuCount = "menuCount"
	// manifestSnapshotFieldBackendHookCount is the published manifest snapshot value key for backend hook count.
	manifestSnapshotFieldBackendHookCount = "backendHookCount"
	// manifestSnapshotFieldResourceSpecCount is the published manifest snapshot value key for resource spec count.
	manifestSnapshotFieldResourceSpecCount = "resourceSpecCount"
	// manifestSnapshotFieldHostServiceAuthNeeded is the published manifest snapshot value key for host-service auth requirements.
	manifestSnapshotFieldHostServiceAuthNeeded = "hostServiceAuthNeeded"
)

// manifestSnapshot is the host-owned immutable view passed to source-plugin
// runtime upgrade callbacks.
type manifestSnapshot struct {
	id      string
	name    string
	version string
	kind    string
	values  map[string]interface{}
}

// NewManifestSnapshot creates one published manifest snapshot wrapper for
// source-plugin upgrade callbacks.
//
// Deprecated: use NewManifestSnapshotFromValues for host-created snapshots so
// the published field contract stays type checked at construction sites.
func NewManifestSnapshot(values map[string]interface{}) ManifestSnapshot {
	copiedValues := cloneValueMap(values)
	return &manifestSnapshot{
		id:      stringValueFromMap(copiedValues, manifestSnapshotFieldID),
		name:    stringValueFromMap(copiedValues, manifestSnapshotFieldName),
		version: stringValueFromMap(copiedValues, manifestSnapshotFieldVersion),
		kind:    stringValueFromMap(copiedValues, manifestSnapshotFieldType),
		values:  copiedValues,
	}
}

// NewManifestSnapshotFromValues creates one published manifest snapshot wrapper
// from the typed source-plugin upgrade callback contract.
func NewManifestSnapshotFromValues(values ManifestSnapshotValues) ManifestSnapshot {
	return &manifestSnapshot{
		id:      values.ID,
		name:    values.Name,
		version: values.Version,
		kind:    values.Type,
		values:  values.toMap(),
	}
}

// ID returns the plugin identifier recorded in the manifest snapshot.
func (s *manifestSnapshot) ID() string {
	if s == nil {
		return ""
	}
	return s.id
}

// Name returns the plugin display name recorded in the manifest snapshot.
func (s *manifestSnapshot) Name() string {
	if s == nil {
		return ""
	}
	return s.name
}

// Version returns the plugin version recorded in the manifest snapshot.
func (s *manifestSnapshot) Version() string {
	if s == nil {
		return ""
	}
	return s.version
}

// Type returns the plugin type recorded in the manifest snapshot.
func (s *manifestSnapshot) Type() string {
	if s == nil {
		return ""
	}
	return s.kind
}

// Values returns a shallow copy of all published manifest snapshot fields.
func (s *manifestSnapshot) Values() map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}
	return cloneValueMap(s.values)
}

// toMap converts typed manifest snapshot values into the published value map.
func (values ManifestSnapshotValues) toMap() map[string]interface{} {
	return map[string]interface{}{
		manifestSnapshotFieldID:                    values.ID,
		manifestSnapshotFieldName:                  values.Name,
		manifestSnapshotFieldVersion:               values.Version,
		manifestSnapshotFieldType:                  values.Type,
		manifestSnapshotFieldScopeNature:           values.ScopeNature,
		manifestSnapshotFieldSupportsMultiTenant:   values.SupportsMultiTenant,
		manifestSnapshotFieldDefaultInstallMode:    values.DefaultInstallMode,
		manifestSnapshotFieldDescription:           values.Description,
		manifestSnapshotFieldInstallSQLCount:       values.InstallSQLCount,
		manifestSnapshotFieldUninstallSQLCount:     values.UninstallSQLCount,
		manifestSnapshotFieldMockSQLCount:          values.MockSQLCount,
		manifestSnapshotFieldMenuCount:             values.MenuCount,
		manifestSnapshotFieldBackendHookCount:      values.BackendHookCount,
		manifestSnapshotFieldResourceSpecCount:     values.ResourceSpecCount,
		manifestSnapshotFieldHostServiceAuthNeeded: values.HostServiceAuthNeeded,
	}
}
