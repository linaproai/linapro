// hostservice_types.go exposes host service payload type aliases through the public protocol facade.
// These aliases preserve the public protocol surface while concrete payload ownership remains in the internal hostservice package.

package protocol

import "lina-core/pkg/plugin/pluginbridge/internal/hostservice"

type (
	HostServiceAITextGenerateRequest             = hostservice.HostServiceAITextGenerateRequest
	HostServiceAIImageGenerateRequest            = hostservice.HostServiceAIImageGenerateRequest
	HostServiceAIImageEditRequest                = hostservice.HostServiceAIImageEditRequest
	HostServiceAIEmbeddingCreateRequest          = hostservice.HostServiceAIEmbeddingCreateRequest
	HostServiceAIAudioTranscribeRequest          = hostservice.HostServiceAIAudioTranscribeRequest
	HostServiceAIAudioSynthesizeRequest          = hostservice.HostServiceAIAudioSynthesizeRequest
	HostServiceAIVisionAnalyzeRequest            = hostservice.HostServiceAIVisionAnalyzeRequest
	HostServiceAIDocumentAnalyzeRequest          = hostservice.HostServiceAIDocumentAnalyzeRequest
	HostServiceAIDocumentCiteRequest             = hostservice.HostServiceAIDocumentCiteRequest
	HostServiceAISafetyModerateRequest           = hostservice.HostServiceAISafetyModerateRequest
	HostServiceAIVideoGenerateRequest            = hostservice.HostServiceAIVideoGenerateRequest
	HostServiceAIVideoEditRequest                = hostservice.HostServiceAIVideoEditRequest
	HostServiceAIVideoExtendRequest              = hostservice.HostServiceAIVideoExtendRequest
	HostServiceAIVideoOperationGetRequest        = hostservice.HostServiceAIVideoOperationGetRequest
	HostServiceAIVideoOperationCancelRequest     = hostservice.HostServiceAIVideoOperationCancelRequest
	HostServiceCacheDeleteRequest                = hostservice.HostServiceCacheDeleteRequest
	HostServiceCacheExpireRequest                = hostservice.HostServiceCacheExpireRequest
	HostServiceCacheExpireResponse               = hostservice.HostServiceCacheExpireResponse
	HostServiceCacheGetRequest                   = hostservice.HostServiceCacheGetRequest
	HostServiceCacheGetResponse                  = hostservice.HostServiceCacheGetResponse
	HostServiceCacheIncrRequest                  = hostservice.HostServiceCacheIncrRequest
	HostServiceCacheIncrResponse                 = hostservice.HostServiceCacheIncrResponse
	HostServiceCacheSetRequest                   = hostservice.HostServiceCacheSetRequest
	HostServiceCacheSetResponse                  = hostservice.HostServiceCacheSetResponse
	HostServiceCacheValue                        = hostservice.HostServiceCacheValue
	HostServiceCapabilityJSONResponse            = hostservice.HostServiceCapabilityJSONResponse
	HostServiceCapabilityTenantRequest           = hostservice.HostServiceCapabilityTenantRequest
	HostServiceCapabilityUserRequest             = hostservice.HostServiceCapabilityUserRequest
	HostServiceCapabilityUsersRequest            = hostservice.HostServiceCapabilityUsersRequest
	HostServiceCapabilityUserTenantRequest       = hostservice.HostServiceCapabilityUserTenantRequest
	HostServiceCapabilityUserTenantSwitchRequest = hostservice.HostServiceCapabilityUserTenantSwitchRequest
	HostServiceConfigKeyRequest                  = hostservice.HostServiceConfigKeyRequest
	HostServiceConfigValueResponse               = hostservice.HostServiceConfigValueResponse
	HostServiceCronRegisterRequest               = hostservice.HostServiceCronRegisterRequest
	HostServiceDataGetRequest                    = hostservice.HostServiceDataGetRequest
	HostServiceDataGetResponse                   = hostservice.HostServiceDataGetResponse
	HostServiceDataListRequest                   = hostservice.HostServiceDataListRequest
	HostServiceDataListResponse                  = hostservice.HostServiceDataListResponse
	HostServiceDataMutationRequest               = hostservice.HostServiceDataMutationRequest
	HostServiceDataMutationResponse              = hostservice.HostServiceDataMutationResponse
	HostServiceDataTransactionOperation          = hostservice.HostServiceDataTransactionOperation
	HostServiceDataTransactionRequest            = hostservice.HostServiceDataTransactionRequest
	HostServiceDataTransactionResponse           = hostservice.HostServiceDataTransactionResponse
	HostServiceLockAcquireRequest                = hostservice.HostServiceLockAcquireRequest
	HostServiceLockAcquireResponse               = hostservice.HostServiceLockAcquireResponse
	HostServiceLockReleaseRequest                = hostservice.HostServiceLockReleaseRequest
	HostServiceLockRenewRequest                  = hostservice.HostServiceLockRenewRequest
	HostServiceLockRenewResponse                 = hostservice.HostServiceLockRenewResponse
	HostServiceManifestGetRequest                = hostservice.HostServiceManifestGetRequest
	HostServiceManifestGetResponse               = hostservice.HostServiceManifestGetResponse
	HostServiceNetworkRequest                    = hostservice.HostServiceNetworkRequest
	HostServiceNetworkResponse                   = hostservice.HostServiceNetworkResponse
	HostServiceNotifySendRequest                 = hostservice.HostServiceNotifySendRequest
	HostServiceNotifySendResponse                = hostservice.HostServiceNotifySendResponse
	HostServiceRequestEnvelope                   = hostservice.HostServiceRequestEnvelope
	HostServiceResourceSpec                      = hostservice.HostServiceResourceSpec
	HostServiceSpec                              = hostservice.HostServiceSpec
	HostServiceStorageDeleteRequest              = hostservice.HostServiceStorageDeleteRequest
	HostServiceStorageGetRequest                 = hostservice.HostServiceStorageGetRequest
	HostServiceStorageGetResponse                = hostservice.HostServiceStorageGetResponse
	HostServiceStorageListRequest                = hostservice.HostServiceStorageListRequest
	HostServiceStorageListResponse               = hostservice.HostServiceStorageListResponse
	HostServiceStorageObject                     = hostservice.HostServiceStorageObject
	HostServiceStoragePutRequest                 = hostservice.HostServiceStoragePutRequest
	HostServiceStoragePutResponse                = hostservice.HostServiceStoragePutResponse
	HostServiceStorageStatRequest                = hostservice.HostServiceStorageStatRequest
	HostServiceStorageStatResponse               = hostservice.HostServiceStorageStatResponse
	HostServiceValueResponse                     = hostservice.HostServiceValueResponse
)
