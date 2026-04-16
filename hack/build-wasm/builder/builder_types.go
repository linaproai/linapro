// This file defines the shared builder constants, manifest DTOs, artifact DTOs,
// and backend contract enums used across the split builder implementation.

package builder

import (
	"regexp"

	"lina-core/pkg/pluginbridge"
)

const (
	pluginTypeDynamic                = "dynamic"
	pluginDynamicKindWasm            = pluginbridge.RuntimeKindWasm
	pluginDynamicSupportedABIVersion = pluginbridge.SupportedABIVersion
	defaultRuntimeOutputDir          = "temp/output"
	runtimeWorkspaceDirName          = ".runtime"

	pluginDynamicWasmSectionManifest            = pluginbridge.WasmSectionManifest
	pluginDynamicWasmSectionDynamic             = pluginbridge.WasmSectionRuntime
	pluginDynamicWasmSectionFrontend            = pluginbridge.WasmSectionFrontendAssets
	pluginDynamicWasmSectionInstallSQL          = pluginbridge.WasmSectionInstallSQL
	pluginDynamicWasmSectionUninstallSQL        = pluginbridge.WasmSectionUninstallSQL
	pluginDynamicWasmSectionBackendHooks        = pluginbridge.WasmSectionBackendHooks
	pluginDynamicWasmSectionBackendRes          = pluginbridge.WasmSectionBackendResources
	pluginDynamicWasmSectionBackendRoutes       = pluginbridge.WasmSectionBackendRoutes
	pluginDynamicWasmSectionBackendBridge       = pluginbridge.WasmSectionBackendBridge
	pluginDynamicWasmSectionBackendHostServices = pluginbridge.WasmSectionBackendHostServices
)

var (
	pluginManifestIDPattern     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	pluginManifestSemverPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?$`)
	safeIdentifierPattern       = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
)

// RuntimeBuildOutput contains the generated dynamic artifact bytes and output path.
type RuntimeBuildOutput struct {
	ArtifactPath string
	Content      []byte
	RuntimePath  string
}

type pluginManifest struct {
	ID          string      `yaml:"id"`
	Name        string      `yaml:"name"`
	Version     string      `yaml:"version"`
	Type        string      `yaml:"type"`
	Description string      `yaml:"description"`
	Menus       []*menuSpec `yaml:"menus"`
	// Capabilities is kept only to reject deprecated author-side manifest input.
	Capabilities []string                        `yaml:"capabilities"`
	HostServices []*pluginbridge.HostServiceSpec `yaml:"hostServices"`
}

type dynamicArtifactManifest struct {
	ID          string      `json:"id" yaml:"id"`
	Name        string      `json:"name" yaml:"name"`
	Version     string      `json:"version" yaml:"version"`
	Type        string      `json:"type" yaml:"type"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Menus       []*menuSpec `json:"menus,omitempty" yaml:"menus,omitempty"`
}

type dynamicArtifactMetadata = pluginbridge.RuntimeArtifactMetadata

type embeddedStaticResourceSet struct {
	files map[string][]byte
}

type frontendAsset struct {
	Path          string `json:"path" yaml:"path"`
	ContentBase64 string `json:"contentBase64" yaml:"contentBase64"`
	ContentType   string `json:"contentType,omitempty" yaml:"contentType,omitempty"`
}

type sqlAsset struct {
	Key     string `json:"key" yaml:"key"`
	Content string `json:"content" yaml:"content"`
}

type menuSpec struct {
	Key        string                 `json:"key" yaml:"key"`
	ParentKey  string                 `json:"parent_key,omitempty" yaml:"parent_key,omitempty"`
	Name       string                 `json:"name" yaml:"name"`
	Path       string                 `json:"path,omitempty" yaml:"path,omitempty"`
	Component  string                 `json:"component,omitempty" yaml:"component,omitempty"`
	Perms      string                 `json:"perms,omitempty" yaml:"perms,omitempty"`
	Icon       string                 `json:"icon,omitempty" yaml:"icon,omitempty"`
	Type       string                 `json:"type,omitempty" yaml:"type,omitempty"`
	Sort       int                    `json:"sort,omitempty" yaml:"sort,omitempty"`
	Visible    *int                   `json:"visible,omitempty" yaml:"visible,omitempty"`
	Status     *int                   `json:"status,omitempty" yaml:"status,omitempty"`
	IsFrame    *int                   `json:"is_frame,omitempty" yaml:"is_frame,omitempty"`
	IsCache    *int                   `json:"is_cache,omitempty" yaml:"is_cache,omitempty"`
	Query      map[string]interface{} `json:"query,omitempty" yaml:"query,omitempty"`
	QueryParam string                 `json:"query_param,omitempty" yaml:"query_param,omitempty"`
	Remark     string                 `json:"remark,omitempty" yaml:"remark,omitempty"`
}

type hookExtensionPoint string
type hookAction string
type callbackExecutionMode string
type resourceSpecType string
type resourceFilterOperator string
type resourceOrderDirection string
type resourceOperation string
type resourceAccessMode string

const (
	callbackExecutionModeBlocking callbackExecutionMode = "blocking"
	callbackExecutionModeAsync    callbackExecutionMode = "async"

	hookActionInsert hookAction = "insert"
	hookActionSleep  hookAction = "sleep"
	hookActionError  hookAction = "error"

	resourceSpecTypeTableList resourceSpecType = "table-list"

	resourceFilterOperatorEQ      resourceFilterOperator = "eq"
	resourceFilterOperatorLike    resourceFilterOperator = "like"
	resourceFilterOperatorGTEDate resourceFilterOperator = "gte-date"
	resourceFilterOperatorLTEDate resourceFilterOperator = "lte-date"

	resourceOrderDirectionASC  resourceOrderDirection = "asc"
	resourceOrderDirectionDESC resourceOrderDirection = "desc"

	resourceOperationQuery       resourceOperation = "query"
	resourceOperationGet         resourceOperation = "get"
	resourceOperationCreate      resourceOperation = "create"
	resourceOperationUpdate      resourceOperation = "update"
	resourceOperationDelete      resourceOperation = "delete"
	resourceOperationTransaction resourceOperation = "transaction"

	resourceAccessModeRequest resourceAccessMode = "request"
	resourceAccessModeSystem  resourceAccessMode = "system"
	resourceAccessModeBoth    resourceAccessMode = "both"

	extensionPointAuthLoginSucceeded  hookExtensionPoint = "auth.login.succeeded"
	extensionPointAuthLoginFailed     hookExtensionPoint = "auth.login.failed"
	extensionPointAuthLogoutSucceeded hookExtensionPoint = "auth.logout.succeeded"
	extensionPointPluginInstalled     hookExtensionPoint = "plugin.installed"
	extensionPointPluginEnabled       hookExtensionPoint = "plugin.enabled"
	extensionPointPluginDisabled      hookExtensionPoint = "plugin.disabled"
	extensionPointPluginUninstalled   hookExtensionPoint = "plugin.uninstalled"
	extensionPointSystemStarted       hookExtensionPoint = "system.started"
)

type hookSpec struct {
	Event        hookExtensionPoint    `json:"event" yaml:"event"`
	Action       hookAction            `json:"action" yaml:"action"`
	Mode         callbackExecutionMode `json:"mode,omitempty" yaml:"mode,omitempty"`
	Table        string                `json:"table,omitempty" yaml:"table,omitempty"`
	Fields       map[string]string     `json:"fields,omitempty" yaml:"fields,omitempty"`
	TimeoutMs    int                   `json:"timeoutMs,omitempty" yaml:"timeoutMs,omitempty"`
	SleepMs      int                   `json:"sleepMs,omitempty" yaml:"sleepMs,omitempty"`
	ErrorMessage string                `json:"errorMessage,omitempty" yaml:"errorMessage,omitempty"`
}

type resourceSpec struct {
	Key            string                 `json:"key" yaml:"key"`
	Type           string                 `json:"type" yaml:"type"`
	Table          string                 `json:"table" yaml:"table"`
	Fields         []*resourceField       `json:"fields" yaml:"fields"`
	Filters        []*resourceQuery       `json:"filters" yaml:"filters"`
	OrderBy        resourceOrderBySpec    `json:"orderBy" yaml:"orderBy"`
	Operations     []string               `json:"operations,omitempty" yaml:"operations,omitempty"`
	KeyField       string                 `json:"keyField,omitempty" yaml:"keyField,omitempty"`
	WritableFields []string               `json:"writableFields,omitempty" yaml:"writableFields,omitempty"`
	Access         string                 `json:"access,omitempty" yaml:"access,omitempty"`
	DataScope      *resourceDataScopeSpec `json:"dataScope,omitempty" yaml:"dataScope,omitempty"`
}

type resourceField struct {
	Name   string `json:"name" yaml:"name"`
	Column string `json:"column" yaml:"column"`
}

type resourceQuery struct {
	Param    string `json:"param" yaml:"param"`
	Column   string `json:"column" yaml:"column"`
	Operator string `json:"operator" yaml:"operator"`
}

type resourceOrderBySpec struct {
	Column    string `json:"column" yaml:"column"`
	Direction string `json:"direction" yaml:"direction"`
}

type resourceDataScopeSpec struct {
	UserColumn string `json:"userColumn,omitempty" yaml:"userColumn,omitempty"`
	DeptColumn string `json:"deptColumn,omitempty" yaml:"deptColumn,omitempty"`
}

var publishedHookPoints = map[hookExtensionPoint]callbackExecutionMode{
	extensionPointAuthLoginSucceeded:  callbackExecutionModeBlocking,
	extensionPointAuthLoginFailed:     callbackExecutionModeBlocking,
	extensionPointAuthLogoutSucceeded: callbackExecutionModeBlocking,
	extensionPointPluginInstalled:     callbackExecutionModeBlocking,
	extensionPointPluginEnabled:       callbackExecutionModeBlocking,
	extensionPointPluginDisabled:      callbackExecutionModeBlocking,
	extensionPointPluginUninstalled:   callbackExecutionModeBlocking,
	extensionPointSystemStarted:       callbackExecutionModeBlocking,
}

var supportedHookModes = map[hookExtensionPoint]map[callbackExecutionMode]struct{}{
	extensionPointAuthLoginSucceeded:  {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointAuthLoginFailed:     {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointAuthLogoutSucceeded: {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointPluginInstalled:     {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointPluginEnabled:       {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointPluginDisabled:      {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointPluginUninstalled:   {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointSystemStarted:       {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
}
