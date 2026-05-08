// Package hostservice defines host-service declarations, capability derivation,
// manifest serialization, and payload codecs for Lina dynamic plugins.
package hostservice

import (
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

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

// Shared host-service lookup tables drive capability derivation and per-service
// validation rules used by manifest normalization.
var (
	hostServiceMethodCapabilityMap = map[string]map[string]string{
		HostServiceRuntime: {
			HostServiceMethodRuntimeLogWrite:    CapabilityRuntime,
			HostServiceMethodRuntimeStateGet:    CapabilityRuntime,
			HostServiceMethodRuntimeStateSet:    CapabilityRuntime,
			HostServiceMethodRuntimeStateDelete: CapabilityRuntime,
			HostServiceMethodRuntimeInfoNow:     CapabilityRuntime,
			HostServiceMethodRuntimeInfoUUID:    CapabilityRuntime,
			HostServiceMethodRuntimeInfoNode:    CapabilityRuntime,
		},
		HostServiceCron: {
			HostServiceMethodCronRegister: CapabilityCron,
		},
		HostServiceStorage: {
			HostServiceMethodStoragePut:    CapabilityStorage,
			HostServiceMethodStorageGet:    CapabilityStorage,
			HostServiceMethodStorageDelete: CapabilityStorage,
			HostServiceMethodStorageList:   CapabilityStorage,
			HostServiceMethodStorageStat:   CapabilityStorage,
		},
		HostServiceNetwork: {
			HostServiceMethodNetworkRequest: CapabilityHTTPRequest,
		},
		HostServiceData: {
			HostServiceMethodDataList:        CapabilityDataRead,
			HostServiceMethodDataGet:         CapabilityDataRead,
			HostServiceMethodDataCreate:      CapabilityDataMutate,
			HostServiceMethodDataUpdate:      CapabilityDataMutate,
			HostServiceMethodDataDelete:      CapabilityDataMutate,
			HostServiceMethodDataTransaction: CapabilityDataMutate,
		},
		HostServiceCache: {
			HostServiceMethodCacheGet:    CapabilityCache,
			HostServiceMethodCacheSet:    CapabilityCache,
			HostServiceMethodCacheDelete: CapabilityCache,
			HostServiceMethodCacheIncr:   CapabilityCache,
			HostServiceMethodCacheExpire: CapabilityCache,
		},
		HostServiceLock: {
			HostServiceMethodLockAcquire: CapabilityLock,
			HostServiceMethodLockRenew:   CapabilityLock,
			HostServiceMethodLockRelease: CapabilityLock,
		},
		HostServiceSecret: {
			"resolve": CapabilitySecret,
		},
		HostServiceEvent: {
			"publish": CapabilityEventPublish,
		},
		HostServiceQueue: {
			"enqueue": CapabilityQueueEnqueue,
		},
		HostServiceNotify: {
			HostServiceMethodNotifySend: CapabilityNotify,
		},
		HostServiceConfig: {
			HostServiceMethodConfigGet:      CapabilityConfig,
			HostServiceMethodConfigExists:   CapabilityConfig,
			HostServiceMethodConfigString:   CapabilityConfig,
			HostServiceMethodConfigBool:     CapabilityConfig,
			HostServiceMethodConfigInt:      CapabilityConfig,
			HostServiceMethodConfigDuration: CapabilityConfig,
		},
	}

	allCapabilities = map[string]struct{}{
		CapabilityRuntime:      {},
		CapabilityCron:         {},
		CapabilityStorage:      {},
		CapabilityHTTPRequest:  {},
		CapabilityDataRead:     {},
		CapabilityDataMutate:   {},
		CapabilityCache:        {},
		CapabilityLock:         {},
		CapabilitySecret:       {},
		CapabilityEventPublish: {},
		CapabilityQueueEnqueue: {},
		CapabilityNotify:       {},
		CapabilityConfig:       {},
	}

	hostServicesWithoutResources = map[string]struct{}{
		HostServiceRuntime: {},
		HostServiceCron:    {},
		HostServiceConfig:  {},
	}

	hostServicesWithTables = map[string]struct{}{
		HostServiceData: {},
	}

	hostServicesWithPaths = map[string]struct{}{
		HostServiceStorage: {},
	}
)

// RequiredCapabilityForHostServiceMethod returns the capability required by one host service method.
func RequiredCapabilityForHostServiceMethod(service string, method string) string {
	service = normalizeHostServiceName(service)
	method = normalizeHostServiceMethod(method)
	methods := hostServiceMethodCapabilityMap[service]
	if methods == nil {
		return ""
	}
	return methods[method]
}

// CapabilitiesFromHostServices returns the sorted capability slice implied by one
// normalized host service declaration set.
func CapabilitiesFromHostServices(specs []*HostServiceSpec) []string {
	capabilityMap := CapabilityMapFromHostServices(specs)
	capabilities := make([]string, 0, len(capabilityMap))
	for capability := range capabilityMap {
		capabilities = append(capabilities, capability)
	}
	sort.Strings(capabilities)
	return capabilities
}

// CapabilityMapFromHostServices returns the capability set implied by one
// normalized host service declaration set.
func CapabilityMapFromHostServices(specs []*HostServiceSpec) map[string]struct{} {
	capabilities := make(map[string]struct{})
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		service := normalizeHostServiceName(spec.Service)
		methods := spec.Methods
		if len(methods) == 0 {
			methods = defaultHostServiceMethods(service)
		}
		for _, rawMethod := range methods {
			method := normalizeHostServiceMethod(rawMethod)
			capability := RequiredCapabilityForHostServiceMethod(service, method)
			if capability != "" {
				capabilities[capability] = struct{}{}
			}
		}
	}
	return capabilities
}

// ValidateHostServiceSpecs validates and normalizes host service declarations in-place.
func ValidateHostServiceSpecs(specs []*HostServiceSpec) error {
	if len(specs) == 0 {
		return nil
	}

	seenServices := make(map[string]struct{}, len(specs))
	for _, spec := range specs {
		if spec == nil {
			return gerror.New("host service declaration cannot be nil")
		}
		spec.Service = normalizeHostServiceName(spec.Service)
		if spec.Service == "" {
			return gerror.New("host service name cannot be empty")
		}
		if _, ok := hostServiceMethodCapabilityMap[spec.Service]; !ok {
			return gerror.Newf("unknown host service declaration: %s", spec.Service)
		}
		if _, exists := seenServices[spec.Service]; exists {
			return gerror.Newf("host service cannot be declared more than once: %s", spec.Service)
		}
		seenServices[spec.Service] = struct{}{}

		if len(spec.Methods) == 0 {
			spec.Methods = defaultHostServiceMethods(spec.Service)
		}
		methodSeen := make(map[string]struct{}, len(spec.Methods))
		methods := make([]string, 0, len(spec.Methods))
		for _, rawMethod := range spec.Methods {
			method := normalizeHostServiceMethod(rawMethod)
			if method == "" {
				return gerror.Newf("host service %s method cannot be empty", spec.Service)
			}
			if RequiredCapabilityForHostServiceMethod(spec.Service, method) == "" {
				return gerror.Newf("host service %s does not support method: %s", spec.Service, method)
			}
			if _, exists := methodSeen[method]; exists {
				return gerror.Newf("host service %s method cannot be duplicated: %s", spec.Service, method)
			}
			methodSeen[method] = struct{}{}
			methods = append(methods, method)
		}
		if len(methods) == 0 {
			return gerror.Newf("host service %s must declare at least one method", spec.Service)
		}
		sort.Strings(methods)
		spec.Methods = methods

		tableSeen := make(map[string]struct{}, len(spec.Tables))
		tables := make([]string, 0, len(spec.Tables))
		for _, rawTable := range spec.Tables {
			table := strings.TrimSpace(rawTable)
			if table == "" {
				return gerror.Newf("host service %s table cannot be empty", spec.Service)
			}
			if _, exists := tableSeen[table]; exists {
				return gerror.Newf("host service %s table cannot be duplicated: %s", spec.Service, table)
			}
			tableSeen[table] = struct{}{}
			tables = append(tables, table)
		}
		sort.Strings(tables)
		spec.Tables = tables

		pathSeen := make(map[string]struct{}, len(spec.Paths))
		paths := make([]string, 0, len(spec.Paths))
		for _, rawPath := range spec.Paths {
			normalizedPath, err := normalizeStorageDeclaredPath(rawPath)
			if err != nil {
				return gerror.Wrapf(err, "host service %s has invalid path", spec.Service)
			}
			if _, exists := pathSeen[normalizedPath]; exists {
				return gerror.Newf("host service %s path cannot be duplicated: %s", spec.Service, normalizedPath)
			}
			pathSeen[normalizedPath] = struct{}{}
			paths = append(paths, normalizedPath)
		}
		sort.Strings(paths)
		spec.Paths = paths

		resourceSeen := make(map[string]struct{}, len(spec.Resources))
		resources := make([]*HostServiceResourceSpec, 0, len(spec.Resources))
		for _, resource := range spec.Resources {
			if resource == nil {
				return gerror.Newf("host service %s resource declaration cannot be nil", spec.Service)
			}
			resource.Ref = strings.TrimSpace(resource.Ref)
			if resource.Ref == "" {
				return gerror.Newf("host service %s resource ref cannot be empty", spec.Service)
			}
			if _, exists := resourceSeen[resource.Ref]; exists {
				return gerror.Newf("host service %s resource ref cannot be duplicated: %s", spec.Service, resource.Ref)
			}
			resourceSeen[resource.Ref] = struct{}{}
			resource.AllowMethods = normalizeUpperStringSlice(resource.AllowMethods)
			resource.HeaderAllowList = normalizeLowerStringSlice(resource.HeaderAllowList)
			resource.Attributes = normalizeStringMap(resource.Attributes)
			resources = append(resources, resource)
		}
		sort.Slice(resources, func(i, j int) bool {
			return resources[i].Ref < resources[j].Ref
		})
		spec.Resources = resources

		if _, ok := hostServicesWithPaths[spec.Service]; ok {
			if len(spec.Tables) > 0 {
				return gerror.Newf("host service %s cannot declare tables", spec.Service)
			}
			if len(spec.Resources) > 0 {
				return gerror.Newf("host service %s cannot declare resource refs", spec.Service)
			}
			if len(spec.Paths) == 0 {
				return gerror.Newf("host service %s must declare at least one path", spec.Service)
			}
			continue
		}

		if _, ok := hostServicesWithTables[spec.Service]; ok {
			if len(spec.Paths) > 0 {
				return gerror.Newf("host service %s cannot declare paths", spec.Service)
			}
			if len(spec.Resources) > 0 {
				return gerror.Newf("host service %s cannot declare resources", spec.Service)
			}
			if len(spec.Tables) == 0 {
				return gerror.Newf("host service %s must declare at least one table", spec.Service)
			}
			continue
		}
		if len(spec.Tables) > 0 {
			return gerror.Newf("host service %s cannot declare tables", spec.Service)
		}
		if len(spec.Paths) > 0 {
			return gerror.Newf("host service %s cannot declare paths", spec.Service)
		}

		if _, ok := hostServicesWithoutResources[spec.Service]; ok {
			if len(spec.Resources) > 0 {
				return gerror.Newf("host service %s cannot declare resources", spec.Service)
			}
			continue
		}
		if len(spec.Resources) == 0 {
			return gerror.Newf("host service %s must declare at least one resource", spec.Service)
		}
		if spec.Service == HostServiceNetwork {
			for _, resource := range spec.Resources {
				if resource == nil {
					continue
				}
				if len(resource.AllowMethods) > 0 || len(resource.HeaderAllowList) > 0 || resource.TimeoutMs > 0 || resource.MaxBodyBytes > 0 || len(resource.Attributes) > 0 {
					return gerror.Newf("host service %s only allows url declarations and cannot include extra governance fields: %s", spec.Service, resource.Ref)
				}
				if err := validateNetworkURLPattern(resource.Ref); err != nil {
					return gerror.Wrapf(err, "host service %s has invalid url", spec.Service)
				}
			}
		}
	}

	sort.Slice(specs, func(i, j int) bool {
		return specs[i].Service < specs[j].Service
	})
	return nil
}

// defaultHostServiceMethods returns service-specific default method grants.
func defaultHostServiceMethods(service string) []string {
	if service == HostServiceConfig {
		return []string{
			HostServiceMethodConfigGet,
			HostServiceMethodConfigExists,
			HostServiceMethodConfigString,
			HostServiceMethodConfigBool,
			HostServiceMethodConfigInt,
			HostServiceMethodConfigDuration,
		}
	}
	return nil
}

// NormalizeHostServiceSpecs returns deep-cloned and normalized host service declarations.
func NormalizeHostServiceSpecs(specs []*HostServiceSpec) ([]*HostServiceSpec, error) {
	if len(specs) == 0 {
		return []*HostServiceSpec{}, nil
	}
	cloned := make([]*HostServiceSpec, 0, len(specs))
	for _, item := range specs {
		if item == nil {
			continue
		}
		next := &HostServiceSpec{
			Service: normalizeHostServiceName(item.Service),
			Methods: normalizeLowerStringSlice(item.Methods),
			Paths:   normalizeStoragePathSlice(item.Paths),
			Tables:  normalizeTableSlice(item.Tables),
		}
		if len(item.Resources) > 0 {
			next.Resources = make([]*HostServiceResourceSpec, 0, len(item.Resources))
			for _, resource := range item.Resources {
				if resource == nil {
					continue
				}
				next.Resources = append(next.Resources, &HostServiceResourceSpec{
					Ref:             strings.TrimSpace(resource.Ref),
					AllowMethods:    normalizeUpperStringSlice(resource.AllowMethods),
					HeaderAllowList: normalizeLowerStringSlice(resource.HeaderAllowList),
					TimeoutMs:       resource.TimeoutMs,
					MaxBodyBytes:    resource.MaxBodyBytes,
					Attributes:      normalizeStringMap(resource.Attributes),
				})
			}
		}
		cloned = append(cloned, next)
	}
	if err := ValidateHostServiceSpecs(cloned); err != nil {
		return nil, err
	}
	return cloned, nil
}

// MustNormalizeHostServiceSpecs returns normalized declarations or panics for
// compile-time constants whose invalid form must fail fast.
func MustNormalizeHostServiceSpecs(specs []*HostServiceSpec) []*HostServiceSpec {
	normalized, err := NormalizeHostServiceSpecs(specs)
	if err != nil {
		panic(err)
	}
	return normalized
}

// AllCapabilities returns a sorted list of all known capability identifiers.
func AllCapabilities() []string {
	result := make([]string, 0, len(allCapabilities))
	for capability := range allCapabilities {
		result = append(result, capability)
	}
	sort.Strings(result)
	return result
}

// ValidateCapabilities checks that every capability string is recognized.
func ValidateCapabilities(capabilities []string) error {
	for _, capability := range capabilities {
		normalized := strings.TrimSpace(capability)
		if normalized == "" {
			return gerror.New("plugin capability declaration cannot be empty")
		}
		if _, ok := allCapabilities[normalized]; !ok {
			return gerror.Newf("unknown plugin capability declaration: %s, supported values: %v", normalized, AllCapabilities())
		}
	}
	return nil
}

// NormalizeCapabilities trims whitespace and removes duplicates from a capability list.
func NormalizeCapabilities(capabilities []string) []string {
	seen := make(map[string]struct{}, len(capabilities))
	result := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		normalized := strings.TrimSpace(capability)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

// CapabilitySliceToMap converts a capability slice to a set for O(1) lookup.
func CapabilitySliceToMap(capabilities []string) map[string]struct{} {
	result := make(map[string]struct{}, len(capabilities))
	for _, capability := range capabilities {
		normalized := strings.TrimSpace(capability)
		if normalized != "" {
			result[normalized] = struct{}{}
		}
	}
	return result
}

// normalizeHostServiceName trims and lowercases one host service identifier.
func normalizeHostServiceName(value string) string {
	return normalizeLower(value, "")
}

// normalizeHostServiceMethod trims and lowercases one host service method name.
func normalizeHostServiceMethod(value string) string {
	return normalizeLower(value, "")
}

// normalizeLower trims and lowercases one string, applying the default when
// the normalized result is empty.
func normalizeLower(value string, defaultValue string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return defaultValue
	}
	return normalized
}

// normalizeStoragePathSlice normalizes declared storage paths and drops invalid
// entries for clone-style normalization flows.
func normalizeStoragePathSlice(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	items := make([]string, 0, len(paths))
	for _, rawPath := range paths {
		normalizedPath, err := normalizeStorageDeclaredPath(rawPath)
		if err != nil {
			continue
		}
		items = append(items, normalizedPath)
	}
	sort.Strings(items)
	return items
}

// normalizeStorageDeclaredPath validates one logical storage path and preserves
// trailing-slash semantics for prefix grants.
func normalizeStorageDeclaredPath(value string) (string, error) {
	raw := strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	if raw == "" {
		return "", gerror.New("path cannot be empty")
	}
	if strings.HasPrefix(raw, "/") {
		return "", gerror.Newf("path cannot be absolute: %s", value)
	}
	if len(raw) >= 2 && ((raw[0] >= 'A' && raw[0] <= 'Z') || (raw[0] >= 'a' && raw[0] <= 'z')) && raw[1] == ':' {
		return "", gerror.Newf("path cannot contain a host drive prefix: %s", value)
	}

	isPrefix := strings.HasSuffix(raw, "/")
	trimmed := strings.TrimSuffix(raw, "/")
	if trimmed == "" {
		return "", gerror.New("path cannot be empty")
	}

	normalized := path.Clean(trimmed)
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		return "", gerror.Newf("path is invalid: %s", value)
	}
	if isPrefix {
		return normalized + "/", nil
	}
	return normalized, nil
}

// validateNetworkURLPattern validates one authorized network URL pattern before
// it is accepted into manifest state.
func validateNetworkURLPattern(rawValue string) error {
	trimmed := strings.TrimSpace(rawValue)
	if trimmed == "" {
		return gerror.New("url cannot be empty")
	}
	if !strings.Contains(trimmed, "://") {
		return gerror.New("url must include a scheme")
	}
	if strings.Contains(trimmed, "?") || strings.Contains(trimmed, "#") {
		return gerror.New("url pattern cannot include query or fragment")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return gerror.Wrap(err, "failed to parse url pattern")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return gerror.New("url scheme only supports http/https")
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return gerror.New("url is missing host")
	}
	return nil
}

// normalizeLowerStringSlice trims, lowercases, de-duplicates, and sorts one
// string slice.
func normalizeLowerStringSlice(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := normalizeLower(item, "")
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

// normalizeUpperStringSlice trims, uppercases, de-duplicates, and sorts one
// string slice.
func normalizeUpperStringSlice(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.ToUpper(strings.TrimSpace(item))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

// normalizeTableSlice trims, de-duplicates, and sorts declared data table
// names.
func normalizeTableSlice(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

// normalizeStringMap trims keys and values while discarding empty keys from a
// metadata map.
func normalizeStringMap(items map[string]string) map[string]string {
	if len(items) == 0 {
		return nil
	}
	result := make(map[string]string, len(items))
	for key, value := range items {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		result[trimmedKey] = strings.TrimSpace(value)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
