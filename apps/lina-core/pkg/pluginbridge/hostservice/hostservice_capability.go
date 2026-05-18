// This file implements host-service capability lookup and capability list normalization.

package hostservice

import (
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

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
