// This file adapts dynamic-plugin AI host-service calls to the ordinary
// aitext.Service consumer contract. The dispatcher keeps provider/model/tier
// storage and external protocol details inside the official AI plugin.

package wasm

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/aicap"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchAIHostService routes one AI host-service method to the same ordinary
// aitext.Service surface exposed to source plugins.
func dispatchAIHostService(
	ctx context.Context,
	hcc *hostCallContext,
	_ string,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	switch method {
	case bridgehostservice.HostServiceMethodAITextGenerate:
		return dispatchAITextGenerate(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAITextMethodStatus:
		return dispatchAITextMethodStatus(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIMethodStatuses:
		return dispatchAIMethodStatuses(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIImageGenerate:
		return dispatchAIImageGenerate(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIImageEdit:
		return dispatchAIImageEdit(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIEmbeddingCreate:
		return dispatchAIEmbeddingCreate(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIAudioTranscribe:
		return dispatchAIAudioTranscribe(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIAudioSynthesize:
		return dispatchAIAudioSynthesize(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIVisionAnalyze:
		return dispatchAIVisionAnalyze(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIDocumentAnalyze:
		return dispatchAIDocumentAnalyze(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIDocumentCite:
		return dispatchAIDocumentCite(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAISafetyModerate:
		return dispatchAISafetyModerate(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIVideoGenerate:
		return dispatchAIVideoGenerate(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIVideoEdit:
		return dispatchAIVideoEdit(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIVideoExtend:
		return dispatchAIVideoExtend(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIVideoOperationGet:
		return dispatchAIVideoOperationGet(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodAIVideoOperationCancel:
		return dispatchAIVideoOperationCancel(ctx, hcc, payload)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"ai host service method not implemented: "+method,
		)
	}
}

// dispatchAITextGenerate decodes, validates, and executes one text generation request.
func dispatchAITextGenerate(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := aiTextServiceForHostCall(hcc)
	if service == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInternalError,
			"ai text host service is not scoped",
		)
	}
	request, err := bridgehostservice.UnmarshalHostServiceAITextGenerateRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	response, err := service.GenerateText(ctx, aitext.GenerateRequest{
		Purpose:         request.Purpose,
		Tier:            request.Tier,
		Messages:        request.Messages,
		MaxOutputTokens: request.MaxOutputTokens,
		Temperature:     request.Temperature,
		ThinkingEffort:  request.ThinkingEffort,
		Metadata:        request.Metadata,
	})
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInvalidRequest,
			sanitizeAIHostServiceError(err),
		)
	}
	return capabilityJSONResponse(response)
}

// dispatchAITextMethodStatus decodes and returns one text AI method status.
func dispatchAITextMethodStatus(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := aiTextServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	var request aicap.MethodStatusQuery
	if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
		return invalidCapabilityRequest(err)
	}
	return capabilityJSONResponse(service.MethodStatus(ctx, request.CapabilityMethod))
}

// dispatchAIMethodStatuses decodes and returns cross-capability AI method statuses.
func dispatchAIMethodStatuses(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	var request aicap.MethodStatusesInput
	if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
		return invalidCapabilityRequest(err)
	}
	response, err := service.MethodStatuses(ctx, request)
	return aiCapabilityResponse(response, err)
}

// dispatchAIImageGenerate decodes, validates, and executes one image generation request.
func dispatchAIImageGenerate(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIImageGenerateRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Image().Generate(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// dispatchAIImageEdit decodes, validates, and executes one image editing request.
func dispatchAIImageEdit(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIImageEditRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Image().Edit(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// dispatchAIEmbeddingCreate decodes, validates, and executes one embedding request.
func dispatchAIEmbeddingCreate(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIEmbeddingCreateRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Embedding().Create(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// dispatchAIAudioTranscribe decodes, validates, and executes one audio transcription request.
func dispatchAIAudioTranscribe(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIAudioTranscribeRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Audio().Transcribe(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// dispatchAIAudioSynthesize decodes, validates, and executes one audio synthesis request.
func dispatchAIAudioSynthesize(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIAudioSynthesizeRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Audio().Synthesize(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// dispatchAIVisionAnalyze decodes, validates, and executes one visual analysis request.
func dispatchAIVisionAnalyze(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIVisionAnalyzeRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Vision().Analyze(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// dispatchAIDocumentAnalyze decodes, validates, and executes one document analysis request.
func dispatchAIDocumentAnalyze(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIDocumentAnalyzeRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Document().Analyze(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// dispatchAIDocumentCite decodes, validates, and executes one document citation request.
func dispatchAIDocumentCite(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIDocumentCiteRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Document().Cite(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// dispatchAISafetyModerate decodes, validates, and executes one safety moderation request.
func dispatchAISafetyModerate(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAISafetyModerateRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Safety().Moderate(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// dispatchAIVideoGenerate decodes, validates, and executes one video generation request.
func dispatchAIVideoGenerate(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIVideoGenerateRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Video().Generate(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// dispatchAIVideoEdit decodes, validates, and executes one video editing request.
func dispatchAIVideoEdit(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIVideoEditRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Video().Edit(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// dispatchAIVideoExtend decodes, validates, and executes one video extension request.
func dispatchAIVideoExtend(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIVideoExtendRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Video().Extend(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// dispatchAIVideoOperationGet decodes, validates, and executes one provider operation lookup.
func dispatchAIVideoOperationGet(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIVideoOperationGetRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Video().OperationGet(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// dispatchAIVideoOperationCancel decodes, validates, and executes one provider operation cancel.
func dispatchAIVideoOperationCancel(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIVideoOperationCancelRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Video().OperationCancel(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// aiTextServiceForHostCall resolves the text AI service for one host call.
func aiTextServiceForHostCall(hcc *hostCallContext) aitext.Service {
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return nil
	}
	return service.Text()
}

// aiServiceForHostCall resolves the AI namespace service for one host call.
func aiServiceForHostCall(hcc *hostCallContext) aicap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.AI()
}

func aiServiceNotScoped() *bridgehostcall.HostCallResponseEnvelope {
	return bridgehostcall.NewHostCallErrorResponse(
		bridgehostcall.HostCallStatusInternalError,
		"ai host service is not scoped",
	)
}

func aiCapabilityResponse(response any, err error) *bridgehostcall.HostCallResponseEnvelope {
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInvalidRequest,
			sanitizeAIHostServiceError(err),
		)
	}
	return capabilityJSONResponse(response)
}

// sanitizeAIHostServiceError returns a compact message that avoids propagating
// request bodies, provider responses, or authorization headers through host-call errors.
func sanitizeAIHostServiceError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	replacements := []string{"authorization", "api-key", "apikey", "bearer", "sk-"}
	lower := strings.ToLower(message)
	for _, marker := range replacements {
		if strings.Contains(lower, marker) {
			return "ai text generation failed with a redacted provider or authorization error"
		}
	}
	if len(message) > 512 {
		return message[:512]
	}
	return message
}
