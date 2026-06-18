// Package hostservices defines the public host-service catalog used by
// protocol governance, guest SDK coverage checks, and host dispatcher coverage
// checks. Runtime validation can derive private lookup tables from this catalog,
// but service and method metadata must be maintained here first.
package hostservices

// ResourceKind describes which authorization resource shape a host service
// declaration uses in plugin manifests.
type ResourceKind string

// Host-service resource kinds used by manifest validation and governance tests.
const (
	ResourceKindNone     ResourceKind = "none"
	ResourceKindPath     ResourceKind = "path"
	ResourceKindTable    ResourceKind = "table"
	ResourceKindKey      ResourceKind = "key"
	ResourceKindRef      ResourceKind = "resource"
	ResourceKindReserved ResourceKind = "reserved"
)

// PayloadKind describes the host-service payload codec family used by one
// method. The field lets governance tests distinguish ordinary JSON envelopes
// from services that intentionally keep dedicated binary codecs.
type PayloadKind string

// Host-service payload codec families.
const (
	PayloadKindNone      PayloadKind = "none"
	PayloadKindJSON      PayloadKind = "json"
	PayloadKindDedicated PayloadKind = "dedicated"
	PayloadKindReserved  PayloadKind = "reserved"
)

// ServiceDescriptor describes one logical host service family.
type ServiceDescriptor struct {
	// Service is the logical host service identifier.
	Service string
	// ResourceKind describes the manifest resource declaration shape.
	ResourceKind ResourceKind
	// Methods lists governed methods under this service.
	Methods []MethodDescriptor
}

// MethodDescriptor describes one governed host service method.
type MethodDescriptor struct {
	// Service is populated when descriptors are flattened.
	Service string
	// Method is the wire method string.
	Method string
	// MethodConst is the stable Go constant name for the method.
	MethodConst string
	// Capability is the capability implied by this method.
	Capability string
	// ResourceKind describes the manifest resource declaration shape.
	ResourceKind ResourceKind
	// RequestPayload names the public request payload type when one exists.
	RequestPayload string
	// ResponsePayload names the public response payload type when one exists.
	ResponsePayload string
	// PayloadKind identifies the codec family used by this method.
	PayloadKind PayloadKind
	// Published reports whether this method is implemented as a guest-callable protocol.
	Published bool
	// GuestClient reports whether a guest SDK helper is expected to call this method.
	GuestClient bool
	// Dispatcher reports whether the wasm host dispatcher must handle this method.
	Dispatcher bool
}

var catalog = []ServiceDescriptor{
	{
		Service:      "runtime",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("log.write", "HostServiceMethodRuntimeLogWrite", "host:runtime", "HostCallLogRequest", ""),
			hostMethod("state.get", "HostServiceMethodRuntimeStateGet", "host:runtime", "HostCallStateGetRequest", "HostCallStateGetResponse"),
			hostMethod("state.get_many", "HostServiceMethodRuntimeStateGetMany", "host:runtime", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("state.set", "HostServiceMethodRuntimeStateSet", "host:runtime", "HostCallStateSetRequest", ""),
			hostMethod("state.set_many", "HostServiceMethodRuntimeStateSetMany", "host:runtime", "HostServiceJSONRequest", ""),
			hostMethod("state.delete", "HostServiceMethodRuntimeStateDelete", "host:runtime", "HostCallStateDeleteRequest", ""),
			hostMethod("state.delete_many", "HostServiceMethodRuntimeStateDeleteMany", "host:runtime", "HostServiceJSONRequest", ""),
			hostMethod("info.now", "HostServiceMethodRuntimeInfoNow", "host:runtime", "", "HostServiceValueResponse"),
			hostMethod("info.uuid", "HostServiceMethodRuntimeInfoUUID", "host:runtime", "", "HostServiceValueResponse"),
			hostMethod("info.node", "HostServiceMethodRuntimeInfoNode", "host:runtime", "", "HostServiceValueResponse"),
		},
	},
	{
		Service:      "storage",
		ResourceKind: ResourceKindPath,
		Methods: []MethodDescriptor{
			hostMethod("put", "HostServiceMethodStoragePut", "host:storage", "HostServiceStoragePutRequest", "HostServiceStoragePutResponse"),
			hostMethod("put.init", "HostServiceMethodStoragePutInit", "host:storage", "HostServiceStoragePutInitRequest", "HostServiceStoragePutInitResponse"),
			hostMethod("put.chunk", "HostServiceMethodStoragePutChunk", "host:storage", "HostServiceStoragePutChunkRequest", "HostServiceStoragePutChunkResponse"),
			hostMethod("put.commit", "HostServiceMethodStoragePutCommit", "host:storage", "HostServiceStoragePutCommitRequest", "HostServiceStoragePutCommitResponse"),
			hostMethod("put.abort", "HostServiceMethodStoragePutAbort", "host:storage", "HostServiceStoragePutAbortRequest", ""),
			hostMethod("get", "HostServiceMethodStorageGet", "host:storage", "HostServiceStorageGetRequest", "HostServiceStorageGetResponse"),
			hostMethod("delete", "HostServiceMethodStorageDelete", "host:storage", "HostServiceStorageDeleteRequest", ""),
			hostMethod("delete.batch", "HostServiceMethodStorageDeleteBatch", "host:storage", "HostServiceJSONRequest", ""),
			hostMethod("list", "HostServiceMethodStorageList", "host:storage", "HostServiceStorageListRequest", "HostServiceStorageListResponse"),
			hostMethod("list.cursor", "HostServiceMethodStorageListCursor", "host:storage", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("stat", "HostServiceMethodStorageStat", "host:storage", "HostServiceStorageStatRequest", "HostServiceStorageStatResponse"),
			hostMethod("stat.batch", "HostServiceMethodStorageStatBatch", "host:storage", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "network",
		ResourceKind: ResourceKindRef,
		Methods: []MethodDescriptor{
			hostMethod("request", "HostServiceMethodNetworkRequest", "host:http:request", "HostServiceNetworkRequest", "HostServiceNetworkResponse"),
		},
	},
	{
		Service:      "data",
		ResourceKind: ResourceKindTable,
		Methods: []MethodDescriptor{
			hostMethod("list", "HostServiceMethodDataList", "host:data:read", "HostServiceDataListRequest", "HostServiceDataListResponse"),
			hostMethod("get", "HostServiceMethodDataGet", "host:data:read", "HostServiceDataGetRequest", "HostServiceDataGetResponse"),
			hostMethod("batch_get", "HostServiceMethodDataBatchGet", "host:data:read", "HostServiceDataBatchGetRequest", "HostServiceDataBatchGetResponse"),
			hostMethod("create", "HostServiceMethodDataCreate", "host:data:mutate", "HostServiceDataMutationRequest", "HostServiceDataMutationResponse"),
			hostMethod("update", "HostServiceMethodDataUpdate", "host:data:mutate", "HostServiceDataMutationRequest", "HostServiceDataMutationResponse"),
			hostMethod("delete", "HostServiceMethodDataDelete", "host:data:mutate", "HostServiceDataMutationRequest", "HostServiceDataMutationResponse"),
			hostMethod("transaction", "HostServiceMethodDataTransaction", "host:data:mutate", "HostServiceDataTransactionRequest", "HostServiceDataTransactionResponse"),
		},
	},
	{
		Service:      "cache",
		ResourceKind: ResourceKindRef,
		Methods: []MethodDescriptor{
			hostMethod("get", "HostServiceMethodCacheGet", "host:cache", "HostServiceCacheGetRequest", "HostServiceCacheGetResponse"),
			hostMethod("get_many", "HostServiceMethodCacheGetMany", "host:cache", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("set", "HostServiceMethodCacheSet", "host:cache", "HostServiceCacheSetRequest", "HostServiceCacheSetResponse"),
			hostMethod("set_many", "HostServiceMethodCacheSetMany", "host:cache", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("delete", "HostServiceMethodCacheDelete", "host:cache", "HostServiceCacheDeleteRequest", ""),
			hostMethod("delete_many", "HostServiceMethodCacheDeleteMany", "host:cache", "HostServiceJSONRequest", ""),
			hostMethod("incr", "HostServiceMethodCacheIncr", "host:cache", "HostServiceCacheIncrRequest", "HostServiceCacheIncrResponse"),
			hostMethod("expire", "HostServiceMethodCacheExpire", "host:cache", "HostServiceCacheExpireRequest", "HostServiceCacheExpireResponse"),
		},
	},
	{
		Service:      "lock",
		ResourceKind: ResourceKindRef,
		Methods: []MethodDescriptor{
			hostMethod("acquire", "HostServiceMethodLockAcquire", "host:lock", "HostServiceLockAcquireRequest", "HostServiceLockAcquireResponse"),
			hostMethod("renew", "HostServiceMethodLockRenew", "host:lock", "HostServiceLockRenewRequest", "HostServiceLockRenewResponse"),
			hostMethod("release", "HostServiceMethodLockRelease", "host:lock", "HostServiceLockReleaseRequest", ""),
		},
	},
	{
		Service:      "secret",
		ResourceKind: ResourceKindRef,
		Methods: []MethodDescriptor{
			reservedHostMethod("resolve", "host:secret"),
		},
	},
	{
		Service:      "event",
		ResourceKind: ResourceKindRef,
		Methods: []MethodDescriptor{
			reservedHostMethod("publish", "host:event:publish"),
		},
	},
	{
		Service:      "queue",
		ResourceKind: ResourceKindRef,
		Methods: []MethodDescriptor{
			reservedHostMethod("enqueue", "host:queue:enqueue"),
		},
	},
	{
		Service:      "hostconfig",
		ResourceKind: ResourceKindKey,
		Methods: []MethodDescriptor{
			hostMethod("get", "HostServiceMethodHostConfigGet", "host:hostconfig", "HostServiceConfigKeyRequest", "HostServiceConfigValueResponse"),
		},
	},
	{
		Service:      "manifest",
		ResourceKind: ResourceKindPath,
		Methods: []MethodDescriptor{
			hostMethod("get", "HostServiceMethodManifestGet", "host:manifest", "HostServiceManifestGetRequest", "HostServiceManifestGetResponse"),
			hostMethod("get_many", "HostServiceMethodManifestGetMany", "host:manifest", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("list", "HostServiceMethodManifestList", "host:manifest", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "apidoc",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("route_text.resolve", "HostServiceMethodAPIDocResolveRouteText", "host:apidoc", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("route_texts.resolve", "HostServiceMethodAPIDocResolveRouteTexts", "host:apidoc", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("route_title_operation_keys.find", "HostServiceMethodAPIDocFindRouteTitleOperationKeys", "host:apidoc", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "auth",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("tenant.select", "HostServiceMethodAuthSelectTenant", "host:auth:token", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("tenant.switch", "HostServiceMethodAuthSwitchTenant", "host:auth:token", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("impersonation_token.issue", "HostServiceMethodAuthIssueImpersonationToken", "host:auth:token", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("impersonation_token.revoke", "HostServiceMethodAuthRevokeImpersonationToken", "host:auth:token", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "authz",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("permissions.batch_get", "HostServiceMethodAuthzBatchGetPermissions", "host:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("permissions.batch_has", "HostServiceMethodAuthzBatchHasPermissions", "host:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("permissions.has", "HostServiceMethodAuthzHasPermission", "host:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.platform_admin.check", "HostServiceMethodAuthzIsPlatformAdmin", "host:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "ai",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("text.generate", "HostServiceMethodAITextGenerate", "host:ai:text", "HostServiceAITextGenerateRequest", "HostServiceJSONResponse"),
			hostMethod("text.method_status.get", "HostServiceMethodAITextMethodStatus", "host:ai:text", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("ai.methods.status.batch_get", "HostServiceMethodAIMethodStatuses", "host:ai", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("image.generate", "HostServiceMethodAIImageGenerate", "host:ai:image", "HostServiceAIImageGenerateRequest", "HostServiceJSONResponse"),
			hostMethod("image.edit", "HostServiceMethodAIImageEdit", "host:ai:image", "HostServiceAIImageEditRequest", "HostServiceJSONResponse"),
			hostMethod("embedding.create", "HostServiceMethodAIEmbeddingCreate", "host:ai:embedding", "HostServiceAIEmbeddingCreateRequest", "HostServiceJSONResponse"),
			hostMethod("audio.transcribe", "HostServiceMethodAIAudioTranscribe", "host:ai:audio", "HostServiceAIAudioTranscribeRequest", "HostServiceJSONResponse"),
			hostMethod("audio.synthesize", "HostServiceMethodAIAudioSynthesize", "host:ai:audio", "HostServiceAIAudioSynthesizeRequest", "HostServiceJSONResponse"),
			hostMethod("vision.analyze", "HostServiceMethodAIVisionAnalyze", "host:ai:vision", "HostServiceAIVisionAnalyzeRequest", "HostServiceJSONResponse"),
			hostMethod("document.analyze", "HostServiceMethodAIDocumentAnalyze", "host:ai:document", "HostServiceAIDocumentAnalyzeRequest", "HostServiceJSONResponse"),
			hostMethod("document.cite", "HostServiceMethodAIDocumentCite", "host:ai:document", "HostServiceAIDocumentCiteRequest", "HostServiceJSONResponse"),
			hostMethod("safety.moderate", "HostServiceMethodAISafetyModerate", "host:ai:safety", "HostServiceAISafetyModerateRequest", "HostServiceJSONResponse"),
			hostMethod("video.generate", "HostServiceMethodAIVideoGenerate", "host:ai:video", "HostServiceAIVideoGenerateRequest", "HostServiceJSONResponse"),
			hostMethod("video.edit", "HostServiceMethodAIVideoEdit", "host:ai:video", "HostServiceAIVideoEditRequest", "HostServiceJSONResponse"),
			hostMethod("video.extend", "HostServiceMethodAIVideoExtend", "host:ai:video", "HostServiceAIVideoExtendRequest", "HostServiceJSONResponse"),
			hostMethod("video.operation.get", "HostServiceMethodAIVideoOperationGet", "host:ai:video", "HostServiceAIVideoOperationGetRequest", "HostServiceJSONResponse"),
			hostMethod("video.operation.cancel", "HostServiceMethodAIVideoOperationCancel", "host:ai:video", "HostServiceAIVideoOperationCancelRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "users",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("users.current.get", "HostServiceMethodUsersCurrent", "host:users", "", "HostServiceJSONResponse"),
			hostMethod("users.batch_get", "HostServiceMethodUsersBatchGet", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.resolve.batch", "HostServiceMethodUsersBatchResolve", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.search", "HostServiceMethodUsersSearch", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.visible.ensure", "HostServiceMethodUsersEnsureVisible", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "bizctx",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("current.get", "HostServiceMethodBizCtxCurrent", "host:bizctx", "", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "dict",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("labels.resolve", "HostServiceMethodDictResolveLabels", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.values.list", "HostServiceMethodDictListValues", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("values.visible.ensure", "HostServiceMethodDictEnsureValuesVisible", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "files",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("files.batch_get", "HostServiceMethodFilesBatchGet", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("files.search", "HostServiceMethodFilesSearch", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("files.visible.ensure", "HostServiceMethodFilesEnsureVisible", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "infra",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("status.batch_get", "HostServiceMethodInfraBatchGetStatus", "host:infra", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "jobs",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("jobs.batch_get", "HostServiceMethodJobsBatchGet", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("jobs.search", "HostServiceMethodJobsSearch", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("jobs.visible.ensure", "HostServiceMethodJobsEnsureVisible", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("jobs.register", "HostServiceMethodJobsRegister", "host:jobs", "HostServiceJobsRegisterRequest", ""),
		},
	},
	{
		Service:      "notifications",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("messages.batch_get", "HostServiceMethodNotificationsBatchGetMessages", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("messages.by_source.batch_get", "HostServiceMethodNotificationsBatchGetBySource", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("messages.visible.ensure", "HostServiceMethodNotificationsEnsureVisible", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethodWithResource("messages.send", "HostServiceMethodNotificationsSend", "host:notifications", ResourceKindRef, "HostServiceNotificationsSendRequest", "HostServiceNotificationsSendResponse"),
		},
	},
	{
		Service:      "plugins",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("plugins.current.get", "HostServiceMethodPluginsCurrent", "host:plugins", "", "HostServiceJSONResponse"),
			hostMethod("plugins.batch_get", "HostServiceMethodPluginsBatchGet", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.search", "HostServiceMethodPluginsSearch", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.tenant.list", "HostServiceMethodPluginsListTenant", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.capabilities.status.batch_get", "HostServiceMethodPluginsBatchGetCapabilityStatus", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.enabled.check", "HostServiceMethodPluginsIsEnabled", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.provider_enabled.check", "HostServiceMethodPluginsIsProviderEnabled", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.enabled_authoritative.check", "HostServiceMethodPluginsIsEnabledAuthoritative", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("config.get", "HostServiceMethodPluginsConfigGet", "host:plugins", "HostServiceConfigKeyRequest", "HostServiceConfigValueResponse"),
			hostMethod("lifecycle.tenant_plugin_disable.ensure", "HostServiceMethodPluginsLifecycleEnsureTenantPluginDisable", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("lifecycle.tenant_plugin_disabled.notify", "HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("lifecycle.tenant_delete.ensure", "HostServiceMethodPluginsLifecycleEnsureTenantDelete", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("lifecycle.tenant_deleted.notify", "HostServiceMethodPluginsLifecycleNotifyTenantDeleted", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "route",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("metadata.get", "HostServiceMethodRouteMetadataGet", "host:route", "", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "sessions",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("sessions.current.get", "HostServiceMethodSessionsCurrent", "host:sessions", "", "HostServiceJSONResponse"),
			hostMethod("sessions.search", "HostServiceMethodSessionsSearch", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("sessions.batch_get", "HostServiceMethodSessionsBatchGet", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("sessions.users.online.batch_get", "HostServiceMethodSessionsBatchGetUserOnlineStatus", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("sessions.visible.ensure", "HostServiceMethodSessionsEnsureVisible", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "org",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("capability.available", "HostServiceMethodOrgAvailable", "host:org", "", "HostServiceJSONResponse"),
			hostMethod("capability.status", "HostServiceMethodOrgStatus", "host:org", "", "HostServiceJSONResponse"),
			hostMethod("users.dept_assignments.list", "HostServiceMethodOrgListUserDeptAssignments", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.org_profiles.batch_get", "HostServiceMethodOrgBatchGetUserOrgProfiles", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.dept_info.get", "HostServiceMethodOrgGetUserDeptInfo", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.dept_name.get", "HostServiceMethodOrgGetUserDeptName", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.dept_ids.get", "HostServiceMethodOrgGetUserDeptIDs", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.post_ids.get", "HostServiceMethodOrgGetUserPostIDs", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("depts.tree.list", "HostServiceMethodOrgListDeptTree", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("depts.search", "HostServiceMethodOrgSearchDepartments", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("posts.options.list", "HostServiceMethodOrgListPostOptions", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("depts.visible.ensure", "HostServiceMethodOrgEnsureDepartmentsVisible", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("posts.visible.ensure", "HostServiceMethodOrgEnsurePostsVisible", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "tenant",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("capability.available", "HostServiceMethodTenantAvailable", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod("capability.status", "HostServiceMethodTenantStatus", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod("tenants.current", "HostServiceMethodTenantCurrent", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod("tenants.current_info.get", "HostServiceMethodTenantCurrentInfo", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod("tenants.platform_bypass", "HostServiceMethodTenantPlatformBypass", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod("tenants.visible.ensure", "HostServiceMethodTenantEnsureVisible", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("tenants.batch_get", "HostServiceMethodTenantBatchGet", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("tenants.search", "HostServiceMethodTenantSearch", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.tenant_membership.validate", "HostServiceMethodTenantValidateUserInTenant", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.tenants.list", "HostServiceMethodTenantListUserTenants", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.tenants.batch_list", "HostServiceMethodTenantBatchListUserTenants", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("tenants.visible.batch_ensure", "HostServiceMethodTenantBatchEnsureVisible", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("tenants.switch.validate", "HostServiceMethodTenantValidateSwitch", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
}

// Catalog returns a deep copy of the governed host-service catalog.
func Catalog() []ServiceDescriptor {
	result := make([]ServiceDescriptor, 0, len(catalog))
	for _, descriptor := range catalog {
		item := descriptor
		item.Methods = make([]MethodDescriptor, 0, len(descriptor.Methods))
		for _, method := range descriptor.Methods {
			method.Service = descriptor.Service
			if method.ResourceKind == "" {
				method.ResourceKind = descriptor.ResourceKind
			}
			if method.PayloadKind == "" {
				method.PayloadKind = inferPayloadKind(method.RequestPayload, method.ResponsePayload)
			}
			item.Methods = append(item.Methods, method)
		}
		result = append(result, item)
	}
	return result
}

// Methods returns all governed host-service method descriptors.
func Methods() []MethodDescriptor {
	methods := make([]MethodDescriptor, 0)
	for _, service := range Catalog() {
		methods = append(methods, service.Methods...)
	}
	return methods
}

func hostMethod(
	method string,
	methodConst string,
	capability string,
	requestPayload string,
	responsePayload string,
) MethodDescriptor {
	return MethodDescriptor{
		Method:          method,
		MethodConst:     methodConst,
		Capability:      capability,
		RequestPayload:  requestPayload,
		ResponsePayload: responsePayload,
		PayloadKind:     inferPayloadKind(requestPayload, responsePayload),
		Published:       true,
		GuestClient:     true,
		Dispatcher:      true,
	}
}

func hostMethodWithResource(
	method string,
	methodConst string,
	capability string,
	resourceKind ResourceKind,
	requestPayload string,
	responsePayload string,
) MethodDescriptor {
	descriptor := hostMethod(method, methodConst, capability, requestPayload, responsePayload)
	descriptor.ResourceKind = resourceKind
	return descriptor
}

func reservedHostMethod(method string, capability string) MethodDescriptor {
	return MethodDescriptor{
		Method:       method,
		Capability:   capability,
		ResourceKind: ResourceKindReserved,
		PayloadKind:  PayloadKindReserved,
	}
}

func inferPayloadKind(requestPayload string, responsePayload string) PayloadKind {
	if requestPayload == "" && responsePayload == "" {
		return PayloadKindNone
	}
	if (requestPayload == "" || requestPayload == "HostServiceJSONRequest") &&
		(responsePayload == "" || responsePayload == "HostServiceJSONResponse") {
		return PayloadKindJSON
	}
	return PayloadKindDedicated
}
