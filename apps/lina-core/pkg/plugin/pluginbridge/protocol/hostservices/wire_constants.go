// This file is the single source of host-service and method wire string
// constants. The public catalog in catalog.go must reference these constants
// instead of repeating wire literals. protocol re-exports the same names for
// guest/host callers.

package hostservices

// Host service identifiers.
const (
	// HostServiceAPIDoc is the catalog wire value for service "apidoc".
	HostServiceAPIDoc = "apidoc"
	// HostServiceAuth is the catalog wire value for service "auth".
	HostServiceAuth = "auth"
	// HostServiceBizCtx is the catalog wire value for service "bizctx".
	HostServiceBizCtx = "bizctx"
	// HostServiceCache is the catalog wire value for service "cache".
	HostServiceCache = "cache"
	// HostServiceData is the catalog wire value for service "data".
	HostServiceData = "data"
	// HostServiceDict is the catalog wire value for service "dict".
	HostServiceDict = "dict"
	// HostServiceFiles is the catalog wire value for service "files".
	HostServiceFiles = "files"
	// HostServiceHostConfig is the catalog wire value for service "hostconfig".
	HostServiceHostConfig = "hostconfig"
	// HostServiceJobs is the catalog wire value for service "jobs".
	HostServiceJobs = "jobs"
	// HostServiceLock is the catalog wire value for service "lock".
	HostServiceLock = "lock"
	// HostServiceManifest is the catalog wire value for service "manifest".
	HostServiceManifest = "manifest"
	// HostServiceNetwork is the catalog wire value for service "network".
	HostServiceNetwork = "network"
	// HostServiceNotifications is the catalog wire value for service "notifications".
	HostServiceNotifications = "notifications"
	// HostServiceOrg is the catalog wire value for service "org".
	HostServiceOrg = "org"
	// HostServicePlugins is the catalog wire value for service "plugins".
	HostServicePlugins = "plugins"
	// HostServiceRoute is the catalog wire value for service "route".
	HostServiceRoute = "route"
	// HostServiceRuntime is the catalog wire value for service "runtime".
	HostServiceRuntime = "runtime"
	// HostServiceSessions is the catalog wire value for service "sessions".
	HostServiceSessions = "sessions"
	// HostServiceStorage is the catalog wire value for service "storage".
	HostServiceStorage = "storage"
	// HostServiceTenant is the catalog wire value for service "tenant".
	HostServiceTenant = "tenant"
	// HostServiceUsers is the catalog wire value for service "users".
	HostServiceUsers = "users"
)

// Host service method identifiers.
const (
	// HostServiceMethodAPIDocFindRouteTitleOperationKeys is the catalog wire value for method "route_title_operation_keys.find".
	HostServiceMethodAPIDocFindRouteTitleOperationKeys = "route_title_operation_keys.find"
	// HostServiceMethodAPIDocResolveRouteText is the catalog wire value for method "route_text.resolve".
	HostServiceMethodAPIDocResolveRouteText = "route_text.resolve"
	// HostServiceMethodAPIDocResolveRouteTexts is the catalog wire value for method "route_texts.resolve".
	HostServiceMethodAPIDocResolveRouteTexts = "route_texts.resolve"
	// HostServiceMethodAuthIssueImpersonationToken is the catalog wire value for method "token.impersonation_token.issue".
	HostServiceMethodAuthIssueImpersonationToken = "token.impersonation_token.issue"
	// HostServiceMethodAuthRevokeImpersonationToken is the catalog wire value for method "token.impersonation_token.revoke".
	HostServiceMethodAuthRevokeImpersonationToken = "token.impersonation_token.revoke"
	// HostServiceMethodAuthSelectTenant is the catalog wire value for method "token.tenant.select".
	HostServiceMethodAuthSelectTenant = "token.tenant.select"
	// HostServiceMethodAuthSwitchTenant is the catalog wire value for method "token.tenant.switch".
	HostServiceMethodAuthSwitchTenant = "token.tenant.switch"
	// HostServiceMethodAuthzBatchGetPermissions is the catalog wire value for method "authz.permissions.batch_get".
	HostServiceMethodAuthzBatchGetPermissions = "authz.permissions.batch_get"
	// HostServiceMethodAuthzBatchHasPermissions is the catalog wire value for method "authz.permissions.batch_has".
	HostServiceMethodAuthzBatchHasPermissions = "authz.permissions.batch_has"
	// HostServiceMethodAuthzHasPermission is the catalog wire value for method "authz.permissions.has".
	HostServiceMethodAuthzHasPermission = "authz.permissions.has"
	// HostServiceMethodAuthzIsPlatformAdmin is the catalog wire value for method "authz.users.platform_admin.check".
	HostServiceMethodAuthzIsPlatformAdmin = "authz.users.platform_admin.check"
	// HostServiceMethodAuthzReplaceRolePermissions is the catalog wire value for method "authz.role_permissions.replace".
	HostServiceMethodAuthzReplaceRolePermissions = "authz.role_permissions.replace"
	// HostServiceMethodBizCtxCurrent is the catalog wire value for method "current.get".
	HostServiceMethodBizCtxCurrent = "current.get"
	// HostServiceMethodCacheDelete is the catalog wire value for method "delete".
	HostServiceMethodCacheDelete = "delete"
	// HostServiceMethodCacheDeleteMany is the catalog wire value for method "delete_many".
	HostServiceMethodCacheDeleteMany = "delete_many"
	// HostServiceMethodCacheExpire is the catalog wire value for method "expire".
	HostServiceMethodCacheExpire = "expire"
	// HostServiceMethodCacheGet is the catalog wire value for method "get".
	HostServiceMethodCacheGet = "get"
	// HostServiceMethodCacheGetMany is the catalog wire value for method "get_many".
	HostServiceMethodCacheGetMany = "get_many"
	// HostServiceMethodCacheIncr is the catalog wire value for method "incr".
	HostServiceMethodCacheIncr = "incr"
	// HostServiceMethodCacheSet is the catalog wire value for method "set".
	HostServiceMethodCacheSet = "set"
	// HostServiceMethodCacheSetMany is the catalog wire value for method "set_many".
	HostServiceMethodCacheSetMany = "set_many"
	// HostServiceMethodDataBatchGet is the catalog wire value for method "batch_get".
	HostServiceMethodDataBatchGet = "batch_get"
	// HostServiceMethodDataCreate is the catalog wire value for method "create".
	HostServiceMethodDataCreate = "create"
	// HostServiceMethodDataDelete is the catalog wire value for method "delete".
	HostServiceMethodDataDelete = "delete"
	// HostServiceMethodDataGet is the catalog wire value for method "get".
	HostServiceMethodDataGet = "get"
	// HostServiceMethodDataList is the catalog wire value for method "list".
	HostServiceMethodDataList = "list"
	// HostServiceMethodDataTransaction is the catalog wire value for method "transaction".
	HostServiceMethodDataTransaction = "transaction"
	// HostServiceMethodDataUpdate is the catalog wire value for method "update".
	HostServiceMethodDataUpdate = "update"
	// HostServiceMethodDictListValues is the catalog wire value for method "dict.value.list".
	HostServiceMethodDictListValues = "dict.value.list"
	// HostServiceMethodDictRefresh is the catalog wire value for method "dict.refresh".
	HostServiceMethodDictRefresh = "dict.refresh"
	// HostServiceMethodDictTypeBatchGet is the catalog wire value for method "dict.type.batch_get".
	HostServiceMethodDictTypeBatchGet = "dict.type.batch_get"
	// HostServiceMethodDictTypeCreate is the catalog wire value for method "dict.type.create".
	HostServiceMethodDictTypeCreate = "dict.type.create"
	// HostServiceMethodDictTypeDelete is the catalog wire value for method "dict.type.delete".
	HostServiceMethodDictTypeDelete = "dict.type.delete"
	// HostServiceMethodDictTypeEnsureKeysVisible is the catalog wire value for method "dict.type.keys.visible.ensure".
	HostServiceMethodDictTypeEnsureKeysVisible = "dict.type.keys.visible.ensure"
	// HostServiceMethodDictTypeEnsureVisible is the catalog wire value for method "dict.type.visible.ensure".
	HostServiceMethodDictTypeEnsureVisible = "dict.type.visible.ensure"
	// HostServiceMethodDictTypeGet is the catalog wire value for method "dict.type.get".
	HostServiceMethodDictTypeGet = "dict.type.get"
	// HostServiceMethodDictTypeList is the catalog wire value for method "dict.type.list".
	HostServiceMethodDictTypeList = "dict.type.list"
	// HostServiceMethodDictTypeUpdate is the catalog wire value for method "dict.type.update".
	HostServiceMethodDictTypeUpdate = "dict.type.update"
	// HostServiceMethodDictValueBatchGet is the catalog wire value for method "dict.value.batch_get".
	HostServiceMethodDictValueBatchGet = "dict.value.batch_get"
	// HostServiceMethodDictValueCreate is the catalog wire value for method "dict.value.create".
	HostServiceMethodDictValueCreate = "dict.value.create"
	// HostServiceMethodDictValueDelete is the catalog wire value for method "dict.value.delete".
	HostServiceMethodDictValueDelete = "dict.value.delete"
	// HostServiceMethodDictValueDeleteByType is the catalog wire value for method "dict.value.by_type.delete".
	HostServiceMethodDictValueDeleteByType = "dict.value.by_type.delete"
	// HostServiceMethodDictValueEnsureValuesVisible is the catalog wire value for method "dict.value.values.visible.ensure".
	HostServiceMethodDictValueEnsureValuesVisible = "dict.value.values.visible.ensure"
	// HostServiceMethodDictValueEnsureVisible is the catalog wire value for method "dict.value.visible.ensure".
	HostServiceMethodDictValueEnsureVisible = "dict.value.visible.ensure"
	// HostServiceMethodDictValueGet is the catalog wire value for method "dict.value.get".
	HostServiceMethodDictValueGet = "dict.value.get"
	// HostServiceMethodDictValueResolveLabels is the catalog wire value for method "dict.value.labels.resolve".
	HostServiceMethodDictValueResolveLabels = "dict.value.labels.resolve"
	// HostServiceMethodDictValueUpdate is the catalog wire value for method "dict.value.update".
	HostServiceMethodDictValueUpdate = "dict.value.update"
	// HostServiceMethodFilesBatchGet is the catalog wire value for method "files.batch_get".
	HostServiceMethodFilesBatchGet = "files.batch_get"
	// HostServiceMethodFilesCreateFromStorage is the catalog wire value for method "files.create_from_storage".
	HostServiceMethodFilesCreateFromStorage = "files.create_from_storage"
	// HostServiceMethodFilesDelete is the catalog wire value for method "files.delete".
	HostServiceMethodFilesDelete = "files.delete"
	// HostServiceMethodFilesDeleteMany is the catalog wire value for method "files.delete_many".
	HostServiceMethodFilesDeleteMany = "files.delete_many"
	// HostServiceMethodFilesEnsureVisible is the catalog wire value for method "files.visible.ensure".
	HostServiceMethodFilesEnsureVisible = "files.visible.ensure"
	// HostServiceMethodFilesList is the catalog wire value for method "files.list".
	HostServiceMethodFilesList = "files.list"
	// HostServiceMethodFilesUpdateMetadata is the catalog wire value for method "files.metadata.update".
	HostServiceMethodFilesUpdateMetadata = "files.metadata.update"
	// HostServiceMethodFilesUpload is the catalog wire value for method "files.upload".
	HostServiceMethodFilesUpload = "files.upload"
	// HostServiceMethodHostConfigGet is the catalog wire value for method "get".
	HostServiceMethodHostConfigGet = "get"
	// HostServiceMethodHostConfigSysConfigGet is the catalog wire value for method "sys_config.get".
	HostServiceMethodHostConfigSysConfigGet = "sys_config.get"
	// HostServiceMethodHostConfigSysConfigReset is the catalog wire value for method "sys_config.reset".
	HostServiceMethodHostConfigSysConfigReset = "sys_config.reset"
	// HostServiceMethodHostConfigSysConfigSetValue is the catalog wire value for method "sys_config.value.set".
	HostServiceMethodHostConfigSysConfigSetValue = "sys_config.value.set"
	// HostServiceMethodJobsBatchGet is the catalog wire value for method "jobs.batch_get".
	HostServiceMethodJobsBatchGet = "jobs.batch_get"
	// HostServiceMethodJobsCreate is the catalog wire value for method "jobs.create".
	HostServiceMethodJobsCreate = "jobs.create"
	// HostServiceMethodJobsDelete is the catalog wire value for method "jobs.delete".
	HostServiceMethodJobsDelete = "jobs.delete"
	// HostServiceMethodJobsEnsureVisible is the catalog wire value for method "jobs.visible.ensure".
	HostServiceMethodJobsEnsureVisible = "jobs.visible.ensure"
	// HostServiceMethodJobsList is the catalog wire value for method "jobs.list".
	HostServiceMethodJobsList = "jobs.list"
	// HostServiceMethodJobsRegister is the catalog wire value for method "jobs.register".
	HostServiceMethodJobsRegister = "jobs.register"
	// HostServiceMethodJobsRun is the catalog wire value for method "jobs.run".
	HostServiceMethodJobsRun = "jobs.run"
	// HostServiceMethodJobsSetStatus is the catalog wire value for method "jobs.status.set".
	HostServiceMethodJobsSetStatus = "jobs.status.set"
	// HostServiceMethodJobsUpdate is the catalog wire value for method "jobs.update".
	HostServiceMethodJobsUpdate = "jobs.update"
	// HostServiceMethodLockAcquire is the catalog wire value for method "acquire".
	HostServiceMethodLockAcquire = "acquire"
	// HostServiceMethodLockRelease is the catalog wire value for method "release".
	HostServiceMethodLockRelease = "release"
	// HostServiceMethodLockRenew is the catalog wire value for method "renew".
	HostServiceMethodLockRenew = "renew"
	// HostServiceMethodManifestGet is the catalog wire value for method "get".
	HostServiceMethodManifestGet = "get"
	// HostServiceMethodManifestGetMany is the catalog wire value for method "get_many".
	HostServiceMethodManifestGetMany = "get_many"
	// HostServiceMethodManifestList is the catalog wire value for method "list".
	HostServiceMethodManifestList = "list"
	// HostServiceMethodNetworkRequest is the catalog wire value for method "request".
	HostServiceMethodNetworkRequest = "request"
	// HostServiceMethodNotificationsBatchGetBySource is the catalog wire value for method "messages.by_source.batch_get".
	HostServiceMethodNotificationsBatchGetBySource = "messages.by_source.batch_get"
	// HostServiceMethodNotificationsBatchGetMessages is the catalog wire value for method "messages.batch_get".
	HostServiceMethodNotificationsBatchGetMessages = "messages.batch_get"
	// HostServiceMethodNotificationsDelete is the catalog wire value for method "messages.delete".
	HostServiceMethodNotificationsDelete = "messages.delete"
	// HostServiceMethodNotificationsDeleteBySource is the catalog wire value for method "messages.by_source.delete".
	HostServiceMethodNotificationsDeleteBySource = "messages.by_source.delete"
	// HostServiceMethodNotificationsEnsureVisible is the catalog wire value for method "messages.visible.ensure".
	HostServiceMethodNotificationsEnsureVisible = "messages.visible.ensure"
	// HostServiceMethodNotificationsList is the catalog wire value for method "messages.list".
	HostServiceMethodNotificationsList = "messages.list"
	// HostServiceMethodNotificationsMarkRead is the catalog wire value for method "messages.mark_read".
	HostServiceMethodNotificationsMarkRead = "messages.mark_read"
	// HostServiceMethodNotificationsMarkUnread is the catalog wire value for method "messages.mark_unread".
	HostServiceMethodNotificationsMarkUnread = "messages.mark_unread"
	// HostServiceMethodNotificationsSend is the catalog wire value for method "messages.send".
	HostServiceMethodNotificationsSend = "messages.send"
	// HostServiceMethodOrgAssignmentCleanupByUser is the catalog wire value for method "org.assignment.by_user.cleanup".
	HostServiceMethodOrgAssignmentCleanupByUser = "org.assignment.by_user.cleanup"
	// HostServiceMethodOrgAssignmentReplaceByUser is the catalog wire value for method "org.assignment.by_user.replace".
	HostServiceMethodOrgAssignmentReplaceByUser = "org.assignment.by_user.replace"
	// HostServiceMethodOrgAvailable is the catalog wire value for method "capability.available".
	HostServiceMethodOrgAvailable = "capability.available"
	// HostServiceMethodOrgBatchGetUserOrgProfiles is the catalog wire value for method "org.assignment.user_profiles.batch_get".
	HostServiceMethodOrgBatchGetUserOrgProfiles = "org.assignment.user_profiles.batch_get"
	// HostServiceMethodOrgDepartmentBatchGet is the catalog wire value for method "org.department.batch_get".
	HostServiceMethodOrgDepartmentBatchGet = "org.department.batch_get"
	// HostServiceMethodOrgDepartmentCreate is the catalog wire value for method "org.department.create".
	HostServiceMethodOrgDepartmentCreate = "org.department.create"
	// HostServiceMethodOrgDepartmentDelete is the catalog wire value for method "org.department.delete".
	HostServiceMethodOrgDepartmentDelete = "org.department.delete"
	// HostServiceMethodOrgDepartmentList is the catalog wire value for method "org.department.list".
	HostServiceMethodOrgDepartmentList = "org.department.list"
	// HostServiceMethodOrgDepartmentUpdate is the catalog wire value for method "org.department.update".
	HostServiceMethodOrgDepartmentUpdate = "org.department.update"
	// HostServiceMethodOrgEnsureDepartmentsVisible is the catalog wire value for method "org.department.visible.ensure_many".
	HostServiceMethodOrgEnsureDepartmentsVisible = "org.department.visible.ensure_many"
	// HostServiceMethodOrgEnsurePostsVisible is the catalog wire value for method "org.post.visible.ensure_many".
	HostServiceMethodOrgEnsurePostsVisible = "org.post.visible.ensure_many"
	// HostServiceMethodOrgListDeptTree is the catalog wire value for method "org.department.tree.list".
	HostServiceMethodOrgListDeptTree = "org.department.tree.list"
	// HostServiceMethodOrgListPostOptions is the catalog wire value for method "org.post.options.list".
	HostServiceMethodOrgListPostOptions = "org.post.options.list"
	// HostServiceMethodOrgPostBatchGet is the catalog wire value for method "org.post.batch_get".
	HostServiceMethodOrgPostBatchGet = "org.post.batch_get"
	// HostServiceMethodOrgPostCreate is the catalog wire value for method "org.post.create".
	HostServiceMethodOrgPostCreate = "org.post.create"
	// HostServiceMethodOrgPostDelete is the catalog wire value for method "org.post.delete".
	HostServiceMethodOrgPostDelete = "org.post.delete"
	// HostServiceMethodOrgPostUpdate is the catalog wire value for method "org.post.update".
	HostServiceMethodOrgPostUpdate = "org.post.update"
	// HostServiceMethodOrgStatus is the catalog wire value for method "capability.status".
	HostServiceMethodOrgStatus = "capability.status"
	// HostServiceMethodPluginsBatchGet is the catalog wire value for method "plugins.batch_get".
	HostServiceMethodPluginsBatchGet = "plugins.batch_get"
	// HostServiceMethodPluginsConfigGet is the catalog wire value for method "config.get".
	HostServiceMethodPluginsConfigGet = "config.get"
	// HostServiceMethodPluginsCurrent is the catalog wire value for method "plugins.current.get".
	HostServiceMethodPluginsCurrent = "plugins.current.get"
	// HostServiceMethodPluginsLifecycleEnsureTenantDeleteAllowed is the catalog wire value for method "plugins.lifecycle.tenant_delete.ensure".
	HostServiceMethodPluginsLifecycleEnsureTenantDeleteAllowed = "plugins.lifecycle.tenant_delete.ensure"
	// HostServiceMethodPluginsLifecycleEnsureTenantPluginDisableAllowed is the catalog wire value for method "plugins.lifecycle.tenant_plugin_disable.ensure".
	HostServiceMethodPluginsLifecycleEnsureTenantPluginDisableAllowed = "plugins.lifecycle.tenant_plugin_disable.ensure"
	// HostServiceMethodPluginsLifecycleNotifyTenantDeleted is the catalog wire value for method "plugins.lifecycle.tenant_deleted.notify".
	HostServiceMethodPluginsLifecycleNotifyTenantDeleted = "plugins.lifecycle.tenant_deleted.notify"
	// HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled is the catalog wire value for method "plugins.lifecycle.tenant_plugin_disabled.notify".
	HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled = "plugins.lifecycle.tenant_plugin_disabled.notify"
	// HostServiceMethodPluginsList is the catalog wire value for method "plugins.registry.list".
	HostServiceMethodPluginsList = "plugins.registry.list"
	// HostServiceMethodPluginsListTenant is the catalog wire value for method "plugins.tenant.list".
	HostServiceMethodPluginsListTenant = "plugins.tenant.list"
	// HostServiceMethodPluginsStateIsEnabled is the catalog wire value for method "plugins.state.enabled.check".
	HostServiceMethodPluginsStateIsEnabled = "plugins.state.enabled.check"
	// HostServiceMethodPluginsStateIsEnabledAuthoritative is the catalog wire value for method "plugins.state.enabled_authoritative.check".
	HostServiceMethodPluginsStateIsEnabledAuthoritative = "plugins.state.enabled_authoritative.check"
	// HostServiceMethodPluginsStateIsProviderEnabled is the catalog wire value for method "plugins.state.provider_enabled.check".
	HostServiceMethodPluginsStateIsProviderEnabled = "plugins.state.provider_enabled.check"
	// HostServiceMethodRouteMetadataGet is the catalog wire value for method "metadata.get".
	HostServiceMethodRouteMetadataGet = "metadata.get"
	// HostServiceMethodRuntimeInfoNode is the catalog wire value for method "info.node".
	HostServiceMethodRuntimeInfoNode = "info.node"
	// HostServiceMethodRuntimeInfoNow is the catalog wire value for method "info.now".
	HostServiceMethodRuntimeInfoNow = "info.now"
	// HostServiceMethodRuntimeInfoUUID is the catalog wire value for method "info.uuid".
	HostServiceMethodRuntimeInfoUUID = "info.uuid"
	// HostServiceMethodRuntimeLogWrite is the catalog wire value for method "log.write".
	HostServiceMethodRuntimeLogWrite = "log.write"
	// HostServiceMethodRuntimeStateDelete is the catalog wire value for method "state.delete".
	HostServiceMethodRuntimeStateDelete = "state.delete"
	// HostServiceMethodRuntimeStateDeleteMany is the catalog wire value for method "state.delete_many".
	HostServiceMethodRuntimeStateDeleteMany = "state.delete_many"
	// HostServiceMethodRuntimeStateGet is the catalog wire value for method "state.get".
	HostServiceMethodRuntimeStateGet = "state.get"
	// HostServiceMethodRuntimeStateGetMany is the catalog wire value for method "state.get_many".
	HostServiceMethodRuntimeStateGetMany = "state.get_many"
	// HostServiceMethodRuntimeStateSet is the catalog wire value for method "state.set".
	HostServiceMethodRuntimeStateSet = "state.set"
	// HostServiceMethodRuntimeStateSetMany is the catalog wire value for method "state.set_many".
	HostServiceMethodRuntimeStateSetMany = "state.set_many"
	// HostServiceMethodSessionsBatchGet is the catalog wire value for method "sessions.batch_get".
	HostServiceMethodSessionsBatchGet = "sessions.batch_get"
	// HostServiceMethodSessionsBatchGetUserOnlineStatus is the catalog wire value for method "sessions.users.online.batch_get".
	HostServiceMethodSessionsBatchGetUserOnlineStatus = "sessions.users.online.batch_get"
	// HostServiceMethodSessionsCurrent is the catalog wire value for method "sessions.current.get".
	HostServiceMethodSessionsCurrent = "sessions.current.get"
	// HostServiceMethodSessionsEnsureVisible is the catalog wire value for method "sessions.visible.ensure".
	HostServiceMethodSessionsEnsureVisible = "sessions.visible.ensure"
	// HostServiceMethodSessionsList is the catalog wire value for method "sessions.list".
	HostServiceMethodSessionsList = "sessions.list"
	// HostServiceMethodSessionsRevoke is the catalog wire value for method "sessions.revoke".
	HostServiceMethodSessionsRevoke = "sessions.revoke"
	// HostServiceMethodSessionsRevokeMany is the catalog wire value for method "sessions.revoke_many".
	HostServiceMethodSessionsRevokeMany = "sessions.revoke_many"
	// HostServiceMethodStorageDelete is the catalog wire value for method "delete".
	HostServiceMethodStorageDelete = "delete"
	// HostServiceMethodStorageDeleteBatch is the catalog wire value for method "delete.batch".
	HostServiceMethodStorageDeleteBatch = "delete.batch"
	// HostServiceMethodStorageGet is the catalog wire value for method "get".
	HostServiceMethodStorageGet = "get"
	// HostServiceMethodStorageList is the catalog wire value for method "list".
	HostServiceMethodStorageList = "list"
	// HostServiceMethodStorageListCursor is the catalog wire value for method "list.cursor".
	HostServiceMethodStorageListCursor = "list.cursor"
	// HostServiceMethodStoragePut is the catalog wire value for method "put".
	HostServiceMethodStoragePut = "put"
	// HostServiceMethodStoragePutAbort is the catalog wire value for method "put.abort".
	HostServiceMethodStoragePutAbort = "put.abort"
	// HostServiceMethodStoragePutChunk is the catalog wire value for method "put.chunk".
	HostServiceMethodStoragePutChunk = "put.chunk"
	// HostServiceMethodStoragePutCommit is the catalog wire value for method "put.commit".
	HostServiceMethodStoragePutCommit = "put.commit"
	// HostServiceMethodStoragePutInit is the catalog wire value for method "put.init".
	HostServiceMethodStoragePutInit = "put.init"
	// HostServiceMethodStorageStat is the catalog wire value for method "stat".
	HostServiceMethodStorageStat = "stat"
	// HostServiceMethodStorageStatBatch is the catalog wire value for method "stat.batch".
	HostServiceMethodStorageStatBatch = "stat.batch"
	// HostServiceMethodTenantAvailable is the catalog wire value for method "capability.available".
	HostServiceMethodTenantAvailable = "capability.available"
	// HostServiceMethodTenantBatchEnsureVisible is the catalog wire value for method "tenant.directory.visible.ensure_many".
	HostServiceMethodTenantBatchEnsureVisible = "tenant.directory.visible.ensure_many"
	// HostServiceMethodTenantBatchGet is the catalog wire value for method "tenant.directory.batch_get".
	HostServiceMethodTenantBatchGet = "tenant.directory.batch_get"
	// HostServiceMethodTenantCurrent is the catalog wire value for method "tenant.context.current".
	HostServiceMethodTenantCurrent = "tenant.context.current"
	// HostServiceMethodTenantCurrentInfo is the catalog wire value for method "tenant.context.info".
	HostServiceMethodTenantCurrentInfo = "tenant.context.info"
	// HostServiceMethodTenantDirectoryList is the catalog wire value for method "tenant.directory.list".
	HostServiceMethodTenantDirectoryList = "tenant.directory.list"
	// HostServiceMethodTenantFilterContext is the catalog wire value for method "tenant.filter.context".
	HostServiceMethodTenantFilterContext = "tenant.filter.context"
	// HostServiceMethodTenantListUserTenants is the catalog wire value for method "tenant.membership.list_by_user".
	HostServiceMethodTenantListUserTenants = "tenant.membership.list_by_user"
	// HostServiceMethodTenantPlatformBypass is the catalog wire value for method "tenant.context.platform_bypass".
	HostServiceMethodTenantPlatformBypass = "tenant.context.platform_bypass"
	// HostServiceMethodTenantPluginProvisionDefaults is the catalog wire value for method "tenant.plugins.defaults.provision".
	HostServiceMethodTenantPluginProvisionDefaults = "tenant.plugins.defaults.provision"
	// HostServiceMethodTenantPluginSetEnabled is the catalog wire value for method "tenant.plugins.enabled.set".
	HostServiceMethodTenantPluginSetEnabled = "tenant.plugins.enabled.set"
	// HostServiceMethodTenantStatus is the catalog wire value for method "capability.status".
	HostServiceMethodTenantStatus = "capability.status"
	// HostServiceMethodTenantValidateUserInTenant is the catalog wire value for method "tenant.membership.validate".
	HostServiceMethodTenantValidateUserInTenant = "tenant.membership.validate"
	// HostServiceMethodUsersBatchGet is the catalog wire value for method "users.batch_get".
	HostServiceMethodUsersBatchGet = "users.batch_get"
	// HostServiceMethodUsersBatchResolve is the catalog wire value for method "users.resolve.batch".
	HostServiceMethodUsersBatchResolve = "users.resolve.batch"
	// HostServiceMethodUsersCreate is the catalog wire value for method "users.create".
	HostServiceMethodUsersCreate = "users.create"
	// HostServiceMethodUsersCurrent is the catalog wire value for method "users.current.get".
	HostServiceMethodUsersCurrent = "users.current.get"
	// HostServiceMethodUsersDelete is the catalog wire value for method "users.delete".
	HostServiceMethodUsersDelete = "users.delete"
	// HostServiceMethodUsersEnsureVisible is the catalog wire value for method "users.visible.ensure".
	HostServiceMethodUsersEnsureVisible = "users.visible.ensure"
	// HostServiceMethodUsersList is the catalog wire value for method "users.list".
	HostServiceMethodUsersList = "users.list"
	// HostServiceMethodUsersReplaceRoles is the catalog wire value for method "users.assignment.roles.replace".
	HostServiceMethodUsersReplaceRoles = "users.assignment.roles.replace"
	// HostServiceMethodUsersResetPassword is the catalog wire value for method "users.password.reset".
	HostServiceMethodUsersResetPassword = "users.password.reset"
	// HostServiceMethodUsersSetStatus is the catalog wire value for method "users.status.set".
	HostServiceMethodUsersSetStatus = "users.status.set"
	// HostServiceMethodUsersUpdate is the catalog wire value for method "users.update".
	HostServiceMethodUsersUpdate = "users.update"
)
