// Package hostservice defines host-service declarations, capability derivation,
// manifest serialization, and payload codecs for Lina dynamic plugins.
package hostservice

// Capability constants describe the coarse-grained permissions implied by host
// service declarations.
const (
	// CapabilityRuntime grants access to runtime log/state/info host services.
	CapabilityRuntime = "host:runtime"
	// CapabilityCron grants access to dynamic-plugin cron registration host services.
	CapabilityCron = "host:cron"
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
	// CapabilitySecret grants access to governed secret resolution services.
	CapabilitySecret = "host:secret"
	// CapabilityEventPublish grants access to governed event publishing.
	CapabilityEventPublish = "host:event:publish"
	// CapabilityQueueEnqueue grants access to governed queue submission.
	CapabilityQueueEnqueue = "host:queue:enqueue"
	// CapabilityNotify grants access to governed notification services.
	CapabilityNotify = "host:notify"
	// CapabilityConfig grants access to read-only host configuration services.
	CapabilityConfig = "host:config"
)

// Host service identifiers declare the logical service families exposed by the
// host runtime to plugins.
const (
	// HostServiceRuntime is the runtime host service identifier.
	HostServiceRuntime = "runtime"
	// HostServiceCron is the cron host service identifier.
	HostServiceCron = "cron"
	// HostServiceStorage is the storage host service identifier.
	HostServiceStorage = "storage"
	// HostServiceNetwork is the network host service identifier.
	HostServiceNetwork = "network"
	// HostServiceData is the data host service identifier.
	HostServiceData = "data"
	// HostServiceCache is the cache host service identifier.
	HostServiceCache = "cache"
	// HostServiceLock is the lock host service identifier.
	HostServiceLock = "lock"
	// HostServiceSecret is the secret host service identifier.
	HostServiceSecret = "secret"
	// HostServiceEvent is the event host service identifier.
	HostServiceEvent = "event"
	// HostServiceQueue is the queue host service identifier.
	HostServiceQueue = "queue"
	// HostServiceNotify is the notify host service identifier.
	HostServiceNotify = "notify"
	// HostServiceConfig is the read-only configuration host service identifier.
	HostServiceConfig = "config"
)

// Runtime host-service methods describe runtime logging, state, and info
// operations available to authorized plugins.
const (
	// HostServiceMethodRuntimeLogWrite writes one structured runtime log entry.
	HostServiceMethodRuntimeLogWrite = "log.write"
	// HostServiceMethodRuntimeStateGet reads one plugin-scoped runtime state value.
	HostServiceMethodRuntimeStateGet = "state.get"
	// HostServiceMethodRuntimeStateSet writes one plugin-scoped runtime state value.
	HostServiceMethodRuntimeStateSet = "state.set"
	// HostServiceMethodRuntimeStateDelete deletes one plugin-scoped runtime state value.
	HostServiceMethodRuntimeStateDelete = "state.delete"
	// HostServiceMethodRuntimeInfoNow returns host time information.
	HostServiceMethodRuntimeInfoNow = "info.now"
	// HostServiceMethodRuntimeInfoUUID returns one host-generated unique identifier.
	HostServiceMethodRuntimeInfoUUID = "info.uuid"
	// HostServiceMethodRuntimeInfoNode returns host node identity information.
	HostServiceMethodRuntimeInfoNode = "info.node"
)

// Cron host-service methods describe dynamic-plugin cron declaration
// operations exposed during host-side discovery.
const (
	// HostServiceMethodCronRegister registers one dynamic-plugin cron contract
	// with the current host-side discovery collector.
	HostServiceMethodCronRegister = "register"
)

// Storage host-service methods describe governed file operations under the
// plugin storage sandbox.
const (
	// HostServiceMethodStoragePut writes one governed storage object.
	HostServiceMethodStoragePut = "put"
	// HostServiceMethodStorageGet reads one governed storage object.
	HostServiceMethodStorageGet = "get"
	// HostServiceMethodStorageDelete deletes one governed storage object.
	HostServiceMethodStorageDelete = "delete"
	// HostServiceMethodStorageList lists governed storage objects under one prefix.
	HostServiceMethodStorageList = "list"
	// HostServiceMethodStorageStat reads metadata for one governed storage object.
	HostServiceMethodStorageStat = "stat"
)

// Network host-service methods describe governed outbound HTTP operations.
const (
	// HostServiceMethodNetworkRequest executes one governed outbound HTTP request.
	HostServiceMethodNetworkRequest = "request"
)

// Data host-service methods describe governed table operations authorized by
// host manifest declarations.
const (
	// HostServiceMethodDataList executes one governed paged list query against an authorized data table.
	HostServiceMethodDataList = "list"
	// HostServiceMethodDataGet reads one governed record by key from an authorized data table.
	HostServiceMethodDataGet = "get"
	// HostServiceMethodDataCreate creates one governed record in an authorized data table.
	HostServiceMethodDataCreate = "create"
	// HostServiceMethodDataUpdate updates one governed record in an authorized data table.
	HostServiceMethodDataUpdate = "update"
	// HostServiceMethodDataDelete deletes one governed record in an authorized data table.
	HostServiceMethodDataDelete = "delete"
	// HostServiceMethodDataTransaction executes one governed transaction over structured data mutations.
	HostServiceMethodDataTransaction = "transaction"
)

// Cache host-service methods describe governed cache access primitives.
const (
	// HostServiceMethodCacheGet reads one governed cache value.
	HostServiceMethodCacheGet = "get"
	// HostServiceMethodCacheSet writes one governed cache value.
	HostServiceMethodCacheSet = "set"
	// HostServiceMethodCacheDelete removes one governed cache value.
	HostServiceMethodCacheDelete = "delete"
	// HostServiceMethodCacheIncr increments one governed cache integer value.
	HostServiceMethodCacheIncr = "incr"
	// HostServiceMethodCacheExpire updates one governed cache expiration policy.
	HostServiceMethodCacheExpire = "expire"
)

// Lock host-service methods describe governed distributed lock operations.
const (
	// HostServiceMethodLockAcquire acquires one governed distributed lock.
	HostServiceMethodLockAcquire = "acquire"
	// HostServiceMethodLockRenew renews one governed distributed lock.
	HostServiceMethodLockRenew = "renew"
	// HostServiceMethodLockRelease releases one governed distributed lock.
	HostServiceMethodLockRelease = "release"
)

// Notify host-service methods describe governed notification dispatch
// operations.
const (
	// HostServiceMethodNotifySend sends one governed notification message.
	HostServiceMethodNotifySend = "send"
)

// Config host-service methods describe read-only host configuration access.
const (
	// HostServiceMethodConfigGet reads one configuration value as JSON.
	HostServiceMethodConfigGet = "get"
	// HostServiceMethodConfigExists reports whether one configuration key exists.
	HostServiceMethodConfigExists = "exists"
	// HostServiceMethodConfigString reads one configuration value as a string.
	HostServiceMethodConfigString = "string"
	// HostServiceMethodConfigBool reads one configuration value as a bool.
	HostServiceMethodConfigBool = "bool"
	// HostServiceMethodConfigInt reads one configuration value as an int.
	HostServiceMethodConfigInt = "int"
	// HostServiceMethodConfigDuration reads one configuration value as a duration string.
	HostServiceMethodConfigDuration = "duration"
)

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
	// Service is the logical host service identifier.
	Service string `json:"service" yaml:"service"`
	// Methods lists the allowed methods under the host service. Config service
	// declarations default to all read-only config methods when methods are omitted.
	Methods []string `json:"methods" yaml:"methods"`
	// Paths lists the authorized logical paths for the storage host service.
	Paths []string `json:"paths,omitempty" yaml:"paths,omitempty"`
	// Tables lists the authorized table names for the data host service.
	Tables []string `json:"tables,omitempty" yaml:"tables,omitempty"`
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
