// This file aliases lower-level bridge component contracts for the guest SDK.

package guest

import (
	"lina-core/pkg/pluginbridge/codec"
	"lina-core/pkg/pluginbridge/contract"
	"lina-core/pkg/pluginbridge/hostcall"
	"lina-core/pkg/pluginbridge/hostservice"
)

type (
	BridgeRequestEnvelopeV1  = contract.BridgeRequestEnvelopeV1
	BridgeResponseEnvelopeV1 = contract.BridgeResponseEnvelopeV1
	RouteMatchSnapshotV1     = contract.RouteMatchSnapshotV1
	HTTPRequestSnapshotV1    = contract.HTTPRequestSnapshotV1
	CronContract             = contract.CronContract

	HostCallLogRequest         = hostcall.HostCallLogRequest
	HostCallStateDeleteRequest = hostcall.HostCallStateDeleteRequest
	HostCallStateGetRequest    = hostcall.HostCallStateGetRequest
	HostCallStateGetResponse   = hostcall.HostCallStateGetResponse
	HostCallStateSetRequest    = hostcall.HostCallStateSetRequest

	HostServiceRequestEnvelope          = hostservice.HostServiceRequestEnvelope
	HostServiceValueResponse            = hostservice.HostServiceValueResponse
	HostServiceStorageObject            = hostservice.HostServiceStorageObject
	HostServiceStoragePutRequest        = hostservice.HostServiceStoragePutRequest
	HostServiceStoragePutResponse       = hostservice.HostServiceStoragePutResponse
	HostServiceStorageGetRequest        = hostservice.HostServiceStorageGetRequest
	HostServiceStorageGetResponse       = hostservice.HostServiceStorageGetResponse
	HostServiceStorageDeleteRequest     = hostservice.HostServiceStorageDeleteRequest
	HostServiceStorageListRequest       = hostservice.HostServiceStorageListRequest
	HostServiceStorageListResponse      = hostservice.HostServiceStorageListResponse
	HostServiceStorageStatRequest       = hostservice.HostServiceStorageStatRequest
	HostServiceStorageStatResponse      = hostservice.HostServiceStorageStatResponse
	HostServiceNetworkRequest           = hostservice.HostServiceNetworkRequest
	HostServiceNetworkResponse          = hostservice.HostServiceNetworkResponse
	HostServiceDataListRequest          = hostservice.HostServiceDataListRequest
	HostServiceDataListResponse         = hostservice.HostServiceDataListResponse
	HostServiceDataGetRequest           = hostservice.HostServiceDataGetRequest
	HostServiceDataMutationRequest      = hostservice.HostServiceDataMutationRequest
	HostServiceDataMutationResponse     = hostservice.HostServiceDataMutationResponse
	HostServiceDataTransactionRequest   = hostservice.HostServiceDataTransactionRequest
	HostServiceDataTransactionOperation = hostservice.HostServiceDataTransactionOperation
	HostServiceDataTransactionResponse  = hostservice.HostServiceDataTransactionResponse
	HostServiceCacheGetRequest          = hostservice.HostServiceCacheGetRequest
	HostServiceCacheGetResponse         = hostservice.HostServiceCacheGetResponse
	HostServiceCacheSetRequest          = hostservice.HostServiceCacheSetRequest
	HostServiceCacheSetResponse         = hostservice.HostServiceCacheSetResponse
	HostServiceCacheDeleteRequest       = hostservice.HostServiceCacheDeleteRequest
	HostServiceCacheIncrRequest         = hostservice.HostServiceCacheIncrRequest
	HostServiceCacheIncrResponse        = hostservice.HostServiceCacheIncrResponse
	HostServiceCacheExpireRequest       = hostservice.HostServiceCacheExpireRequest
	HostServiceCacheExpireResponse      = hostservice.HostServiceCacheExpireResponse
	HostServiceCacheValue               = hostservice.HostServiceCacheValue
	HostServiceLockAcquireRequest       = hostservice.HostServiceLockAcquireRequest
	HostServiceLockAcquireResponse      = hostservice.HostServiceLockAcquireResponse
	HostServiceLockRenewRequest         = hostservice.HostServiceLockRenewRequest
	HostServiceLockRenewResponse        = hostservice.HostServiceLockRenewResponse
	HostServiceLockReleaseRequest       = hostservice.HostServiceLockReleaseRequest
	HostServiceConfigKeyRequest         = hostservice.HostServiceConfigKeyRequest
	HostServiceConfigValueResponse      = hostservice.HostServiceConfigValueResponse
	HostServiceNotifySendRequest        = hostservice.HostServiceNotifySendRequest
	HostServiceNotifySendResponse       = hostservice.HostServiceNotifySendResponse
	HostServiceCronRegisterRequest      = hostservice.HostServiceCronRegisterRequest
)

const (
	OpcodeServiceInvoke                 = hostcall.OpcodeServiceInvoke
	HostCallStatusSuccess               = hostcall.HostCallStatusSuccess
	LogLevelDebug                       = hostcall.LogLevelDebug
	LogLevelInfo                        = hostcall.LogLevelInfo
	LogLevelWarning                     = hostcall.LogLevelWarning
	LogLevelError                       = hostcall.LogLevelError
	HostServiceRuntime                  = hostservice.HostServiceRuntime
	HostServiceCron                     = hostservice.HostServiceCron
	HostServiceStorage                  = hostservice.HostServiceStorage
	HostServiceNetwork                  = hostservice.HostServiceNetwork
	HostServiceData                     = hostservice.HostServiceData
	HostServiceCache                    = hostservice.HostServiceCache
	HostServiceLock                     = hostservice.HostServiceLock
	HostServiceNotify                   = hostservice.HostServiceNotify
	HostServiceConfig                   = hostservice.HostServiceConfig
	HostServiceMethodRuntimeLogWrite    = hostservice.HostServiceMethodRuntimeLogWrite
	HostServiceMethodRuntimeStateGet    = hostservice.HostServiceMethodRuntimeStateGet
	HostServiceMethodRuntimeStateSet    = hostservice.HostServiceMethodRuntimeStateSet
	HostServiceMethodRuntimeStateDelete = hostservice.HostServiceMethodRuntimeStateDelete
	HostServiceMethodRuntimeInfoNow     = hostservice.HostServiceMethodRuntimeInfoNow
	HostServiceMethodRuntimeInfoUUID    = hostservice.HostServiceMethodRuntimeInfoUUID
	HostServiceMethodRuntimeInfoNode    = hostservice.HostServiceMethodRuntimeInfoNode
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
	HostServiceMethodConfigGet          = hostservice.HostServiceMethodConfigGet
	HostServiceMethodConfigExists       = hostservice.HostServiceMethodConfigExists
	HostServiceMethodConfigString       = hostservice.HostServiceMethodConfigString
	HostServiceMethodConfigBool         = hostservice.HostServiceMethodConfigBool
	HostServiceMethodConfigInt          = hostservice.HostServiceMethodConfigInt
	HostServiceMethodConfigDuration     = hostservice.HostServiceMethodConfigDuration
	HostServiceMethodNotifySend         = hostservice.HostServiceMethodNotifySend
	HostServiceMethodCronRegister       = hostservice.HostServiceMethodCronRegister
)

var (
	EncodeRequestEnvelope                       = codec.EncodeRequestEnvelope
	DecodeRequestEnvelope                       = codec.DecodeRequestEnvelope
	EncodeResponseEnvelope                      = codec.EncodeResponseEnvelope
	DecodeResponseEnvelope                      = codec.DecodeResponseEnvelope
	NewSuccessResponse                          = codec.NewSuccessResponse
	NewJSONResponse                             = codec.NewJSONResponse
	NewFailureResponse                          = codec.NewFailureResponse
	NewUnauthorizedResponse                     = codec.NewUnauthorizedResponse
	NewForbiddenResponse                        = codec.NewForbiddenResponse
	NewBadRequestResponse                       = codec.NewBadRequestResponse
	NewNotFoundResponse                         = codec.NewNotFoundResponse
	NewInternalErrorResponse                    = codec.NewInternalErrorResponse
	MarshalHostCallResponse                     = hostcall.MarshalHostCallResponse
	UnmarshalHostCallResponse                   = hostcall.UnmarshalHostCallResponse
	MarshalHostCallLogRequest                   = hostcall.MarshalHostCallLogRequest
	MarshalHostCallStateGetRequest              = hostcall.MarshalHostCallStateGetRequest
	UnmarshalHostCallStateGetResponse           = hostcall.UnmarshalHostCallStateGetResponse
	MarshalHostCallStateSetRequest              = hostcall.MarshalHostCallStateSetRequest
	MarshalHostCallStateDeleteRequest           = hostcall.MarshalHostCallStateDeleteRequest
	MarshalHostServiceRequestEnvelope           = hostservice.MarshalHostServiceRequestEnvelope
	MarshalHostServiceValueResponse             = hostservice.MarshalHostServiceValueResponse
	UnmarshalHostServiceValueResponse           = hostservice.UnmarshalHostServiceValueResponse
	MarshalHostServiceStoragePutRequest         = hostservice.MarshalHostServiceStoragePutRequest
	UnmarshalHostServiceStoragePutResponse      = hostservice.UnmarshalHostServiceStoragePutResponse
	MarshalHostServiceStorageGetRequest         = hostservice.MarshalHostServiceStorageGetRequest
	UnmarshalHostServiceStorageGetResponse      = hostservice.UnmarshalHostServiceStorageGetResponse
	MarshalHostServiceStorageDeleteRequest      = hostservice.MarshalHostServiceStorageDeleteRequest
	MarshalHostServiceStorageListRequest        = hostservice.MarshalHostServiceStorageListRequest
	UnmarshalHostServiceStorageListResponse     = hostservice.UnmarshalHostServiceStorageListResponse
	MarshalHostServiceStorageStatRequest        = hostservice.MarshalHostServiceStorageStatRequest
	UnmarshalHostServiceStorageStatResponse     = hostservice.UnmarshalHostServiceStorageStatResponse
	MarshalHostServiceNetworkRequest            = hostservice.MarshalHostServiceNetworkRequest
	UnmarshalHostServiceNetworkResponse         = hostservice.UnmarshalHostServiceNetworkResponse
	MarshalHostServiceDataListRequest           = hostservice.MarshalHostServiceDataListRequest
	UnmarshalHostServiceDataListResponse        = hostservice.UnmarshalHostServiceDataListResponse
	MarshalHostServiceDataGetRequest            = hostservice.MarshalHostServiceDataGetRequest
	UnmarshalHostServiceDataGetResponse         = hostservice.UnmarshalHostServiceDataGetResponse
	MarshalHostServiceDataMutationRequest       = hostservice.MarshalHostServiceDataMutationRequest
	UnmarshalHostServiceDataMutationResponse    = hostservice.UnmarshalHostServiceDataMutationResponse
	MarshalHostServiceDataTransactionRequest    = hostservice.MarshalHostServiceDataTransactionRequest
	UnmarshalHostServiceDataTransactionResponse = hostservice.UnmarshalHostServiceDataTransactionResponse
	MarshalHostServiceCacheGetRequest           = hostservice.MarshalHostServiceCacheGetRequest
	UnmarshalHostServiceCacheGetResponse        = hostservice.UnmarshalHostServiceCacheGetResponse
	MarshalHostServiceCacheSetRequest           = hostservice.MarshalHostServiceCacheSetRequest
	UnmarshalHostServiceCacheSetResponse        = hostservice.UnmarshalHostServiceCacheSetResponse
	MarshalHostServiceCacheDeleteRequest        = hostservice.MarshalHostServiceCacheDeleteRequest
	MarshalHostServiceCacheIncrRequest          = hostservice.MarshalHostServiceCacheIncrRequest
	UnmarshalHostServiceCacheIncrResponse       = hostservice.UnmarshalHostServiceCacheIncrResponse
	MarshalHostServiceCacheExpireRequest        = hostservice.MarshalHostServiceCacheExpireRequest
	UnmarshalHostServiceCacheExpireResponse     = hostservice.UnmarshalHostServiceCacheExpireResponse
	MarshalHostServiceLockAcquireRequest        = hostservice.MarshalHostServiceLockAcquireRequest
	UnmarshalHostServiceLockAcquireResponse     = hostservice.UnmarshalHostServiceLockAcquireResponse
	MarshalHostServiceLockRenewRequest          = hostservice.MarshalHostServiceLockRenewRequest
	UnmarshalHostServiceLockRenewResponse       = hostservice.UnmarshalHostServiceLockRenewResponse
	MarshalHostServiceLockReleaseRequest        = hostservice.MarshalHostServiceLockReleaseRequest
	MarshalHostServiceConfigKeyRequest          = hostservice.MarshalHostServiceConfigKeyRequest
	UnmarshalHostServiceConfigValueResponse     = hostservice.UnmarshalHostServiceConfigValueResponse
	MarshalHostServiceNotifySendRequest         = hostservice.MarshalHostServiceNotifySendRequest
	UnmarshalHostServiceNotifySendResponse      = hostservice.UnmarshalHostServiceNotifySendResponse
	MarshalHostServiceCronRegisterRequest       = hostservice.MarshalHostServiceCronRegisterRequest
)
