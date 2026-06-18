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
	// CapabilityAuthz grants access to authorization-domain capability services.
	CapabilityAuthz = "host:authz"
	// CapabilityUsers grants access to host-defined user-domain capability services.
	CapabilityUsers = "host:users"
	// CapabilityBizCtx grants access to request business-context projections.
	CapabilityBizCtx = "host:bizctx"
	// CapabilityDict grants access to dictionary-domain capability services.
	CapabilityDict = "host:dict"
	// CapabilityFiles grants access to file-domain capability services.
	CapabilityFiles = "host:files"
	// CapabilityInfra grants access to infrastructure-domain capability services.
	CapabilityInfra = "host:infra"
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
	// HostServiceAuth is the authentication token capability host service identifier.
	HostServiceAuth = "auth"
	// HostServiceAuthz is the authorization-domain capability host service identifier.
	HostServiceAuthz = "authz"
	// HostServiceUsers is the user-domain capability host service identifier.
	HostServiceUsers = "users"
	// HostServiceBizCtx is the business-context capability host service identifier.
	HostServiceBizCtx = "bizctx"
	// HostServiceDict is the dictionary-domain capability host service identifier.
	HostServiceDict = "dict"
	// HostServiceFiles is the file-domain capability host service identifier.
	HostServiceFiles = "files"
	// HostServiceInfra is the infrastructure-domain capability host service identifier.
	HostServiceInfra = "infra"
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

// Auth host-service methods describe authentication token handoff operations.
const (
	// HostServiceMethodAuthSelectTenant issues a tenant token from a pre-login token.
	HostServiceMethodAuthSelectTenant = "tenant.select"
	// HostServiceMethodAuthSwitchTenant switches the current bearer token to another tenant.
	HostServiceMethodAuthSwitchTenant = "tenant.switch"
	// HostServiceMethodAuthIssueImpersonationToken issues one host-owned impersonation token.
	HostServiceMethodAuthIssueImpersonationToken = "impersonation_token.issue"
	// HostServiceMethodAuthRevokeImpersonationToken revokes one host-owned impersonation token.
	HostServiceMethodAuthRevokeImpersonationToken = "impersonation_token.revoke"
)

// Authz host-service methods describe authorization-domain ordinary reads.
const (
	// HostServiceMethodAuthzBatchGetPermissions reads visible permission projections.
	HostServiceMethodAuthzBatchGetPermissions = "permissions.batch_get"
	// HostServiceMethodAuthzBatchHasPermissions checks multiple permissions in one call.
	HostServiceMethodAuthzBatchHasPermissions = "permissions.batch_has"
	// HostServiceMethodAuthzHasPermission checks whether the current actor has one permission.
	HostServiceMethodAuthzHasPermission = "permissions.has"
	// HostServiceMethodAuthzIsPlatformAdmin checks whether one user has platform-admin scope.
	HostServiceMethodAuthzIsPlatformAdmin = "users.platform_admin.check"
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
	// HostServiceMethodUsersSearch searches visible user candidates with bounded paging.
	HostServiceMethodUsersSearch = "users.search"
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
	// HostServiceMethodDictResolveLabels resolves dictionary labels for requested values.
	HostServiceMethodDictResolveLabels = "labels.resolve"
	// HostServiceMethodDictListValues lists visible dictionary value candidates.
	HostServiceMethodDictListValues = "dict.values.list"
	// HostServiceMethodDictEnsureValuesVisible validates visible dictionary values.
	HostServiceMethodDictEnsureValuesVisible = "values.visible.ensure"
)

// Files host-service methods describe ordinary file-domain reads and checks.
const (
	// HostServiceMethodFilesBatchGet reads visible file projections in batch.
	HostServiceMethodFilesBatchGet = "files.batch_get"
	// HostServiceMethodFilesSearch searches visible file candidates.
	HostServiceMethodFilesSearch = "files.search"
	// HostServiceMethodFilesEnsureVisible validates that requested files are visible.
	HostServiceMethodFilesEnsureVisible = "files.visible.ensure"
)

// Infrastructure host-service methods describe ordinary infrastructure reads.
const (
	// HostServiceMethodInfraBatchGetStatus reads component status projections.
	HostServiceMethodInfraBatchGetStatus = "status.batch_get"
)

// Jobs host-service methods describe ordinary scheduled-job reads and
// discovery-time plugin job declarations.
const (
	// HostServiceMethodJobsBatchGet reads visible job projections in batch.
	HostServiceMethodJobsBatchGet = "jobs.batch_get"
	// HostServiceMethodJobsSearch searches visible scheduled-job candidates.
	HostServiceMethodJobsSearch = "jobs.search"
	// HostServiceMethodJobsEnsureVisible validates that requested jobs are visible.
	HostServiceMethodJobsEnsureVisible = "jobs.visible.ensure"
	// HostServiceMethodJobsRegister registers one dynamic-plugin job declaration during discovery.
	HostServiceMethodJobsRegister = "jobs.register"
)

// Notifications host-service methods describe notification reads and sends.
const (
	// HostServiceMethodNotificationsBatchGetMessages reads visible notification message projections.
	HostServiceMethodNotificationsBatchGetMessages = "messages.batch_get"
	// HostServiceMethodNotificationsBatchGetBySource reads visible messages by source IDs.
	HostServiceMethodNotificationsBatchGetBySource = "messages.by_source.batch_get"
	// HostServiceMethodNotificationsEnsureVisible validates notification message visibility.
	HostServiceMethodNotificationsEnsureVisible = "messages.visible.ensure"
	// HostServiceMethodNotificationsSend sends one governed notification message.
	HostServiceMethodNotificationsSend = "messages.send"
)

// Plugins host-service methods describe plugin-governance ordinary capability reads.
const (
	// HostServiceMethodPluginsCurrent reads the current caller plugin projection.
	HostServiceMethodPluginsCurrent = "plugins.current.get"
	// HostServiceMethodPluginsBatchGet reads visible plugin projections.
	HostServiceMethodPluginsBatchGet = "plugins.batch_get"
	// HostServiceMethodPluginsSearch searches visible plugin projections.
	HostServiceMethodPluginsSearch = "plugins.search"
	// HostServiceMethodPluginsListTenant lists tenant-controllable plugin projections.
	HostServiceMethodPluginsListTenant = "plugins.tenant.list"
	// HostServiceMethodPluginsBatchGetCapabilityStatus reads framework capability statuses.
	HostServiceMethodPluginsBatchGetCapabilityStatus = "plugins.capabilities.status.batch_get"
	// HostServiceMethodPluginsIsEnabled checks regular plugin enablement.
	HostServiceMethodPluginsIsEnabled = "plugins.enabled.check"
	// HostServiceMethodPluginsIsProviderEnabled checks provider enablement.
	HostServiceMethodPluginsIsProviderEnabled = "plugins.provider_enabled.check"
	// HostServiceMethodPluginsIsEnabledAuthoritative checks authoritative plugin enablement.
	HostServiceMethodPluginsIsEnabledAuthoritative = "plugins.enabled_authoritative.check"
	// HostServiceMethodPluginsConfigGet reads one plugin-scoped config value as JSON.
	HostServiceMethodPluginsConfigGet = "config.get"
	// HostServiceMethodPluginsLifecycleEnsureTenantPluginDisable runs tenant-plugin disable preconditions.
	HostServiceMethodPluginsLifecycleEnsureTenantPluginDisable = "lifecycle.tenant_plugin_disable.ensure"
	// HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled runs tenant-plugin disable notifications.
	HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled = "lifecycle.tenant_plugin_disabled.notify"
	// HostServiceMethodPluginsLifecycleEnsureTenantDelete runs tenant-delete preconditions.
	HostServiceMethodPluginsLifecycleEnsureTenantDelete = "lifecycle.tenant_delete.ensure"
	// HostServiceMethodPluginsLifecycleNotifyTenantDeleted runs tenant-delete notifications.
	HostServiceMethodPluginsLifecycleNotifyTenantDeleted = "lifecycle.tenant_deleted.notify"
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
	// HostServiceMethodSessionsSearch searches visible online sessions.
	HostServiceMethodSessionsSearch = "sessions.search"
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
	// HostServiceMethodOrgListUserDeptAssignments lists user department assignments in batch.
	HostServiceMethodOrgListUserDeptAssignments = "users.dept_assignments.list"
	// HostServiceMethodOrgBatchGetUserOrgProfiles reads user organization profiles in batch.
	HostServiceMethodOrgBatchGetUserOrgProfiles = "users.org_profiles.batch_get"
	// HostServiceMethodOrgGetUserDeptInfo reads one user's department identifier and name.
	HostServiceMethodOrgGetUserDeptInfo = "users.dept_info.get"
	// HostServiceMethodOrgGetUserDeptName reads one user's department name.
	HostServiceMethodOrgGetUserDeptName = "users.dept_name.get"
	// HostServiceMethodOrgGetUserDeptIDs reads one user's department identifiers.
	HostServiceMethodOrgGetUserDeptIDs = "users.dept_ids.get"
	// HostServiceMethodOrgGetUserPostIDs reads one user's post identifiers.
	HostServiceMethodOrgGetUserPostIDs = "users.post_ids.get"
	// HostServiceMethodOrgListDeptTree reads one bounded department tree projection.
	HostServiceMethodOrgListDeptTree = "depts.tree.list"
	// HostServiceMethodOrgSearchDepartments searches visible department projections.
	HostServiceMethodOrgSearchDepartments = "depts.search"
	// HostServiceMethodOrgListPostOptions lists bounded visible post candidates.
	HostServiceMethodOrgListPostOptions = "posts.options.list"
	// HostServiceMethodOrgEnsureDepartmentsVisible validates department visibility.
	HostServiceMethodOrgEnsureDepartmentsVisible = "depts.visible.ensure"
	// HostServiceMethodOrgEnsurePostsVisible validates post visibility.
	HostServiceMethodOrgEnsurePostsVisible = "posts.visible.ensure"
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
	HostServiceMethodTenantCurrent = "tenants.current"
	// HostServiceMethodTenantCurrentInfo reads the current request tenant projection.
	HostServiceMethodTenantCurrentInfo = "tenants.current_info.get"
	// HostServiceMethodTenantPlatformBypass reports whether tenant filtering may be bypassed.
	HostServiceMethodTenantPlatformBypass = "tenants.platform_bypass"
	// HostServiceMethodTenantEnsureVisible validates that the current user can access one tenant.
	HostServiceMethodTenantEnsureVisible = "tenants.visible.ensure"
	// HostServiceMethodTenantBatchGet reads visible tenant projections in batch.
	HostServiceMethodTenantBatchGet = "tenants.batch_get"
	// HostServiceMethodTenantSearch searches visible tenant candidates.
	HostServiceMethodTenantSearch = "tenants.search"
	// HostServiceMethodTenantValidateUserInTenant validates one user's tenant membership.
	HostServiceMethodTenantValidateUserInTenant = "users.tenant_membership.validate"
	// HostServiceMethodTenantListUserTenants lists tenants visible to one user.
	HostServiceMethodTenantListUserTenants = "users.tenants.list"
	// HostServiceMethodTenantBatchListUserTenants lists visible tenants for users in batch.
	HostServiceMethodTenantBatchListUserTenants = "users.tenants.batch_list"
	// HostServiceMethodTenantBatchEnsureVisible validates tenant visibility in batch.
	HostServiceMethodTenantBatchEnsureVisible = "tenants.visible.batch_ensure"
	// HostServiceMethodTenantValidateSwitch validates one tenant switch target.
	HostServiceMethodTenantValidateSwitch = "tenants.switch.validate"
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
