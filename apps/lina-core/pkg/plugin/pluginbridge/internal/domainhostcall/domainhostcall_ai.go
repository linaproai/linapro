// This file implements guest-side AI capability calls that cross the
// pluginbridge host-service transport. The implementation stays internal so
// the public guest package can return the ordinary aicap.Service contract.

package domainhostcall

import (
	"context"
	"encoding/json"

	"lina-core/pkg/plugin/capability/aicap"
	"lina-core/pkg/plugin/capability/aicap/aiaudio"
	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/aicap/aidocument"
	"lina-core/pkg/plugin/capability/aicap/aiembedding"
	"lina-core/pkg/plugin/capability/aicap/aiimage"
	"lina-core/pkg/plugin/capability/aicap/aisafety"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/aicap/aivideo"
	"lina-core/pkg/plugin/capability/aicap/aivision"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

var _ aicap.Service = (*aiService)(nil)
var _ aitext.Service = (*aiTextService)(nil)
var _ aiimage.Service = (*aiImageService)(nil)
var _ aiembedding.Service = (*aiEmbeddingService)(nil)
var _ aiaudio.Service = (*aiAudioService)(nil)
var _ aivision.Service = (*aiVisionService)(nil)
var _ aidocument.Service = (*aiDocumentService)(nil)
var _ aisafety.Service = (*aiSafetyService)(nil)
var _ aivideo.Service = (*aiVideoService)(nil)

// aiService implements the guest-side AI namespace.
type aiService struct{ baseService }

// aiTextService implements guest text AI generation calls.
type aiTextService struct{ baseService }

// aiImageService implements guest image AI calls.
type aiImageService struct{ baseService }

// aiEmbeddingService implements guest embedding AI calls.
type aiEmbeddingService struct{ baseService }

// aiAudioService implements guest audio AI calls.
type aiAudioService struct{ baseService }

// aiVisionService implements guest vision AI calls.
type aiVisionService struct{ baseService }

// aiDocumentService implements guest document AI calls.
type aiDocumentService struct{ baseService }

// aiSafetyService implements guest safety AI calls.
type aiSafetyService struct{ baseService }

// aiVideoService implements guest video AI calls.
type aiVideoService struct{ baseService }

// AI creates the AI capability guest client.
func AI(invoker Invoker) aicap.Service {
	return aiService{baseService: newBaseService(invoker)}
}

// Text returns the governed text AI guest client.
func (s aiService) Text() aitext.Service {
	return aiTextService{baseService: s.baseService}
}

// Image returns the governed image AI guest client.
func (s aiService) Image() aiimage.Service { return aiImageService{baseService: s.baseService} }

// Embedding returns the governed embedding AI guest client.
func (s aiService) Embedding() aiembedding.Service {
	return aiEmbeddingService{baseService: s.baseService}
}

// Audio returns the governed audio AI guest client.
func (s aiService) Audio() aiaudio.Service { return aiAudioService{baseService: s.baseService} }

// Vision returns the governed vision AI guest client.
func (s aiService) Vision() aivision.Service { return aiVisionService{baseService: s.baseService} }

// Document returns the governed document AI guest client.
func (s aiService) Document() aidocument.Service {
	return aiDocumentService{baseService: s.baseService}
}

// Safety returns the governed safety AI guest client.
func (s aiService) Safety() aisafety.Service { return aiSafetyService{baseService: s.baseService} }

// Video returns the governed video AI guest client.
func (s aiService) Video() aivideo.Service { return aiVideoService{baseService: s.baseService} }

// MethodStatuses reads AI method availability across sub-capabilities.
func (s aiService) MethodStatuses(_ context.Context, input aicap.MethodStatusesInput) (*aicap.MethodStatusesResult, error) {
	var response aicap.MethodStatusesResult
	payload, err := marshalAIJSONRequest(input)
	if err != nil {
		return nil, err
	}
	err = s.invokeAIJSON(
		protocol.HostServiceMethodAIMethodStatuses,
		payload,
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Available reports that guest transport does not expose text provider status.
func (aiTextService) Available(context.Context) bool { return false }

// Status returns a safe text AI unavailable projection for guest-side status checks.
func (aiTextService) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(aitext.CapabilityAITextV1)
}

// MethodStatus reads one text AI method availability projection through host transport.
func (s aiTextService) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	var response aicommon.MethodStatus
	payload, err := marshalAIJSONRequest(aicap.MethodStatusQuery{
		CapabilityType:   aicommon.CapabilityTypeText,
		CapabilityMethod: method,
	})
	if err != nil {
		return aicommon.UnavailableMethodStatus(aitext.CapabilityAITextV1, aicommon.CapabilityTypeText, method)
	}
	err = s.invokeAIJSON(
		protocol.HostServiceMethodAITextMethodStatus,
		payload,
		&response,
	)
	if err != nil {
		return aicommon.UnavailableMethodStatus(aitext.CapabilityAITextV1, aicommon.CapabilityTypeText, method)
	}
	return response
}

// Available reports that guest transport does not expose image provider status.
func (aiImageService) Available(context.Context) bool { return false }

// Status returns a safe image AI unavailable projection for guest-side status checks.
func (aiImageService) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(aiimage.CapabilityAIImageV1)
}

// MethodStatus returns a safe image method status for guest-side status checks.
func (aiImageService) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(aiimage.CapabilityAIImageV1, aiimage.CapabilityType, method)
}

// Available reports that guest transport does not expose embedding provider status.
func (aiEmbeddingService) Available(context.Context) bool { return false }

// Status returns a safe embedding AI unavailable projection for guest-side status checks.
func (aiEmbeddingService) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(aiembedding.CapabilityAIEmbeddingV1)
}

// MethodStatus returns a safe embedding method status for guest-side status checks.
func (aiEmbeddingService) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(aiembedding.CapabilityAIEmbeddingV1, aiembedding.CapabilityType, method)
}

// Available reports that guest transport does not expose audio provider status.
func (aiAudioService) Available(context.Context) bool { return false }

// Status returns a safe audio AI unavailable projection for guest-side status checks.
func (aiAudioService) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(aiaudio.CapabilityAIAudioV1)
}

// MethodStatus returns a safe audio method status for guest-side status checks.
func (aiAudioService) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(aiaudio.CapabilityAIAudioV1, aiaudio.CapabilityType, method)
}

// Available reports that guest transport does not expose vision provider status.
func (aiVisionService) Available(context.Context) bool { return false }

// Status returns a safe vision AI unavailable projection for guest-side status checks.
func (aiVisionService) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(aivision.CapabilityAIVisionV1)
}

// MethodStatus returns a safe vision method status for guest-side status checks.
func (aiVisionService) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(aivision.CapabilityAIVisionV1, aivision.CapabilityType, method)
}

// Available reports that guest transport does not expose document provider status.
func (aiDocumentService) Available(context.Context) bool { return false }

// Status returns a safe document AI unavailable projection for guest-side status checks.
func (aiDocumentService) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(aidocument.CapabilityAIDocumentV1)
}

// MethodStatus returns a safe document method status for guest-side status checks.
func (aiDocumentService) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(aidocument.CapabilityAIDocumentV1, aidocument.CapabilityType, method)
}

// Available reports that guest transport does not expose safety provider status.
func (aiSafetyService) Available(context.Context) bool { return false }

// Status returns a safe safety AI unavailable projection for guest-side status checks.
func (aiSafetyService) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(aisafety.CapabilityAISafetyV1)
}

// MethodStatus returns a safe safety method status for guest-side status checks.
func (aiSafetyService) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(aisafety.CapabilityAISafetyV1, aisafety.CapabilityType, method)
}

// Available reports that guest transport does not expose video provider status.
func (aiVideoService) Available(context.Context) bool { return false }

// Status returns a safe video AI unavailable projection for guest-side status checks.
func (aiVideoService) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(aivideo.CapabilityAIVideoV1)
}

// MethodStatus returns a safe video method status for guest-side status checks.
func (aiVideoService) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(aivideo.CapabilityAIVideoV1, aivideo.CapabilityType, method)
}

// GenerateText executes one governed text generation call.
func (s aiTextService) GenerateText(_ context.Context, request aitext.GenerateRequest) (*aitext.GenerateResponse, error) {
	var response aitext.GenerateResponse
	payload := protocol.MarshalHostServiceAITextGenerateRequest(
		&protocol.HostServiceAITextGenerateRequest{
			Purpose:         request.Purpose,
			Tier:            request.Tier,
			Messages:        request.Messages,
			MaxOutputTokens: request.MaxOutputTokens,
			Temperature:     request.Temperature,
			ThinkingEffort:  request.ThinkingEffort,
			Metadata:        request.Metadata,
		},
	)
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAITextGenerate,
		payload,
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Generate executes one governed image generation call.
func (s aiImageService) Generate(_ context.Context, request aiimage.GenerateRequest) (*aiimage.Response, error) {
	var response aiimage.Response
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAIImageGenerate,
		protocol.MarshalHostServiceAIImageGenerateRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Edit executes one governed image editing call.
func (s aiImageService) Edit(_ context.Context, request aiimage.EditRequest) (*aiimage.Response, error) {
	var response aiimage.Response
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAIImageEdit,
		protocol.MarshalHostServiceAIImageEditRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Create executes one governed embedding creation call.
func (s aiEmbeddingService) Create(_ context.Context, request aiembedding.CreateRequest) (*aiembedding.CreateResponse, error) {
	var response aiembedding.CreateResponse
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAIEmbeddingCreate,
		protocol.MarshalHostServiceAIEmbeddingCreateRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Transcribe executes one governed audio transcription call.
func (s aiAudioService) Transcribe(_ context.Context, request aiaudio.TranscribeRequest) (*aiaudio.TranscribeResponse, error) {
	var response aiaudio.TranscribeResponse
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAIAudioTranscribe,
		protocol.MarshalHostServiceAIAudioTranscribeRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Synthesize executes one governed audio synthesis call.
func (s aiAudioService) Synthesize(_ context.Context, request aiaudio.SynthesizeRequest) (*aiaudio.SynthesizeResponse, error) {
	var response aiaudio.SynthesizeResponse
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAIAudioSynthesize,
		protocol.MarshalHostServiceAIAudioSynthesizeRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Analyze executes one governed visual analysis call.
func (s aiVisionService) Analyze(_ context.Context, request aivision.AnalyzeRequest) (*aivision.AnalyzeResponse, error) {
	var response aivision.AnalyzeResponse
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAIVisionAnalyze,
		protocol.MarshalHostServiceAIVisionAnalyzeRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Analyze executes one governed document analysis call.
func (s aiDocumentService) Analyze(_ context.Context, request aidocument.AnalyzeRequest) (*aidocument.Response, error) {
	var response aidocument.Response
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAIDocumentAnalyze,
		protocol.MarshalHostServiceAIDocumentAnalyzeRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Cite executes one governed citation-aware document call.
func (s aiDocumentService) Cite(_ context.Context, request aidocument.CiteRequest) (*aidocument.Response, error) {
	var response aidocument.Response
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAIDocumentCite,
		protocol.MarshalHostServiceAIDocumentCiteRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Moderate executes one governed safety moderation call.
func (s aiSafetyService) Moderate(_ context.Context, request aisafety.ModerateRequest) (*aisafety.ModerateResponse, error) {
	var response aisafety.ModerateResponse
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAISafetyModerate,
		protocol.MarshalHostServiceAISafetyModerateRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Generate executes one governed video generation call.
func (s aiVideoService) Generate(_ context.Context, request aivideo.GenerateRequest) (*aivideo.Response, error) {
	var response aivideo.Response
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAIVideoGenerate,
		protocol.MarshalHostServiceAIVideoGenerateRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Edit executes one governed video editing call.
func (s aiVideoService) Edit(_ context.Context, request aivideo.EditRequest) (*aivideo.Response, error) {
	var response aivideo.Response
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAIVideoEdit,
		protocol.MarshalHostServiceAIVideoEditRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Extend executes one governed video extension call.
func (s aiVideoService) Extend(_ context.Context, request aivideo.ExtendRequest) (*aivideo.Response, error) {
	var response aivideo.Response
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAIVideoExtend,
		protocol.MarshalHostServiceAIVideoExtendRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// OperationGet reads one provider operation projection.
func (s aiVideoService) OperationGet(_ context.Context, request aivideo.OperationGetRequest) (*aivideo.Response, error) {
	var response aivideo.Response
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAIVideoOperationGet,
		protocol.MarshalHostServiceAIVideoOperationGetRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// OperationCancel cancels one provider operation when authorized and supported.
func (s aiVideoService) OperationCancel(_ context.Context, request aivideo.OperationCancelRequest) (*aicommon.ProviderOperationRef, error) {
	var response aicommon.ProviderOperationRef
	err := s.invokeAIJSON(
		protocol.HostServiceMethodAIVideoOperationCancel,
		protocol.MarshalHostServiceAIVideoOperationCancelRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// invokeAIJSON dispatches one AI host-service request through the injected invoker.
func (s baseService) invokeAIJSON(method string, payload []byte, out any) error {
	return s.call(protocol.HostServiceAI, method, payload, out)
}

func marshalAIJSONRequest(value any) ([]byte, error) {
	content, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return protocol.MarshalHostServiceJSONRequest(&protocol.HostServiceJSONRequest{Value: content}), nil
}
