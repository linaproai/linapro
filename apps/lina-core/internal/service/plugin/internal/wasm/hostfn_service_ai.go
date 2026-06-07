// This file adapts dynamic-plugin AI host-service calls to the ordinary
// aitext.Service consumer contract. The dispatcher keeps provider/model/tier
// storage and external protocol details inside the official AI plugin.

package wasm

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/aicap"
	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// aiTextHostServices stores the runtime-owned capability services used by AI
// host-service dispatch.
var aiTextHostServices capability.Services

// ConfigureAITextHostService replaces the text AI capability service directory
// used by dynamic-plugin host calls.
func ConfigureAITextHostService(services capability.Services) error {
	if services == nil {
		return gerror.New("ai text host services directory is nil")
	}
	aiTextHostServices = services
	return nil
}

// dispatchAIHostService routes one AI host-service method to the same ordinary
// aitext.Service surface exposed to source plugins.
func dispatchAIHostService(
	ctx context.Context,
	hcc *hostCallContext,
	resourceRef string,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	switch method {
	case bridgehostservice.HostServiceMethodAITextGenerate:
		return dispatchAITextGenerate(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAIImageGenerate:
		return dispatchAIImageGenerate(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAIImageEdit:
		return dispatchAIImageEdit(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAIEmbeddingCreate:
		return dispatchAIEmbeddingCreate(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAIAudioTranscribe:
		return dispatchAIAudioTranscribe(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAIAudioSynthesize:
		return dispatchAIAudioSynthesize(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAIVisionAnalyze:
		return dispatchAIVisionAnalyze(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAIDocumentAnalyze:
		return dispatchAIDocumentAnalyze(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAIDocumentCite:
		return dispatchAIDocumentCite(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAISafetyModerate:
		return dispatchAISafetyModerate(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAIVideoGenerate:
		return dispatchAIVideoGenerate(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAIVideoEdit:
		return dispatchAIVideoEdit(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAIVideoExtend:
		return dispatchAIVideoExtend(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAIVideoOperationGet:
		return dispatchAIVideoOperationGet(ctx, hcc, resourceRef, payload)
	case bridgehostservice.HostServiceMethodAIVideoOperationCancel:
		return dispatchAIVideoOperationCancel(ctx, hcc, resourceRef, payload)
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
	resourceRef string,
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
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	resource := hcc.hostServiceResource(bridgehostservice.HostServiceAI, resourceRef)
	if resource == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			"ai text purpose is not authorized",
		)
	}
	normalizedPurpose := strings.TrimSpace(request.Purpose)
	expectedRef := aitext.PurposeResourceRef(normalizedPurpose)
	if expectedRef == "" {
		normalizedPurpose = strings.TrimSpace(strings.TrimPrefix(resourceRef, "purpose:"))
		expectedRef = aitext.PurposeResourceRef(normalizedPurpose)
	}
	if expectedRef == "" || expectedRef != strings.TrimSpace(resource.Ref) {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			fmt.Sprintf("ai text purpose %s is not authorized", normalizedPurpose),
		)
	}
	request.Purpose = normalizedPurpose
	if request.Tier == "" {
		request.Tier = aitext.Tier(resource.Attributes["defaultTier"])
	}
	if request.Tier == "" {
		request.Tier = aitext.TierStandard
	}
	maxOutputTokens, err := applyAIOutputPolicy(resource, request.MaxOutputTokens)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	response, err := service.GenerateText(ctx, aitext.GenerateRequest{
		Purpose:         request.Purpose,
		Tier:            request.Tier,
		Messages:        request.Messages,
		MaxOutputTokens: maxOutputTokens,
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

// dispatchAIImageGenerate decodes, validates, and executes one image generation request.
func dispatchAIImageGenerate(
	ctx context.Context,
	hcc *hostCallContext,
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIImageGenerateRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	resource, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, nil, request.Count, false, false)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
	request.Tier = applyAIDefaultTier(resource, request.Tier)
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
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIImageEditRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	assets := []aicommon.AssetRef{request.Image}
	if request.Mask != nil {
		assets = append(assets, *request.Mask)
	}
	resource, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, assets, request.Count, false, false)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
	request.Tier = applyAIDefaultTier(resource, request.Tier)
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
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIEmbeddingCreateRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	assets := make([]aicommon.AssetRef, 0, len(request.Inputs))
	for _, input := range request.Inputs {
		if input.AssetRef != nil {
			assets = append(assets, *input.AssetRef)
		}
	}
	resource, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, assets, 0, false, false)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
	request.Tier = applyAIDefaultTier(resource, request.Tier)
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
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIAudioTranscribeRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	resource, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, []aicommon.AssetRef{request.Audio}, 0, false, false)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
	request.Tier = applyAIDefaultTier(resource, request.Tier)
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
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIAudioSynthesizeRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	resource, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, nil, 1, false, false)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
	request.Tier = applyAIDefaultTier(resource, request.Tier)
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
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIVisionAnalyzeRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	resource, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, request.Images, 0, false, false)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
	request.Tier = applyAIDefaultTier(resource, request.Tier)
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
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIDocumentAnalyzeRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	resource, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, request.Documents, 0, false, false)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
	request.Tier = applyAIDefaultTier(resource, request.Tier)
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
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIDocumentCiteRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	resource, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, request.Documents, 0, false, false)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
	request.Tier = applyAIDefaultTier(resource, request.Tier)
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
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAISafetyModerateRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	resource, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, request.Assets, 0, false, false)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
	request.Tier = applyAIDefaultTier(resource, request.Tier)
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
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIVideoGenerateRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	resource, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, nil, 1, false, false)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
	request.Tier = applyAIDefaultTier(resource, request.Tier)
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
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIVideoEditRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	resource, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, []aicommon.AssetRef{request.Video}, 1, false, false)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
	request.Tier = applyAIDefaultTier(resource, request.Tier)
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
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIVideoExtendRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	resource, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, []aicommon.AssetRef{request.Video}, 1, false, false)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
	request.Tier = applyAIDefaultTier(resource, request.Tier)
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
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIVideoOperationGetRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	_, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, nil, 0, true, false)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
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
	resourceRef string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceAIVideoOperationCancelRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	_, normalizedPurpose, policyErr := authorizeAIRequest(hcc, resourceRef, request.Purpose, payload, nil, 0, true, true)
	if policyErr != nil {
		return policyErr
	}
	request.Purpose = normalizedPurpose
	service := aiServiceForHostCall(hcc)
	if service == nil {
		return aiServiceNotScoped()
	}
	response, err := service.Video().OperationCancel(ctx, *request)
	return aiCapabilityResponse(response, err)
}

// aiTextServiceForPlugin returns the runtime-owned text AI capability bound to
// one dynamic plugin.
func aiTextServiceForPlugin(pluginID string) aitext.Service {
	service := aiServiceForPlugin(pluginID)
	if service == nil {
		return nil
	}
	return service.Text()
}

// aiServiceForPlugin returns the runtime-owned AI namespace bound to one dynamic plugin.
func aiServiceForPlugin(pluginID string) aicap.Service {
	if aiTextHostServices == nil {
		return nil
	}
	services := capability.ServicesForPlugin(aiTextHostServices, pluginID)
	if services == nil {
		return nil
	}
	return services.AI()
}

// aiTextServiceForHostCall resolves the text AI service for one host call.
func aiTextServiceForHostCall(hcc *hostCallContext) aitext.Service {
	if hcc == nil {
		return nil
	}
	return aiTextServiceForPlugin(hcc.pluginID)
}

// aiServiceForHostCall resolves the AI namespace service for one host call.
func aiServiceForHostCall(hcc *hostCallContext) aicap.Service {
	if hcc == nil {
		return nil
	}
	return aiServiceForPlugin(hcc.pluginID)
}

// authorizeAIRequest checks resource, purpose, payload, asset, and operation policies.
func authorizeAIRequest(
	hcc *hostCallContext,
	resourceRef string,
	purpose string,
	payload []byte,
	assets []aicommon.AssetRef,
	requestedOutputAssets int,
	requiresOperation bool,
	requiresCancel bool,
) (*bridgehostservice.HostServiceResourceSpec, string, *bridgehostcall.HostCallResponseEnvelope) {
	resource := hcc.hostServiceResource(bridgehostservice.HostServiceAI, resourceRef)
	if resource == nil {
		return nil, "", bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			"ai purpose is not authorized",
		)
	}
	normalizedPurpose := strings.TrimSpace(purpose)
	expectedRef := aicommon.PurposeResourceRef(normalizedPurpose)
	if expectedRef == "" {
		normalizedPurpose = strings.TrimSpace(strings.TrimPrefix(resourceRef, "purpose:"))
		expectedRef = aicommon.PurposeResourceRef(normalizedPurpose)
	}
	if expectedRef == "" || expectedRef != strings.TrimSpace(resource.Ref) {
		return nil, "", bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			fmt.Sprintf("ai purpose %s is not authorized", normalizedPurpose),
		)
	}
	if err := applyAIPayloadPolicy(resource, payload); err != nil {
		return nil, "", bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if err := applyAIAssetPolicy(resource, assets, requestedOutputAssets); err != nil {
		return nil, "", bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}
	if err := applyAIOperationPolicy(resource, requiresOperation, requiresCancel); err != nil {
		return nil, "", bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusCapabilityDenied, err.Error())
	}
	return resource, normalizedPurpose, nil
}

// applyAIDefaultTier applies the purpose default tier and stable fallback tier.
func applyAIDefaultTier(resource *bridgehostservice.HostServiceResourceSpec, requested aicommon.Tier) aicommon.Tier {
	if requested != "" {
		return requested
	}
	if resource != nil && resource.Attributes != nil {
		if tier := aicommon.Tier(strings.TrimSpace(resource.Attributes["defaultTier"])); tier != "" {
			return tier
		}
	}
	return aicommon.TierStandard
}

// applyAIOutputPolicy checks purpose-level maxOutputTokens before provider
// dispatch and applies the authorized cap when the caller omits an explicit limit.
func applyAIOutputPolicy(resource *bridgehostservice.HostServiceResourceSpec, requested int) (int, error) {
	if requested < 0 {
		return 0, gerror.New("maxOutputTokens must be greater than or equal to zero")
	}
	if resource == nil || resource.Attributes == nil || strings.TrimSpace(resource.Attributes["maxOutputTokens"]) == "" {
		return requested, nil
	}
	maxValue, err := strconv.Atoi(strings.TrimSpace(resource.Attributes["maxOutputTokens"]))
	if err != nil || maxValue <= 0 {
		return 0, gerror.New("authorized maxOutputTokens is invalid")
	}
	if requested == 0 {
		return maxValue, nil
	}
	if requested > maxValue {
		return 0, gerror.Newf("maxOutputTokens exceeds authorized purpose limit: requested=%d max=%d", requested, maxValue)
	}
	return requested, nil
}

// applyAIPayloadPolicy checks the resource payload byte limit.
func applyAIPayloadPolicy(resource *bridgehostservice.HostServiceResourceSpec, payload []byte) error {
	limit := aiIntAttribute(resource, "maxPayloadBytes")
	if limit <= 0 {
		return nil
	}
	if len(payload) > limit {
		return gerror.Newf("ai payload exceeds authorized limit: size=%d max=%d", len(payload), limit)
	}
	return nil
}

// applyAIAssetPolicy checks input/output asset count, byte, and MIME limits.
func applyAIAssetPolicy(resource *bridgehostservice.HostServiceResourceSpec, assets []aicommon.AssetRef, requestedOutputAssets int) error {
	maxInputAssets := aiIntAttribute(resource, "maxInputAssets")
	if maxInputAssets > 0 && len(assets) > maxInputAssets {
		return gerror.Newf("ai input assets exceed authorized limit: count=%d max=%d", len(assets), maxInputAssets)
	}
	maxOutputAssets := aiIntAttribute(resource, "maxOutputAssets")
	if maxOutputAssets > 0 && requestedOutputAssets > maxOutputAssets {
		return gerror.Newf("ai output assets exceed authorized limit: count=%d max=%d", requestedOutputAssets, maxOutputAssets)
	}
	maxAssetBytes := int64(aiIntAttribute(resource, "maxAssetBytes"))
	allowedMIMEs := aiAllowedMIMEs(resource)
	for _, asset := range assets {
		if strings.TrimSpace(asset.Ref) == "" {
			return gerror.New("ai asset reference is not visible to this plugin")
		}
		if maxAssetBytes > 0 && asset.SizeBytes > maxAssetBytes {
			return gerror.Newf("ai asset exceeds authorized byte limit: size=%d max=%d", asset.SizeBytes, maxAssetBytes)
		}
		if len(allowedMIMEs) > 0 && !aiMIMEAllowed(asset.MimeType, allowedMIMEs) {
			return gerror.Newf("ai asset mime type is not authorized: %s", asset.MimeType)
		}
	}
	return nil
}

// applyAIOperationPolicy checks provider operation access policy.
func applyAIOperationPolicy(resource *bridgehostservice.HostServiceResourceSpec, requiresOperation bool, requiresCancel bool) error {
	if requiresOperation && !aiBoolAttribute(resource, "allowOperation") {
		return gerror.New("ai provider operation access is not authorized")
	}
	if requiresCancel && !aiBoolAttribute(resource, "allowOperationCancel") {
		return gerror.New("ai provider operation cancellation is not authorized")
	}
	return nil
}

func aiIntAttribute(resource *bridgehostservice.HostServiceResourceSpec, key string) int {
	if resource == nil || resource.Attributes == nil {
		return 0
	}
	value, err := strconv.Atoi(strings.TrimSpace(resource.Attributes[key]))
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func aiBoolAttribute(resource *bridgehostservice.HostServiceResourceSpec, key string) bool {
	if resource == nil || resource.Attributes == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(resource.Attributes[key]), "true")
}

func aiAllowedMIMEs(resource *bridgehostservice.HostServiceResourceSpec) []string {
	if resource == nil || resource.Attributes == nil {
		return nil
	}
	raw := strings.TrimSpace(resource.Attributes["allowedMimeTypes"])
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.ToLower(strings.TrimSpace(part)); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func aiMIMEAllowed(mimeType string, allowed []string) bool {
	normalized := strings.ToLower(strings.TrimSpace(mimeType))
	for _, item := range allowed {
		if item == "*" || item == normalized {
			return true
		}
		if strings.HasSuffix(item, "/*") && strings.HasPrefix(normalized, strings.TrimSuffix(item, "*")) {
			return true
		}
	}
	return false
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
