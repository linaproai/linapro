// Package aivision defines the typed vision AI capability contract exposed
// under AI().Vision().
package aivision

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	// CapabilityAIVisionV1 identifies the versioned vision AI framework capability.
	CapabilityAIVisionV1 = "framework.ai.vision.v1"
	// CapabilityType identifies the vision capability family.
	CapabilityType = aicommon.CapabilityTypeVision
	// CapabilityMethodAnalyze identifies visual understanding.
	CapabilityMethodAnalyze = aicommon.CapabilityMethodVisionAnalyze
)

// AnalyzeRequest carries one governed visual analysis request.
type AnalyzeRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aicommon.Tier `json:"tier"`
	// Images references input images, screenshots, diagrams, or frames.
	Images []aicommon.AssetRef `json:"images"`
	// Prompt is the analysis instruction.
	Prompt string `json:"prompt"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// AnalyzeResponse carries visual analysis text and provider identity.
type AnalyzeResponse struct {
	// Text is the visual analysis result.
	Text string `json:"text"`
	// Provider contains public provider/model identity.
	Provider aicommon.ProviderProjection `json:"provider"`
	// Usage contains minimal usage details.
	Usage aicommon.Usage `json:"usage,omitempty"`
	// LatencyMs is the provider call latency in milliseconds.
	LatencyMs int `json:"latencyMs,omitempty"`
	// AnalyzedAt is a Unix timestamp in milliseconds.
	AnalyzedAt int64 `json:"analyzedAt,omitempty"`
}

// ProviderRequest carries provider-internal vision requests with source identity.
type ProviderRequest struct {
	AnalyzeRequest
	// SourcePluginID identifies the dynamic or source plugin that initiated the call.
	SourcePluginID string `json:"sourcePluginId,omitempty"`
}

// Provider defines vision AI capability implemented by provider plugins.
type Provider interface {
	// AnalyzeVision executes one governed visual analysis request.
	AnalyzeVision(ctx context.Context, request ProviderRequest) (*AnalyzeResponse, error)
}

// Service defines the vision AI capability consumed by host services and plugins.
type Service interface {
	// Available reports whether an active vision provider is available.
	Available(ctx context.Context) bool
	// Status returns the current vision AI capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// MethodStatus returns the current activation state for one vision method.
	MethodStatus(ctx context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus
	// Analyze executes one governed visual analysis request.
	Analyze(ctx context.Context, request AnalyzeRequest) (*AnalyzeResponse, error)
}

type serviceImpl struct {
	sourcePluginID string
}

var _ Service = (*serviceImpl)(nil)

// New creates a fallback vision AI service.
func New() Service { return &serviceImpl{} }

// ForPlugin returns a plugin-scoped vision service.
func ForPlugin(service Service, pluginID string) Service {
	if _, ok := service.(*serviceImpl); service == nil || ok {
		return &serviceImpl{sourcePluginID: strings.TrimSpace(pluginID)}
	}
	return service
}

// Available reports whether an active vision provider is available.
func (*serviceImpl) Available(context.Context) bool { return false }

// Status returns the fallback vision capability status.
func (*serviceImpl) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(CapabilityAIVisionV1)
}

// MethodStatus returns a fallback vision method status.
func (*serviceImpl) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(CapabilityAIVisionV1, CapabilityType, method)
}

// Analyze validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Analyze(_ context.Context, request AnalyzeRequest) (*AnalyzeResponse, error) {
	if strings.TrimSpace(request.Purpose) == "" {
		return nil, bizerr.NewCode(aicommon.CodePurposeRequired)
	}
	if !request.Tier.Valid() {
		return nil, bizerr.NewCode(aicommon.CodeTierInvalid, bizerr.P("tier", string(request.Tier)))
	}
	if len(request.Images) == 0 || strings.TrimSpace(request.Images[0].Ref) == "" {
		return nil, bizerr.NewCode(aicommon.CodeAssetRefRequired)
	}
	return nil, bizerr.NewCode(
		aicommon.CodeProviderUnavailable,
		bizerr.P("capabilityType", string(CapabilityType)),
		bizerr.P("capabilityMethod", string(CapabilityMethodAnalyze)),
	)
}
