// Package hostservice defines host-service declarations, capability derivation,
// manifest serialization, validation, normalization, and descriptor governance
// for Lina dynamic plugins.
package hostservice

// Capability constants describe the coarse-grained permissions implied by host
// service declarations.
const (
	// CapabilityRuntime grants access to runtime log/state/info host services.
	CapabilityRuntime = "host:runtime"
	// CapabilityStorage grants access to governed storage host services.
	CapabilityStorage = "host:storage"
	// CapabilityHTTPRequest grants access to governed outbound HTTP requests.
	CapabilityHTTPRequest = "host:http:request"
	// CapabilityDataRead grants access to read-oriented data service methods.
	CapabilityDataRead = "host:data:read"
	// CapabilityDataMutate grants access to write-oriented data service methods.
	CapabilityDataMutate = "host:data:mutate"
	// CapabilityCache grants access to governed cache host services.
	CapabilityCache = "host:cache"
	// CapabilityLock grants access to governed lock host services.
	CapabilityLock = "host:lock"
	// CapabilityHostConfig grants access to authorized host config keys.
	CapabilityHostConfig = "host:hostconfig"
	// CapabilityManifest grants access to plugin-scoped manifest resources.
	CapabilityManifest = "host:manifest"
	// CapabilityAPIDoc grants access to API-documentation localization services.
	CapabilityAPIDoc = "host:apidoc"
	// CapabilityAuthToken grants access to authentication token handoff services.
	CapabilityAuthToken = "host:auth:token"
	// CapabilityAuthz grants access to authorization-domain methods under auth.
	CapabilityAuthz = "host:auth:authz"
	// CapabilityUsers grants access to host-defined user-domain capability services.
	CapabilityUsers = "host:users"
	// CapabilityBizCtx grants access to request business-context projections.
	CapabilityBizCtx = "host:bizctx"
	// CapabilityDict grants access to dictionary-domain capability services.
	CapabilityDict = "host:dict"
	// CapabilityFiles grants access to file-domain capability services.
	CapabilityFiles = "host:files"
	// CapabilityJobs grants access to scheduled-job domain capability services.
	CapabilityJobs = "host:jobs"
	// CapabilityNotifications grants access to notification-domain capability services.
	CapabilityNotifications = "host:notifications"
	// CapabilityPlugins grants access to plugin-governance capability services.
	CapabilityPlugins = "host:plugins"
	// CapabilityRoute grants access to current dynamic-route metadata services.
	CapabilityRoute = "host:route"
	// CapabilitySessions grants access to online-session domain capability services.
	CapabilitySessions = "host:sessions"
	// CapabilityOrg grants access to host-defined organization capability services.
	CapabilityOrg = "host:org"
	// CapabilityTenant grants access to host-defined tenant capability services.
	CapabilityTenant = "host:tenant"
)

// Service and method wire constants live in protocol/hostservices (wire_constants.go)
// and are referenced through the hostservices package.

// Storage visibility constants describe the serving posture attached to plugin
// storage objects.
const (
	// HostServiceStorageVisibilityPrivate keeps storage objects internal to host-call access only.
	HostServiceStorageVisibilityPrivate = "private"
	// HostServiceStorageVisibilityPublic marks storage objects as eligible for future public serving.
	HostServiceStorageVisibilityPublic = "public"
)

// HostServiceSpec declares one structured host service authorization block in plugin.yaml.
type HostServiceSpec struct {
	// Owner is the plugin ID that owns this host service when the declaration
	// targets a plugin-owned capability. Core-owned host services leave it empty.
	Owner string `json:"owner,omitempty" yaml:"owner,omitempty"`
	// Service is the logical host service identifier.
	Service string `json:"service" yaml:"service"`
	// Version is the owner capability protocol version for plugin-owned host services.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// Methods lists the allowed methods under the host service.
	Methods []string `json:"methods" yaml:"methods"`
	// Paths lists the authorized logical paths for the storage host service.
	Paths []string `json:"paths,omitempty" yaml:"paths,omitempty"`
	// Tables lists the authorized table names for the data host service.
	Tables []string `json:"tables,omitempty" yaml:"tables,omitempty"`
	// Keys lists the authorized host config keys for the hostConfig service.
	Keys []string `json:"keys,omitempty" yaml:"keys,omitempty"`
	// Resources lists governed resource declarations bound to the host service.
	// For network service, Ref stores the authorized URL pattern.
	Resources []*HostServiceResourceSpec `json:"resources,omitempty" yaml:"resources,omitempty"`
}

// HostServiceResourceSpec declares one governed logical resource reference.
type HostServiceResourceSpec struct {
	// Ref is the stable governed target visible to the plugin. For network
	// service it stores one authorized URL pattern.
	Ref string `json:"ref" yaml:"ref"`
	// AllowMethods optionally restricts nested business methods such as HTTP verbs.
	AllowMethods []string `json:"allowMethods,omitempty" yaml:"allowMethods,omitempty"`
	// HeaderAllowList optionally whitelists request headers the plugin may set.
	HeaderAllowList []string `json:"headerAllowList,omitempty" yaml:"headerAllowList,omitempty"`
	// TimeoutMs optionally overrides the default timeout for this resource.
	TimeoutMs int `json:"timeoutMs,omitempty" yaml:"timeoutMs,omitempty"`
	// MaxBodyBytes optionally caps request or response payload size.
	MaxBodyBytes int `json:"maxBodyBytes,omitempty" yaml:"maxBodyBytes,omitempty"`
	// Attributes carries additional string-based governance metadata.
	Attributes map[string]string `json:"attributes,omitempty" yaml:"attributes,omitempty"`
}
