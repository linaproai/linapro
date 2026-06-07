// hostservice_contract.go exposes host service names, method names, capabilities, and validation helpers.
// Keep this file limited to service contract metadata; payload codecs live in the per-domain hostservice codec files.

package protocol

import "lina-core/pkg/plugin/pluginbridge/internal/hostservice"

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
	CapabilityHostConfig   = hostservice.CapabilityHostConfig
	CapabilityManifest     = hostservice.CapabilityManifest
	CapabilityAIText       = hostservice.CapabilityAIText
	CapabilityAIImage      = hostservice.CapabilityAIImage
	CapabilityAIEmbedding  = hostservice.CapabilityAIEmbedding
	CapabilityAIAudio      = hostservice.CapabilityAIAudio
	CapabilityAIVision     = hostservice.CapabilityAIVision
	CapabilityAIDocument   = hostservice.CapabilityAIDocument
	CapabilityAISafety     = hostservice.CapabilityAISafety
	CapabilityAIVideo      = hostservice.CapabilityAIVideo
	CapabilityOrg          = hostservice.CapabilityOrg
	CapabilityTenant       = hostservice.CapabilityTenant

	HostServiceRuntime    = hostservice.HostServiceRuntime
	HostServiceCron       = hostservice.HostServiceCron
	HostServiceStorage    = hostservice.HostServiceStorage
	HostServiceNetwork    = hostservice.HostServiceNetwork
	HostServiceData       = hostservice.HostServiceData
	HostServiceCache      = hostservice.HostServiceCache
	HostServiceLock       = hostservice.HostServiceLock
	HostServiceSecret     = hostservice.HostServiceSecret
	HostServiceEvent      = hostservice.HostServiceEvent
	HostServiceQueue      = hostservice.HostServiceQueue
	HostServiceNotify     = hostservice.HostServiceNotify
	HostServiceConfig     = hostservice.HostServiceConfig
	HostServiceHostConfig = hostservice.HostServiceHostConfig
	HostServiceManifest   = hostservice.HostServiceManifest
	HostServiceAI         = hostservice.HostServiceAI
	HostServiceOrg        = hostservice.HostServiceOrg
	HostServiceTenant     = hostservice.HostServiceTenant

	HostServiceMethodRuntimeLogWrite            = hostservice.HostServiceMethodRuntimeLogWrite
	HostServiceMethodRuntimeStateGet            = hostservice.HostServiceMethodRuntimeStateGet
	HostServiceMethodRuntimeStateSet            = hostservice.HostServiceMethodRuntimeStateSet
	HostServiceMethodRuntimeStateDelete         = hostservice.HostServiceMethodRuntimeStateDelete
	HostServiceMethodRuntimeInfoNow             = hostservice.HostServiceMethodRuntimeInfoNow
	HostServiceMethodRuntimeInfoUUID            = hostservice.HostServiceMethodRuntimeInfoUUID
	HostServiceMethodRuntimeInfoNode            = hostservice.HostServiceMethodRuntimeInfoNode
	HostServiceMethodCronRegister               = hostservice.HostServiceMethodCronRegister
	HostServiceMethodStoragePut                 = hostservice.HostServiceMethodStoragePut
	HostServiceMethodStorageGet                 = hostservice.HostServiceMethodStorageGet
	HostServiceMethodStorageDelete              = hostservice.HostServiceMethodStorageDelete
	HostServiceMethodStorageList                = hostservice.HostServiceMethodStorageList
	HostServiceMethodStorageStat                = hostservice.HostServiceMethodStorageStat
	HostServiceMethodNetworkRequest             = hostservice.HostServiceMethodNetworkRequest
	HostServiceMethodDataList                   = hostservice.HostServiceMethodDataList
	HostServiceMethodDataGet                    = hostservice.HostServiceMethodDataGet
	HostServiceMethodDataCreate                 = hostservice.HostServiceMethodDataCreate
	HostServiceMethodDataUpdate                 = hostservice.HostServiceMethodDataUpdate
	HostServiceMethodDataDelete                 = hostservice.HostServiceMethodDataDelete
	HostServiceMethodDataTransaction            = hostservice.HostServiceMethodDataTransaction
	HostServiceMethodCacheGet                   = hostservice.HostServiceMethodCacheGet
	HostServiceMethodCacheSet                   = hostservice.HostServiceMethodCacheSet
	HostServiceMethodCacheDelete                = hostservice.HostServiceMethodCacheDelete
	HostServiceMethodCacheIncr                  = hostservice.HostServiceMethodCacheIncr
	HostServiceMethodCacheExpire                = hostservice.HostServiceMethodCacheExpire
	HostServiceMethodLockAcquire                = hostservice.HostServiceMethodLockAcquire
	HostServiceMethodLockRenew                  = hostservice.HostServiceMethodLockRenew
	HostServiceMethodLockRelease                = hostservice.HostServiceMethodLockRelease
	HostServiceMethodNotifySend                 = hostservice.HostServiceMethodNotifySend
	HostServiceMethodConfigGet                  = hostservice.HostServiceMethodConfigGet
	HostServiceMethodConfigExists               = hostservice.HostServiceMethodConfigExists
	HostServiceMethodConfigString               = hostservice.HostServiceMethodConfigString
	HostServiceMethodConfigBool                 = hostservice.HostServiceMethodConfigBool
	HostServiceMethodConfigInt                  = hostservice.HostServiceMethodConfigInt
	HostServiceMethodConfigDuration             = hostservice.HostServiceMethodConfigDuration
	HostServiceMethodHostConfigGet              = hostservice.HostServiceMethodHostConfigGet
	HostServiceMethodManifestGet                = hostservice.HostServiceMethodManifestGet
	HostServiceMethodAITextGenerate             = hostservice.HostServiceMethodAITextGenerate
	HostServiceMethodAIImageGenerate            = hostservice.HostServiceMethodAIImageGenerate
	HostServiceMethodAIImageEdit                = hostservice.HostServiceMethodAIImageEdit
	HostServiceMethodAIEmbeddingCreate          = hostservice.HostServiceMethodAIEmbeddingCreate
	HostServiceMethodAIAudioTranscribe          = hostservice.HostServiceMethodAIAudioTranscribe
	HostServiceMethodAIAudioSynthesize          = hostservice.HostServiceMethodAIAudioSynthesize
	HostServiceMethodAIVisionAnalyze            = hostservice.HostServiceMethodAIVisionAnalyze
	HostServiceMethodAIDocumentAnalyze          = hostservice.HostServiceMethodAIDocumentAnalyze
	HostServiceMethodAIDocumentCite             = hostservice.HostServiceMethodAIDocumentCite
	HostServiceMethodAISafetyModerate           = hostservice.HostServiceMethodAISafetyModerate
	HostServiceMethodAIVideoGenerate            = hostservice.HostServiceMethodAIVideoGenerate
	HostServiceMethodAIVideoEdit                = hostservice.HostServiceMethodAIVideoEdit
	HostServiceMethodAIVideoExtend              = hostservice.HostServiceMethodAIVideoExtend
	HostServiceMethodAIVideoOperationGet        = hostservice.HostServiceMethodAIVideoOperationGet
	HostServiceMethodAIVideoOperationCancel     = hostservice.HostServiceMethodAIVideoOperationCancel
	HostServiceMethodOrgAvailable               = hostservice.HostServiceMethodOrgAvailable
	HostServiceMethodOrgStatus                  = hostservice.HostServiceMethodOrgStatus
	HostServiceMethodOrgListUserDeptAssignments = hostservice.HostServiceMethodOrgListUserDeptAssignments
	HostServiceMethodOrgGetUserDeptInfo         = hostservice.HostServiceMethodOrgGetUserDeptInfo
	HostServiceMethodOrgGetUserDeptName         = hostservice.HostServiceMethodOrgGetUserDeptName
	HostServiceMethodOrgGetUserDeptIDs          = hostservice.HostServiceMethodOrgGetUserDeptIDs
	HostServiceMethodOrgGetUserPostIDs          = hostservice.HostServiceMethodOrgGetUserPostIDs
	HostServiceMethodTenantAvailable            = hostservice.HostServiceMethodTenantAvailable
	HostServiceMethodTenantStatus               = hostservice.HostServiceMethodTenantStatus
	HostServiceMethodTenantCurrent              = hostservice.HostServiceMethodTenantCurrent
	HostServiceMethodTenantPlatformBypass       = hostservice.HostServiceMethodTenantPlatformBypass
	HostServiceMethodTenantEnsureVisible        = hostservice.HostServiceMethodTenantEnsureVisible
	HostServiceMethodTenantValidateUserInTenant = hostservice.HostServiceMethodTenantValidateUserInTenant
	HostServiceMethodTenantListUserTenants      = hostservice.HostServiceMethodTenantListUserTenants
	HostServiceMethodTenantValidateSwitch       = hostservice.HostServiceMethodTenantValidateSwitch

	HostServiceStorageVisibilityPrivate = hostservice.HostServiceStorageVisibilityPrivate
	HostServiceStorageVisibilityPublic  = hostservice.HostServiceStorageVisibilityPublic
	HostServiceCacheValueKindString     = hostservice.HostServiceCacheValueKindString
	HostServiceCacheValueKindInt        = hostservice.HostServiceCacheValueKindInt
)

var (
	RequiredCapabilityForHostServiceMethod = hostservice.RequiredCapabilityForHostServiceMethod
	CapabilitiesFromHostServices           = hostservice.CapabilitiesFromHostServices
	CapabilityMapFromHostServices          = hostservice.CapabilityMapFromHostServices
	ValidateHostServiceSpecs               = hostservice.ValidateHostServiceSpecs
	ValidateHostServiceSpecsForPlugin      = hostservice.ValidateHostServiceSpecsForPlugin
	NormalizeHostServiceSpecs              = hostservice.NormalizeHostServiceSpecs
	NormalizeHostServiceSpecsForPlugin     = hostservice.NormalizeHostServiceSpecsForPlugin
	MustNormalizeHostServiceSpecs          = hostservice.MustNormalizeHostServiceSpecs
	MustNormalizeHostServiceSpecsForPlugin = hostservice.MustNormalizeHostServiceSpecsForPlugin
	AllCapabilities                        = hostservice.AllCapabilities
	ValidateCapabilities                   = hostservice.ValidateCapabilities
	NormalizeCapabilities                  = hostservice.NormalizeCapabilities
	CapabilitySliceToMap                   = hostservice.CapabilitySliceToMap
)
