// This file defines the plugin manifest type and all nested data types shared
// across catalog, runtime, integration, and frontend sub-packages.

package catalog

import (
	"strings"

	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginhost"
)

// MenuType defines the canonical display category for one plugin-declared menu entry.
type MenuType string

const (
	// MenuTypeDirectory marks a directory/group menu.
	MenuTypeDirectory MenuType = "D"
	// MenuTypePage marks a page/router menu.
	MenuTypePage MenuType = "M"
	// MenuTypeButton marks a hidden button/permission menu.
	MenuTypeButton MenuType = "B"
)

// String returns the canonical menu type value.
func (t MenuType) String() string { return string(t) }

// NormalizeMenuType converts a raw menu type string to the canonical MenuType value.
func NormalizeMenuType(value string) MenuType {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "":
		return MenuTypePage
	case MenuTypeDirectory.String():
		return MenuTypeDirectory
	case MenuTypePage.String():
		return MenuTypePage
	case MenuTypeButton.String():
		return MenuTypeButton
	default:
		return ""
	}
}

// IsSupportedMenuType reports whether value is a recognized menu type.
func IsSupportedMenuType(value MenuType) bool {
	return value == MenuTypeDirectory || value == MenuTypePage || value == MenuTypeButton
}

// Manifest defines plugin metadata loaded from plugin.yaml or wasm custom sections.
type Manifest struct {
	// ID is the unique kebab-case plugin identifier.
	ID string `yaml:"id"`
	// Name is the human-readable plugin name.
	Name string `yaml:"name"`
	// Version is the semantic version string.
	Version string `yaml:"version"`
	// Type is the normalized plugin type ("source" or "dynamic").
	Type string `yaml:"type"`
	// Description is an optional human-readable description.
	Description string `yaml:"description"`
	// Author is an optional author string.
	Author string `yaml:"author"`
	// Homepage is an optional URL.
	Homepage string `yaml:"homepage"`
	// License is an optional license identifier.
	License string `yaml:"license"`
	// Menus holds manifest-declared host menu entries.
	Menus []*MenuSpec `yaml:"menus"`
	// ManifestPath is the filesystem path to the plugin.yaml file (source plugins).
	ManifestPath string
	// RootDir is the plugin root directory path.
	RootDir string
	// Hooks holds plugin-declared hook handler specifications.
	Hooks []*HookSpec
	// BackendResources holds plugin-declared backend resource specifications keyed by resource ID.
	BackendResources map[string]*ResourceSpec
	// Routes holds plugin-declared bridge route contracts.
	Routes []*pluginbridge.RouteContract
	// BridgeSpec carries the WASM bridge ABI metadata.
	BridgeSpec *pluginbridge.BridgeSpec
	// HostCapabilities is the set of granted host call capabilities.
	HostCapabilities map[string]struct{}
	// HostServices holds the structured host service declarations restored from release metadata.
	HostServices []*pluginbridge.HostServiceSpec
	// RuntimeArtifact holds the validated WASM artifact for dynamic plugins.
	RuntimeArtifact *ArtifactSpec
	// SourcePlugin is the embedded source-plugin registration for source plugins.
	SourcePlugin *pluginhost.SourcePlugin
}

// MenuSpec defines one manifest-declared host menu entry.
type MenuSpec struct {
	// Key is the unique kebab-case menu key within this plugin.
	Key string `yaml:"key" json:"key"`
	// ParentKey is the parent menu key for nested menus.
	ParentKey string `yaml:"parent_key,omitempty" json:"parent_key,omitempty"`
	// Name is the display name.
	Name string `yaml:"name" json:"name"`
	// Path is the frontend route path.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	// Component is the frontend component path.
	Component string `yaml:"component,omitempty" json:"component,omitempty"`
	// Perms is the permission code string.
	Perms string `yaml:"perms,omitempty" json:"perms,omitempty"`
	// Icon is the menu icon identifier.
	Icon string `yaml:"icon,omitempty" json:"icon,omitempty"`
	// Type is the menu type (D=directory, M=page, B=button).
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
	// Sort is the display sort order.
	Sort int `yaml:"sort,omitempty" json:"sort,omitempty"`
	// Visible overrides the default visibility.
	Visible *int `yaml:"visible,omitempty" json:"visible,omitempty"`
	// Status overrides the default menu status.
	Status *int `yaml:"status,omitempty" json:"status,omitempty"`
	// IsFrame overrides the default iframe flag.
	IsFrame *int `yaml:"is_frame,omitempty" json:"is_frame,omitempty"`
	// IsCache overrides the default cache flag.
	IsCache *int `yaml:"is_cache,omitempty" json:"is_cache,omitempty"`
	// Query holds arbitrary query parameters attached to the menu.
	Query map[string]interface{} `yaml:"query,omitempty" json:"query,omitempty"`
	// QueryParam is an alternative query parameter shorthand.
	QueryParam string `yaml:"query_param,omitempty" json:"query_param,omitempty"`
	// Remark is an optional description.
	Remark string `yaml:"remark,omitempty" json:"remark,omitempty"`
}

// HookSpec defines a plugin-declared hook handler.
type HookSpec struct {
	// Event is the extension point this hook listens on.
	Event pluginhost.ExtensionPoint `json:"event" yaml:"event"`
	// Action is the hook action type.
	Action pluginhost.HookAction `json:"action" yaml:"action"`
	// Mode is the optional execution mode (sync/async).
	Mode pluginhost.CallbackExecutionMode `json:"mode,omitempty" yaml:"mode,omitempty"`
	// Table is the target table name for data hooks.
	Table string `json:"table,omitempty" yaml:"table,omitempty"`
	// Fields maps output field names to column expressions.
	Fields map[string]string `json:"fields,omitempty" yaml:"fields,omitempty"`
	// TimeoutMs is the hook invocation timeout in milliseconds.
	TimeoutMs int `json:"timeoutMs,omitempty" yaml:"timeoutMs,omitempty"`
	// SleepMs is an optional delay before hook invocation.
	SleepMs int `json:"sleepMs,omitempty" yaml:"sleepMs,omitempty"`
	// ErrorMessage is the user-facing message returned on hook failure.
	ErrorMessage string `json:"errorMessage,omitempty" yaml:"errorMessage,omitempty"`
}

// ResourceSpec defines a plugin-declared backend resource.
type ResourceSpec struct {
	// Key is the unique resource identifier within this plugin.
	Key string `json:"key" yaml:"key"`
	// Type is the resource type (e.g. "table-list").
	Type string `json:"type" yaml:"type"`
	// Table is the backing database table name.
	Table string `json:"table" yaml:"table"`
	// Fields is the ordered list of output fields.
	Fields []*ResourceField `json:"fields" yaml:"fields"`
	// Filters is the list of supported query filters.
	Filters []*ResourceQuery `json:"filters" yaml:"filters"`
	// OrderBy defines default result ordering.
	OrderBy ResourceOrderBySpec `json:"orderBy" yaml:"orderBy"`
	// Operations lists the structured data methods that may operate on this resource.
	Operations []string `json:"operations,omitempty" yaml:"operations,omitempty"`
	// KeyField declares the API field name used as the primary identity for get/update/delete operations.
	KeyField string `json:"keyField,omitempty" yaml:"keyField,omitempty"`
	// WritableFields lists the API field names the guest may submit for create/update operations.
	WritableFields []string `json:"writableFields,omitempty" yaml:"writableFields,omitempty"`
	// Access limits which execution contexts may invoke this resource.
	Access string `json:"access,omitempty" yaml:"access,omitempty"`
	// DataScope optionally restricts results by role data scope.
	DataScope *ResourceDataScopeSpec `json:"dataScope,omitempty" yaml:"dataScope,omitempty"`
}

// ResourceField defines one output column for a plugin resource.
type ResourceField struct {
	// Name is the API field name.
	Name string `json:"name" yaml:"name"`
	// Column is the database column expression.
	Column string `json:"column" yaml:"column"`
}

// ResourceQuery defines one query filter parameter for a plugin resource.
type ResourceQuery struct {
	// Param is the query parameter name.
	Param string `json:"param" yaml:"param"`
	// Column is the database column to filter on.
	Column string `json:"column" yaml:"column"`
	// Operator is the filter operator (eq, like, gte-date, lte-date).
	Operator string `json:"operator" yaml:"operator"`
}

// ResourceOrderBySpec defines the ordering configuration for a plugin resource.
type ResourceOrderBySpec struct {
	// Column is the column to order by.
	Column string `json:"column" yaml:"column"`
	// Direction is the order direction ("asc" or "desc").
	Direction string `json:"direction" yaml:"direction"`
}

// ResourceDataScopeSpec defines how a plugin resource binds to host role data scopes.
type ResourceDataScopeSpec struct {
	// UserColumn is the user-ID column name for user-scope filtering.
	UserColumn string `json:"userColumn,omitempty" yaml:"userColumn,omitempty"`
	// DeptColumn is the dept-ID column name for dept-scope filtering.
	DeptColumn string `json:"deptColumn,omitempty" yaml:"deptColumn,omitempty"`
}

// ArtifactSpec describes one validated runtime WASM artifact loaded from disk.
type ArtifactSpec struct {
	// Path is the filesystem path to the WASM file.
	Path string
	// Checksum is the hex-encoded SHA-256 of the artifact content.
	Checksum string
	// RuntimeKind identifies the WASM runtime type (e.g. "wasm").
	RuntimeKind string
	// ABIVersion is the bridge ABI version string declared in the artifact.
	ABIVersion string
	// FrontendAssetCount is the count of embedded frontend static assets.
	FrontendAssetCount int
	// SQLAssetCount is the count of embedded SQL migration assets.
	SQLAssetCount int
	// RouteCount is the count of declared bridge routes.
	RouteCount int
	// Manifest is the embedded plugin identity manifest.
	Manifest *ArtifactManifest
	// FrontendAssets holds the embedded frontend static assets.
	FrontendAssets []*ArtifactFrontendAsset
	// InstallSQLAssets holds the embedded install SQL migration steps.
	InstallSQLAssets []*ArtifactSQLAsset
	// UninstallSQLAssets holds the embedded uninstall SQL migration steps.
	UninstallSQLAssets []*ArtifactSQLAsset
	// HookSpecs holds the embedded hook handler declarations.
	HookSpecs []*HookSpec
	// ResourceSpecs holds the embedded resource declarations.
	ResourceSpecs []*ResourceSpec
	// RouteContracts holds the embedded bridge route contracts.
	RouteContracts []*pluginbridge.RouteContract
	// BridgeSpec carries the WASM bridge ABI metadata.
	BridgeSpec *pluginbridge.BridgeSpec
	// Capabilities lists the coarse host capability identifiers derived from HostServices.
	Capabilities []string
	// HostServices lists the structured host service declarations embedded in the artifact.
	HostServices []*pluginbridge.HostServiceSpec
}

// ArtifactManifest stores the plugin identity embedded in WASM custom sections.
type ArtifactManifest struct {
	// ID is the unique plugin identifier.
	ID string `json:"id" yaml:"id"`
	// Name is the human-readable plugin name.
	Name string `json:"name" yaml:"name"`
	// Version is the semantic version string.
	Version string `json:"version" yaml:"version"`
	// Type is the normalized plugin type.
	Type string `json:"type" yaml:"type"`
	// Description is an optional human-readable description.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Menus holds manifest-declared host menu entries.
	Menus []*MenuSpec `json:"menus,omitempty" yaml:"menus,omitempty"`
}

// ArtifactFrontendAsset stores one embedded frontend static asset.
type ArtifactFrontendAsset struct {
	// Path is the asset path relative to the frontend root.
	Path string `json:"path" yaml:"path"`
	// ContentBase64 is the base64-encoded asset content.
	ContentBase64 string `json:"contentBase64" yaml:"contentBase64"`
	// ContentType is the MIME type of the asset.
	ContentType string `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	// Content is the decoded asset bytes (not serialized).
	Content []byte `json:"-" yaml:"-"`
}

// ArtifactSQLAsset stores one embedded SQL migration step.
type ArtifactSQLAsset struct {
	// Key identifies this SQL step within the artifact.
	Key string `json:"key" yaml:"key"`
	// Content is the raw SQL text.
	Content string `json:"content" yaml:"content"`
}
