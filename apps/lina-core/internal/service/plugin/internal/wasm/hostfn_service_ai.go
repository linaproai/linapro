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
	"lina-core/pkg/plugin/capability/ai/aitext"
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

// aiTextServiceForPlugin returns the runtime-owned text AI capability bound to
// one dynamic plugin.
func aiTextServiceForPlugin(pluginID string) aitext.Service {
	if aiTextHostServices == nil {
		return nil
	}
	services := capability.ServicesForPlugin(aiTextHostServices, pluginID)
	if services == nil {
		return nil
	}
	aiService := services.AI()
	if aiService == nil {
		return nil
	}
	return aiService.Text()
}

// aiTextServiceForHostCall resolves the text AI service for one host call.
func aiTextServiceForHostCall(hcc *hostCallContext) aitext.Service {
	if hcc == nil {
		return nil
	}
	return aiTextServiceForPlugin(hcc.pluginID)
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
