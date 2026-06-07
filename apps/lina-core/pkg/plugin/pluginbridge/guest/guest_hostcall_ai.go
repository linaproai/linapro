// This file implements guest-side text AI capability calls that cross the
// pluginbridge host-service transport. Purpose resource authorization is
// expressed through the host-service resourceRef.

package guest

import (
	"context"

	"lina-core/pkg/plugin/capability/aicap/aiaudio"
	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/aicap/aidocument"
	"lina-core/pkg/plugin/capability/aicap/aiembedding"
	"lina-core/pkg/plugin/capability/aicap/aiimage"
	"lina-core/pkg/plugin/capability/aicap/aisafety"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/aicap/aivideo"
	"lina-core/pkg/plugin/capability/aicap/aivision"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// AITextService exposes guest-side governed text AI generation.
type AITextService interface {
	// GenerateText executes one governed text generation call through the host
	// AI service. The call is authorized by purpose resource and returns the
	// same stable DTO used by source-plugin text AI consumers.
	GenerateText(ctx context.Context, request aitext.GenerateRequest) (*aitext.GenerateResponse, error)
}

// AIImageService exposes guest-side governed image generation and editing.
type AIImageService interface {
	// Generate executes one governed image generation call.
	Generate(ctx context.Context, request aiimage.GenerateRequest) (*aiimage.Response, error)
	// Edit executes one governed image editing call.
	Edit(ctx context.Context, request aiimage.EditRequest) (*aiimage.Response, error)
}

// AIEmbeddingService exposes guest-side governed embedding creation.
type AIEmbeddingService interface {
	// Create executes one governed embedding creation call.
	Create(ctx context.Context, request aiembedding.CreateRequest) (*aiembedding.CreateResponse, error)
}

// AIAudioService exposes guest-side governed audio transcription and synthesis.
type AIAudioService interface {
	// Transcribe executes one governed audio transcription call.
	Transcribe(ctx context.Context, request aiaudio.TranscribeRequest) (*aiaudio.TranscribeResponse, error)
	// Synthesize executes one governed audio synthesis call.
	Synthesize(ctx context.Context, request aiaudio.SynthesizeRequest) (*aiaudio.SynthesizeResponse, error)
}

// AIVisionService exposes guest-side governed visual analysis.
type AIVisionService interface {
	// Analyze executes one governed visual analysis call.
	Analyze(ctx context.Context, request aivision.AnalyzeRequest) (*aivision.AnalyzeResponse, error)
}

// AIDocumentService exposes guest-side governed document analysis and citation.
type AIDocumentService interface {
	// Analyze executes one governed document analysis call.
	Analyze(ctx context.Context, request aidocument.AnalyzeRequest) (*aidocument.Response, error)
	// Cite executes one governed citation-aware document call.
	Cite(ctx context.Context, request aidocument.CiteRequest) (*aidocument.Response, error)
}

// AISafetyService exposes guest-side governed safety moderation.
type AISafetyService interface {
	// Moderate executes one governed safety moderation call.
	Moderate(ctx context.Context, request aisafety.ModerateRequest) (*aisafety.ModerateResponse, error)
}

// AIVideoService exposes guest-side governed video and provider operation calls.
type AIVideoService interface {
	// Generate executes one governed video generation call.
	Generate(ctx context.Context, request aivideo.GenerateRequest) (*aivideo.Response, error)
	// Edit executes one governed video editing call.
	Edit(ctx context.Context, request aivideo.EditRequest) (*aivideo.Response, error)
	// Extend executes one governed video extension call.
	Extend(ctx context.Context, request aivideo.ExtendRequest) (*aivideo.Response, error)
	// OperationGet reads one provider operation projection.
	OperationGet(ctx context.Context, request aivideo.OperationGetRequest) (*aivideo.Response, error)
	// OperationCancel cancels one provider operation when authorized and supported.
	OperationCancel(ctx context.Context, request aivideo.OperationCancelRequest) (*aicommon.ProviderOperationRef, error)
}

// AIService exposes guest-side AI sub capabilities.
type AIService interface {
	// Text returns the governed text AI guest client.
	Text() AITextService
	// Image returns the governed image AI guest client.
	Image() AIImageService
	// Embedding returns the governed embedding AI guest client.
	Embedding() AIEmbeddingService
	// Audio returns the governed audio AI guest client.
	Audio() AIAudioService
	// Vision returns the governed vision AI guest client.
	Vision() AIVisionService
	// Document returns the governed document AI guest client.
	Document() AIDocumentService
	// Safety returns the governed safety AI guest client.
	Safety() AISafetyService
	// Video returns the governed video AI guest client.
	Video() AIVideoService
}

var _ AIService = (*aiService)(nil)
var _ AITextService = (*aiTextService)(nil)
var _ AIImageService = (*aiImageService)(nil)
var _ AIEmbeddingService = (*aiEmbeddingService)(nil)
var _ AIAudioService = (*aiAudioService)(nil)
var _ AIVisionService = (*aiVisionService)(nil)
var _ AIDocumentService = (*aiDocumentService)(nil)
var _ AISafetyService = (*aiSafetyService)(nil)
var _ AIVideoService = (*aiVideoService)(nil)

// aiService implements the guest-side AI namespace.
type aiService struct{}

// aiTextService implements guest text AI generation calls.
type aiTextService struct{}

// aiImageService implements guest image AI calls.
type aiImageService struct{}

// aiEmbeddingService implements guest embedding AI calls.
type aiEmbeddingService struct{}

// aiAudioService implements guest audio AI calls.
type aiAudioService struct{}

// aiVisionService implements guest vision AI calls.
type aiVisionService struct{}

// aiDocumentService implements guest document AI calls.
type aiDocumentService struct{}

// aiSafetyService implements guest safety AI calls.
type aiSafetyService struct{}

// aiVideoService implements guest video AI calls.
type aiVideoService struct{}

// Text returns the governed text AI guest client.
func (aiService) Text() AITextService {
	return aiTextService{}
}

// Image returns the governed image AI guest client.
func (aiService) Image() AIImageService { return aiImageService{} }

// Embedding returns the governed embedding AI guest client.
func (aiService) Embedding() AIEmbeddingService { return aiEmbeddingService{} }

// Audio returns the governed audio AI guest client.
func (aiService) Audio() AIAudioService { return aiAudioService{} }

// Vision returns the governed vision AI guest client.
func (aiService) Vision() AIVisionService { return aiVisionService{} }

// Document returns the governed document AI guest client.
func (aiService) Document() AIDocumentService { return aiDocumentService{} }

// Safety returns the governed safety AI guest client.
func (aiService) Safety() AISafetyService { return aiSafetyService{} }

// Video returns the governed video AI guest client.
func (aiService) Video() AIVideoService { return aiVideoService{} }

// GenerateText executes one governed text generation call.
func (aiTextService) GenerateText(_ context.Context, request aitext.GenerateRequest) (*aitext.GenerateResponse, error) {
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
	err := invokeCapabilityJSONWithResource(
		protocol.HostServiceAI,
		protocol.HostServiceMethodAITextGenerate,
		aitext.PurposeResourceRef(request.Purpose),
		payload,
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Generate executes one governed image generation call.
func (aiImageService) Generate(_ context.Context, request aiimage.GenerateRequest) (*aiimage.Response, error) {
	var response aiimage.Response
	err := invokeAIJSON(
		protocol.HostServiceMethodAIImageGenerate,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAIImageGenerateRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Edit executes one governed image editing call.
func (aiImageService) Edit(_ context.Context, request aiimage.EditRequest) (*aiimage.Response, error) {
	var response aiimage.Response
	err := invokeAIJSON(
		protocol.HostServiceMethodAIImageEdit,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAIImageEditRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Create executes one governed embedding creation call.
func (aiEmbeddingService) Create(_ context.Context, request aiembedding.CreateRequest) (*aiembedding.CreateResponse, error) {
	var response aiembedding.CreateResponse
	err := invokeAIJSON(
		protocol.HostServiceMethodAIEmbeddingCreate,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAIEmbeddingCreateRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Transcribe executes one governed audio transcription call.
func (aiAudioService) Transcribe(_ context.Context, request aiaudio.TranscribeRequest) (*aiaudio.TranscribeResponse, error) {
	var response aiaudio.TranscribeResponse
	err := invokeAIJSON(
		protocol.HostServiceMethodAIAudioTranscribe,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAIAudioTranscribeRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Synthesize executes one governed audio synthesis call.
func (aiAudioService) Synthesize(_ context.Context, request aiaudio.SynthesizeRequest) (*aiaudio.SynthesizeResponse, error) {
	var response aiaudio.SynthesizeResponse
	err := invokeAIJSON(
		protocol.HostServiceMethodAIAudioSynthesize,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAIAudioSynthesizeRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Analyze executes one governed visual analysis call.
func (aiVisionService) Analyze(_ context.Context, request aivision.AnalyzeRequest) (*aivision.AnalyzeResponse, error) {
	var response aivision.AnalyzeResponse
	err := invokeAIJSON(
		protocol.HostServiceMethodAIVisionAnalyze,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAIVisionAnalyzeRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Analyze executes one governed document analysis call.
func (aiDocumentService) Analyze(_ context.Context, request aidocument.AnalyzeRequest) (*aidocument.Response, error) {
	var response aidocument.Response
	err := invokeAIJSON(
		protocol.HostServiceMethodAIDocumentAnalyze,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAIDocumentAnalyzeRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Cite executes one governed citation-aware document call.
func (aiDocumentService) Cite(_ context.Context, request aidocument.CiteRequest) (*aidocument.Response, error) {
	var response aidocument.Response
	err := invokeAIJSON(
		protocol.HostServiceMethodAIDocumentCite,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAIDocumentCiteRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Moderate executes one governed safety moderation call.
func (aiSafetyService) Moderate(_ context.Context, request aisafety.ModerateRequest) (*aisafety.ModerateResponse, error) {
	var response aisafety.ModerateResponse
	err := invokeAIJSON(
		protocol.HostServiceMethodAISafetyModerate,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAISafetyModerateRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Generate executes one governed video generation call.
func (aiVideoService) Generate(_ context.Context, request aivideo.GenerateRequest) (*aivideo.Response, error) {
	var response aivideo.Response
	err := invokeAIJSON(
		protocol.HostServiceMethodAIVideoGenerate,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAIVideoGenerateRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Edit executes one governed video editing call.
func (aiVideoService) Edit(_ context.Context, request aivideo.EditRequest) (*aivideo.Response, error) {
	var response aivideo.Response
	err := invokeAIJSON(
		protocol.HostServiceMethodAIVideoEdit,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAIVideoEditRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Extend executes one governed video extension call.
func (aiVideoService) Extend(_ context.Context, request aivideo.ExtendRequest) (*aivideo.Response, error) {
	var response aivideo.Response
	err := invokeAIJSON(
		protocol.HostServiceMethodAIVideoExtend,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAIVideoExtendRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// OperationGet reads one provider operation projection.
func (aiVideoService) OperationGet(_ context.Context, request aivideo.OperationGetRequest) (*aivideo.Response, error) {
	var response aivideo.Response
	err := invokeAIJSON(
		protocol.HostServiceMethodAIVideoOperationGet,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAIVideoOperationGetRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// OperationCancel cancels one provider operation when authorized and supported.
func (aiVideoService) OperationCancel(_ context.Context, request aivideo.OperationCancelRequest) (*aicommon.ProviderOperationRef, error) {
	var response aicommon.ProviderOperationRef
	err := invokeAIJSON(
		protocol.HostServiceMethodAIVideoOperationCancel,
		aicommon.PurposeResourceRef(request.Purpose),
		protocol.MarshalHostServiceAIVideoOperationCancelRequest(&request),
		&response,
	)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func invokeAIJSON(method string, resourceRef string, payload []byte, out any) error {
	return invokeCapabilityJSONWithResource(protocol.HostServiceAI, method, resourceRef, payload, out)
}
