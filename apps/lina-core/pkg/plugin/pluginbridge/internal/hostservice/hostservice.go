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
	// CapabilitySecret grants access to governed secret resolution services.
	CapabilitySecret = "host:secret"
	// CapabilityEventPublish grants access to governed event publishing.
	CapabilityEventPublish = "host:event:publish"
	// CapabilityQueueEnqueue grants access to governed queue submission.
	CapabilityQueueEnqueue = "host:queue:enqueue"
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
	// CapabilityAI grants access to cross-sub-capability AI status projections.
	CapabilityAI = "host:ai"
	// CapabilityAIText grants access to host-defined text AI capability services.
	CapabilityAIText = "host:ai:text"
	// CapabilityAIImage grants access to host-defined image AI capability services.
	CapabilityAIImage = "host:ai:image"
	// CapabilityAIEmbedding grants access to host-defined embedding AI capability services.
	CapabilityAIEmbedding = "host:ai:embedding"
	// CapabilityAIAudio grants access to host-defined audio AI capability services.
	CapabilityAIAudio = "host:ai:audio"
	// CapabilityAIVision grants access to host-defined vision AI capability services.
	CapabilityAIVision = "host:ai:vision"
	// CapabilityAIDocument grants access to host-defined document AI capability services.
	CapabilityAIDocument = "host:ai:document"
	// CapabilityAISafety grants access to host-defined safety AI capability services.
	CapabilityAISafety = "host:ai:safety"
	// CapabilityAIVideo grants access to host-defined video AI capability services.
	CapabilityAIVideo = "host:ai:video"
	// CapabilityOrg grants access to host-defined organization capability services.
	CapabilityOrg = "host:org"
	// CapabilityTenant grants access to host-defined tenant capability services.
	CapabilityTenant = "host:tenant"
)

// Host service identifiers declare the logical service families exposed by the
// host runtime to plugins.
const (
	// HostServiceRuntime is the runtime host service identifier.
	HostServiceRuntime = "runtime"
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
	// HostServiceHostConfig is the host config service identifier.
	HostServiceHostConfig = "hostconfig"
	// HostServiceManifest is the plugin-scoped manifest resource service identifier.
	HostServiceManifest = "manifest"
	// HostServiceAPIDoc is the API-documentation capability host service identifier.
	HostServiceAPIDoc = "apidoc"
	// HostServiceAuth is the authentication and authorization capability host service identifier.
	HostServiceAuth = "auth"
	// HostServiceUsers is the user-domain capability host service identifier.
	HostServiceUsers = "users"
	// HostServiceBizCtx is the business-context capability host service identifier.
	HostServiceBizCtx = "bizctx"
	// HostServiceDict is the dictionary-domain capability host service identifier.
	HostServiceDict = "dict"
	// HostServiceFiles is the file-domain capability host service identifier.
	HostServiceFiles = "files"
	// HostServiceJobs is the scheduled-job capability host service identifier.
	HostServiceJobs = "jobs"
	// HostServiceNotifications is the notification-domain capability host service identifier.
	HostServiceNotifications = "notifications"
	// HostServicePlugins is the plugin-governance capability host service identifier.
	HostServicePlugins = "plugins"
	// HostServiceRoute is the dynamic-route metadata capability host service identifier.
	HostServiceRoute = "route"
	// HostServiceSessions is the online-session capability host service identifier.
	HostServiceSessions = "sessions"
	// HostServiceAI is the AI capability host service identifier.
	HostServiceAI = "ai"
	// HostServiceOrg is the organization capability host service identifier.
	HostServiceOrg = "org"
	// HostServiceTenant is the tenant capability host service identifier.
	HostServiceTenant = "tenant"
)

// Runtime host-service methods describe runtime logging, state, and info
// operations available to authorized plugins.
const (
	// HostServiceMethodRuntimeLogWrite writes one structured runtime log entry.
	HostServiceMethodRuntimeLogWrite = "log.write"
	// HostServiceMethodRuntimeStateGet reads one plugin-scoped runtime state value.
	HostServiceMethodRuntimeStateGet = "state.get"
	// HostServiceMethodRuntimeStateGetMany reads plugin-scoped runtime state values.
	HostServiceMethodRuntimeStateGetMany = "state.get_many"
	// HostServiceMethodRuntimeStateSet writes one plugin-scoped runtime state value.
	HostServiceMethodRuntimeStateSet = "state.set"
	// HostServiceMethodRuntimeStateSetMany writes plugin-scoped runtime state values.
	HostServiceMethodRuntimeStateSetMany = "state.set_many"
	// HostServiceMethodRuntimeStateDelete deletes one plugin-scoped runtime state value.
	HostServiceMethodRuntimeStateDelete = "state.delete"
	// HostServiceMethodRuntimeStateDeleteMany deletes plugin-scoped runtime state values.
	HostServiceMethodRuntimeStateDeleteMany = "state.delete_many"
	// HostServiceMethodRuntimeInfoNow returns host time information.
	HostServiceMethodRuntimeInfoNow = "info.now"
	// HostServiceMethodRuntimeInfoUUID returns one host-generated unique identifier.
	HostServiceMethodRuntimeInfoUUID = "info.uuid"
	// HostServiceMethodRuntimeInfoNode returns host node identity information.
	HostServiceMethodRuntimeInfoNode = "info.node"
)

// Storage host-service methods describe governed file operations under the
// plugin storage sandbox.
const (
	// HostServiceMethodStoragePut writes one governed storage object.
	HostServiceMethodStoragePut = "put"
	// HostServiceMethodStoragePutInit starts one governed storage upload session.
	HostServiceMethodStoragePutInit = "put.init"
	// HostServiceMethodStoragePutChunk appends one governed storage upload chunk.
	HostServiceMethodStoragePutChunk = "put.chunk"
	// HostServiceMethodStoragePutCommit commits one governed storage upload session.
	HostServiceMethodStoragePutCommit = "put.commit"
	// HostServiceMethodStoragePutAbort aborts one governed storage upload session.
	HostServiceMethodStoragePutAbort = "put.abort"
	// HostServiceMethodStorageGet reads one governed storage object.
	HostServiceMethodStorageGet = "get"
	// HostServiceMethodStorageDelete deletes one governed storage object.
	HostServiceMethodStorageDelete = "delete"
	// HostServiceMethodStorageDeleteBatch deletes governed storage objects by explicit paths.
	HostServiceMethodStorageDeleteBatch = "delete.batch"
	// HostServiceMethodStorageList lists governed storage objects under one prefix.
	HostServiceMethodStorageList = "list"
	// HostServiceMethodStorageListCursor lists governed storage objects under one prefix with cursor pagination.
	HostServiceMethodStorageListCursor = "list.cursor"
	// HostServiceMethodStorageStat reads metadata for one governed storage object.
	HostServiceMethodStorageStat = "stat"
	// HostServiceMethodStorageStatBatch reads metadata for governed storage objects by explicit paths.
	HostServiceMethodStorageStatBatch = "stat.batch"
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
	// HostServiceMethodDataBatchGet reads governed records by keys from an authorized data table.
	HostServiceMethodDataBatchGet = "batch_get"
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
	// HostServiceMethodCacheGetMany reads governed cache values by explicit keys.
	HostServiceMethodCacheGetMany = "get_many"
	// HostServiceMethodCacheSet writes one governed cache value.
	HostServiceMethodCacheSet = "set"
	// HostServiceMethodCacheSetMany writes governed cache values by explicit keys.
	HostServiceMethodCacheSetMany = "set_many"
	// HostServiceMethodCacheDelete removes one governed cache value.
	HostServiceMethodCacheDelete = "delete"
	// HostServiceMethodCacheDeleteMany removes governed cache values by explicit keys.
	HostServiceMethodCacheDeleteMany = "delete_many"
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

// HostConfig host-service methods describe authorized host config reads.
const (
	// HostServiceMethodHostConfigGet reads one authorized host config value.
	HostServiceMethodHostConfigGet = "get"
)

// Manifest host-service methods describe plugin-scoped manifest resource reads.
const (
	// HostServiceMethodManifestGet reads one plugin-scoped manifest resource.
	HostServiceMethodManifestGet = "get"
	// HostServiceMethodManifestGetMany reads plugin-scoped manifest resources by explicit paths.
	HostServiceMethodManifestGetMany = "get_many"
	// HostServiceMethodManifestList lists plugin-scoped manifest resource metadata.
	HostServiceMethodManifestList = "list"
)

// APIDoc host-service methods describe API-documentation localization reads.
const (
	// HostServiceMethodAPIDocResolveRouteText resolves one route text projection.
	HostServiceMethodAPIDocResolveRouteText = "route_text.resolve"
	// HostServiceMethodAPIDocResolveRouteTexts resolves multiple route text projections.
	HostServiceMethodAPIDocResolveRouteTexts = "route_texts.resolve"
	// HostServiceMethodAPIDocFindRouteTitleOperationKeys finds route title operation keys.
	HostServiceMethodAPIDocFindRouteTitleOperationKeys = "route_title_operation_keys.find"
)

// Auth host-service methods describe authentication token and authorization
// sub-capability operations under one auth domain.
const (
	// HostServiceMethodAuthSelectTenant issues a tenant token from a pre-login token.
	HostServiceMethodAuthSelectTenant = "token.tenant.select"
	// HostServiceMethodAuthSwitchTenant switches the current bearer token to another tenant.
	HostServiceMethodAuthSwitchTenant = "token.tenant.switch"
	// HostServiceMethodAuthIssueImpersonationToken issues one host-owned impersonation token.
	HostServiceMethodAuthIssueImpersonationToken = "token.impersonation_token.issue"
	// HostServiceMethodAuthRevokeImpersonationToken revokes one host-owned impersonation token.
	HostServiceMethodAuthRevokeImpersonationToken = "token.impersonation_token.revoke"
	// HostServiceMethodAuthzBatchGetPermissions reads visible permission projections.
	HostServiceMethodAuthzBatchGetPermissions = "authz.permissions.batch_get"
	// HostServiceMethodAuthzBatchHasPermissions checks multiple permissions in one call.
	HostServiceMethodAuthzBatchHasPermissions = "authz.permissions.batch_has"
	// HostServiceMethodAuthzHasPermission checks whether the current actor has one permission.
	HostServiceMethodAuthzHasPermission = "authz.permissions.has"
	// HostServiceMethodAuthzIsPlatformAdmin checks whether one user has platform-admin scope.
	HostServiceMethodAuthzIsPlatformAdmin = "authz.users.platform_admin.check"
)

// Users host-service methods describe the ordinary user-domain capability
// surface available to authorized dynamic plugins.
const (
	// HostServiceMethodUsersCurrent reads the current actor user projection.
	HostServiceMethodUsersCurrent = "users.current.get"
	// HostServiceMethodUsersBatchGet reads visible user projections in batch.
	HostServiceMethodUsersBatchGet = "users.batch_get"
	// HostServiceMethodUsersBatchResolve resolves visible users by stable identifiers.
	HostServiceMethodUsersBatchResolve = "users.resolve.batch"
	// HostServiceMethodUsersList lists visible user candidates with bounded paging.
	HostServiceMethodUsersList = "users.list"
	// HostServiceMethodUsersEnsureVisible validates that all requested users are visible.
	HostServiceMethodUsersEnsureVisible = "users.visible.ensure"
)

// Business-context host-service methods describe current request projections.
const (
	// HostServiceMethodBizCtxCurrent reads the current request business context.
	HostServiceMethodBizCtxCurrent = "current.get"
)

// Dictionary host-service methods describe ordinary dictionary reads.
const (
	// HostServiceMethodDictValueResolveLabels resolves dictionary labels for requested values.
	HostServiceMethodDictValueResolveLabels = "dict.value.labels.resolve"
	// HostServiceMethodDictListValues lists visible dictionary value candidates.
	HostServiceMethodDictListValues = "dict.value.list"
	// HostServiceMethodDictValueEnsureValuesVisible validates visible dictionary values.
	HostServiceMethodDictValueEnsureValuesVisible = "dict.value.values.visible.ensure"
)

// Files host-service methods describe ordinary file-domain reads and checks.
const (
	// HostServiceMethodFilesBatchGet reads visible file projections in batch.
	HostServiceMethodFilesBatchGet = "files.batch_get"
	// HostServiceMethodFilesList lists visible file candidates.
	HostServiceMethodFilesList = "files.list"
	// HostServiceMethodFilesEnsureVisible validates that requested files are visible.
	HostServiceMethodFilesEnsureVisible = "files.visible.ensure"
	// HostServiceMethodFilesUpload creates one host file-center record from direct content.
	HostServiceMethodFilesUpload = "files.upload"
	// HostServiceMethodFilesCreateFromStorage creates one host file-center record from plugin storage.
	HostServiceMethodFilesCreateFromStorage = "files.create_from_storage"
)

// Jobs host-service methods describe ordinary scheduled-job reads and
// discovery-time plugin job declarations.
const (
	// HostServiceMethodJobsBatchGet reads visible job projections in batch.
	HostServiceMethodJobsBatchGet = "jobs.batch_get"
	// HostServiceMethodJobsList lists visible scheduled-job candidates.
	HostServiceMethodJobsList = "jobs.list"
	// HostServiceMethodJobsEnsureVisible validates that requested jobs are visible.
	HostServiceMethodJobsEnsureVisible = "jobs.visible.ensure"
	// HostServiceMethodJobsRegister registers one dynamic-plugin job declaration during discovery.
	HostServiceMethodJobsRegister = "jobs.register"
)

// Notifications host-service methods describe notification reads and sends.
const (
	// HostServiceMethodNotificationsBatchGetMessages reads visible notification message projections.
	HostServiceMethodNotificationsBatchGetMessages = "messages.batch_get"
	// HostServiceMethodNotificationsList lists visible notification message projections.
	HostServiceMethodNotificationsList = "messages.list"
	// HostServiceMethodNotificationsBatchGetBySource reads visible messages by source IDs.
	HostServiceMethodNotificationsBatchGetBySource = "messages.by_source.batch_get"
	// HostServiceMethodNotificationsEnsureVisible validates notification message visibility.
	HostServiceMethodNotificationsEnsureVisible = "messages.visible.ensure"
	// HostServiceMethodNotificationsSend sends one governed notification message.
	HostServiceMethodNotificationsSend = "messages.send"
	// HostServiceMethodNotificationsDelete removes visible notification messages.
	HostServiceMethodNotificationsDelete = "messages.delete"
	// HostServiceMethodNotificationsDeleteBySource removes visible messages by source IDs.
	HostServiceMethodNotificationsDeleteBySource = "messages.by_source.delete"
	// HostServiceMethodNotificationsMarkRead marks visible notification messages read.
	HostServiceMethodNotificationsMarkRead = "messages.mark_read"
	// HostServiceMethodNotificationsMarkUnread marks visible notification messages unread.
	HostServiceMethodNotificationsMarkUnread = "messages.mark_unread"
)

// Plugins host-service methods describe plugin-governance ordinary capability reads.
const (
	// HostServiceMethodPluginsCurrent reads the current caller plugin projection.
	HostServiceMethodPluginsCurrent = "plugins.current.get"
	// HostServiceMethodPluginsBatchGet reads visible plugin projections.
	HostServiceMethodPluginsBatchGet = "plugins.batch_get"
	// HostServiceMethodPluginsList lists visible plugin projections.
	HostServiceMethodPluginsList = "plugins.registry.list"
	// HostServiceMethodPluginsListTenant lists tenant-controllable plugin projections.
	HostServiceMethodPluginsListTenant = "plugins.tenant.list"
	// HostServiceMethodPluginsConfigGet reads one plugin-scoped config value as JSON.
	HostServiceMethodPluginsConfigGet = "config.get"
	// HostServiceMethodPluginsStateIsEnabled checks plugin business-entry enablement.
	HostServiceMethodPluginsStateIsEnabled = "plugins.state.enabled.check"
	// HostServiceMethodPluginsStateIsProviderEnabled checks provider enablement.
	HostServiceMethodPluginsStateIsProviderEnabled = "plugins.state.provider_enabled.check"
	// HostServiceMethodPluginsStateIsEnabledAuthoritative checks persisted plugin enablement.
	HostServiceMethodPluginsStateIsEnabledAuthoritative = "plugins.state.enabled_authoritative.check"
	// HostServiceMethodPluginsLifecycleEnsureTenantPluginDisableAllowed runs tenant-plugin disable preconditions.
	HostServiceMethodPluginsLifecycleEnsureTenantPluginDisableAllowed = "plugins.lifecycle.tenant_plugin_disable.ensure"
	// HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled runs tenant-plugin disabled notifications.
	HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled = "plugins.lifecycle.tenant_plugin_disabled.notify"
	// HostServiceMethodPluginsLifecycleEnsureTenantDeleteAllowed runs tenant-delete preconditions.
	HostServiceMethodPluginsLifecycleEnsureTenantDeleteAllowed = "plugins.lifecycle.tenant_delete.ensure"
	// HostServiceMethodPluginsLifecycleNotifyTenantDeleted runs tenant-deleted notifications.
	HostServiceMethodPluginsLifecycleNotifyTenantDeleted = "plugins.lifecycle.tenant_deleted.notify"
)

// Route host-service methods describe current dynamic-route metadata reads.
const (
	// HostServiceMethodRouteMetadataGet reads current dynamic-route metadata.
	HostServiceMethodRouteMetadataGet = "metadata.get"
)

// Sessions host-service methods describe ordinary online-session reads.
const (
	// HostServiceMethodSessionsCurrent reads the current online-session projection.
	HostServiceMethodSessionsCurrent = "sessions.current.get"
	// HostServiceMethodSessionsList lists visible online sessions.
	HostServiceMethodSessionsList = "sessions.list"
	// HostServiceMethodSessionsBatchGet reads visible online sessions in batch.
	HostServiceMethodSessionsBatchGet = "sessions.batch_get"
	// HostServiceMethodSessionsBatchGetUserOnlineStatus reads visible user online states.
	HostServiceMethodSessionsBatchGetUserOnlineStatus = "sessions.users.online.batch_get"
	// HostServiceMethodSessionsEnsureVisible validates that requested sessions are visible.
	HostServiceMethodSessionsEnsureVisible = "sessions.visible.ensure"
)

// Organization host-service methods describe the ordinary organization
// capability surface available to authorized dynamic plugins. Capability
// business DTOs are owned by capability/orgcap and adapted by guest clients.
const (
	// HostServiceMethodOrgAvailable reports whether organization capability is available.
	HostServiceMethodOrgAvailable = "capability.available"
	// HostServiceMethodOrgStatus reads organization capability status.
	HostServiceMethodOrgStatus = "capability.status"
	// HostServiceMethodOrgBatchGetUserOrgProfiles reads user organization profiles in batch.
	HostServiceMethodOrgBatchGetUserOrgProfiles = "org.assignment.user_profiles.batch_get"
	// HostServiceMethodOrgListDeptTree reads one bounded department tree projection.
	HostServiceMethodOrgListDeptTree = "org.department.tree.list"
	// HostServiceMethodOrgDepartmentBatchGet reads visible department projections in batch.
	HostServiceMethodOrgDepartmentBatchGet = "org.department.batch_get"
	// HostServiceMethodOrgDepartmentList lists visible department projections.
	HostServiceMethodOrgDepartmentList = "org.department.list"
	// HostServiceMethodOrgPostBatchGet reads visible post projections in batch.
	HostServiceMethodOrgPostBatchGet = "org.post.batch_get"
	// HostServiceMethodOrgListPostOptions lists bounded visible post candidates.
	HostServiceMethodOrgListPostOptions = "org.post.options.list"
	// HostServiceMethodOrgEnsureDepartmentsVisible validates department visibility.
	HostServiceMethodOrgEnsureDepartmentsVisible = "org.department.visible.ensure_many"
	// HostServiceMethodOrgEnsurePostsVisible validates post visibility.
	HostServiceMethodOrgEnsurePostsVisible = "org.post.visible.ensure_many"
)

// Tenant host-service methods describe the ordinary tenant capability surface
// available to authorized dynamic plugins. Request resolution and query builders
// stay out of this protocol; plugin lifecycle governance belongs to plugins.
const (
	// HostServiceMethodTenantAvailable reports whether tenant capability is available.
	HostServiceMethodTenantAvailable = "capability.available"
	// HostServiceMethodTenantStatus reads tenant capability status.
	HostServiceMethodTenantStatus = "capability.status"
	// HostServiceMethodTenantCurrent reads the current request tenant.
	HostServiceMethodTenantCurrent = "tenant.context.current"
	// HostServiceMethodTenantCurrentInfo reads the current request tenant projection.
	HostServiceMethodTenantCurrentInfo = "tenant.context.info"
	// HostServiceMethodTenantPlatformBypass reports whether tenant filtering may be bypassed.
	HostServiceMethodTenantPlatformBypass = "tenant.context.platform_bypass"
	// HostServiceMethodTenantBatchGet reads visible tenant projections in batch.
	HostServiceMethodTenantBatchGet = "tenant.directory.batch_get"
	// HostServiceMethodTenantDirectoryList lists visible tenant candidates.
	HostServiceMethodTenantDirectoryList = "tenant.directory.list"
	// HostServiceMethodTenantValidateUserInTenant validates one user's tenant membership.
	HostServiceMethodTenantValidateUserInTenant = "tenant.membership.validate"
	// HostServiceMethodTenantListUserTenants lists tenants visible to one user.
	HostServiceMethodTenantListUserTenants = "tenant.membership.list_by_user"
	// HostServiceMethodTenantBatchEnsureVisible validates tenant visibility in batch.
	HostServiceMethodTenantBatchEnsureVisible = "tenant.directory.visible.ensure_many"
	// HostServiceMethodTenantPluginSetEnabled updates tenant plugin enablement.
	HostServiceMethodTenantPluginSetEnabled = "tenant.plugins.enabled.set"
	// HostServiceMethodTenantPluginProvisionDefaults provisions default tenant plugins.
	HostServiceMethodTenantPluginProvisionDefaults = "tenant.plugins.defaults.provision"
	// HostServiceMethodTenantFilterContext reads plugin table tenant-filter context.
	HostServiceMethodTenantFilterContext = "tenant.filter.context"
)

// AI host-service methods describe the governed typed AI capability surface
// available to authorized dynamic plugins.
const (
	// HostServiceMethodAITextGenerate executes one governed text generation request.
	HostServiceMethodAITextGenerate = "text.generate"
	// HostServiceMethodAITextMethodStatus reads one text AI method availability projection.
	HostServiceMethodAITextMethodStatus = "text.method_status.get"
	// HostServiceMethodAIMethodStatuses reads AI method availability projections across sub-capabilities.
	HostServiceMethodAIMethodStatuses = "ai.methods.status.batch_get"
	// HostServiceMethodAIImageGenerate executes one governed image generation request.
	HostServiceMethodAIImageGenerate = "image.generate"
	// HostServiceMethodAIImageEdit executes one governed image editing request.
	HostServiceMethodAIImageEdit = "image.edit"
	// HostServiceMethodAIEmbeddingCreate executes one governed embedding request.
	HostServiceMethodAIEmbeddingCreate = "embedding.create"
	// HostServiceMethodAIAudioTranscribe executes one governed audio transcription request.
	HostServiceMethodAIAudioTranscribe = "audio.transcribe"
	// HostServiceMethodAIAudioSynthesize executes one governed audio synthesis request.
	HostServiceMethodAIAudioSynthesize = "audio.synthesize"
	// HostServiceMethodAIVisionAnalyze executes one governed visual analysis request.
	HostServiceMethodAIVisionAnalyze = "vision.analyze"
	// HostServiceMethodAIDocumentAnalyze executes one governed document analysis request.
	HostServiceMethodAIDocumentAnalyze = "document.analyze"
	// HostServiceMethodAIDocumentCite executes one governed citation-aware document request.
	HostServiceMethodAIDocumentCite = "document.cite"
	// HostServiceMethodAISafetyModerate executes one governed safety moderation request.
	HostServiceMethodAISafetyModerate = "safety.moderate"
	// HostServiceMethodAIVideoGenerate executes one governed video generation request.
	HostServiceMethodAIVideoGenerate = "video.generate"
	// HostServiceMethodAIVideoEdit executes one governed video editing request.
	HostServiceMethodAIVideoEdit = "video.edit"
	// HostServiceMethodAIVideoExtend executes one governed video extension request.
	HostServiceMethodAIVideoExtend = "video.extend"
	// HostServiceMethodAIVideoOperationGet reads one governed provider operation.
	HostServiceMethodAIVideoOperationGet = "video.operation.get"
	// HostServiceMethodAIVideoOperationCancel cancels one governed provider operation.
	HostServiceMethodAIVideoOperationCancel = "video.operation.cancel"
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
