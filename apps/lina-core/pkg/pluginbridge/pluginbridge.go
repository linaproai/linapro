// Package pluginbridge defines the shared runtime bridge contracts, codecs,
// and guest helpers used by Lina dynamic plugins.
package pluginbridge

import (
	"lina-core/pkg/pluginbridge/artifact"
	"lina-core/pkg/pluginbridge/codec"
	"lina-core/pkg/pluginbridge/contract"
	"lina-core/pkg/pluginbridge/guest"
	"lina-core/pkg/pluginbridge/hostcall"
	"lina-core/pkg/pluginbridge/hostservice"
)

// Public bridge types are preserved as aliases for existing callers while new
// code can import the focused subcomponents directly.
type (
	BridgeFailureV1          = contract.BridgeFailureV1
	BridgeRequestEnvelopeV1  = contract.BridgeRequestEnvelopeV1
	BridgeResponseEnvelopeV1 = contract.BridgeResponseEnvelopeV1
	BridgeSpec               = contract.BridgeSpec
	CronConcurrency          = contract.CronConcurrency
	CronContract             = contract.CronContract
	CronScope                = contract.CronScope
	ExecutionSource          = contract.ExecutionSource
	HTTPRequestSnapshotV1    = contract.HTTPRequestSnapshotV1
	IdentitySnapshotV1       = contract.IdentitySnapshotV1
	LifecycleContract        = contract.LifecycleContract
	LifecycleDecision        = contract.LifecycleDecision
	LifecycleOperation       = contract.LifecycleOperation
	LifecycleRequest         = contract.LifecycleRequest
	RouteContract            = contract.RouteContract
	RouteMatchSnapshotV1     = contract.RouteMatchSnapshotV1
	RuntimeArtifactMetadata  = artifact.RuntimeArtifactMetadata

	HostCallLogRequest         = hostcall.HostCallLogRequest
	HostCallResponseEnvelope   = hostcall.HostCallResponseEnvelope
	HostCallStateDeleteRequest = hostcall.HostCallStateDeleteRequest
	HostCallStateGetRequest    = hostcall.HostCallStateGetRequest
	HostCallStateGetResponse   = hostcall.HostCallStateGetResponse
	HostCallStateSetRequest    = hostcall.HostCallStateSetRequest

	HostServiceCacheDeleteRequest       = hostservice.HostServiceCacheDeleteRequest
	HostServiceCacheExpireRequest       = hostservice.HostServiceCacheExpireRequest
	HostServiceCacheExpireResponse      = hostservice.HostServiceCacheExpireResponse
	HostServiceCacheGetRequest          = hostservice.HostServiceCacheGetRequest
	HostServiceCacheGetResponse         = hostservice.HostServiceCacheGetResponse
	HostServiceCacheIncrRequest         = hostservice.HostServiceCacheIncrRequest
	HostServiceCacheIncrResponse        = hostservice.HostServiceCacheIncrResponse
	HostServiceCacheSetRequest          = hostservice.HostServiceCacheSetRequest
	HostServiceCacheSetResponse         = hostservice.HostServiceCacheSetResponse
	HostServiceCacheValue               = hostservice.HostServiceCacheValue
	HostServiceConfigKeyRequest         = hostservice.HostServiceConfigKeyRequest
	HostServiceConfigValueResponse      = hostservice.HostServiceConfigValueResponse
	HostServiceCronRegisterRequest      = hostservice.HostServiceCronRegisterRequest
	HostServiceDataGetRequest           = hostservice.HostServiceDataGetRequest
	HostServiceDataGetResponse          = hostservice.HostServiceDataGetResponse
	HostServiceDataListRequest          = hostservice.HostServiceDataListRequest
	HostServiceDataListResponse         = hostservice.HostServiceDataListResponse
	HostServiceDataMutationRequest      = hostservice.HostServiceDataMutationRequest
	HostServiceDataMutationResponse     = hostservice.HostServiceDataMutationResponse
	HostServiceDataTransactionOperation = hostservice.HostServiceDataTransactionOperation
	HostServiceDataTransactionRequest   = hostservice.HostServiceDataTransactionRequest
	HostServiceDataTransactionResponse  = hostservice.HostServiceDataTransactionResponse
	HostServiceLockAcquireRequest       = hostservice.HostServiceLockAcquireRequest
	HostServiceLockAcquireResponse      = hostservice.HostServiceLockAcquireResponse
	HostServiceLockReleaseRequest       = hostservice.HostServiceLockReleaseRequest
	HostServiceLockRenewRequest         = hostservice.HostServiceLockRenewRequest
	HostServiceLockRenewResponse        = hostservice.HostServiceLockRenewResponse
	HostServiceNetworkRequest           = hostservice.HostServiceNetworkRequest
	HostServiceNetworkResponse          = hostservice.HostServiceNetworkResponse
	HostServiceNotifySendRequest        = hostservice.HostServiceNotifySendRequest
	HostServiceNotifySendResponse       = hostservice.HostServiceNotifySendResponse
	HostServiceRequestEnvelope          = hostservice.HostServiceRequestEnvelope
	HostServiceResourceSpec             = hostservice.HostServiceResourceSpec
	HostServiceSpec                     = hostservice.HostServiceSpec
	HostServiceStorageDeleteRequest     = hostservice.HostServiceStorageDeleteRequest
	HostServiceStorageGetRequest        = hostservice.HostServiceStorageGetRequest
	HostServiceStorageGetResponse       = hostservice.HostServiceStorageGetResponse
	HostServiceStorageListRequest       = hostservice.HostServiceStorageListRequest
	HostServiceStorageListResponse      = hostservice.HostServiceStorageListResponse
	HostServiceStorageObject            = hostservice.HostServiceStorageObject
	HostServiceStoragePutRequest        = hostservice.HostServiceStoragePutRequest
	HostServiceStoragePutResponse       = hostservice.HostServiceStoragePutResponse
	HostServiceStorageStatRequest       = hostservice.HostServiceStorageStatRequest
	HostServiceStorageStatResponse      = hostservice.HostServiceStorageStatResponse
	HostServiceValueResponse            = hostservice.HostServiceValueResponse
	GuestControllerRouteDispatcher      = guest.GuestControllerRouteDispatcher
	GuestControllerHandlerKind          = guest.GuestControllerHandlerKind
	GuestControllerHandlerMetadata      = guest.GuestControllerHandlerMetadata
	GuestHandler                        = guest.GuestHandler
	GuestRuntime                        = guest.GuestRuntime
	ErrorCase                           = guest.ErrorCase
	ErrorClassifier                     = guest.ErrorClassifier
	ErrorMatcher                        = guest.ErrorMatcher
	ErrorResponseBuilder                = guest.ErrorResponseBuilder
	ResponseError                       = guest.ResponseError
)

// Bridge contract constants are kept on the root package for compatibility.
const (
	CodecProtobuf                 = contract.CodecProtobuf
	AccessPublic                  = contract.AccessPublic
	AccessLogin                   = contract.AccessLogin
	RuntimeKindWasm               = contract.RuntimeKindWasm
	ABIVersionV1                  = contract.ABIVersionV1
	SupportedABIVersion           = contract.SupportedABIVersion
	DefaultGuestAllocExport       = contract.DefaultGuestAllocExport
	DefaultGuestExecuteExport     = contract.DefaultGuestExecuteExport
	BridgeFailureCodeUnauthorized = contract.BridgeFailureCodeUnauthorized
	BridgeFailureCodeForbidden    = contract.BridgeFailureCodeForbidden
	BridgeFailureCodeBadRequest   = contract.BridgeFailureCodeBadRequest
	BridgeFailureCodeNotFound     = contract.BridgeFailureCodeNotFound
	BridgeFailureCodeInternal     = contract.BridgeFailureCodeInternal

	DefaultCronContractTimezone          = contract.DefaultCronContractTimezone
	DefaultCronContractTimeoutSeconds    = contract.DefaultCronContractTimeoutSeconds
	DeclaredCronRouteBasePath            = contract.DeclaredCronRouteBasePath
	DeclaredCronRegistrationInternalPath = contract.DeclaredCronRegistrationInternalPath
	DeclaredCronRegistrationRequestType  = contract.DeclaredCronRegistrationRequestType
	CronScopeMasterOnly                  = contract.CronScopeMasterOnly
	CronScopeAllNode                     = contract.CronScopeAllNode
	CronConcurrencySingleton             = contract.CronConcurrencySingleton
	CronConcurrencyParallel              = contract.CronConcurrencyParallel

	ExecutionSourceRoute         = contract.ExecutionSourceRoute
	ExecutionSourceHook          = contract.ExecutionSourceHook
	ExecutionSourceCron          = contract.ExecutionSourceCron
	ExecutionSourceCronDiscovery = contract.ExecutionSourceCronDiscovery
	ExecutionSourceLifecycle     = contract.ExecutionSourceLifecycle

	LifecycleOperationBeforeInstall           = contract.LifecycleOperationBeforeInstall
	LifecycleOperationAfterInstall            = contract.LifecycleOperationAfterInstall
	LifecycleOperationBeforeUpgrade           = contract.LifecycleOperationBeforeUpgrade
	LifecycleOperationAfterUpgrade            = contract.LifecycleOperationAfterUpgrade
	LifecycleOperationBeforeDisable           = contract.LifecycleOperationBeforeDisable
	LifecycleOperationAfterDisable            = contract.LifecycleOperationAfterDisable
	LifecycleOperationBeforeUninstall         = contract.LifecycleOperationBeforeUninstall
	LifecycleOperationAfterUninstall          = contract.LifecycleOperationAfterUninstall
	LifecycleOperationBeforeTenantDisable     = contract.LifecycleOperationBeforeTenantDisable
	LifecycleOperationAfterTenantDisable      = contract.LifecycleOperationAfterTenantDisable
	LifecycleOperationBeforeTenantDelete      = contract.LifecycleOperationBeforeTenantDelete
	LifecycleOperationAfterTenantDelete       = contract.LifecycleOperationAfterTenantDelete
	LifecycleOperationBeforeInstallModeChange = contract.LifecycleOperationBeforeInstallModeChange
	LifecycleOperationAfterInstallModeChange  = contract.LifecycleOperationAfterInstallModeChange

	GuestControllerHandlerKindEnvelope = guest.GuestControllerHandlerKindEnvelope
	GuestControllerHandlerKindTyped    = guest.GuestControllerHandlerKindTyped
)

// WASM artifact constants are kept on the root package for compatibility.
const (
	WasmSectionManifest            = artifact.WasmSectionManifest
	WasmSectionRuntime             = artifact.WasmSectionRuntime
	WasmSectionLegacyRuntime       = artifact.WasmSectionLegacyRuntime
	WasmSectionFrontendAssets      = artifact.WasmSectionFrontendAssets
	WasmSectionI18NAssets          = artifact.WasmSectionI18NAssets
	WasmSectionAPIDocI18NAssets    = artifact.WasmSectionAPIDocI18NAssets
	WasmSectionInstallSQL          = artifact.WasmSectionInstallSQL
	WasmSectionUninstallSQL        = artifact.WasmSectionUninstallSQL
	WasmSectionMockSQL             = artifact.WasmSectionMockSQL
	WasmSectionBackendHooks        = artifact.WasmSectionBackendHooks
	WasmSectionBackendLifecycle    = artifact.WasmSectionBackendLifecycle
	WasmSectionBackendResources    = artifact.WasmSectionBackendResources
	WasmSectionBackendCrons        = artifact.WasmSectionBackendCrons
	WasmSectionBackendRoutes       = artifact.WasmSectionBackendRoutes
	WasmSectionBackendBridge       = artifact.WasmSectionBackendBridge
	WasmSectionBackendCapabilities = artifact.WasmSectionBackendCapabilities
	WasmSectionBackendHostServices = artifact.WasmSectionBackendHostServices
)

// Host call constants are kept on the root package for compatibility.
const (
	HostModuleName                  = hostcall.HostModuleName
	HostCallFunctionName            = hostcall.HostCallFunctionName
	DefaultGuestHostCallAllocExport = hostcall.DefaultGuestHostCallAllocExport
	HostCallStatusSuccess           = hostcall.HostCallStatusSuccess
	HostCallStatusCapabilityDenied  = hostcall.HostCallStatusCapabilityDenied
	HostCallStatusNotFound          = hostcall.HostCallStatusNotFound
	HostCallStatusInvalidRequest    = hostcall.HostCallStatusInvalidRequest
	HostCallStatusInternalError     = hostcall.HostCallStatusInternalError
	OpcodeServiceInvoke             = hostcall.OpcodeServiceInvoke
	LogLevelDebug                   = hostcall.LogLevelDebug
	LogLevelInfo                    = hostcall.LogLevelInfo
	LogLevelWarning                 = hostcall.LogLevelWarning
	LogLevelError                   = hostcall.LogLevelError
)

// Host service constants are kept on the root package for compatibility.
const (
	CapabilityRuntime      = hostservice.CapabilityRuntime
	CapabilityCron         = hostservice.CapabilityCron
	CapabilityStorage      = hostservice.CapabilityStorage
	CapabilityHTTPRequest  = hostservice.CapabilityHTTPRequest
	CapabilityDataRead     = hostservice.CapabilityDataRead
	CapabilityDataMutate   = hostservice.CapabilityDataMutate
	CapabilityCache        = hostservice.CapabilityCache
	CapabilityLock         = hostservice.CapabilityLock
	CapabilitySecret       = hostservice.CapabilitySecret
	CapabilityEventPublish = hostservice.CapabilityEventPublish
	CapabilityQueueEnqueue = hostservice.CapabilityQueueEnqueue
	CapabilityNotify       = hostservice.CapabilityNotify
	CapabilityConfig       = hostservice.CapabilityConfig

	HostServiceRuntime = hostservice.HostServiceRuntime
	HostServiceCron    = hostservice.HostServiceCron
	HostServiceStorage = hostservice.HostServiceStorage
	HostServiceNetwork = hostservice.HostServiceNetwork
	HostServiceData    = hostservice.HostServiceData
	HostServiceCache   = hostservice.HostServiceCache
	HostServiceLock    = hostservice.HostServiceLock
	HostServiceSecret  = hostservice.HostServiceSecret
	HostServiceEvent   = hostservice.HostServiceEvent
	HostServiceQueue   = hostservice.HostServiceQueue
	HostServiceNotify  = hostservice.HostServiceNotify
	HostServiceConfig  = hostservice.HostServiceConfig

	HostServiceMethodRuntimeLogWrite    = hostservice.HostServiceMethodRuntimeLogWrite
	HostServiceMethodRuntimeStateGet    = hostservice.HostServiceMethodRuntimeStateGet
	HostServiceMethodRuntimeStateSet    = hostservice.HostServiceMethodRuntimeStateSet
	HostServiceMethodRuntimeStateDelete = hostservice.HostServiceMethodRuntimeStateDelete
	HostServiceMethodRuntimeInfoNow     = hostservice.HostServiceMethodRuntimeInfoNow
	HostServiceMethodRuntimeInfoUUID    = hostservice.HostServiceMethodRuntimeInfoUUID
	HostServiceMethodRuntimeInfoNode    = hostservice.HostServiceMethodRuntimeInfoNode
	HostServiceMethodCronRegister       = hostservice.HostServiceMethodCronRegister
	HostServiceMethodStoragePut         = hostservice.HostServiceMethodStoragePut
	HostServiceMethodStorageGet         = hostservice.HostServiceMethodStorageGet
	HostServiceMethodStorageDelete      = hostservice.HostServiceMethodStorageDelete
	HostServiceMethodStorageList        = hostservice.HostServiceMethodStorageList
	HostServiceMethodStorageStat        = hostservice.HostServiceMethodStorageStat
	HostServiceMethodNetworkRequest     = hostservice.HostServiceMethodNetworkRequest
	HostServiceMethodDataList           = hostservice.HostServiceMethodDataList
	HostServiceMethodDataGet            = hostservice.HostServiceMethodDataGet
	HostServiceMethodDataCreate         = hostservice.HostServiceMethodDataCreate
	HostServiceMethodDataUpdate         = hostservice.HostServiceMethodDataUpdate
	HostServiceMethodDataDelete         = hostservice.HostServiceMethodDataDelete
	HostServiceMethodDataTransaction    = hostservice.HostServiceMethodDataTransaction
	HostServiceMethodCacheGet           = hostservice.HostServiceMethodCacheGet
	HostServiceMethodCacheSet           = hostservice.HostServiceMethodCacheSet
	HostServiceMethodCacheDelete        = hostservice.HostServiceMethodCacheDelete
	HostServiceMethodCacheIncr          = hostservice.HostServiceMethodCacheIncr
	HostServiceMethodCacheExpire        = hostservice.HostServiceMethodCacheExpire
	HostServiceMethodLockAcquire        = hostservice.HostServiceMethodLockAcquire
	HostServiceMethodLockRenew          = hostservice.HostServiceMethodLockRenew
	HostServiceMethodLockRelease        = hostservice.HostServiceMethodLockRelease
	HostServiceMethodNotifySend         = hostservice.HostServiceMethodNotifySend
	HostServiceMethodConfigGet          = hostservice.HostServiceMethodConfigGet
	HostServiceMethodConfigExists       = hostservice.HostServiceMethodConfigExists
	HostServiceMethodConfigString       = hostservice.HostServiceMethodConfigString
	HostServiceMethodConfigBool         = hostservice.HostServiceMethodConfigBool
	HostServiceMethodConfigInt          = hostservice.HostServiceMethodConfigInt
	HostServiceMethodConfigDuration     = hostservice.HostServiceMethodConfigDuration

	HostServiceStorageVisibilityPrivate = hostservice.HostServiceStorageVisibilityPrivate
	HostServiceStorageVisibilityPublic  = hostservice.HostServiceStorageVisibilityPublic
	HostServiceCacheValueKindString     = hostservice.HostServiceCacheValueKindString
	HostServiceCacheValueKindInt        = hostservice.HostServiceCacheValueKindInt
)

// Bridge and artifact functions forward to focused subcomponents.
var (
	ValidateRouteContracts        = contract.ValidateRouteContracts
	NormalizeBridgeSpec           = contract.NormalizeBridgeSpec
	ValidateBridgeSpec            = contract.ValidateBridgeSpec
	NormalizeLifecycleContract    = contract.NormalizeLifecycleContract
	ValidateLifecycleContracts    = contract.ValidateLifecycleContracts
	IsSupportedLifecycleOperation = contract.IsSupportedLifecycleOperation
	NormalizeCronScope            = contract.NormalizeCronScope
	NormalizeCronConcurrency      = contract.NormalizeCronConcurrency
	NormalizeCronContract         = contract.NormalizeCronContract
	BuildPluginCronHandlerRef     = contract.BuildPluginCronHandlerRef
	BuildDeclaredCronRoutePath    = contract.BuildDeclaredCronRoutePath
	ValidateCronContracts         = contract.ValidateCronContracts
	NormalizeExecutionSource      = contract.NormalizeExecutionSource
	ReadCustomSection             = artifact.ReadCustomSection
	ListCustomSections            = artifact.ListCustomSections
	EncodeRequestEnvelope         = codec.EncodeRequestEnvelope
	DecodeRequestEnvelope         = codec.DecodeRequestEnvelope
	EncodeResponseEnvelope        = codec.EncodeResponseEnvelope
	DecodeResponseEnvelope        = codec.DecodeResponseEnvelope
	EncodeBodyBase64              = codec.EncodeBodyBase64
	NewSuccessResponse            = codec.NewSuccessResponse
	NewJSONResponse               = codec.NewJSONResponse
	NewFailureResponse            = codec.NewFailureResponse
	NewUnauthorizedResponse       = codec.NewUnauthorizedResponse
	NewForbiddenResponse          = codec.NewForbiddenResponse
	NewBadRequestResponse         = codec.NewBadRequestResponse
	NewNotFoundResponse           = codec.NewNotFoundResponse
	NewInternalErrorResponse      = codec.NewInternalErrorResponse
)

// Host call functions forward to the hostcall subcomponent.
var (
	MarshalHostCallResponse             = hostcall.MarshalHostCallResponse
	UnmarshalHostCallResponse           = hostcall.UnmarshalHostCallResponse
	NewHostCallSuccessResponse          = hostcall.NewHostCallSuccessResponse
	NewHostCallEmptySuccessResponse     = hostcall.NewHostCallEmptySuccessResponse
	NewHostCallErrorResponse            = hostcall.NewHostCallErrorResponse
	MarshalHostCallLogRequest           = hostcall.MarshalHostCallLogRequest
	UnmarshalHostCallLogRequest         = hostcall.UnmarshalHostCallLogRequest
	MarshalHostCallStateGetRequest      = hostcall.MarshalHostCallStateGetRequest
	UnmarshalHostCallStateGetRequest    = hostcall.UnmarshalHostCallStateGetRequest
	MarshalHostCallStateGetResponse     = hostcall.MarshalHostCallStateGetResponse
	UnmarshalHostCallStateGetResponse   = hostcall.UnmarshalHostCallStateGetResponse
	MarshalHostCallStateSetRequest      = hostcall.MarshalHostCallStateSetRequest
	UnmarshalHostCallStateSetRequest    = hostcall.UnmarshalHostCallStateSetRequest
	MarshalHostCallStateDeleteRequest   = hostcall.MarshalHostCallStateDeleteRequest
	UnmarshalHostCallStateDeleteRequest = hostcall.UnmarshalHostCallStateDeleteRequest
)

// Host service functions forward to the hostservice subcomponent.
var (
	RequiredCapabilityForHostServiceMethod      = hostservice.RequiredCapabilityForHostServiceMethod
	CapabilitiesFromHostServices                = hostservice.CapabilitiesFromHostServices
	CapabilityMapFromHostServices               = hostservice.CapabilityMapFromHostServices
	ValidateHostServiceSpecs                    = hostservice.ValidateHostServiceSpecs
	NormalizeHostServiceSpecs                   = hostservice.NormalizeHostServiceSpecs
	MustNormalizeHostServiceSpecs               = hostservice.MustNormalizeHostServiceSpecs
	AllCapabilities                             = hostservice.AllCapabilities
	ValidateCapabilities                        = hostservice.ValidateCapabilities
	NormalizeCapabilities                       = hostservice.NormalizeCapabilities
	CapabilitySliceToMap                        = hostservice.CapabilitySliceToMap
	MarshalHostServiceRequestEnvelope           = hostservice.MarshalHostServiceRequestEnvelope
	UnmarshalHostServiceRequestEnvelope         = hostservice.UnmarshalHostServiceRequestEnvelope
	MarshalHostServiceValueResponse             = hostservice.MarshalHostServiceValueResponse
	UnmarshalHostServiceValueResponse           = hostservice.UnmarshalHostServiceValueResponse
	MarshalHostServiceStoragePutRequest         = hostservice.MarshalHostServiceStoragePutRequest
	UnmarshalHostServiceStoragePutRequest       = hostservice.UnmarshalHostServiceStoragePutRequest
	MarshalHostServiceStoragePutResponse        = hostservice.MarshalHostServiceStoragePutResponse
	UnmarshalHostServiceStoragePutResponse      = hostservice.UnmarshalHostServiceStoragePutResponse
	MarshalHostServiceStorageGetRequest         = hostservice.MarshalHostServiceStorageGetRequest
	UnmarshalHostServiceStorageGetRequest       = hostservice.UnmarshalHostServiceStorageGetRequest
	MarshalHostServiceStorageGetResponse        = hostservice.MarshalHostServiceStorageGetResponse
	UnmarshalHostServiceStorageGetResponse      = hostservice.UnmarshalHostServiceStorageGetResponse
	MarshalHostServiceStorageDeleteRequest      = hostservice.MarshalHostServiceStorageDeleteRequest
	UnmarshalHostServiceStorageDeleteRequest    = hostservice.UnmarshalHostServiceStorageDeleteRequest
	MarshalHostServiceStorageListRequest        = hostservice.MarshalHostServiceStorageListRequest
	UnmarshalHostServiceStorageListRequest      = hostservice.UnmarshalHostServiceStorageListRequest
	MarshalHostServiceStorageListResponse       = hostservice.MarshalHostServiceStorageListResponse
	UnmarshalHostServiceStorageListResponse     = hostservice.UnmarshalHostServiceStorageListResponse
	MarshalHostServiceStorageStatRequest        = hostservice.MarshalHostServiceStorageStatRequest
	UnmarshalHostServiceStorageStatRequest      = hostservice.UnmarshalHostServiceStorageStatRequest
	MarshalHostServiceStorageStatResponse       = hostservice.MarshalHostServiceStorageStatResponse
	UnmarshalHostServiceStorageStatResponse     = hostservice.UnmarshalHostServiceStorageStatResponse
	MarshalHostServiceNetworkRequest            = hostservice.MarshalHostServiceNetworkRequest
	UnmarshalHostServiceNetworkRequest          = hostservice.UnmarshalHostServiceNetworkRequest
	MarshalHostServiceNetworkResponse           = hostservice.MarshalHostServiceNetworkResponse
	UnmarshalHostServiceNetworkResponse         = hostservice.UnmarshalHostServiceNetworkResponse
	MarshalHostServiceDataListRequest           = hostservice.MarshalHostServiceDataListRequest
	UnmarshalHostServiceDataListRequest         = hostservice.UnmarshalHostServiceDataListRequest
	MarshalHostServiceDataListResponse          = hostservice.MarshalHostServiceDataListResponse
	UnmarshalHostServiceDataListResponse        = hostservice.UnmarshalHostServiceDataListResponse
	MarshalHostServiceDataGetRequest            = hostservice.MarshalHostServiceDataGetRequest
	UnmarshalHostServiceDataGetRequest          = hostservice.UnmarshalHostServiceDataGetRequest
	MarshalHostServiceDataGetResponse           = hostservice.MarshalHostServiceDataGetResponse
	UnmarshalHostServiceDataGetResponse         = hostservice.UnmarshalHostServiceDataGetResponse
	MarshalHostServiceDataMutationRequest       = hostservice.MarshalHostServiceDataMutationRequest
	UnmarshalHostServiceDataMutationRequest     = hostservice.UnmarshalHostServiceDataMutationRequest
	MarshalHostServiceDataMutationResponse      = hostservice.MarshalHostServiceDataMutationResponse
	UnmarshalHostServiceDataMutationResponse    = hostservice.UnmarshalHostServiceDataMutationResponse
	MarshalHostServiceDataTransactionRequest    = hostservice.MarshalHostServiceDataTransactionRequest
	UnmarshalHostServiceDataTransactionRequest  = hostservice.UnmarshalHostServiceDataTransactionRequest
	MarshalHostServiceDataTransactionResponse   = hostservice.MarshalHostServiceDataTransactionResponse
	UnmarshalHostServiceDataTransactionResponse = hostservice.UnmarshalHostServiceDataTransactionResponse
	DecodeHostServiceDataListPlan               = hostservice.DecodeHostServiceDataListPlan
	DecodeHostServiceDataGetPlan                = hostservice.DecodeHostServiceDataGetPlan
	MarshalHostServiceCacheGetRequest           = hostservice.MarshalHostServiceCacheGetRequest
	UnmarshalHostServiceCacheGetRequest         = hostservice.UnmarshalHostServiceCacheGetRequest
	MarshalHostServiceCacheGetResponse          = hostservice.MarshalHostServiceCacheGetResponse
	UnmarshalHostServiceCacheGetResponse        = hostservice.UnmarshalHostServiceCacheGetResponse
	MarshalHostServiceCacheSetRequest           = hostservice.MarshalHostServiceCacheSetRequest
	UnmarshalHostServiceCacheSetRequest         = hostservice.UnmarshalHostServiceCacheSetRequest
	MarshalHostServiceCacheSetResponse          = hostservice.MarshalHostServiceCacheSetResponse
	UnmarshalHostServiceCacheSetResponse        = hostservice.UnmarshalHostServiceCacheSetResponse
	MarshalHostServiceCacheDeleteRequest        = hostservice.MarshalHostServiceCacheDeleteRequest
	UnmarshalHostServiceCacheDeleteRequest      = hostservice.UnmarshalHostServiceCacheDeleteRequest
	MarshalHostServiceCacheIncrRequest          = hostservice.MarshalHostServiceCacheIncrRequest
	UnmarshalHostServiceCacheIncrRequest        = hostservice.UnmarshalHostServiceCacheIncrRequest
	MarshalHostServiceCacheIncrResponse         = hostservice.MarshalHostServiceCacheIncrResponse
	UnmarshalHostServiceCacheIncrResponse       = hostservice.UnmarshalHostServiceCacheIncrResponse
	MarshalHostServiceCacheExpireRequest        = hostservice.MarshalHostServiceCacheExpireRequest
	UnmarshalHostServiceCacheExpireRequest      = hostservice.UnmarshalHostServiceCacheExpireRequest
	MarshalHostServiceCacheExpireResponse       = hostservice.MarshalHostServiceCacheExpireResponse
	UnmarshalHostServiceCacheExpireResponse     = hostservice.UnmarshalHostServiceCacheExpireResponse
	MarshalHostServiceLockAcquireRequest        = hostservice.MarshalHostServiceLockAcquireRequest
	UnmarshalHostServiceLockAcquireRequest      = hostservice.UnmarshalHostServiceLockAcquireRequest
	MarshalHostServiceLockAcquireResponse       = hostservice.MarshalHostServiceLockAcquireResponse
	UnmarshalHostServiceLockAcquireResponse     = hostservice.UnmarshalHostServiceLockAcquireResponse
	MarshalHostServiceLockRenewRequest          = hostservice.MarshalHostServiceLockRenewRequest
	UnmarshalHostServiceLockRenewRequest        = hostservice.UnmarshalHostServiceLockRenewRequest
	MarshalHostServiceLockRenewResponse         = hostservice.MarshalHostServiceLockRenewResponse
	UnmarshalHostServiceLockRenewResponse       = hostservice.UnmarshalHostServiceLockRenewResponse
	MarshalHostServiceLockReleaseRequest        = hostservice.MarshalHostServiceLockReleaseRequest
	UnmarshalHostServiceLockReleaseRequest      = hostservice.UnmarshalHostServiceLockReleaseRequest
	MarshalHostServiceConfigKeyRequest          = hostservice.MarshalHostServiceConfigKeyRequest
	UnmarshalHostServiceConfigKeyRequest        = hostservice.UnmarshalHostServiceConfigKeyRequest
	MarshalHostServiceConfigValueResponse       = hostservice.MarshalHostServiceConfigValueResponse
	UnmarshalHostServiceConfigValueResponse     = hostservice.UnmarshalHostServiceConfigValueResponse
	MarshalHostServiceNotifySendRequest         = hostservice.MarshalHostServiceNotifySendRequest
	UnmarshalHostServiceNotifySendRequest       = hostservice.UnmarshalHostServiceNotifySendRequest
	MarshalHostServiceNotifySendResponse        = hostservice.MarshalHostServiceNotifySendResponse
	UnmarshalHostServiceNotifySendResponse      = hostservice.UnmarshalHostServiceNotifySendResponse
	MarshalHostServiceCronRegisterRequest       = hostservice.MarshalHostServiceCronRegisterRequest
	UnmarshalHostServiceCronRegisterRequest     = hostservice.UnmarshalHostServiceCronRegisterRequest
)

// Guest SDK functions forward to the guest subcomponent.
var (
	NewGuestRuntime                       = guest.NewGuestRuntime
	NewGuestControllerRouteDispatcher     = guest.NewGuestControllerRouteDispatcher
	MustNewGuestControllerRouteDispatcher = guest.MustNewGuestControllerRouteDispatcher
	DiscoverGuestControllerHandlers       = guest.DiscoverGuestControllerHandlers
	BuildGuestControllerInternalPath      = guest.BuildGuestControllerInternalPath
	IsGuestBindJSONError                  = guest.IsGuestBindJSONError
	ClassifyBindJSONError                 = guest.ClassifyBindJSONError
	WriteJSON                             = guest.WriteJSON
	PathParam                             = guest.PathParam
	QueryValue                            = guest.QueryValue
	QueryInt                              = guest.QueryInt
	QueryFlag                             = guest.QueryFlag
	NewErrorCase                          = guest.NewErrorCase
	NewErrorClassifier                    = guest.NewErrorClassifier
	RequestEnvelopeFromContext            = guest.RequestEnvelopeFromContext
	SetResponseHeader                     = guest.SetResponseHeader
	SetResponseStatusCode                 = guest.SetResponseStatusCode
	WriteResponse                         = guest.WriteResponse
	WriteNoContent                        = guest.WriteNoContent
	NewResponseError                      = guest.NewResponseError
	ResponseFromError                     = guest.ResponseFromError
)

// BindJSON decodes the envelope request body into T through the guest SDK.
func BindJSON[T any](request *BridgeRequestEnvelopeV1) (*T, error) {
	return guest.BindJSON[T](request)
}
