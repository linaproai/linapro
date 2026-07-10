// Package hostservices defines the public host-service catalog used by
// protocol governance, guest SDK coverage checks, and host dispatcher coverage
// checks. Runtime validation can derive private lookup tables from this catalog,
// but service and method metadata must be maintained here first.
//
// Host service/method wire Go constants live in this package (wire_constants.go).
// Catalog entries must reference those constants for service and method wire values.
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
		Service:      HostServiceRuntime,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodRuntimeLogWrite, "HostServiceMethodRuntimeLogWrite", "host:runtime", "HostCallLogRequest", ""),
			hostMethod(HostServiceMethodRuntimeStateGet, "HostServiceMethodRuntimeStateGet", "host:runtime", "HostCallStateGetRequest", "HostCallStateGetResponse"),
			hostMethod(HostServiceMethodRuntimeStateGetMany, "HostServiceMethodRuntimeStateGetMany", "host:runtime", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodRuntimeStateSet, "HostServiceMethodRuntimeStateSet", "host:runtime", "HostCallStateSetRequest", ""),
			hostMethod(HostServiceMethodRuntimeStateSetMany, "HostServiceMethodRuntimeStateSetMany", "host:runtime", "HostServiceJSONRequest", ""),
			hostMethod(HostServiceMethodRuntimeStateDelete, "HostServiceMethodRuntimeStateDelete", "host:runtime", "HostCallStateDeleteRequest", ""),
			hostMethod(HostServiceMethodRuntimeStateDeleteMany, "HostServiceMethodRuntimeStateDeleteMany", "host:runtime", "HostServiceJSONRequest", ""),
			hostMethod(HostServiceMethodRuntimeInfoNow, "HostServiceMethodRuntimeInfoNow", "host:runtime", "", "HostServiceValueResponse"),
			hostMethod(HostServiceMethodRuntimeInfoUUID, "HostServiceMethodRuntimeInfoUUID", "host:runtime", "", "HostServiceValueResponse"),
			hostMethod(HostServiceMethodRuntimeInfoNode, "HostServiceMethodRuntimeInfoNode", "host:runtime", "", "HostServiceValueResponse"),
		},
	},
	{
		Service:      HostServiceStorage,
		ResourceKind: ResourceKindPath,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodStoragePut, "HostServiceMethodStoragePut", "host:storage", "HostServiceStoragePutRequest", "HostServiceStoragePutResponse"),
			hostMethod(HostServiceMethodStoragePutInit, "HostServiceMethodStoragePutInit", "host:storage", "HostServiceStoragePutInitRequest", "HostServiceStoragePutInitResponse"),
			hostMethod(HostServiceMethodStoragePutChunk, "HostServiceMethodStoragePutChunk", "host:storage", "HostServiceStoragePutChunkRequest", "HostServiceStoragePutChunkResponse"),
			hostMethod(HostServiceMethodStoragePutCommit, "HostServiceMethodStoragePutCommit", "host:storage", "HostServiceStoragePutCommitRequest", "HostServiceStoragePutCommitResponse"),
			hostMethod(HostServiceMethodStoragePutAbort, "HostServiceMethodStoragePutAbort", "host:storage", "HostServiceStoragePutAbortRequest", ""),
			hostMethod(HostServiceMethodStorageGet, "HostServiceMethodStorageGet", "host:storage", "HostServiceStorageGetRequest", "HostServiceStorageGetResponse"),
			hostMethod(HostServiceMethodStorageDelete, "HostServiceMethodStorageDelete", "host:storage", "HostServiceStorageDeleteRequest", ""),
			hostMethod(HostServiceMethodStorageDeleteBatch, "HostServiceMethodStorageDeleteBatch", "host:storage", "HostServiceJSONRequest", ""),
			hostMethod(HostServiceMethodStorageList, "HostServiceMethodStorageList", "host:storage", "HostServiceStorageListRequest", "HostServiceStorageListResponse"),
			hostMethod(HostServiceMethodStorageListCursor, "HostServiceMethodStorageListCursor", "host:storage", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodStorageStat, "HostServiceMethodStorageStat", "host:storage", "HostServiceStorageStatRequest", "HostServiceStorageStatResponse"),
			hostMethod(HostServiceMethodStorageStatBatch, "HostServiceMethodStorageStatBatch", "host:storage", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServiceNetwork,
		ResourceKind: ResourceKindRef,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodNetworkRequest, "HostServiceMethodNetworkRequest", "host:http:request", "HostServiceNetworkRequest", "HostServiceNetworkResponse"),
		},
	},
	{
		Service:      HostServiceData,
		ResourceKind: ResourceKindTable,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodDataList, "HostServiceMethodDataList", "host:data:read", "HostServiceDataListRequest", "HostServiceDataListResponse"),
			hostMethod(HostServiceMethodDataGet, "HostServiceMethodDataGet", "host:data:read", "HostServiceDataGetRequest", "HostServiceDataGetResponse"),
			hostMethod(HostServiceMethodDataBatchGet, "HostServiceMethodDataBatchGet", "host:data:read", "HostServiceDataBatchGetRequest", "HostServiceDataBatchGetResponse"),
			hostMethod(HostServiceMethodDataCreate, "HostServiceMethodDataCreate", "host:data:mutate", "HostServiceDataMutationRequest", "HostServiceDataMutationResponse"),
			hostMethod(HostServiceMethodDataUpdate, "HostServiceMethodDataUpdate", "host:data:mutate", "HostServiceDataMutationRequest", "HostServiceDataMutationResponse"),
			hostMethod(HostServiceMethodDataDelete, "HostServiceMethodDataDelete", "host:data:mutate", "HostServiceDataMutationRequest", "HostServiceDataMutationResponse"),
			hostMethod(HostServiceMethodDataTransaction, "HostServiceMethodDataTransaction", "host:data:mutate", "HostServiceDataTransactionRequest", "HostServiceDataTransactionResponse"),
		},
	},
	{
		Service:      HostServiceCache,
		ResourceKind: ResourceKindRef,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodCacheGet, "HostServiceMethodCacheGet", "host:cache", "HostServiceCacheGetRequest", "HostServiceCacheGetResponse"),
			hostMethod(HostServiceMethodCacheGetMany, "HostServiceMethodCacheGetMany", "host:cache", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodCacheSet, "HostServiceMethodCacheSet", "host:cache", "HostServiceCacheSetRequest", "HostServiceCacheSetResponse"),
			hostMethod(HostServiceMethodCacheSetMany, "HostServiceMethodCacheSetMany", "host:cache", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodCacheDelete, "HostServiceMethodCacheDelete", "host:cache", "HostServiceCacheDeleteRequest", ""),
			hostMethod(HostServiceMethodCacheDeleteMany, "HostServiceMethodCacheDeleteMany", "host:cache", "HostServiceJSONRequest", ""),
			hostMethod(HostServiceMethodCacheIncr, "HostServiceMethodCacheIncr", "host:cache", "HostServiceCacheIncrRequest", "HostServiceCacheIncrResponse"),
			hostMethod(HostServiceMethodCacheExpire, "HostServiceMethodCacheExpire", "host:cache", "HostServiceCacheExpireRequest", "HostServiceCacheExpireResponse"),
		},
	},
	{
		Service:      HostServiceLock,
		ResourceKind: ResourceKindRef,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodLockAcquire, "HostServiceMethodLockAcquire", "host:lock", "HostServiceLockAcquireRequest", "HostServiceLockAcquireResponse"),
			hostMethod(HostServiceMethodLockRenew, "HostServiceMethodLockRenew", "host:lock", "HostServiceLockRenewRequest", "HostServiceLockRenewResponse"),
			hostMethod(HostServiceMethodLockRelease, "HostServiceMethodLockRelease", "host:lock", "HostServiceLockReleaseRequest", ""),
		},
	},
	{
		Service:      HostServiceHostConfig,
		ResourceKind: ResourceKindKey,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodHostConfigGet, "HostServiceMethodHostConfigGet", "host:hostconfig", "HostServiceConfigKeyRequest", "HostServiceConfigValueResponse"),
			hostMethod(HostServiceMethodHostConfigSysConfigGet, "HostServiceMethodHostConfigSysConfigGet", "host:hostconfig", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodHostConfigSysConfigSetValue, "HostServiceMethodHostConfigSysConfigSetValue", "host:hostconfig", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodHostConfigSysConfigReset, "HostServiceMethodHostConfigSysConfigReset", "host:hostconfig", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServiceManifest,
		ResourceKind: ResourceKindPath,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodManifestGet, "HostServiceMethodManifestGet", "host:manifest", "HostServiceManifestGetRequest", "HostServiceManifestGetResponse"),
			hostMethod(HostServiceMethodManifestGetMany, "HostServiceMethodManifestGetMany", "host:manifest", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodManifestList, "HostServiceMethodManifestList", "host:manifest", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServiceAPIDoc,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodAPIDocResolveRouteText, "HostServiceMethodAPIDocResolveRouteText", "host:apidoc", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodAPIDocResolveRouteTexts, "HostServiceMethodAPIDocResolveRouteTexts", "host:apidoc", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodAPIDocFindRouteTitleOperationKeys, "HostServiceMethodAPIDocFindRouteTitleOperationKeys", "host:apidoc", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServiceAuth,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodAuthSelectTenant, "HostServiceMethodAuthSelectTenant", "host:auth:token", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodAuthSwitchTenant, "HostServiceMethodAuthSwitchTenant", "host:auth:token", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodAuthIssueImpersonationToken, "HostServiceMethodAuthIssueImpersonationToken", "host:auth:token", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodAuthRevokeImpersonationToken, "HostServiceMethodAuthRevokeImpersonationToken", "host:auth:token", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodAuthzBatchGetPermissions, "HostServiceMethodAuthzBatchGetPermissions", "host:auth:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodAuthzBatchHasPermissions, "HostServiceMethodAuthzBatchHasPermissions", "host:auth:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodAuthzHasPermission, "HostServiceMethodAuthzHasPermission", "host:auth:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodAuthzIsPlatformAdmin, "HostServiceMethodAuthzIsPlatformAdmin", "host:auth:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodAuthzReplaceRolePermissions, "HostServiceMethodAuthzReplaceRolePermissions", "host:auth:authz", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServiceUsers,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodUsersCurrent, "HostServiceMethodUsersCurrent", "host:users", "", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodUsersBatchGet, "HostServiceMethodUsersBatchGet", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodUsersBatchResolve, "HostServiceMethodUsersBatchResolve", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodUsersList, "HostServiceMethodUsersList", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodUsersEnsureVisible, "HostServiceMethodUsersEnsureVisible", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodUsersCreate, "HostServiceMethodUsersCreate", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodUsersUpdate, "HostServiceMethodUsersUpdate", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodUsersDelete, "HostServiceMethodUsersDelete", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodUsersSetStatus, "HostServiceMethodUsersSetStatus", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodUsersResetPassword, "HostServiceMethodUsersResetPassword", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodUsersReplaceRoles, "HostServiceMethodUsersReplaceRoles", "host:users", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServiceBizCtx,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodBizCtxCurrent, "HostServiceMethodBizCtxCurrent", "host:bizctx", "", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServiceDict,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodDictRefresh, "HostServiceMethodDictRefresh", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictTypeGet, "HostServiceMethodDictTypeGet", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictTypeBatchGet, "HostServiceMethodDictTypeBatchGet", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictTypeList, "HostServiceMethodDictTypeList", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictTypeEnsureVisible, "HostServiceMethodDictTypeEnsureVisible", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictTypeEnsureKeysVisible, "HostServiceMethodDictTypeEnsureKeysVisible", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictTypeCreate, "HostServiceMethodDictTypeCreate", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictTypeUpdate, "HostServiceMethodDictTypeUpdate", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictTypeDelete, "HostServiceMethodDictTypeDelete", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictValueGet, "HostServiceMethodDictValueGet", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictValueBatchGet, "HostServiceMethodDictValueBatchGet", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictValueResolveLabels, "HostServiceMethodDictValueResolveLabels", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictListValues, "HostServiceMethodDictListValues", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictValueEnsureVisible, "HostServiceMethodDictValueEnsureVisible", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictValueEnsureValuesVisible, "HostServiceMethodDictValueEnsureValuesVisible", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictValueCreate, "HostServiceMethodDictValueCreate", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictValueUpdate, "HostServiceMethodDictValueUpdate", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictValueDelete, "HostServiceMethodDictValueDelete", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodDictValueDeleteByType, "HostServiceMethodDictValueDeleteByType", "host:dict", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServiceFiles,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodFilesBatchGet, "HostServiceMethodFilesBatchGet", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodFilesList, "HostServiceMethodFilesList", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodFilesEnsureVisible, "HostServiceMethodFilesEnsureVisible", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodFilesUpload, "HostServiceMethodFilesUpload", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodFilesCreateFromStorage, "HostServiceMethodFilesCreateFromStorage", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodFilesUpdateMetadata, "HostServiceMethodFilesUpdateMetadata", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodFilesDelete, "HostServiceMethodFilesDelete", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodFilesDeleteMany, "HostServiceMethodFilesDeleteMany", "host:files", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServiceJobs,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodJobsBatchGet, "HostServiceMethodJobsBatchGet", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodJobsList, "HostServiceMethodJobsList", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodJobsEnsureVisible, "HostServiceMethodJobsEnsureVisible", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodJobsCreate, "HostServiceMethodJobsCreate", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodJobsUpdate, "HostServiceMethodJobsUpdate", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodJobsDelete, "HostServiceMethodJobsDelete", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodJobsRun, "HostServiceMethodJobsRun", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodJobsSetStatus, "HostServiceMethodJobsSetStatus", "host:jobs", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodJobsRegister, "HostServiceMethodJobsRegister", "host:jobs", "HostServiceJobsRegisterRequest", ""),
		},
	},
	{
		Service:      HostServiceNotifications,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodNotificationsBatchGetMessages, "HostServiceMethodNotificationsBatchGetMessages", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodNotificationsList, "HostServiceMethodNotificationsList", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodNotificationsBatchGetBySource, "HostServiceMethodNotificationsBatchGetBySource", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodNotificationsEnsureVisible, "HostServiceMethodNotificationsEnsureVisible", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethodWithResource(HostServiceMethodNotificationsSend, "HostServiceMethodNotificationsSend", "host:notifications", ResourceKindRef, "HostServiceNotificationsSendRequest", "HostServiceNotificationsSendResponse"),
			hostMethod(HostServiceMethodNotificationsDelete, "HostServiceMethodNotificationsDelete", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodNotificationsDeleteBySource, "HostServiceMethodNotificationsDeleteBySource", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodNotificationsMarkRead, "HostServiceMethodNotificationsMarkRead", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodNotificationsMarkUnread, "HostServiceMethodNotificationsMarkUnread", "host:notifications", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServicePlugins,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodPluginsCurrent, "HostServiceMethodPluginsCurrent", "host:plugins", "", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodPluginsBatchGet, "HostServiceMethodPluginsBatchGet", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodPluginsList, "HostServiceMethodPluginsList", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodPluginsListTenant, "HostServiceMethodPluginsListTenant", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodPluginsConfigGet, "HostServiceMethodPluginsConfigGet", "host:plugins", "HostServiceConfigKeyRequest", "HostServiceConfigValueResponse"),
			hostMethod(HostServiceMethodPluginsStateIsEnabled, "HostServiceMethodPluginsStateIsEnabled", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodPluginsStateIsProviderEnabled, "HostServiceMethodPluginsStateIsProviderEnabled", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodPluginsStateIsEnabledAuthoritative, "HostServiceMethodPluginsStateIsEnabledAuthoritative", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodPluginsLifecycleEnsureTenantPluginDisableAllowed, "HostServiceMethodPluginsLifecycleEnsureTenantPluginDisableAllowed", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled, "HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodPluginsLifecycleEnsureTenantDeleteAllowed, "HostServiceMethodPluginsLifecycleEnsureTenantDeleteAllowed", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodPluginsLifecycleNotifyTenantDeleted, "HostServiceMethodPluginsLifecycleNotifyTenantDeleted", "host:plugins", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServiceRoute,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodRouteMetadataGet, "HostServiceMethodRouteMetadataGet", "host:route", "", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServiceSessions,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodSessionsCurrent, "HostServiceMethodSessionsCurrent", "host:sessions", "", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodSessionsList, "HostServiceMethodSessionsList", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodSessionsBatchGet, "HostServiceMethodSessionsBatchGet", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodSessionsBatchGetUserOnlineStatus, "HostServiceMethodSessionsBatchGetUserOnlineStatus", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodSessionsEnsureVisible, "HostServiceMethodSessionsEnsureVisible", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodSessionsRevoke, "HostServiceMethodSessionsRevoke", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodSessionsRevokeMany, "HostServiceMethodSessionsRevokeMany", "host:sessions", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServiceOrg,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodOrgAvailable, "HostServiceMethodOrgAvailable", "host:org", "", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgStatus, "HostServiceMethodOrgStatus", "host:org", "", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgBatchGetUserOrgProfiles, "HostServiceMethodOrgBatchGetUserOrgProfiles", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgListDeptTree, "HostServiceMethodOrgListDeptTree", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgDepartmentBatchGet, "HostServiceMethodOrgDepartmentBatchGet", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgDepartmentList, "HostServiceMethodOrgDepartmentList", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgPostBatchGet, "HostServiceMethodOrgPostBatchGet", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgListPostOptions, "HostServiceMethodOrgListPostOptions", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgEnsureDepartmentsVisible, "HostServiceMethodOrgEnsureDepartmentsVisible", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgEnsurePostsVisible, "HostServiceMethodOrgEnsurePostsVisible", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgDepartmentCreate, "HostServiceMethodOrgDepartmentCreate", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgDepartmentUpdate, "HostServiceMethodOrgDepartmentUpdate", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgDepartmentDelete, "HostServiceMethodOrgDepartmentDelete", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgPostCreate, "HostServiceMethodOrgPostCreate", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgPostUpdate, "HostServiceMethodOrgPostUpdate", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgPostDelete, "HostServiceMethodOrgPostDelete", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgAssignmentReplaceByUser, "HostServiceMethodOrgAssignmentReplaceByUser", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodOrgAssignmentCleanupByUser, "HostServiceMethodOrgAssignmentCleanupByUser", "host:org", "HostServiceJSONRequest", "HostServiceJSONResponse"),
		},
	},
	{
		Service:      HostServiceTenant,
		ResourceKind: ResourceKindNone,
		Methods: []MethodDescriptor{
			hostMethod(HostServiceMethodTenantAvailable, "HostServiceMethodTenantAvailable", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodTenantStatus, "HostServiceMethodTenantStatus", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodTenantCurrent, "HostServiceMethodTenantCurrent", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodTenantCurrentInfo, "HostServiceMethodTenantCurrentInfo", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodTenantPlatformBypass, "HostServiceMethodTenantPlatformBypass", "host:tenant", "", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodTenantBatchGet, "HostServiceMethodTenantBatchGet", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodTenantDirectoryList, "HostServiceMethodTenantDirectoryList", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodTenantValidateUserInTenant, "HostServiceMethodTenantValidateUserInTenant", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodTenantListUserTenants, "HostServiceMethodTenantListUserTenants", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodTenantBatchEnsureVisible, "HostServiceMethodTenantBatchEnsureVisible", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodTenantPluginSetEnabled, "HostServiceMethodTenantPluginSetEnabled", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodTenantPluginProvisionDefaults, "HostServiceMethodTenantPluginProvisionDefaults", "host:tenant", "HostServiceJSONRequest", "HostServiceJSONResponse"),
			hostMethod(HostServiceMethodTenantFilterContext, "HostServiceMethodTenantFilterContext", "host:tenant", "", "HostServiceJSONResponse"),
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
