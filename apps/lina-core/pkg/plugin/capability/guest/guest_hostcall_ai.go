// This file implements guest-side text AI capability calls that cross the
// pluginbridge host-service transport. Purpose resource authorization is
// expressed through the host-service resourceRef.

package guest

import (
	"context"

	"lina-core/pkg/plugin/capability/ai/aitext"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// AITextService exposes guest-side governed text AI generation.
type AITextService interface {
	// GenerateText executes one governed text generation call through the host
	// AI service. The call is authorized by purpose resource and returns the
	// same stable DTO used by source-plugin text AI consumers.
	GenerateText(ctx context.Context, request aitext.GenerateRequest) (*aitext.GenerateResponse, error)
}

// AIService exposes guest-side AI sub capabilities.
type AIService interface {
	// Text returns the governed text AI guest client.
	Text() AITextService
}

var _ AIService = (*aiService)(nil)
var _ AITextService = (*aiTextService)(nil)

// aiService implements the guest-side AI namespace.
type aiService struct{}

// aiTextService implements guest text AI generation calls.
type aiTextService struct{}

// Text returns the governed text AI guest client.
func (aiService) Text() AITextService {
	return aiTextService{}
}

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
