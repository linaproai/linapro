// Package aiimage defines the typed image AI capability contract exposed under
// AI().Image(). It owns image-specific DTOs and fallback behavior; provider
// storage and external protocol details stay in provider plugins.
package aiimage

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	// CapabilityAIImageV1 identifies the versioned image AI framework capability.
	CapabilityAIImageV1 = "framework.ai.image.v1"
)

const (
	// CapabilityType identifies the image capability family.
	CapabilityType = aicommon.CapabilityTypeImage
	// CapabilityMethodGenerate identifies image generation.
	CapabilityMethodGenerate = aicommon.CapabilityMethodImageGenerate
	// CapabilityMethodEdit identifies image editing.
	CapabilityMethodEdit = aicommon.CapabilityMethodImageEdit
)

// GenerateRequest carries one governed image generation request.
type GenerateRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aicommon.Tier `json:"tier"`
	// Prompt is the text instruction used to generate images.
	Prompt string `json:"prompt"`
	// Size optionally requests a provider-supported image size.
	Size string `json:"size,omitempty"`
	// Count caps the number of generated assets.
	Count int `json:"count,omitempty"`
	// Metadata carries short audit keys and must not include prompt or response bodies.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// EditRequest carries one governed image editing request.
type EditRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aicommon.Tier `json:"tier"`
	// Image references the image to edit.
	Image aicommon.AssetRef `json:"image"`
	// Mask optionally references an edit mask.
	Mask *aicommon.AssetRef `json:"mask,omitempty"`
	// Prompt is the edit instruction.
	Prompt string `json:"prompt"`
	// Count caps the number of generated assets.
	Count int `json:"count,omitempty"`
	// Metadata carries short audit keys and must not include prompt or response bodies.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Response carries image assets or a provider operation reference.
type Response struct {
	// Assets contains generated or edited image asset projections.
	Assets []aicommon.AssetResult `json:"assets,omitempty"`
	// Operation contains provider async operation state when the result is not ready.
	Operation *aicommon.ProviderOperationRef `json:"operation,omitempty"`
	// Provider contains public provider/model identity.
	Provider aicommon.ProviderProjection `json:"provider"`
	// Usage contains minimal usage details.
	Usage aicommon.Usage `json:"usage,omitempty"`
	// LatencyMs is the provider call latency in milliseconds.
	LatencyMs int `json:"latencyMs,omitempty"`
	// CreatedAt is a Unix timestamp in milliseconds.
	CreatedAt int64 `json:"createdAt,omitempty"`
}

// ProviderRequest carries provider-internal image requests with governed source identity.
type ProviderRequest struct {
	// SourcePluginID identifies the dynamic or source plugin that initiated the call.
	SourcePluginID string `json:"sourcePluginId,omitempty"`
	// Generate contains the image generation request when Method is generate.
	Generate *GenerateRequest `json:"generate,omitempty"`
	// Edit contains the image editing request when Method is edit.
	Edit *EditRequest `json:"edit,omitempty"`
}

// Provider defines image AI capability implemented by provider plugins.
type Provider interface {
	// GenerateImage executes one governed image generation request.
	GenerateImage(ctx context.Context, request ProviderRequest) (*Response, error)
	// EditImage executes one governed image editing request.
	EditImage(ctx context.Context, request ProviderRequest) (*Response, error)
}

// Service defines the image AI capability consumed by host services and plugins.
type Service interface {
	// Available reports whether an active image provider is available.
	Available(ctx context.Context) bool
	// Status returns the current image AI capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// MethodStatus returns the current activation state for one image method.
	MethodStatus(ctx context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus
	// Generate executes one governed image generation request.
	Generate(ctx context.Context, request GenerateRequest) (*Response, error)
	// Edit executes one governed image editing request.
	Edit(ctx context.Context, request EditRequest) (*Response, error)
}

// serviceImpl is the fallback image service used until a provider plugin binds the method.
type serviceImpl struct {
	sourcePluginID string
}

var _ Service = (*serviceImpl)(nil)

// New creates a fallback image AI service.
func New() Service {
	return &serviceImpl{}
}

// ForPlugin returns a plugin-scoped image service.
func ForPlugin(service Service, pluginID string) Service {
	if service == nil {
		return &serviceImpl{sourcePluginID: strings.TrimSpace(pluginID)}
	}
	if _, ok := service.(*serviceImpl); ok {
		return &serviceImpl{sourcePluginID: strings.TrimSpace(pluginID)}
	}
	return service
}

// Available reports whether an active image provider is available.
func (*serviceImpl) Available(context.Context) bool {
	return false
}

// Status returns the fallback image capability status.
func (*serviceImpl) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(CapabilityAIImageV1)
}

// MethodStatus returns a fallback image method status.
func (*serviceImpl) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(CapabilityAIImageV1, CapabilityType, method)
}

// Generate validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Generate(_ context.Context, request GenerateRequest) (*Response, error) {
	if err := validatePurposeTier(request.Purpose, request.Tier); err != nil {
		return nil, err
	}
	return nil, unavailable(CapabilityMethodGenerate)
}

// Edit validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Edit(_ context.Context, request EditRequest) (*Response, error) {
	if err := validatePurposeTier(request.Purpose, request.Tier); err != nil {
		return nil, err
	}
	if strings.TrimSpace(request.Image.Ref) == "" {
		return nil, bizerr.NewCode(aicommon.CodeAssetRefRequired)
	}
	return nil, unavailable(CapabilityMethodEdit)
}

func validatePurposeTier(purpose string, tier aicommon.Tier) error {
	if strings.TrimSpace(purpose) == "" {
		return bizerr.NewCode(aicommon.CodePurposeRequired)
	}
	if !tier.Valid() {
		return bizerr.NewCode(aicommon.CodeTierInvalid, bizerr.P("tier", string(tier)))
	}
	return nil
}

func unavailable(method aicommon.CapabilityMethod) error {
	return bizerr.NewCode(
		aicommon.CodeProviderUnavailable,
		bizerr.P("capabilityType", string(CapabilityType)),
		bizerr.P("capabilityMethod", string(method)),
	)
}
