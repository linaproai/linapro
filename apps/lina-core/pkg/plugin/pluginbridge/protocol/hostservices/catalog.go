// Package hostservices defines the public host-service catalog used by
// protocol governance, guest SDK coverage checks, and host dispatcher coverage
// checks. Runtime validation can derive private lookup tables from this catalog,
// but service and method metadata must be maintained here first.
package hostservices

import (
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/capregistry"
)

// ResourceKind describes which authorization resource shape a host service
// declaration uses in plugin manifests.
type ResourceKind string

// Host-service resource kinds used by manifest validation and governance tests.
const (
	ResourceKindNone  ResourceKind = "none"
	ResourceKindPath  ResourceKind = "path"
	ResourceKindTable ResourceKind = "table"
	ResourceKindKey   ResourceKind = "key"
	ResourceKindRef   ResourceKind = "resource"
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
)

// RiskLevel classifies one owner method for authorization and upgrade previews.
type RiskLevel string

// Host-service risk levels projected from plugin-owned capability descriptors.
const (
	RiskLevelRead    RiskLevel = "read"
	RiskLevelWrite   RiskLevel = "write"
	RiskLevelExecute RiskLevel = "execute"
)

// ServiceDescriptor describes one logical host service family.
type ServiceDescriptor struct {
	// Owner is empty for core-owned services and contains the owner plugin ID
	// for plugin-owned service projections.
	Owner string
	// Service is the logical host service identifier.
	Service string
	// Version is empty for core-owned services and contains the plugin-owned
	// capability protocol version for owner projections.
	Version string
	// ResourceKind describes the manifest resource declaration shape.
	ResourceKind ResourceKind
	// SourceContract names the public Go source-plugin contract package when
	// the service comes from a plugin-owned capability descriptor.
	SourceContract string
	// DynamicContract names the public dynamic-plugin bridge SDK package when
	// the service comes from a plugin-owned capability descriptor.
	DynamicContract string
	// Methods lists governed methods under this service.
	Methods []MethodDescriptor
}

// MethodDescriptor describes one governed host service method.
type MethodDescriptor struct {
	// Owner is empty for core-owned services and contains the owner plugin ID
	// for plugin-owned method projections.
	Owner string
	// Service is populated when descriptors are flattened.
	Service string
	// Version is empty for core-owned services and contains the plugin-owned
	// capability protocol version for owner projections.
	Version string
	// Method is the wire method string.
	Method string
	// MethodConst is the stable Go constant name for the method.
	MethodConst string
	// Capability is the capability implied by this method.
	Capability string
	// Risk classifies plugin-owned methods for authorization and upgrade
	// previews. Core-owned static methods leave this empty.
	Risk RiskLevel
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
		Service:      "hostconfig",
		ResourceKind: ResourceKindKey,
		Methods: []MethodDescriptor{
			hostMethod("get", "HostServiceMethodHostConfigGet", "host:hostconfig", "HostServiceConfigKeyRequest", "HostServiceConfigValueResponse"),
			hostMethod("sys_config.get", "HostServiceMethodHostConfigSysConfigGet", "host:hostconfig", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("sys_config.value.set", "HostServiceMethodHostConfigSysConfigSetValue", "host:hostconfig", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("sys_config.reset", "HostServiceMethodHostConfigSysConfigReset", "host:hostconfig", "HostServiceJSONRequest", "HostServiceJSONResponse"),
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
			hostMethod("token.tenant.select", "HostServiceMethodAuthSelectTenant", "host:auth:token", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("token.tenant.switch", "HostServiceMethodAuthSwitchTenant", "host:auth:token", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("token.impersonation_token.issue", "HostServiceMethodAuthIssueImpersonationToken", "host:auth:token", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("token.impersonation_token.revoke", "HostServiceMethodAuthRevokeImpersonationToken", "host:auth:token", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("authz.permissions.batch_get", "HostServiceMethodAuthzBatchGetPermissions", "host:auth:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("authz.permissions.batch_has", "HostServiceMethodAuthzBatchHasPermissions", "host:auth:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("authz.permissions.has", "HostServiceMethodAuthzHasPermission", "host:auth:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("authz.users.platform_admin.check", "HostServiceMethodAuthzIsPlatformAdmin", "host:auth:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("authz.role_permissions.replace", "HostServiceMethodAuthzReplaceRolePermissions", "host:auth:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "users",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("users.current.get", "HostServiceMethodUsersCurrent", "host:users", "", "HostServiceJSONResponse"),
			hostMethod("users.batch_get", "HostServiceMethodUsersBatchGet", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.resolve.batch", "HostServiceMethodUsersBatchResolve", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.list", "HostServiceMethodUsersList", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.visible.ensure", "HostServiceMethodUsersEnsureVisible", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.create", "HostServiceMethodUsersCreate", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.update", "HostServiceMethodUsersUpdate", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.delete", "HostServiceMethodUsersDelete", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.status.set", "HostServiceMethodUsersSetStatus", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.password.reset", "HostServiceMethodUsersResetPassword", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("users.assignment.roles.replace", "HostServiceMethodUsersReplaceRoles", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
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
			hostMethod("dict.refresh", "HostServiceMethodDictRefresh", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.type.get", "HostServiceMethodDictTypeGet", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.type.batch_get", "HostServiceMethodDictTypeBatchGet", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.type.list", "HostServiceMethodDictTypeList", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.type.visible.ensure", "HostServiceMethodDictTypeEnsureVisible", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.type.keys.visible.ensure", "HostServiceMethodDictTypeEnsureKeysVisible", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.type.create", "HostServiceMethodDictTypeCreate", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.type.update", "HostServiceMethodDictTypeUpdate", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.type.delete", "HostServiceMethodDictTypeDelete", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.value.get", "HostServiceMethodDictValueGet", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.value.batch_get", "HostServiceMethodDictValueBatchGet", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.value.labels.resolve", "HostServiceMethodDictValueResolveLabels", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.value.list", "HostServiceMethodDictListValues", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.value.visible.ensure", "HostServiceMethodDictValueEnsureVisible", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.value.values.visible.ensure", "HostServiceMethodDictValueEnsureValuesVisible", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.value.create", "HostServiceMethodDictValueCreate", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.value.update", "HostServiceMethodDictValueUpdate", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.value.delete", "HostServiceMethodDictValueDelete", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("dict.value.by_type.delete", "HostServiceMethodDictValueDeleteByType", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "files",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("files.batch_get", "HostServiceMethodFilesBatchGet", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("files.list", "HostServiceMethodFilesList", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("files.visible.ensure", "HostServiceMethodFilesEnsureVisible", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("files.upload", "HostServiceMethodFilesUpload", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("files.create_from_storage", "HostServiceMethodFilesCreateFromStorage", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("files.metadata.update", "HostServiceMethodFilesUpdateMetadata", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("files.delete", "HostServiceMethodFilesDelete", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("files.delete_many", "HostServiceMethodFilesDeleteMany", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "jobs",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("jobs.batch_get", "HostServiceMethodJobsBatchGet", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("jobs.list", "HostServiceMethodJobsList", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("jobs.visible.ensure", "HostServiceMethodJobsEnsureVisible", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("jobs.create", "HostServiceMethodJobsCreate", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("jobs.update", "HostServiceMethodJobsUpdate", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("jobs.delete", "HostServiceMethodJobsDelete", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("jobs.run", "HostServiceMethodJobsRun", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("jobs.status.set", "HostServiceMethodJobsSetStatus", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("jobs.register", "HostServiceMethodJobsRegister", "host:jobs", "HostServiceJobsRegisterRequest", ""),
		},
	},
	{
		Service:      "notifications",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("messages.batch_get", "HostServiceMethodNotificationsBatchGetMessages", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("messages.list", "HostServiceMethodNotificationsList", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("messages.by_source.batch_get", "HostServiceMethodNotificationsBatchGetBySource", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("messages.visible.ensure", "HostServiceMethodNotificationsEnsureVisible", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethodWithResource("messages.send", "HostServiceMethodNotificationsSend", "host:notifications", ResourceKindRef, "HostServiceNotificationsSendRequest", "HostServiceNotificationsSendResponse"),
			hostMethod("messages.delete", "HostServiceMethodNotificationsDelete", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("messages.by_source.delete", "HostServiceMethodNotificationsDeleteBySource", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("messages.mark_read", "HostServiceMethodNotificationsMarkRead", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("messages.mark_unread", "HostServiceMethodNotificationsMarkUnread", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "plugins",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("plugins.current.get", "HostServiceMethodPluginsCurrent", "host:plugins", "", "HostServiceJSONResponse"),
			hostMethod("plugins.batch_get", "HostServiceMethodPluginsBatchGet", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.registry.list", "HostServiceMethodPluginsList", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.tenant.list", "HostServiceMethodPluginsListTenant", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("config.get", "HostServiceMethodPluginsConfigGet", "host:plugins", "HostServiceConfigKeyRequest", "HostServiceConfigValueResponse"),
			hostMethod("plugins.state.enabled.check", "HostServiceMethodPluginsStateIsEnabled", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.state.provider_enabled.check", "HostServiceMethodPluginsStateIsProviderEnabled", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.state.enabled_authoritative.check", "HostServiceMethodPluginsStateIsEnabledAuthoritative", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.lifecycle.tenant_plugin_disable.ensure", "HostServiceMethodPluginsLifecycleEnsureTenantPluginDisableAllowed", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.lifecycle.tenant_plugin_disabled.notify", "HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.lifecycle.tenant_delete.ensure", "HostServiceMethodPluginsLifecycleEnsureTenantDeleteAllowed", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("plugins.lifecycle.tenant_deleted.notify", "HostServiceMethodPluginsLifecycleNotifyTenantDeleted", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
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
			hostMethod("sessions.list", "HostServiceMethodSessionsList", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("sessions.batch_get", "HostServiceMethodSessionsBatchGet", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("sessions.users.online.batch_get", "HostServiceMethodSessionsBatchGetUserOnlineStatus", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("sessions.visible.ensure", "HostServiceMethodSessionsEnsureVisible", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("sessions.revoke", "HostServiceMethodSessionsRevoke", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("sessions.revoke_many", "HostServiceMethodSessionsRevokeMany", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "org",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("capability.available", "HostServiceMethodOrgAvailable", "host:org", "", "HostServiceJSONResponse"),
			hostMethod("capability.status", "HostServiceMethodOrgStatus", "host:org", "", "HostServiceJSONResponse"),
			hostMethod("org.assignment.user_profiles.batch_get", "HostServiceMethodOrgBatchGetUserOrgProfiles", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.department.tree.list", "HostServiceMethodOrgListDeptTree", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.department.batch_get", "HostServiceMethodOrgDepartmentBatchGet", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.department.list", "HostServiceMethodOrgDepartmentList", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.post.batch_get", "HostServiceMethodOrgPostBatchGet", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.post.options.list", "HostServiceMethodOrgListPostOptions", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.department.visible.ensure_many", "HostServiceMethodOrgEnsureDepartmentsVisible", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.post.visible.ensure_many", "HostServiceMethodOrgEnsurePostsVisible", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.department.create", "HostServiceMethodOrgDepartmentCreate", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.department.update", "HostServiceMethodOrgDepartmentUpdate", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.department.delete", "HostServiceMethodOrgDepartmentDelete", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.post.create", "HostServiceMethodOrgPostCreate", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.post.update", "HostServiceMethodOrgPostUpdate", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.post.delete", "HostServiceMethodOrgPostDelete", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.assignment.by_user.replace", "HostServiceMethodOrgAssignmentReplaceByUser", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("org.assignment.by_user.cleanup", "HostServiceMethodOrgAssignmentCleanupByUser", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      "tenant",
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod("capability.available", "HostServiceMethodTenantAvailable", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod("capability.status", "HostServiceMethodTenantStatus", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod("tenant.context.current", "HostServiceMethodTenantCurrent", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod("tenant.context.info", "HostServiceMethodTenantCurrentInfo", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod("tenant.context.platform_bypass", "HostServiceMethodTenantPlatformBypass", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod("tenant.directory.batch_get", "HostServiceMethodTenantBatchGet", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("tenant.directory.list", "HostServiceMethodTenantDirectoryList", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("tenant.membership.validate", "HostServiceMethodTenantValidateUserInTenant", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("tenant.membership.list_by_user", "HostServiceMethodTenantListUserTenants", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("tenant.directory.visible.ensure_many", "HostServiceMethodTenantBatchEnsureVisible", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("tenant.plugins.enabled.set", "HostServiceMethodTenantPluginSetEnabled", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("tenant.plugins.defaults.provision", "HostServiceMethodTenantPluginProvisionDefaults", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod("tenant.filter.context", "HostServiceMethodTenantFilterContext", "host:tenant", "", "HostServiceJSONResponse"),
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

// CatalogWithDescriptors returns the static core-owned catalog merged with
// plugin-owned owner descriptors projected into host-service metadata.
func CatalogWithDescriptors(descriptors []capregistry.Descriptor) ([]ServiceDescriptor, error) {
	catalog := Catalog()
	ownerCatalog, err := CatalogFromDescriptors(descriptors)
	if err != nil {
		return nil, err
	}
	catalog = append(catalog, ownerCatalog...)
	return catalog, nil
}

// CatalogFromDescriptors projects plugin-owned capability descriptors into
// owner-aware host-service catalog entries. Projection is a pure value transform
// with lightweight identity validation; runtime registration still happens in
// capregistry when the host starts.
func CatalogFromDescriptors(descriptors []capregistry.Descriptor) ([]ServiceDescriptor, error) {
	if len(descriptors) == 0 {
		return []ServiceDescriptor{}, nil
	}
	seen := make(map[string]struct{}, len(descriptors))
	result := make([]ServiceDescriptor, 0, len(descriptors))
	for _, descriptor := range descriptors {
		service, err := projectOwnerDescriptor(descriptor)
		if err != nil {
			return nil, err
		}
		key := service.Owner + "\x00" + service.Service + "\x00" + service.Version
		if _, exists := seen[key]; exists {
			return nil, gerror.Newf(
				"capability descriptor already registered: owner=%s service=%s version=%s",
				service.Owner,
				service.Service,
				service.Version,
			)
		}
		seen[key] = struct{}{}
		result = append(result, service)
	}
	sort.Slice(result, func(i, j int) bool {
		left := result[i].Owner + "\x00" + result[i].Service + "\x00" + result[i].Version
		right := result[j].Owner + "\x00" + result[j].Service + "\x00" + result[j].Version
		return left < right
	})
	return result, nil
}

func projectOwnerDescriptor(descriptor capregistry.Descriptor) (ServiceDescriptor, error) {
	owner := strings.TrimSpace(descriptor.OwnerPluginID)
	service := strings.TrimSpace(descriptor.Service)
	version := strings.TrimSpace(descriptor.Version)
	if owner == "" {
		return ServiceDescriptor{}, gerror.New("capability descriptor owner plugin id is required")
	}
	if service == "" {
		return ServiceDescriptor{}, gerror.Newf("capability descriptor service is required: owner=%s", owner)
	}
	if version == "" {
		return ServiceDescriptor{}, gerror.Newf(
			"capability descriptor version is required: owner=%s service=%s",
			owner,
			service,
		)
	}
	if len(descriptor.Methods) == 0 {
		return ServiceDescriptor{}, gerror.Newf(
			"capability descriptor methods are required: owner=%s service=%s version=%s",
			owner,
			service,
			version,
		)
	}

	projected := ServiceDescriptor{
		Owner:           owner,
		Service:         service,
		Version:         version,
		ResourceKind:    ownerDescriptorResourceKind(descriptor.Methods),
		SourceContract:  strings.TrimSpace(descriptor.SourceContract),
		DynamicContract: strings.TrimSpace(descriptor.DynamicContract),
		Methods:         make([]MethodDescriptor, 0, len(descriptor.Methods)),
	}
	seenMethods := make(map[string]struct{}, len(descriptor.Methods))
	for _, method := range descriptor.Methods {
		name := strings.TrimSpace(method.Method)
		if name == "" {
			return ServiceDescriptor{}, gerror.Newf(
				"capability descriptor method is required: owner=%s service=%s version=%s",
				owner,
				service,
				version,
			)
		}
		if _, exists := seenMethods[name]; exists {
			return ServiceDescriptor{}, gerror.Newf(
				"capability descriptor contains duplicate method: owner=%s service=%s version=%s method=%s",
				owner,
				service,
				version,
				name,
			)
		}
		seenMethods[name] = struct{}{}
		projected.Methods = append(projected.Methods, MethodDescriptor{
			Owner:           owner,
			Service:         service,
			Version:         version,
			Method:          name,
			Capability:      strings.TrimSpace(method.Capability),
			Risk:            ownerRiskLevel(method.Risk),
			ResourceKind:    ownerResourceKind(method.ResourceKind),
			RequestPayload:  strings.TrimSpace(method.RequestPayload),
			ResponsePayload: strings.TrimSpace(method.ResponsePayload),
			PayloadKind:     PayloadKindJSON,
			Published:       true,
			GuestClient:     true,
			Dispatcher:      true,
		})
	}
	sort.Slice(projected.Methods, func(i, j int) bool {
		return projected.Methods[i].Method < projected.Methods[j].Method
	})
	return projected, nil
}

// Methods returns all governed host-service method descriptors.
func Methods() []MethodDescriptor {
	methods := make([]MethodDescriptor, 0)
	for _, service := range Catalog() {
		methods = append(methods, service.Methods...)
	}
	return methods
}

// MethodsWithDescriptors returns all core-owned and plugin-owned method
// descriptors using the merged catalog projection.
func MethodsWithDescriptors(descriptors []capregistry.Descriptor) ([]MethodDescriptor, error) {
	catalog, err := CatalogWithDescriptors(descriptors)
	if err != nil {
		return nil, err
	}
	return methodsFromCatalog(catalog), nil
}

// MethodsFromDescriptors returns plugin-owned method descriptors projected from
// owner capability descriptors.
func MethodsFromDescriptors(descriptors []capregistry.Descriptor) ([]MethodDescriptor, error) {
	catalog, err := CatalogFromDescriptors(descriptors)
	if err != nil {
		return nil, err
	}
	return methodsFromCatalog(catalog), nil
}

func methodsFromCatalog(catalog []ServiceDescriptor) []MethodDescriptor {
	methods := make([]MethodDescriptor, 0)
	for _, service := range catalog {
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

func ownerDescriptorResourceKind(methods []capregistry.MethodDescriptor) ResourceKind {
	if len(methods) == 0 {
		return ResourceKindNone
	}
	first := ownerResourceKind(methods[0].ResourceKind)
	for _, method := range methods[1:] {
		if ownerResourceKind(method.ResourceKind) != first {
			return ResourceKindNone
		}
	}
	return first
}

func ownerResourceKind(kind capregistry.ResourceKind) ResourceKind {
	switch kind {
	case capregistry.ResourceKindPath:
		return ResourceKindPath
	case capregistry.ResourceKindTable:
		return ResourceKindTable
	case capregistry.ResourceKindKey:
		return ResourceKindKey
	case capregistry.ResourceKindRef:
		return ResourceKindRef
	default:
		return ResourceKindNone
	}
}

func ownerRiskLevel(risk capregistry.RiskLevel) RiskLevel {
	switch risk {
	case capregistry.RiskLevelRead:
		return RiskLevelRead
	case capregistry.RiskLevelWrite:
		return RiskLevelWrite
	case capregistry.RiskLevelExecute:
		return RiskLevelExecute
	default:
		return RiskLevel("")
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
