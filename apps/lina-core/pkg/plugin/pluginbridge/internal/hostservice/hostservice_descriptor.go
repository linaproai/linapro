// This file keeps host-service method governance metadata in one place so
// capability derivation, resource validation, public protocol aliases, guest
// clients, stubs, and host dispatchers cannot drift silently.

package hostservice

// HostServiceResourceKind describes which authorization resource shape a host
// service declaration uses in plugin manifests.
type HostServiceResourceKind string

// Host-service resource kinds used by manifest validation and governance tests.
const (
	HostServiceResourceNone     HostServiceResourceKind = "none"
	HostServiceResourcePath     HostServiceResourceKind = "path"
	HostServiceResourceTable    HostServiceResourceKind = "table"
	HostServiceResourceKey      HostServiceResourceKind = "key"
	HostServiceResourceRef      HostServiceResourceKind = "resource"
	HostServiceResourceReserved HostServiceResourceKind = "reserved"
)

// HostServiceDescriptor describes one logical host service family.
type HostServiceDescriptor struct {
	// Service is the logical host service identifier.
	Service string
	// ResourceKind describes the manifest resource declaration shape.
	ResourceKind HostServiceResourceKind
	// DefaultMethods are applied when a manifest declares the service without methods.
	DefaultMethods []string
	// Methods lists governed methods under this service.
	Methods []HostServiceMethodDescriptor
}

// HostServiceMethodDescriptor describes one governed host service method.
type HostServiceMethodDescriptor struct {
	// Service is populated when descriptors are flattened.
	Service string
	// Method is the wire method string.
	Method string
	// MethodConst is the stable Go constant name for the method.
	MethodConst string
	// Capability is the capability implied by this method.
	Capability string
	// ResourceKind describes the manifest resource declaration shape.
	ResourceKind HostServiceResourceKind
	// RequestPayload names the public request payload type when one exists.
	RequestPayload string
	// ResponsePayload names the public response payload type when one exists.
	ResponsePayload string
	// Published reports whether this method is implemented as a guest-callable protocol.
	Published bool
	// GuestClient reports whether a guest SDK helper is expected to call this method.
	GuestClient bool
	// Dispatcher reports whether the wasm host dispatcher must handle this method.
	Dispatcher bool
}

var hostServiceDescriptors = []HostServiceDescriptor{
	{
		Service:      HostServiceRuntime,
		ResourceKind: HostServiceResourceNone,
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodRuntimeLogWrite, "HostServiceMethodRuntimeLogWrite", CapabilityRuntime, "HostCallLogRequest", ""),
			hostMethod(HostServiceMethodRuntimeStateGet, "HostServiceMethodRuntimeStateGet", CapabilityRuntime, "HostCallStateGetRequest", "HostCallStateGetResponse"),
			hostMethod(HostServiceMethodRuntimeStateSet, "HostServiceMethodRuntimeStateSet", CapabilityRuntime, "HostCallStateSetRequest", ""),
			hostMethod(HostServiceMethodRuntimeStateDelete, "HostServiceMethodRuntimeStateDelete", CapabilityRuntime, "HostCallStateDeleteRequest", ""),
			hostMethod(HostServiceMethodRuntimeInfoNow, "HostServiceMethodRuntimeInfoNow", CapabilityRuntime, "", "HostServiceValueResponse"),
			hostMethod(HostServiceMethodRuntimeInfoUUID, "HostServiceMethodRuntimeInfoUUID", CapabilityRuntime, "", "HostServiceValueResponse"),
			hostMethod(HostServiceMethodRuntimeInfoNode, "HostServiceMethodRuntimeInfoNode", CapabilityRuntime, "", "HostServiceValueResponse"),
		},
	},
	{
		Service:      HostServiceCron,
		ResourceKind: HostServiceResourceNone,
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodCronRegister, "HostServiceMethodCronRegister", CapabilityCron, "HostServiceCronRegisterRequest", ""),
		},
	},
	{
		Service:      HostServiceStorage,
		ResourceKind: HostServiceResourcePath,
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodStoragePut, "HostServiceMethodStoragePut", CapabilityStorage, "HostServiceStoragePutRequest", "HostServiceStoragePutResponse"),
			hostMethod(HostServiceMethodStorageGet, "HostServiceMethodStorageGet", CapabilityStorage, "HostServiceStorageGetRequest", "HostServiceStorageGetResponse"),
			hostMethod(HostServiceMethodStorageDelete, "HostServiceMethodStorageDelete", CapabilityStorage, "HostServiceStorageDeleteRequest", ""),
			hostMethod(HostServiceMethodStorageList, "HostServiceMethodStorageList", CapabilityStorage, "HostServiceStorageListRequest", "HostServiceStorageListResponse"),
			hostMethod(HostServiceMethodStorageStat, "HostServiceMethodStorageStat", CapabilityStorage, "HostServiceStorageStatRequest", "HostServiceStorageStatResponse"),
		},
	},
	{
		Service:      HostServiceNetwork,
		ResourceKind: HostServiceResourceRef,
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodNetworkRequest, "HostServiceMethodNetworkRequest", CapabilityHTTPRequest, "HostServiceNetworkRequest", "HostServiceNetworkResponse"),
		},
	},
	{
		Service:      HostServiceData,
		ResourceKind: HostServiceResourceTable,
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodDataList, "HostServiceMethodDataList", CapabilityDataRead, "HostServiceDataListRequest", "HostServiceDataListResponse"),
			hostMethod(HostServiceMethodDataGet, "HostServiceMethodDataGet", CapabilityDataRead, "HostServiceDataGetRequest", "HostServiceDataGetResponse"),
			hostMethod(HostServiceMethodDataCreate, "HostServiceMethodDataCreate", CapabilityDataMutate, "HostServiceDataMutationRequest", "HostServiceDataMutationResponse"),
			hostMethod(HostServiceMethodDataUpdate, "HostServiceMethodDataUpdate", CapabilityDataMutate, "HostServiceDataMutationRequest", "HostServiceDataMutationResponse"),
			hostMethod(HostServiceMethodDataDelete, "HostServiceMethodDataDelete", CapabilityDataMutate, "HostServiceDataMutationRequest", "HostServiceDataMutationResponse"),
			hostMethod(HostServiceMethodDataTransaction, "HostServiceMethodDataTransaction", CapabilityDataMutate, "HostServiceDataTransactionRequest", "HostServiceDataTransactionResponse"),
		},
	},
	{
		Service:      HostServiceCache,
		ResourceKind: HostServiceResourceRef,
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodCacheGet, "HostServiceMethodCacheGet", CapabilityCache, "HostServiceCacheGetRequest", "HostServiceCacheGetResponse"),
			hostMethod(HostServiceMethodCacheSet, "HostServiceMethodCacheSet", CapabilityCache, "HostServiceCacheSetRequest", "HostServiceCacheSetResponse"),
			hostMethod(HostServiceMethodCacheDelete, "HostServiceMethodCacheDelete", CapabilityCache, "HostServiceCacheDeleteRequest", ""),
			hostMethod(HostServiceMethodCacheIncr, "HostServiceMethodCacheIncr", CapabilityCache, "HostServiceCacheIncrRequest", "HostServiceCacheIncrResponse"),
			hostMethod(HostServiceMethodCacheExpire, "HostServiceMethodCacheExpire", CapabilityCache, "HostServiceCacheExpireRequest", "HostServiceCacheExpireResponse"),
		},
	},
	{
		Service:      HostServiceLock,
		ResourceKind: HostServiceResourceRef,
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodLockAcquire, "HostServiceMethodLockAcquire", CapabilityLock, "HostServiceLockAcquireRequest", "HostServiceLockAcquireResponse"),
			hostMethod(HostServiceMethodLockRenew, "HostServiceMethodLockRenew", CapabilityLock, "HostServiceLockRenewRequest", "HostServiceLockRenewResponse"),
			hostMethod(HostServiceMethodLockRelease, "HostServiceMethodLockRelease", CapabilityLock, "HostServiceLockReleaseRequest", ""),
		},
	},
	{
		Service:      HostServiceSecret,
		ResourceKind: HostServiceResourceRef,
		Methods: []HostServiceMethodDescriptor{
			reservedHostMethod("resolve", CapabilitySecret),
		},
	},
	{
		Service:      HostServiceEvent,
		ResourceKind: HostServiceResourceRef,
		Methods: []HostServiceMethodDescriptor{
			reservedHostMethod("publish", CapabilityEventPublish),
		},
	},
	{
		Service:      HostServiceQueue,
		ResourceKind: HostServiceResourceRef,
		Methods: []HostServiceMethodDescriptor{
			reservedHostMethod("enqueue", CapabilityQueueEnqueue),
		},
	},
	{
		Service:      HostServiceNotify,
		ResourceKind: HostServiceResourceRef,
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodNotifySend, "HostServiceMethodNotifySend", CapabilityNotify, "HostServiceNotifySendRequest", "HostServiceNotifySendResponse"),
		},
	},
	{
		Service:        HostServiceConfig,
		ResourceKind:   HostServiceResourceNone,
		DefaultMethods: []string{HostServiceMethodConfigGet},
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodConfigGet, "HostServiceMethodConfigGet", CapabilityConfig, "HostServiceConfigKeyRequest", "HostServiceConfigValueResponse"),
		},
	},
	{
		Service:        HostServiceHostConfig,
		ResourceKind:   HostServiceResourceKey,
		DefaultMethods: []string{HostServiceMethodHostConfigGet},
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodHostConfigGet, "HostServiceMethodHostConfigGet", CapabilityHostConfig, "HostServiceConfigKeyRequest", "HostServiceConfigValueResponse"),
		},
	},
	{
		Service:        HostServiceManifest,
		ResourceKind:   HostServiceResourcePath,
		DefaultMethods: []string{HostServiceMethodManifestGet},
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodManifestGet, "HostServiceMethodManifestGet", CapabilityManifest, "HostServiceManifestGetRequest", "HostServiceManifestGetResponse"),
		},
	},
	{
		Service:      HostServiceAI,
		ResourceKind: HostServiceResourceRef,
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodAITextGenerate, "HostServiceMethodAITextGenerate", CapabilityAIText, "HostServiceAITextGenerateRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAIImageGenerate, "HostServiceMethodAIImageGenerate", CapabilityAIImage, "HostServiceAIImageGenerateRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAIImageEdit, "HostServiceMethodAIImageEdit", CapabilityAIImage, "HostServiceAIImageEditRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAIEmbeddingCreate, "HostServiceMethodAIEmbeddingCreate", CapabilityAIEmbedding, "HostServiceAIEmbeddingCreateRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAIAudioTranscribe, "HostServiceMethodAIAudioTranscribe", CapabilityAIAudio, "HostServiceAIAudioTranscribeRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAIAudioSynthesize, "HostServiceMethodAIAudioSynthesize", CapabilityAIAudio, "HostServiceAIAudioSynthesizeRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAIVisionAnalyze, "HostServiceMethodAIVisionAnalyze", CapabilityAIVision, "HostServiceAIVisionAnalyzeRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAIDocumentAnalyze, "HostServiceMethodAIDocumentAnalyze", CapabilityAIDocument, "HostServiceAIDocumentAnalyzeRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAIDocumentCite, "HostServiceMethodAIDocumentCite", CapabilityAIDocument, "HostServiceAIDocumentCiteRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAISafetyModerate, "HostServiceMethodAISafetyModerate", CapabilityAISafety, "HostServiceAISafetyModerateRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAIVideoGenerate, "HostServiceMethodAIVideoGenerate", CapabilityAIVideo, "HostServiceAIVideoGenerateRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAIVideoEdit, "HostServiceMethodAIVideoEdit", CapabilityAIVideo, "HostServiceAIVideoEditRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAIVideoExtend, "HostServiceMethodAIVideoExtend", CapabilityAIVideo, "HostServiceAIVideoExtendRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAIVideoOperationGet, "HostServiceMethodAIVideoOperationGet", CapabilityAIVideo, "HostServiceAIVideoOperationGetRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodAIVideoOperationCancel, "HostServiceMethodAIVideoOperationCancel", CapabilityAIVideo, "HostServiceAIVideoOperationCancelRequest", "HostServiceCapabilityJSONResponse"),
		},
	},
	{
		Service:      HostServiceOrg,
		ResourceKind: HostServiceResourceNone,
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodOrgAvailable, "HostServiceMethodOrgAvailable", CapabilityOrg, "", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodOrgStatus, "HostServiceMethodOrgStatus", CapabilityOrg, "", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodOrgListUserDeptAssignments, "HostServiceMethodOrgListUserDeptAssignments", CapabilityOrg, "HostServiceCapabilityUsersRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodOrgGetUserDeptInfo, "HostServiceMethodOrgGetUserDeptInfo", CapabilityOrg, "HostServiceCapabilityUserRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodOrgGetUserDeptName, "HostServiceMethodOrgGetUserDeptName", CapabilityOrg, "HostServiceCapabilityUserRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodOrgGetUserDeptIDs, "HostServiceMethodOrgGetUserDeptIDs", CapabilityOrg, "HostServiceCapabilityUserRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodOrgGetUserPostIDs, "HostServiceMethodOrgGetUserPostIDs", CapabilityOrg, "HostServiceCapabilityUserRequest", "HostServiceCapabilityJSONResponse"),
		},
	},
	{
		Service:      HostServiceTenant,
		ResourceKind: HostServiceResourceNone,
		Methods: []HostServiceMethodDescriptor{
			hostMethod(HostServiceMethodTenantAvailable, "HostServiceMethodTenantAvailable", CapabilityTenant, "", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodTenantStatus, "HostServiceMethodTenantStatus", CapabilityTenant, "", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodTenantCurrent, "HostServiceMethodTenantCurrent", CapabilityTenant, "", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodTenantPlatformBypass, "HostServiceMethodTenantPlatformBypass", CapabilityTenant, "", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodTenantEnsureVisible, "HostServiceMethodTenantEnsureVisible", CapabilityTenant, "HostServiceCapabilityTenantRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodTenantValidateUserInTenant, "HostServiceMethodTenantValidateUserInTenant", CapabilityTenant, "HostServiceCapabilityUserTenantRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodTenantListUserTenants, "HostServiceMethodTenantListUserTenants", CapabilityTenant, "HostServiceCapabilityUserRequest", "HostServiceCapabilityJSONResponse"),
			hostMethod(HostServiceMethodTenantValidateSwitch, "HostServiceMethodTenantValidateSwitch", CapabilityTenant, "HostServiceCapabilityUserTenantSwitchRequest", "HostServiceCapabilityJSONResponse"),
		},
	},
}

// HostServiceDescriptors returns a copy of the governed host service descriptor table.
func HostServiceDescriptors() []HostServiceDescriptor {
	result := make([]HostServiceDescriptor, 0, len(hostServiceDescriptors))
	for _, descriptor := range hostServiceDescriptors {
		item := descriptor
		item.DefaultMethods = append([]string(nil), descriptor.DefaultMethods...)
		item.Methods = make([]HostServiceMethodDescriptor, 0, len(descriptor.Methods))
		for _, method := range descriptor.Methods {
			method.Service = descriptor.Service
			if method.ResourceKind == "" {
				method.ResourceKind = descriptor.ResourceKind
			}
			item.Methods = append(item.Methods, method)
		}
		result = append(result, item)
	}
	return result
}

// HostServiceMethodDescriptors returns all governed host-service method descriptors.
func HostServiceMethodDescriptors() []HostServiceMethodDescriptor {
	return append([]HostServiceMethodDescriptor(nil), hostServiceMethodDescriptors()...)
}

func hostServiceMethodDescriptors() []HostServiceMethodDescriptor {
	methods := make([]HostServiceMethodDescriptor, 0)
	for _, service := range hostServiceDescriptors {
		for _, method := range service.Methods {
			item := method
			item.Service = service.Service
			if item.ResourceKind == "" {
				item.ResourceKind = service.ResourceKind
			}
			methods = append(methods, item)
		}
	}
	return methods
}

func hostMethod(
	method string,
	methodConst string,
	capability string,
	requestPayload string,
	responsePayload string,
) HostServiceMethodDescriptor {
	return HostServiceMethodDescriptor{
		Method:          method,
		MethodConst:     methodConst,
		Capability:      capability,
		RequestPayload:  requestPayload,
		ResponsePayload: responsePayload,
		Published:       true,
		GuestClient:     true,
		Dispatcher:      true,
	}
}

func reservedHostMethod(method string, capability string) HostServiceMethodDescriptor {
	return HostServiceMethodDescriptor{
		Method:       method,
		Capability:   capability,
		ResourceKind: HostServiceResourceReserved,
	}
}

func buildHostServiceMethodCapabilityMap() map[string]map[string]string {
	result := make(map[string]map[string]string)
	for _, descriptor := range hostServiceMethodDescriptors() {
		if descriptor.Service == "" || descriptor.Method == "" || descriptor.Capability == "" {
			continue
		}
		if result[descriptor.Service] == nil {
			result[descriptor.Service] = make(map[string]string)
		}
		result[descriptor.Service][descriptor.Method] = descriptor.Capability
	}
	return result
}

func buildHostServiceCapabilitySet() map[string]struct{} {
	result := make(map[string]struct{})
	for _, descriptor := range hostServiceMethodDescriptors() {
		if descriptor.Capability != "" {
			result[descriptor.Capability] = struct{}{}
		}
	}
	return result
}

func buildHostServiceDefaultMethods() map[string][]string {
	result := make(map[string][]string)
	for _, descriptor := range hostServiceDescriptors {
		if len(descriptor.DefaultMethods) == 0 {
			continue
		}
		result[descriptor.Service] = append([]string(nil), descriptor.DefaultMethods...)
	}
	return result
}

func buildHostServiceResourceKindSet(kind HostServiceResourceKind) map[string]struct{} {
	result := make(map[string]struct{})
	for _, descriptor := range hostServiceDescriptors {
		if descriptor.ResourceKind == kind {
			result[descriptor.Service] = struct{}{}
		}
	}
	return result
}
