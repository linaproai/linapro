// Package aisafety defines the typed safety AI capability contract exposed
// under AI().Safety().
package aisafety

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	// CapabilityAISafetyV1 identifies the versioned safety AI framework capability.
	CapabilityAISafetyV1 = "framework.ai.safety.v1"
	// CapabilityType identifies the safety capability family.
	CapabilityType = aicommon.CapabilityTypeSafety
	// CapabilityMethodModerate identifies safety moderation.
	CapabilityMethodModerate = aicommon.CapabilityMethodSafetyModerate
)

// ModerateRequest carries one governed safety moderation request.
type ModerateRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aicommon.Tier `json:"tier"`
	// Text optionally carries short text to moderate.
	Text string `json:"text,omitempty"`
	// Assets references non-text inputs for moderation.
	Assets []aicommon.AssetRef `json:"assets,omitempty"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CategoryScore contains a provider-neutral safety category result.
type CategoryScore struct {
	// Category is the stable moderation category.
	Category string `json:"category"`
	// Flagged reports whether this category was flagged.
	Flagged bool `json:"flagged"`
	// Score is the provider-normalized confidence score when available.
	Score float64 `json:"score,omitempty"`
}

// ModerateResponse carries moderation results and provider identity.
type ModerateResponse struct {
	// Flagged reports whether the input violates the applied safety policy.
	Flagged bool `json:"flagged"`
	// Categories contains provider-neutral category scores.
	Categories []CategoryScore `json:"categories,omitempty"`
	// Provider contains public provider/model identity.
	Provider aicommon.ProviderProjection `json:"provider"`
	// Usage contains minimal usage details.
	Usage aicommon.Usage `json:"usage,omitempty"`
	// LatencyMs is the provider call latency in milliseconds.
	LatencyMs int `json:"latencyMs,omitempty"`
	// ModeratedAt is a Unix timestamp in milliseconds.
	ModeratedAt int64 `json:"moderatedAt,omitempty"`
}

// ProviderRequest carries provider-internal safety requests with source identity.
type ProviderRequest struct {
	ModerateRequest
	// SourcePluginID identifies the dynamic or source plugin that initiated the call.
	SourcePluginID string `json:"sourcePluginId,omitempty"`
}

// Provider defines safety AI capability implemented by provider plugins.
type Provider interface {
	// Moderate executes one governed safety moderation request.
	Moderate(ctx context.Context, request ProviderRequest) (*ModerateResponse, error)
}

// Service defines the safety AI capability consumed by host services and plugins.
type Service interface {
	// Available reports whether an active safety provider is available.
	Available(ctx context.Context) bool
	// Status returns the current safety AI capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// MethodStatus returns the current activation state for one safety method.
	MethodStatus(ctx context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus
	// Moderate executes one governed safety moderation request.
	Moderate(ctx context.Context, request ModerateRequest) (*ModerateResponse, error)
}

type serviceImpl struct {
	sourcePluginID string
}

var _ Service = (*serviceImpl)(nil)

// New creates a fallback safety AI service.
func New() Service { return &serviceImpl{} }

// ForPlugin returns a plugin-scoped safety service.
func ForPlugin(service Service, pluginID string) Service {
	if _, ok := service.(*serviceImpl); service == nil || ok {
		return &serviceImpl{sourcePluginID: strings.TrimSpace(pluginID)}
	}
	return service
}

// Available reports whether an active safety provider is available.
func (*serviceImpl) Available(context.Context) bool { return false }

// Status returns the fallback safety capability status.
func (*serviceImpl) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(CapabilityAISafetyV1)
}

// MethodStatus returns a fallback safety method status.
func (*serviceImpl) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(CapabilityAISafetyV1, CapabilityType, method)
}

// Moderate validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Moderate(_ context.Context, request ModerateRequest) (*ModerateResponse, error) {
	if strings.TrimSpace(request.Purpose) == "" {
		return nil, bizerr.NewCode(aicommon.CodePurposeRequired)
	}
	if !request.Tier.Valid() {
		return nil, bizerr.NewCode(aicommon.CodeTierInvalid, bizerr.P("tier", string(request.Tier)))
	}
	if strings.TrimSpace(request.Text) == "" && len(request.Assets) == 0 {
		return nil, bizerr.NewCode(aicommon.CodeAssetRefRequired)
	}
	return nil, bizerr.NewCode(
		aicommon.CodeProviderUnavailable,
		bizerr.P("capabilityType", string(CapabilityType)),
		bizerr.P("capabilityMethod", string(CapabilityMethodModerate)),
	)
}
