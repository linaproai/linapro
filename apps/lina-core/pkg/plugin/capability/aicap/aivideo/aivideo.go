// Package aivideo defines the typed video AI capability contract exposed under
// AI().Video(). Provider operations are protocol references only, not business
// jobs or user-facing progress records.
package aivideo

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	// CapabilityAIVideoV1 identifies the versioned video AI framework capability.
	CapabilityAIVideoV1 = "framework.ai.video.v1"
	// CapabilityType identifies the video capability family.
	CapabilityType = aicommon.CapabilityTypeVideo
	// CapabilityMethodGenerate identifies video generation.
	CapabilityMethodGenerate = aicommon.CapabilityMethodVideoGenerate
	// CapabilityMethodEdit identifies video editing.
	CapabilityMethodEdit = aicommon.CapabilityMethodVideoEdit
	// CapabilityMethodExtend identifies video extension.
	CapabilityMethodExtend = aicommon.CapabilityMethodVideoExtend
	// CapabilityMethodOperationGet identifies provider operation status lookup.
	CapabilityMethodOperationGet = aicommon.CapabilityMethodVideoOperationGet
	// CapabilityMethodOperationCancel identifies provider operation cancellation.
	CapabilityMethodOperationCancel = aicommon.CapabilityMethodVideoOperationCancel
)

// GenerateRequest carries one governed video generation request.
type GenerateRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aicommon.Tier `json:"tier"`
	// Prompt is the video generation instruction.
	Prompt string `json:"prompt"`
	// DurationMs optionally requests output duration in milliseconds.
	DurationMs int64 `json:"durationMs,omitempty"`
	// AspectRatio optionally requests a provider-supported aspect ratio.
	AspectRatio string `json:"aspectRatio,omitempty"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// EditRequest carries one governed video editing request.
type EditRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aicommon.Tier `json:"tier"`
	// Video references the video to edit.
	Video aicommon.AssetRef `json:"video"`
	// Prompt is the edit instruction.
	Prompt string `json:"prompt"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ExtendRequest carries one governed video extension request.
type ExtendRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aicommon.Tier `json:"tier"`
	// Video references the video to extend.
	Video aicommon.AssetRef `json:"video"`
	// Prompt is the extension instruction.
	Prompt string `json:"prompt,omitempty"`
	// DurationMs optionally requests additional duration in milliseconds.
	DurationMs int64 `json:"durationMs,omitempty"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// OperationGetRequest carries one provider operation status lookup.
type OperationGetRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// OperationRef is the opaque provider operation reference.
	OperationRef string `json:"operationRef"`
}

// OperationCancelRequest carries one provider operation cancel request.
type OperationCancelRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// OperationRef is the opaque provider operation reference.
	OperationRef string `json:"operationRef"`
}

// Response carries video assets or a provider operation reference.
type Response struct {
	// Assets contains output video assets when available.
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

// ProviderRequest carries provider-internal video requests with source identity.
type ProviderRequest struct {
	// SourcePluginID identifies the dynamic or source plugin that initiated the call.
	SourcePluginID string `json:"sourcePluginId,omitempty"`
	// Generate contains the video generation request when Method is generate.
	Generate *GenerateRequest `json:"generate,omitempty"`
	// Edit contains the video edit request when Method is edit.
	Edit *EditRequest `json:"edit,omitempty"`
	// Extend contains the video extension request when Method is extend.
	Extend *ExtendRequest `json:"extend,omitempty"`
}

// Provider defines video AI capability implemented by provider plugins.
type Provider interface {
	// GenerateVideo executes one governed video generation request.
	GenerateVideo(ctx context.Context, request ProviderRequest) (*Response, error)
	// EditVideo executes one governed video editing request.
	EditVideo(ctx context.Context, request ProviderRequest) (*Response, error)
	// ExtendVideo executes one governed video extension request.
	ExtendVideo(ctx context.Context, request ProviderRequest) (*Response, error)
	// GetOperation returns one provider operation projection.
	GetOperation(ctx context.Context, request OperationGetRequest) (*Response, error)
	// CancelOperation cancels one provider operation when authorized and supported.
	CancelOperation(ctx context.Context, request OperationCancelRequest) (*aicommon.ProviderOperationRef, error)
}

// Service defines the video AI capability consumed by host services and plugins.
type Service interface {
	// Available reports whether an active video provider is available.
	Available(ctx context.Context) bool
	// Status returns the current video AI capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// MethodStatus returns the current activation state for one video method.
	MethodStatus(ctx context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus
	// Generate executes one governed video generation request.
	Generate(ctx context.Context, request GenerateRequest) (*Response, error)
	// Edit executes one governed video editing request.
	Edit(ctx context.Context, request EditRequest) (*Response, error)
	// Extend executes one governed video extension request.
	Extend(ctx context.Context, request ExtendRequest) (*Response, error)
	// OperationGet returns one provider operation projection.
	OperationGet(ctx context.Context, request OperationGetRequest) (*Response, error)
	// OperationCancel cancels one provider operation when authorized and supported.
	OperationCancel(ctx context.Context, request OperationCancelRequest) (*aicommon.ProviderOperationRef, error)
}

type serviceImpl struct {
	sourcePluginID string
}

var _ Service = (*serviceImpl)(nil)

// New creates a fallback video AI service.
func New() Service { return &serviceImpl{} }

// ForPlugin returns a plugin-scoped video service.
func ForPlugin(service Service, pluginID string) Service {
	if _, ok := service.(*serviceImpl); service == nil || ok {
		return &serviceImpl{sourcePluginID: strings.TrimSpace(pluginID)}
	}
	return service
}

// Available reports whether an active video provider is available.
func (*serviceImpl) Available(context.Context) bool { return false }

// Status returns the fallback video capability status.
func (*serviceImpl) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(CapabilityAIVideoV1)
}

// MethodStatus returns a fallback video method status.
func (*serviceImpl) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(CapabilityAIVideoV1, CapabilityType, method)
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
	if strings.TrimSpace(request.Video.Ref) == "" {
		return nil, bizerr.NewCode(aicommon.CodeAssetRefRequired)
	}
	return nil, unavailable(CapabilityMethodEdit)
}

// Extend validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Extend(_ context.Context, request ExtendRequest) (*Response, error) {
	if err := validatePurposeTier(request.Purpose, request.Tier); err != nil {
		return nil, err
	}
	if strings.TrimSpace(request.Video.Ref) == "" {
		return nil, bizerr.NewCode(aicommon.CodeAssetRefRequired)
	}
	return nil, unavailable(CapabilityMethodExtend)
}

// OperationGet validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) OperationGet(_ context.Context, request OperationGetRequest) (*Response, error) {
	if strings.TrimSpace(request.Purpose) == "" {
		return nil, bizerr.NewCode(aicommon.CodePurposeRequired)
	}
	if strings.TrimSpace(request.OperationRef) == "" {
		return nil, bizerr.NewCode(aicommon.CodeOperationRefRequired)
	}
	return nil, unavailable(CapabilityMethodOperationGet)
}

// OperationCancel validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) OperationCancel(_ context.Context, request OperationCancelRequest) (*aicommon.ProviderOperationRef, error) {
	if strings.TrimSpace(request.Purpose) == "" {
		return nil, bizerr.NewCode(aicommon.CodePurposeRequired)
	}
	if strings.TrimSpace(request.OperationRef) == "" {
		return nil, bizerr.NewCode(aicommon.CodeOperationRefRequired)
	}
	return nil, unavailable(CapabilityMethodOperationCancel)
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
