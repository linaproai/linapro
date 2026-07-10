// This file wires all published dynamic-plugin host-service methods into the
// explicit dispatch registry. It deliberately keeps dependency ownership in the
// wasm package while removing the service-level switch from the entrypoint.

package wasm

import (
	"context"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"

	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

var (
	hostServiceDispatchRegistryOnce sync.Once
	hostServiceDispatchRegistryMemo *hostServiceDispatchRegistry
	errHostServiceDispatchRegistry  error
)

// hostServiceDispatchContext carries one authorized host-service invocation
// into a registered handler.
type hostServiceDispatchContext struct {
	hostContext *hostCallContext
	service     string
	method      string
	resourceRef string
	table       string
	payload     []byte
}

// hostServiceDispatchHandler dispatches one authorized host-service invocation.
type hostServiceDispatchHandler func(context.Context, hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope

// hostServiceDispatchRegistry stores explicitly registered host-service handlers.
type hostServiceDispatchRegistry struct {
	handlers map[string]hostServiceDispatchHandler
	methods  []hostServiceDispatchMethod
}

// hostServiceDispatchMethod describes one registered service/method pair.
type hostServiceDispatchMethod struct {
	service string
	method  string
}

func newEmptyHostServiceDispatchRegistry() *hostServiceDispatchRegistry {
	return &hostServiceDispatchRegistry{handlers: make(map[string]hostServiceDispatchHandler)}
}

func (r *hostServiceDispatchRegistry) register(service string, method string, handler hostServiceDispatchHandler) error {
	if r == nil {
		return gerror.New("host service dispatch registry is nil")
	}
	service = strings.TrimSpace(service)
	method = strings.TrimSpace(method)
	if service == "" || method == "" {
		return gerror.New("host service dispatch registration requires service and method")
	}
	if handler == nil {
		return gerror.Newf("host service dispatch handler is nil: %s.%s", service, method)
	}
	key := hostServiceDispatchRegistryKey(service, method)
	if _, ok := r.handlers[key]; ok {
		return gerror.Newf("host service dispatch handler already registered: %s.%s", service, method)
	}
	r.handlers[key] = handler
	r.methods = append(r.methods, hostServiceDispatchMethod{service: service, method: method})
	return nil
}

func (r *hostServiceDispatchRegistry) lookup(service string, method string) (hostServiceDispatchHandler, bool) {
	if r == nil {
		return nil, false
	}
	handler, ok := r.handlers[hostServiceDispatchRegistryKey(service, method)]
	return handler, ok
}

func (r *hostServiceDispatchRegistry) dispatch(
	ctx context.Context,
	input hostServiceDispatchContext,
) *bridgehostcall.HostCallResponseEnvelope {
	handler, ok := r.lookup(input.service, input.method)
	if !ok {
		return hostServiceDispatchNotFound(input.service, input.method)
	}
	return handler(ctx, input)
}

func hostServiceDispatchNotFound(service string, method string) *bridgehostcall.HostCallResponseEnvelope {
	return bridgehostcall.NewHostCallErrorResponse(
		bridgehostcall.HostCallStatusNotFound,
		"host service method not registered: "+strings.TrimSpace(service)+"."+strings.TrimSpace(method),
	)
}

func hostServiceDispatchRegistryKey(service string, method string) string {
	return strings.TrimSpace(service) + "\x00" + strings.TrimSpace(method)
}

func defaultHostServiceDispatchRegistry() (*hostServiceDispatchRegistry, error) {
	hostServiceDispatchRegistryOnce.Do(func() {
		hostServiceDispatchRegistryMemo, errHostServiceDispatchRegistry = newHostServiceDispatchRegistry()
	})
	return hostServiceDispatchRegistryMemo, errHostServiceDispatchRegistry
}

func newHostServiceDispatchRegistry() (*hostServiceDispatchRegistry, error) {
	registry := newEmptyHostServiceDispatchRegistry()
	for _, register := range []func(*hostServiceDispatchRegistry) error{
		registerRuntimeHostService,
		registerStorageHostService,
		registerNetworkHostService,
		registerDataHostService,
		registerCacheHostService,
		registerLockHostService,
		registerHostConfigHostService,
		registerManifestHostService,
		registerAPIDocHostService,
		registerAuthHostService,
		registerUsersHostService,
		registerBizCtxHostService,
		registerDictHostService,
		registerFilesHostService,
		registerJobsHostService,
		registerNotificationsHostService,
		registerPluginsHostService,
		registerRouteHostService,
		registerSessionsHostService,
		registerOrgHostService,
		registerTenantHostService,
	} {
		if err := register(registry); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func dispatchRegisteredHostService(
	ctx context.Context,
	hcc *hostCallContext,
	request *bridgehostservice.HostServiceRequestEnvelope,
) *bridgehostcall.HostCallResponseEnvelope {
	registry, err := defaultHostServiceDispatchRegistry()
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	if request == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInvalidRequest,
			"host service request is nil",
		)
	}
	return registry.dispatch(ctx, hostServiceDispatchContext{
		hostContext: hcc,
		service:     request.Service,
		method:      request.Method,
		resourceRef: request.ResourceRef,
		table:       request.Table,
		payload:     request.Payload,
	})
}

type hostServiceDispatchAdapter func(context.Context, *hostCallContext, hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope

func registerHostServiceMethods(
	registry *hostServiceDispatchRegistry,
	service string,
	methods []string,
	adapter hostServiceDispatchAdapter,
) error {
	for _, method := range methods {
		if err := registerHostServiceMethod(registry, service, method, adapter); err != nil {
			return err
		}
	}
	return nil
}

func registerHostServiceMethod(
	registry *hostServiceDispatchRegistry,
	service string,
	method string,
	adapter hostServiceDispatchAdapter,
) error {
	if adapter == nil {
		return gerror.Newf("host service dispatch adapter is nil: %s.%s", service, method)
	}
	return registry.register(service, method, func(ctx context.Context, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		if input.hostContext == nil {
			return bridgehostcall.NewHostCallErrorResponse(
				bridgehostcall.HostCallStatusInternalError,
				"host service call context is missing",
			)
		}
		return adapter(ctx, input.hostContext, input)
	})
}

func registerRuntimeHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceRuntime, []string{
		bridgehostservice.HostServiceMethodRuntimeLogWrite,
		bridgehostservice.HostServiceMethodRuntimeStateGet,
		bridgehostservice.HostServiceMethodRuntimeStateGetMany,
		bridgehostservice.HostServiceMethodRuntimeStateSet,
		bridgehostservice.HostServiceMethodRuntimeStateSetMany,
		bridgehostservice.HostServiceMethodRuntimeStateDelete,
		bridgehostservice.HostServiceMethodRuntimeStateDeleteMany,
		bridgehostservice.HostServiceMethodRuntimeInfoNow,
		bridgehostservice.HostServiceMethodRuntimeInfoUUID,
		bridgehostservice.HostServiceMethodRuntimeInfoNode,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchRuntimeHostService(ctx, hcc, input.method, input.payload)
	})
}

func registerStorageHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceStorage, []string{
		bridgehostservice.HostServiceMethodStoragePut,
		bridgehostservice.HostServiceMethodStoragePutInit,
		bridgehostservice.HostServiceMethodStoragePutChunk,
		bridgehostservice.HostServiceMethodStoragePutCommit,
		bridgehostservice.HostServiceMethodStoragePutAbort,
		bridgehostservice.HostServiceMethodStorageGet,
		bridgehostservice.HostServiceMethodStorageDelete,
		bridgehostservice.HostServiceMethodStorageDeleteBatch,
		bridgehostservice.HostServiceMethodStorageList,
		bridgehostservice.HostServiceMethodStorageListCursor,
		bridgehostservice.HostServiceMethodStorageStat,
		bridgehostservice.HostServiceMethodStorageStatBatch,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchStorageHostService(ctx, hcc, input.resourceRef, input.method, input.payload)
	})
}

func registerNetworkHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethod(registry, bridgehostservice.HostServiceNetwork, bridgehostservice.HostServiceMethodNetworkRequest,
		func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchNetworkHostService(ctx, hcc, input.resourceRef, input.method, input.payload)
		})
}

func registerDataHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceData, []string{
		bridgehostservice.HostServiceMethodDataList,
		bridgehostservice.HostServiceMethodDataGet,
		bridgehostservice.HostServiceMethodDataBatchGet,
		bridgehostservice.HostServiceMethodDataCreate,
		bridgehostservice.HostServiceMethodDataUpdate,
		bridgehostservice.HostServiceMethodDataDelete,
		bridgehostservice.HostServiceMethodDataTransaction,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchDataHostService(ctx, hcc, input.table, input.method, input.payload)
	})
}

func registerCacheHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceCache, []string{
		bridgehostservice.HostServiceMethodCacheGet,
		bridgehostservice.HostServiceMethodCacheGetMany,
		bridgehostservice.HostServiceMethodCacheSet,
		bridgehostservice.HostServiceMethodCacheSetMany,
		bridgehostservice.HostServiceMethodCacheDelete,
		bridgehostservice.HostServiceMethodCacheDeleteMany,
		bridgehostservice.HostServiceMethodCacheIncr,
		bridgehostservice.HostServiceMethodCacheExpire,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchCacheHostService(ctx, hcc, input.resourceRef, input.method, input.payload)
	})
}

func registerLockHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceLock, []string{
		bridgehostservice.HostServiceMethodLockAcquire,
		bridgehostservice.HostServiceMethodLockRenew,
		bridgehostservice.HostServiceMethodLockRelease,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchLockHostService(ctx, hcc, input.resourceRef, input.method, input.payload)
	})
}

func registerHostConfigHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceHostConfig, []string{
		bridgehostservice.HostServiceMethodHostConfigGet,
		bridgehostservice.HostServiceMethodHostConfigSysConfigGet,
		bridgehostservice.HostServiceMethodHostConfigSysConfigSetValue,
		bridgehostservice.HostServiceMethodHostConfigSysConfigReset,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchHostConfigService(ctx, hcc, input.resourceRef, input.method, input.payload)
	})
}

func registerManifestHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceManifest, []string{
		bridgehostservice.HostServiceMethodManifestGet,
		bridgehostservice.HostServiceMethodManifestGetMany,
		bridgehostservice.HostServiceMethodManifestList,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchManifestHostService(ctx, hcc, input.resourceRef, input.method, input.payload)
	})
}

func registerAPIDocHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceAPIDoc, []string{
		bridgehostservice.HostServiceMethodAPIDocResolveRouteText,
		bridgehostservice.HostServiceMethodAPIDocResolveRouteTexts,
		bridgehostservice.HostServiceMethodAPIDocFindRouteTitleOperationKeys,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchAPIDocHostService(ctx, hcc, input.method, input.payload)
	})
}

func registerAuthHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceAuth, []string{
		bridgehostservice.HostServiceMethodAuthSelectTenant,
		bridgehostservice.HostServiceMethodAuthSwitchTenant,
		bridgehostservice.HostServiceMethodAuthIssueImpersonationToken,
		bridgehostservice.HostServiceMethodAuthRevokeImpersonationToken,
		bridgehostservice.HostServiceMethodAuthzBatchGetPermissions,
		bridgehostservice.HostServiceMethodAuthzBatchHasPermissions,
		bridgehostservice.HostServiceMethodAuthzHasPermission,
		bridgehostservice.HostServiceMethodAuthzIsPlatformAdmin,
		bridgehostservice.HostServiceMethodAuthzReplaceRolePermissions,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchAuthHostService(ctx, hcc, input.method, input.payload)
	})
}

func registerUsersHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceUsers, []string{
		bridgehostservice.HostServiceMethodUsersCurrent,
		bridgehostservice.HostServiceMethodUsersBatchGet,
		bridgehostservice.HostServiceMethodUsersBatchResolve,
		bridgehostservice.HostServiceMethodUsersList,
		bridgehostservice.HostServiceMethodUsersEnsureVisible,
		bridgehostservice.HostServiceMethodUsersCreate,
		bridgehostservice.HostServiceMethodUsersUpdate,
		bridgehostservice.HostServiceMethodUsersDelete,
		bridgehostservice.HostServiceMethodUsersSetStatus,
		bridgehostservice.HostServiceMethodUsersResetPassword,
		bridgehostservice.HostServiceMethodUsersReplaceRoles,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchUsersHostService(ctx, hcc, input.method, input.payload)
	})
}

func registerBizCtxHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethod(registry, bridgehostservice.HostServiceBizCtx, bridgehostservice.HostServiceMethodBizCtxCurrent,
		func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchBizCtxHostService(ctx, hcc, input.method, input.payload)
		})
}

func registerDictHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceDict, []string{
		bridgehostservice.HostServiceMethodDictRefresh,
		bridgehostservice.HostServiceMethodDictTypeGet,
		bridgehostservice.HostServiceMethodDictTypeBatchGet,
		bridgehostservice.HostServiceMethodDictTypeList,
		bridgehostservice.HostServiceMethodDictTypeEnsureVisible,
		bridgehostservice.HostServiceMethodDictTypeEnsureKeysVisible,
		bridgehostservice.HostServiceMethodDictTypeCreate,
		bridgehostservice.HostServiceMethodDictTypeUpdate,
		bridgehostservice.HostServiceMethodDictTypeDelete,
		bridgehostservice.HostServiceMethodDictValueGet,
		bridgehostservice.HostServiceMethodDictValueBatchGet,
		bridgehostservice.HostServiceMethodDictValueResolveLabels,
		bridgehostservice.HostServiceMethodDictListValues,
		bridgehostservice.HostServiceMethodDictValueEnsureVisible,
		bridgehostservice.HostServiceMethodDictValueEnsureValuesVisible,
		bridgehostservice.HostServiceMethodDictValueCreate,
		bridgehostservice.HostServiceMethodDictValueUpdate,
		bridgehostservice.HostServiceMethodDictValueDelete,
		bridgehostservice.HostServiceMethodDictValueDeleteByType,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchDictHostService(ctx, hcc, input.method, input.payload)
	})
}

func registerFilesHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceFiles, []string{
		bridgehostservice.HostServiceMethodFilesBatchGet,
		bridgehostservice.HostServiceMethodFilesList,
		bridgehostservice.HostServiceMethodFilesEnsureVisible,
		bridgehostservice.HostServiceMethodFilesUpload,
		bridgehostservice.HostServiceMethodFilesCreateFromStorage,
		bridgehostservice.HostServiceMethodFilesUpdateMetadata,
		bridgehostservice.HostServiceMethodFilesDelete,
		bridgehostservice.HostServiceMethodFilesDeleteMany,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchFilesHostService(ctx, hcc, input.method, input.payload)
	})
}

func registerJobsHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceJobs, []string{
		bridgehostservice.HostServiceMethodJobsBatchGet,
		bridgehostservice.HostServiceMethodJobsList,
		bridgehostservice.HostServiceMethodJobsEnsureVisible,
		bridgehostservice.HostServiceMethodJobsCreate,
		bridgehostservice.HostServiceMethodJobsUpdate,
		bridgehostservice.HostServiceMethodJobsDelete,
		bridgehostservice.HostServiceMethodJobsRun,
		bridgehostservice.HostServiceMethodJobsSetStatus,
		bridgehostservice.HostServiceMethodJobsRegister,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchJobsHostService(ctx, hcc, input.method, input.payload)
	})
}

func registerNotificationsHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceNotifications, []string{
		bridgehostservice.HostServiceMethodNotificationsBatchGetMessages,
		bridgehostservice.HostServiceMethodNotificationsList,
		bridgehostservice.HostServiceMethodNotificationsBatchGetBySource,
		bridgehostservice.HostServiceMethodNotificationsEnsureVisible,
		bridgehostservice.HostServiceMethodNotificationsSend,
		bridgehostservice.HostServiceMethodNotificationsDelete,
		bridgehostservice.HostServiceMethodNotificationsDeleteBySource,
		bridgehostservice.HostServiceMethodNotificationsMarkRead,
		bridgehostservice.HostServiceMethodNotificationsMarkUnread,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchNotificationsHostService(ctx, hcc, input.resourceRef, input.method, input.payload)
	})
}

func registerPluginsHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServicePlugins, []string{
		bridgehostservice.HostServiceMethodPluginsCurrent,
		bridgehostservice.HostServiceMethodPluginsBatchGet,
		bridgehostservice.HostServiceMethodPluginsList,
		bridgehostservice.HostServiceMethodPluginsListTenant,
		bridgehostservice.HostServiceMethodPluginsConfigGet,
		bridgehostservice.HostServiceMethodPluginsStateIsEnabled,
		bridgehostservice.HostServiceMethodPluginsStateIsProviderEnabled,
		bridgehostservice.HostServiceMethodPluginsStateIsEnabledAuthoritative,
		bridgehostservice.HostServiceMethodPluginsLifecycleEnsureTenantPluginDisableAllowed,
		bridgehostservice.HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled,
		bridgehostservice.HostServiceMethodPluginsLifecycleEnsureTenantDeleteAllowed,
		bridgehostservice.HostServiceMethodPluginsLifecycleNotifyTenantDeleted,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchPluginsHostService(ctx, hcc, input.method, input.payload)
	})
}

func registerRouteHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethod(registry, bridgehostservice.HostServiceRoute, bridgehostservice.HostServiceMethodRouteMetadataGet,
		func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchRouteHostService(ctx, hcc, input.method, input.payload)
		})
}

func registerSessionsHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceSessions, []string{
		bridgehostservice.HostServiceMethodSessionsCurrent,
		bridgehostservice.HostServiceMethodSessionsList,
		bridgehostservice.HostServiceMethodSessionsBatchGet,
		bridgehostservice.HostServiceMethodSessionsBatchGetUserOnlineStatus,
		bridgehostservice.HostServiceMethodSessionsEnsureVisible,
		bridgehostservice.HostServiceMethodSessionsRevoke,
		bridgehostservice.HostServiceMethodSessionsRevokeMany,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchSessionsHostService(ctx, hcc, input.method, input.payload)
	})
}

func registerOrgHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceOrg, []string{
		bridgehostservice.HostServiceMethodOrgAvailable,
		bridgehostservice.HostServiceMethodOrgStatus,
		bridgehostservice.HostServiceMethodOrgBatchGetUserOrgProfiles,
		bridgehostservice.HostServiceMethodOrgListDeptTree,
		bridgehostservice.HostServiceMethodOrgDepartmentBatchGet,
		bridgehostservice.HostServiceMethodOrgDepartmentList,
		bridgehostservice.HostServiceMethodOrgPostBatchGet,
		bridgehostservice.HostServiceMethodOrgListPostOptions,
		bridgehostservice.HostServiceMethodOrgEnsureDepartmentsVisible,
		bridgehostservice.HostServiceMethodOrgEnsurePostsVisible,
		bridgehostservice.HostServiceMethodOrgDepartmentCreate,
		bridgehostservice.HostServiceMethodOrgDepartmentUpdate,
		bridgehostservice.HostServiceMethodOrgDepartmentDelete,
		bridgehostservice.HostServiceMethodOrgPostCreate,
		bridgehostservice.HostServiceMethodOrgPostUpdate,
		bridgehostservice.HostServiceMethodOrgPostDelete,
		bridgehostservice.HostServiceMethodOrgAssignmentReplaceByUser,
		bridgehostservice.HostServiceMethodOrgAssignmentCleanupByUser,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchOrgHostService(ctx, hcc, input.method, input.payload)
	})
}

func registerTenantHostService(registry *hostServiceDispatchRegistry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceTenant, []string{
		bridgehostservice.HostServiceMethodTenantAvailable,
		bridgehostservice.HostServiceMethodTenantStatus,
		bridgehostservice.HostServiceMethodTenantCurrent,
		bridgehostservice.HostServiceMethodTenantCurrentInfo,
		bridgehostservice.HostServiceMethodTenantPlatformBypass,
		bridgehostservice.HostServiceMethodTenantBatchGet,
		bridgehostservice.HostServiceMethodTenantDirectoryList,
		bridgehostservice.HostServiceMethodTenantValidateUserInTenant,
		bridgehostservice.HostServiceMethodTenantListUserTenants,
		bridgehostservice.HostServiceMethodTenantBatchEnsureVisible,
		bridgehostservice.HostServiceMethodTenantPluginSetEnabled,
		bridgehostservice.HostServiceMethodTenantPluginProvisionDefaults,
		bridgehostservice.HostServiceMethodTenantFilterContext,
	}, func(ctx context.Context, hcc *hostCallContext, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchTenantHostService(ctx, hcc, input.method, input.payload)
	})
}
