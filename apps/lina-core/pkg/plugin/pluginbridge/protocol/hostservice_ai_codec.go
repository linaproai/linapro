// hostservice_ai_codec.go exposes AI host-service payload codecs through the
// public protocol facade.

package protocol

import "lina-core/pkg/plugin/pluginbridge/internal/hostservice"

var (
	MarshalHostServiceAITextGenerateRequest           = hostservice.MarshalHostServiceAITextGenerateRequest
	UnmarshalHostServiceAITextGenerateRequest         = hostservice.UnmarshalHostServiceAITextGenerateRequest
	MarshalHostServiceAIImageGenerateRequest          = hostservice.MarshalHostServiceAIImageGenerateRequest
	UnmarshalHostServiceAIImageGenerateRequest        = hostservice.UnmarshalHostServiceAIImageGenerateRequest
	MarshalHostServiceAIImageEditRequest              = hostservice.MarshalHostServiceAIImageEditRequest
	UnmarshalHostServiceAIImageEditRequest            = hostservice.UnmarshalHostServiceAIImageEditRequest
	MarshalHostServiceAIEmbeddingCreateRequest        = hostservice.MarshalHostServiceAIEmbeddingCreateRequest
	UnmarshalHostServiceAIEmbeddingCreateRequest      = hostservice.UnmarshalHostServiceAIEmbeddingCreateRequest
	MarshalHostServiceAIAudioTranscribeRequest        = hostservice.MarshalHostServiceAIAudioTranscribeRequest
	UnmarshalHostServiceAIAudioTranscribeRequest      = hostservice.UnmarshalHostServiceAIAudioTranscribeRequest
	MarshalHostServiceAIAudioSynthesizeRequest        = hostservice.MarshalHostServiceAIAudioSynthesizeRequest
	UnmarshalHostServiceAIAudioSynthesizeRequest      = hostservice.UnmarshalHostServiceAIAudioSynthesizeRequest
	MarshalHostServiceAIVisionAnalyzeRequest          = hostservice.MarshalHostServiceAIVisionAnalyzeRequest
	UnmarshalHostServiceAIVisionAnalyzeRequest        = hostservice.UnmarshalHostServiceAIVisionAnalyzeRequest
	MarshalHostServiceAIDocumentAnalyzeRequest        = hostservice.MarshalHostServiceAIDocumentAnalyzeRequest
	UnmarshalHostServiceAIDocumentAnalyzeRequest      = hostservice.UnmarshalHostServiceAIDocumentAnalyzeRequest
	MarshalHostServiceAIDocumentCiteRequest           = hostservice.MarshalHostServiceAIDocumentCiteRequest
	UnmarshalHostServiceAIDocumentCiteRequest         = hostservice.UnmarshalHostServiceAIDocumentCiteRequest
	MarshalHostServiceAISafetyModerateRequest         = hostservice.MarshalHostServiceAISafetyModerateRequest
	UnmarshalHostServiceAISafetyModerateRequest       = hostservice.UnmarshalHostServiceAISafetyModerateRequest
	MarshalHostServiceAIVideoGenerateRequest          = hostservice.MarshalHostServiceAIVideoGenerateRequest
	UnmarshalHostServiceAIVideoGenerateRequest        = hostservice.UnmarshalHostServiceAIVideoGenerateRequest
	MarshalHostServiceAIVideoEditRequest              = hostservice.MarshalHostServiceAIVideoEditRequest
	UnmarshalHostServiceAIVideoEditRequest            = hostservice.UnmarshalHostServiceAIVideoEditRequest
	MarshalHostServiceAIVideoExtendRequest            = hostservice.MarshalHostServiceAIVideoExtendRequest
	UnmarshalHostServiceAIVideoExtendRequest          = hostservice.UnmarshalHostServiceAIVideoExtendRequest
	MarshalHostServiceAIVideoOperationGetRequest      = hostservice.MarshalHostServiceAIVideoOperationGetRequest
	UnmarshalHostServiceAIVideoOperationGetRequest    = hostservice.UnmarshalHostServiceAIVideoOperationGetRequest
	MarshalHostServiceAIVideoOperationCancelRequest   = hostservice.MarshalHostServiceAIVideoOperationCancelRequest
	UnmarshalHostServiceAIVideoOperationCancelRequest = hostservice.UnmarshalHostServiceAIVideoOperationCancelRequest
)
