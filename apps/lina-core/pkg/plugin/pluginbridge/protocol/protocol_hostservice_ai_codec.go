// This file implements transport codecs for governed AI host-service calls.
// The bridge owns only payload serialization; typed capability semantics remain
// in capability/ai subpackages.

package protocol

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/aicap/aiaudio"
	"lina-core/pkg/plugin/capability/aicap/aidocument"
	"lina-core/pkg/plugin/capability/aicap/aiembedding"
	"lina-core/pkg/plugin/capability/aicap/aiimage"
	"lina-core/pkg/plugin/capability/aicap/aisafety"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/aicap/aivideo"
	"lina-core/pkg/plugin/capability/aicap/aivision"
)

// HostServiceAITextGenerateRequest carries one text generation host-service request.
type HostServiceAITextGenerateRequest struct {
	// Purpose is the governed calling scenario passed to the AI capability.
	Purpose string `json:"purpose"`
	// Tier is the requested text AI tier; empty delegates defaulting to AI policy.
	Tier aitext.Tier `json:"tier,omitempty"`
	// Messages carries ordered plain-text generation context.
	Messages []aitext.Message `json:"messages"`
	// MaxOutputTokens optionally caps generated output tokens.
	MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
	// Temperature optionally controls sampling.
	Temperature *float64 `json:"temperature,omitempty"`
	// ThinkingEffort optionally requests abstract model reasoning effort.
	ThinkingEffort *aitext.ThinkingEffort `json:"thinkingEffort,omitempty"`
	// Metadata carries short audit keys and must not include prompt or response bodies.
	Metadata map[string]string `json:"metadata,omitempty"`
}

type (
	// HostServiceAIImageGenerateRequest carries one image generation host-service request.
	HostServiceAIImageGenerateRequest = aiimage.GenerateRequest
	// HostServiceAIImageEditRequest carries one image editing host-service request.
	HostServiceAIImageEditRequest = aiimage.EditRequest
	// HostServiceAIEmbeddingCreateRequest carries one embedding host-service request.
	HostServiceAIEmbeddingCreateRequest = aiembedding.CreateRequest
	// HostServiceAIAudioTranscribeRequest carries one audio transcription host-service request.
	HostServiceAIAudioTranscribeRequest = aiaudio.TranscribeRequest
	// HostServiceAIAudioSynthesizeRequest carries one audio synthesis host-service request.
	HostServiceAIAudioSynthesizeRequest = aiaudio.SynthesizeRequest
	// HostServiceAIVisionAnalyzeRequest carries one visual analysis host-service request.
	HostServiceAIVisionAnalyzeRequest = aivision.AnalyzeRequest
	// HostServiceAIDocumentAnalyzeRequest carries one document analysis host-service request.
	HostServiceAIDocumentAnalyzeRequest = aidocument.AnalyzeRequest
	// HostServiceAIDocumentCiteRequest carries one citation-aware document host-service request.
	HostServiceAIDocumentCiteRequest = aidocument.CiteRequest
	// HostServiceAISafetyModerateRequest carries one safety moderation host-service request.
	HostServiceAISafetyModerateRequest = aisafety.ModerateRequest
	// HostServiceAIVideoGenerateRequest carries one video generation host-service request.
	HostServiceAIVideoGenerateRequest = aivideo.GenerateRequest
	// HostServiceAIVideoEditRequest carries one video editing host-service request.
	HostServiceAIVideoEditRequest = aivideo.EditRequest
	// HostServiceAIVideoExtendRequest carries one video extension host-service request.
	HostServiceAIVideoExtendRequest = aivideo.ExtendRequest
	// HostServiceAIVideoOperationGetRequest carries one provider operation lookup request.
	HostServiceAIVideoOperationGetRequest = aivideo.OperationGetRequest
	// HostServiceAIVideoOperationCancelRequest carries one provider operation cancel request.
	HostServiceAIVideoOperationCancelRequest = aivideo.OperationCancelRequest
)

// MarshalHostServiceAITextGenerateRequest encodes one AI text request as JSON.
func MarshalHostServiceAITextGenerateRequest(req *HostServiceAITextGenerateRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAITextGenerateRequest decodes one AI text request.
func UnmarshalHostServiceAITextGenerateRequest(data []byte) (*HostServiceAITextGenerateRequest, error) {
	return unmarshalAIRequest[HostServiceAITextGenerateRequest](data, "ai text generate request is empty", "decode ai text generate request failed")
}

// MarshalHostServiceAIImageGenerateRequest encodes one AI image generate request as JSON.
func MarshalHostServiceAIImageGenerateRequest(req *HostServiceAIImageGenerateRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAIImageGenerateRequest decodes one AI image generate request.
func UnmarshalHostServiceAIImageGenerateRequest(data []byte) (*HostServiceAIImageGenerateRequest, error) {
	return unmarshalAIRequest[HostServiceAIImageGenerateRequest](data, "ai image generate request is empty", "decode ai image generate request failed")
}

// MarshalHostServiceAIImageEditRequest encodes one AI image edit request as JSON.
func MarshalHostServiceAIImageEditRequest(req *HostServiceAIImageEditRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAIImageEditRequest decodes one AI image edit request.
func UnmarshalHostServiceAIImageEditRequest(data []byte) (*HostServiceAIImageEditRequest, error) {
	return unmarshalAIRequest[HostServiceAIImageEditRequest](data, "ai image edit request is empty", "decode ai image edit request failed")
}

// MarshalHostServiceAIEmbeddingCreateRequest encodes one AI embedding request as JSON.
func MarshalHostServiceAIEmbeddingCreateRequest(req *HostServiceAIEmbeddingCreateRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAIEmbeddingCreateRequest decodes one AI embedding request.
func UnmarshalHostServiceAIEmbeddingCreateRequest(data []byte) (*HostServiceAIEmbeddingCreateRequest, error) {
	return unmarshalAIRequest[HostServiceAIEmbeddingCreateRequest](data, "ai embedding create request is empty", "decode ai embedding create request failed")
}

// MarshalHostServiceAIAudioTranscribeRequest encodes one AI audio transcribe request as JSON.
func MarshalHostServiceAIAudioTranscribeRequest(req *HostServiceAIAudioTranscribeRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAIAudioTranscribeRequest decodes one AI audio transcribe request.
func UnmarshalHostServiceAIAudioTranscribeRequest(data []byte) (*HostServiceAIAudioTranscribeRequest, error) {
	return unmarshalAIRequest[HostServiceAIAudioTranscribeRequest](data, "ai audio transcribe request is empty", "decode ai audio transcribe request failed")
}

// MarshalHostServiceAIAudioSynthesizeRequest encodes one AI audio synthesize request as JSON.
func MarshalHostServiceAIAudioSynthesizeRequest(req *HostServiceAIAudioSynthesizeRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAIAudioSynthesizeRequest decodes one AI audio synthesize request.
func UnmarshalHostServiceAIAudioSynthesizeRequest(data []byte) (*HostServiceAIAudioSynthesizeRequest, error) {
	return unmarshalAIRequest[HostServiceAIAudioSynthesizeRequest](data, "ai audio synthesize request is empty", "decode ai audio synthesize request failed")
}

// MarshalHostServiceAIVisionAnalyzeRequest encodes one AI vision analyze request as JSON.
func MarshalHostServiceAIVisionAnalyzeRequest(req *HostServiceAIVisionAnalyzeRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAIVisionAnalyzeRequest decodes one AI vision analyze request.
func UnmarshalHostServiceAIVisionAnalyzeRequest(data []byte) (*HostServiceAIVisionAnalyzeRequest, error) {
	return unmarshalAIRequest[HostServiceAIVisionAnalyzeRequest](data, "ai vision analyze request is empty", "decode ai vision analyze request failed")
}

// MarshalHostServiceAIDocumentAnalyzeRequest encodes one AI document analyze request as JSON.
func MarshalHostServiceAIDocumentAnalyzeRequest(req *HostServiceAIDocumentAnalyzeRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAIDocumentAnalyzeRequest decodes one AI document analyze request.
func UnmarshalHostServiceAIDocumentAnalyzeRequest(data []byte) (*HostServiceAIDocumentAnalyzeRequest, error) {
	return unmarshalAIRequest[HostServiceAIDocumentAnalyzeRequest](data, "ai document analyze request is empty", "decode ai document analyze request failed")
}

// MarshalHostServiceAIDocumentCiteRequest encodes one AI document cite request as JSON.
func MarshalHostServiceAIDocumentCiteRequest(req *HostServiceAIDocumentCiteRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAIDocumentCiteRequest decodes one AI document cite request.
func UnmarshalHostServiceAIDocumentCiteRequest(data []byte) (*HostServiceAIDocumentCiteRequest, error) {
	return unmarshalAIRequest[HostServiceAIDocumentCiteRequest](data, "ai document cite request is empty", "decode ai document cite request failed")
}

// MarshalHostServiceAISafetyModerateRequest encodes one AI safety moderate request as JSON.
func MarshalHostServiceAISafetyModerateRequest(req *HostServiceAISafetyModerateRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAISafetyModerateRequest decodes one AI safety moderate request.
func UnmarshalHostServiceAISafetyModerateRequest(data []byte) (*HostServiceAISafetyModerateRequest, error) {
	return unmarshalAIRequest[HostServiceAISafetyModerateRequest](data, "ai safety moderate request is empty", "decode ai safety moderate request failed")
}

// MarshalHostServiceAIVideoGenerateRequest encodes one AI video generate request as JSON.
func MarshalHostServiceAIVideoGenerateRequest(req *HostServiceAIVideoGenerateRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAIVideoGenerateRequest decodes one AI video generate request.
func UnmarshalHostServiceAIVideoGenerateRequest(data []byte) (*HostServiceAIVideoGenerateRequest, error) {
	return unmarshalAIRequest[HostServiceAIVideoGenerateRequest](data, "ai video generate request is empty", "decode ai video generate request failed")
}

// MarshalHostServiceAIVideoEditRequest encodes one AI video edit request as JSON.
func MarshalHostServiceAIVideoEditRequest(req *HostServiceAIVideoEditRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAIVideoEditRequest decodes one AI video edit request.
func UnmarshalHostServiceAIVideoEditRequest(data []byte) (*HostServiceAIVideoEditRequest, error) {
	return unmarshalAIRequest[HostServiceAIVideoEditRequest](data, "ai video edit request is empty", "decode ai video edit request failed")
}

// MarshalHostServiceAIVideoExtendRequest encodes one AI video extend request as JSON.
func MarshalHostServiceAIVideoExtendRequest(req *HostServiceAIVideoExtendRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAIVideoExtendRequest decodes one AI video extend request.
func UnmarshalHostServiceAIVideoExtendRequest(data []byte) (*HostServiceAIVideoExtendRequest, error) {
	return unmarshalAIRequest[HostServiceAIVideoExtendRequest](data, "ai video extend request is empty", "decode ai video extend request failed")
}

// MarshalHostServiceAIVideoOperationGetRequest encodes one AI video operation get request as JSON.
func MarshalHostServiceAIVideoOperationGetRequest(req *HostServiceAIVideoOperationGetRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAIVideoOperationGetRequest decodes one AI video operation get request.
func UnmarshalHostServiceAIVideoOperationGetRequest(data []byte) (*HostServiceAIVideoOperationGetRequest, error) {
	return unmarshalAIRequest[HostServiceAIVideoOperationGetRequest](data, "ai video operation get request is empty", "decode ai video operation get request failed")
}

// MarshalHostServiceAIVideoOperationCancelRequest encodes one AI video operation cancel request as JSON.
func MarshalHostServiceAIVideoOperationCancelRequest(req *HostServiceAIVideoOperationCancelRequest) []byte {
	return marshalAIRequest(req)
}

// UnmarshalHostServiceAIVideoOperationCancelRequest decodes one AI video operation cancel request.
func UnmarshalHostServiceAIVideoOperationCancelRequest(data []byte) (*HostServiceAIVideoOperationCancelRequest, error) {
	return unmarshalAIRequest[HostServiceAIVideoOperationCancelRequest](data, "ai video operation cancel request is empty", "decode ai video operation cancel request failed")
}

func marshalAIRequest(req any) []byte {
	if req == nil {
		return nil
	}
	content, err := json.Marshal(req)
	if err != nil {
		return nil
	}
	return content
}

func unmarshalAIRequest[T any](data []byte, emptyMessage string, wrapMessage string) (*T, error) {
	if len(data) == 0 {
		return nil, gerror.New(emptyMessage)
	}
	out := new(T)
	if err := json.Unmarshal(data, out); err != nil {
		return nil, gerror.Wrap(err, wrapMessage)
	}
	return out, nil
}
