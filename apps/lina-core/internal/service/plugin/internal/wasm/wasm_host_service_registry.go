// This file wires all published dynamic-plugin host-service methods into the
// explicit dispatch registry. It deliberately keeps dependency ownership in the
// wasm package while removing the service-level switch from the entrypoint.

package wasm

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/wasm/hostservicedispatch"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

var (
	hostServiceDispatchRegistryOnce sync.Once
	hostServiceDispatchRegistry     *hostservicedispatch.Registry
	hostServiceDispatchRegistryErr  error
)

func defaultHostServiceDispatchRegistry() (*hostservicedispatch.Registry, error) {
	hostServiceDispatchRegistryOnce.Do(func() {
		hostServiceDispatchRegistry, hostServiceDispatchRegistryErr = newHostServiceDispatchRegistry()
	})
	return hostServiceDispatchRegistry, hostServiceDispatchRegistryErr
}

func newHostServiceDispatchRegistry() (*hostservicedispatch.Registry, error) {
	registry := hostservicedispatch.NewRegistry()
	for _, register := range []func(*hostservicedispatch.Registry) error{
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
		registerAuthzHostService,
		registerAIHostService,
		registerUsersHostService,
		registerBizCtxHostService,
		registerDictHostService,
		registerFilesHostService,
		registerInfraHostService,
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
	return registry.Dispatch(ctx, hostservicedispatch.Context{
		HostContext: hcc,
		Service:     request.Service,
		Method:      request.Method,
		ResourceRef: request.ResourceRef,
		Table:       request.Table,
		Payload:     request.Payload,
	})
}

type hostServiceDispatchAdapter func(context.Context, *hostCallContext, hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope

func registerHostServiceMethods(
	registry *hostservicedispatch.Registry,
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
	registry *hostservicedispatch.Registry,
	service string,
	method string,
	adapter hostServiceDispatchAdapter,
) error {
	if adapter == nil {
		return gerror.Newf("host service dispatch adapter is nil: %s.%s", service, method)
	}
	return registry.Register(service, method, func(ctx context.Context, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		hcc, ok := input.HostContext.(*hostCallContext)
		if !ok || hcc == nil {
			return bridgehostcall.NewHostCallErrorResponse(
				bridgehostcall.HostCallStatusInternalError,
				"host service call context is missing",
			)
		}
		return adapter(ctx, hcc, input)
	})
}

func registerRuntimeHostService(registry *hostservicedispatch.Registry) error {
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
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchRuntimeHostService(ctx, hcc, input.Method, input.Payload)
	})
}

func registerStorageHostService(registry *hostservicedispatch.Registry) error {
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
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchStorageHostService(ctx, hcc, input.ResourceRef, input.Method, input.Payload)
	})
}

func registerNetworkHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethod(registry, bridgehostservice.HostServiceNetwork, bridgehostservice.HostServiceMethodNetworkRequest,
		func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchNetworkHostService(ctx, hcc, input.ResourceRef, input.Method, input.Payload)
		})
}

func registerDataHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceData, []string{
		bridgehostservice.HostServiceMethodDataList,
		bridgehostservice.HostServiceMethodDataGet,
		bridgehostservice.HostServiceMethodDataBatchGet,
		bridgehostservice.HostServiceMethodDataCreate,
		bridgehostservice.HostServiceMethodDataUpdate,
		bridgehostservice.HostServiceMethodDataDelete,
		bridgehostservice.HostServiceMethodDataTransaction,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchDataHostService(ctx, hcc, input.Table, input.Method, input.Payload)
	})
}

func registerCacheHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceCache, []string{
		bridgehostservice.HostServiceMethodCacheGet,
		bridgehostservice.HostServiceMethodCacheGetMany,
		bridgehostservice.HostServiceMethodCacheSet,
		bridgehostservice.HostServiceMethodCacheSetMany,
		bridgehostservice.HostServiceMethodCacheDelete,
		bridgehostservice.HostServiceMethodCacheDeleteMany,
		bridgehostservice.HostServiceMethodCacheIncr,
		bridgehostservice.HostServiceMethodCacheExpire,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchCacheHostService(ctx, hcc, input.ResourceRef, input.Method, input.Payload)
	})
}

func registerLockHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceLock, []string{
		bridgehostservice.HostServiceMethodLockAcquire,
		bridgehostservice.HostServiceMethodLockRenew,
		bridgehostservice.HostServiceMethodLockRelease,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchLockHostService(ctx, hcc, input.ResourceRef, input.Method, input.Payload)
	})
}

func registerHostConfigHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethod(registry, bridgehostservice.HostServiceHostConfig, bridgehostservice.HostServiceMethodHostConfigGet,
		func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchHostConfigService(ctx, hcc, input.Method, input.Payload)
		})
}

func registerManifestHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceManifest, []string{
		bridgehostservice.HostServiceMethodManifestGet,
		bridgehostservice.HostServiceMethodManifestGetMany,
		bridgehostservice.HostServiceMethodManifestList,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchManifestHostService(ctx, hcc, input.ResourceRef, input.Method, input.Payload)
	})
}

func registerAPIDocHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceAPIDoc, []string{
		bridgehostservice.HostServiceMethodAPIDocResolveRouteText,
		bridgehostservice.HostServiceMethodAPIDocResolveRouteTexts,
		bridgehostservice.HostServiceMethodAPIDocFindRouteTitleOperationKeys,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchAPIDocHostService(ctx, hcc, input.Method, input.Payload)
	})
}

func registerAuthHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceAuth, []string{
		bridgehostservice.HostServiceMethodAuthSelectTenant,
		bridgehostservice.HostServiceMethodAuthSwitchTenant,
		bridgehostservice.HostServiceMethodAuthIssueImpersonationToken,
		bridgehostservice.HostServiceMethodAuthRevokeImpersonationToken,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchAuthHostService(ctx, hcc, input.Method, input.Payload)
	})
}

func registerAuthzHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceAuthz, []string{
		bridgehostservice.HostServiceMethodAuthzBatchGetPermissions,
		bridgehostservice.HostServiceMethodAuthzBatchHasPermissions,
		bridgehostservice.HostServiceMethodAuthzHasPermission,
		bridgehostservice.HostServiceMethodAuthzIsPlatformAdmin,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchAuthzHostService(ctx, hcc, input.Method, input.Payload)
	})
}

func registerAIHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceAI, []string{
		bridgehostservice.HostServiceMethodAITextGenerate,
		bridgehostservice.HostServiceMethodAITextMethodStatus,
		bridgehostservice.HostServiceMethodAIMethodStatuses,
		bridgehostservice.HostServiceMethodAIImageGenerate,
		bridgehostservice.HostServiceMethodAIImageEdit,
		bridgehostservice.HostServiceMethodAIEmbeddingCreate,
		bridgehostservice.HostServiceMethodAIAudioTranscribe,
		bridgehostservice.HostServiceMethodAIAudioSynthesize,
		bridgehostservice.HostServiceMethodAIVisionAnalyze,
		bridgehostservice.HostServiceMethodAIDocumentAnalyze,
		bridgehostservice.HostServiceMethodAIDocumentCite,
		bridgehostservice.HostServiceMethodAISafetyModerate,
		bridgehostservice.HostServiceMethodAIVideoGenerate,
		bridgehostservice.HostServiceMethodAIVideoEdit,
		bridgehostservice.HostServiceMethodAIVideoExtend,
		bridgehostservice.HostServiceMethodAIVideoOperationGet,
		bridgehostservice.HostServiceMethodAIVideoOperationCancel,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchAIHostService(ctx, hcc, input.ResourceRef, input.Method, input.Payload)
	})
}

func registerUsersHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceUsers, []string{
		bridgehostservice.HostServiceMethodUsersCurrent,
		bridgehostservice.HostServiceMethodUsersBatchGet,
		bridgehostservice.HostServiceMethodUsersBatchResolve,
		bridgehostservice.HostServiceMethodUsersSearch,
		bridgehostservice.HostServiceMethodUsersEnsureVisible,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchUsersHostService(ctx, hcc, input.Method, input.Payload)
	})
}

func registerBizCtxHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethod(registry, bridgehostservice.HostServiceBizCtx, bridgehostservice.HostServiceMethodBizCtxCurrent,
		func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchBizCtxHostService(ctx, hcc, input.Method, input.Payload)
		})
}

func registerDictHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceDict, []string{
		bridgehostservice.HostServiceMethodDictResolveLabels,
		bridgehostservice.HostServiceMethodDictListValues,
		bridgehostservice.HostServiceMethodDictEnsureValuesVisible,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchDictHostService(ctx, hcc, input.Method, input.Payload)
	})
}

func registerFilesHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceFiles, []string{
		bridgehostservice.HostServiceMethodFilesBatchGet,
		bridgehostservice.HostServiceMethodFilesSearch,
		bridgehostservice.HostServiceMethodFilesEnsureVisible,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchFilesHostService(ctx, hcc, input.Method, input.Payload)
	})
}

func registerInfraHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethod(registry, bridgehostservice.HostServiceInfra, bridgehostservice.HostServiceMethodInfraBatchGetStatus,
		func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchInfraHostService(ctx, hcc, input.Method, input.Payload)
		})
}

func registerJobsHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceJobs, []string{
		bridgehostservice.HostServiceMethodJobsBatchGet,
		bridgehostservice.HostServiceMethodJobsSearch,
		bridgehostservice.HostServiceMethodJobsEnsureVisible,
		bridgehostservice.HostServiceMethodJobsRegister,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchJobsHostService(ctx, hcc, input.Method, input.Payload)
	})
}

func registerNotificationsHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceNotifications, []string{
		bridgehostservice.HostServiceMethodNotificationsBatchGetMessages,
		bridgehostservice.HostServiceMethodNotificationsBatchGetBySource,
		bridgehostservice.HostServiceMethodNotificationsEnsureVisible,
		bridgehostservice.HostServiceMethodNotificationsSend,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchNotificationsHostService(ctx, hcc, input.ResourceRef, input.Method, input.Payload)
	})
}

func registerPluginsHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServicePlugins, []string{
		bridgehostservice.HostServiceMethodPluginsCurrent,
		bridgehostservice.HostServiceMethodPluginsBatchGet,
		bridgehostservice.HostServiceMethodPluginsSearch,
		bridgehostservice.HostServiceMethodPluginsListTenant,
		bridgehostservice.HostServiceMethodPluginsBatchGetCapabilityStatus,
		bridgehostservice.HostServiceMethodPluginsIsEnabled,
		bridgehostservice.HostServiceMethodPluginsIsProviderEnabled,
		bridgehostservice.HostServiceMethodPluginsIsEnabledAuthoritative,
		bridgehostservice.HostServiceMethodPluginsConfigGet,
		bridgehostservice.HostServiceMethodPluginsLifecycleEnsureTenantPluginDisable,
		bridgehostservice.HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled,
		bridgehostservice.HostServiceMethodPluginsLifecycleEnsureTenantDelete,
		bridgehostservice.HostServiceMethodPluginsLifecycleNotifyTenantDeleted,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchPluginsHostService(ctx, hcc, input.Method, input.Payload)
	})
}

func registerRouteHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethod(registry, bridgehostservice.HostServiceRoute, bridgehostservice.HostServiceMethodRouteMetadataGet,
		func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
			return dispatchRouteHostService(ctx, hcc, input.Method, input.Payload)
		})
}

func registerSessionsHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceSessions, []string{
		bridgehostservice.HostServiceMethodSessionsCurrent,
		bridgehostservice.HostServiceMethodSessionsSearch,
		bridgehostservice.HostServiceMethodSessionsBatchGet,
		bridgehostservice.HostServiceMethodSessionsBatchGetUserOnlineStatus,
		bridgehostservice.HostServiceMethodSessionsEnsureVisible,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchSessionsHostService(ctx, hcc, input.Method, input.Payload)
	})
}

func registerOrgHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceOrg, []string{
		bridgehostservice.HostServiceMethodOrgAvailable,
		bridgehostservice.HostServiceMethodOrgStatus,
		bridgehostservice.HostServiceMethodOrgListUserDeptAssignments,
		bridgehostservice.HostServiceMethodOrgBatchGetUserOrgProfiles,
		bridgehostservice.HostServiceMethodOrgGetUserDeptInfo,
		bridgehostservice.HostServiceMethodOrgGetUserDeptName,
		bridgehostservice.HostServiceMethodOrgGetUserDeptIDs,
		bridgehostservice.HostServiceMethodOrgGetUserPostIDs,
		bridgehostservice.HostServiceMethodOrgListDeptTree,
		bridgehostservice.HostServiceMethodOrgSearchDepartments,
		bridgehostservice.HostServiceMethodOrgListPostOptions,
		bridgehostservice.HostServiceMethodOrgEnsureDepartmentsVisible,
		bridgehostservice.HostServiceMethodOrgEnsurePostsVisible,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchOrgHostService(ctx, hcc, input.Method, input.Payload)
	})
}

func registerTenantHostService(registry *hostservicedispatch.Registry) error {
	return registerHostServiceMethods(registry, bridgehostservice.HostServiceTenant, []string{
		bridgehostservice.HostServiceMethodTenantAvailable,
		bridgehostservice.HostServiceMethodTenantStatus,
		bridgehostservice.HostServiceMethodTenantCurrent,
		bridgehostservice.HostServiceMethodTenantCurrentInfo,
		bridgehostservice.HostServiceMethodTenantPlatformBypass,
		bridgehostservice.HostServiceMethodTenantEnsureVisible,
		bridgehostservice.HostServiceMethodTenantBatchGet,
		bridgehostservice.HostServiceMethodTenantSearch,
		bridgehostservice.HostServiceMethodTenantValidateUserInTenant,
		bridgehostservice.HostServiceMethodTenantListUserTenants,
		bridgehostservice.HostServiceMethodTenantBatchListUserTenants,
		bridgehostservice.HostServiceMethodTenantBatchEnsureVisible,
		bridgehostservice.HostServiceMethodTenantValidateSwitch,
	}, func(ctx context.Context, hcc *hostCallContext, input hostservicedispatch.Context) *bridgehostcall.HostCallResponseEnvelope {
		return dispatchTenantHostService(ctx, hcc, input.Method, input.Payload)
	})
}
