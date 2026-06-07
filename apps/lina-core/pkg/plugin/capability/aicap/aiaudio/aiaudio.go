// Package aiaudio defines the typed audio AI capability contract exposed under
// AI().Audio().
package aiaudio

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	// CapabilityAIAudioV1 identifies the versioned audio AI framework capability.
	CapabilityAIAudioV1 = "framework.ai.audio.v1"
	// CapabilityType identifies the audio capability family.
	CapabilityType = aicommon.CapabilityTypeAudio
	// CapabilityMethodTranscribe identifies speech-to-text transcription.
	CapabilityMethodTranscribe = aicommon.CapabilityMethodAudioTranscribe
	// CapabilityMethodSynthesize identifies text-to-speech synthesis.
	CapabilityMethodSynthesize = aicommon.CapabilityMethodAudioSynthesize
)

// TranscribeRequest carries one governed audio transcription request.
type TranscribeRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aicommon.Tier `json:"tier"`
	// Audio references the audio asset to transcribe.
	Audio aicommon.AssetRef `json:"audio"`
	// Language optionally hints the spoken language.
	Language string `json:"language,omitempty"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// TranscribeResponse carries a transcript and provider identity.
type TranscribeResponse struct {
	// Text is the transcript text.
	Text string `json:"text"`
	// Language is the detected or applied language.
	Language string `json:"language,omitempty"`
	// Provider contains public provider/model identity.
	Provider aicommon.ProviderProjection `json:"provider"`
	// Usage contains minimal usage details.
	Usage aicommon.Usage `json:"usage,omitempty"`
	// LatencyMs is the provider call latency in milliseconds.
	LatencyMs int `json:"latencyMs,omitempty"`
	// TranscribedAt is a Unix timestamp in milliseconds.
	TranscribedAt int64 `json:"transcribedAt,omitempty"`
}

// SynthesizeRequest carries one governed audio synthesis request.
type SynthesizeRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aicommon.Tier `json:"tier"`
	// Text is the text to synthesize.
	Text string `json:"text"`
	// Voice optionally selects a provider-supported voice.
	Voice string `json:"voice,omitempty"`
	// Format optionally selects an output audio format.
	Format string `json:"format,omitempty"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SynthesizeResponse carries a synthesized audio asset or provider operation.
type SynthesizeResponse struct {
	// Asset contains the output audio reference when available.
	Asset *aicommon.AssetResult `json:"asset,omitempty"`
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

// ProviderRequest carries provider-internal audio requests with source identity.
type ProviderRequest struct {
	// SourcePluginID identifies the dynamic or source plugin that initiated the call.
	SourcePluginID string `json:"sourcePluginId,omitempty"`
	// Transcribe contains the transcription request when Method is transcribe.
	Transcribe *TranscribeRequest `json:"transcribe,omitempty"`
	// Synthesize contains the synthesis request when Method is synthesize.
	Synthesize *SynthesizeRequest `json:"synthesize,omitempty"`
}

// Provider defines audio AI capability implemented by provider plugins.
type Provider interface {
	// TranscribeAudio executes one governed transcription request.
	TranscribeAudio(ctx context.Context, request ProviderRequest) (*TranscribeResponse, error)
	// SynthesizeAudio executes one governed synthesis request.
	SynthesizeAudio(ctx context.Context, request ProviderRequest) (*SynthesizeResponse, error)
}

// Service defines the audio AI capability consumed by host services and plugins.
type Service interface {
	// Available reports whether an active audio provider is available.
	Available(ctx context.Context) bool
	// Status returns the current audio AI capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// MethodStatus returns the current activation state for one audio method.
	MethodStatus(ctx context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus
	// Transcribe executes one governed audio transcription request.
	Transcribe(ctx context.Context, request TranscribeRequest) (*TranscribeResponse, error)
	// Synthesize executes one governed audio synthesis request.
	Synthesize(ctx context.Context, request SynthesizeRequest) (*SynthesizeResponse, error)
}

type serviceImpl struct {
	sourcePluginID string
}

var _ Service = (*serviceImpl)(nil)

// New creates a fallback audio AI service.
func New() Service { return &serviceImpl{} }

// ForPlugin returns a plugin-scoped audio service.
func ForPlugin(service Service, pluginID string) Service {
	if _, ok := service.(*serviceImpl); service == nil || ok {
		return &serviceImpl{sourcePluginID: strings.TrimSpace(pluginID)}
	}
	return service
}

// Available reports whether an active audio provider is available.
func (*serviceImpl) Available(context.Context) bool { return false }

// Status returns the fallback audio capability status.
func (*serviceImpl) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(CapabilityAIAudioV1)
}

// MethodStatus returns a fallback audio method status.
func (*serviceImpl) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(CapabilityAIAudioV1, CapabilityType, method)
}

// Transcribe validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Transcribe(_ context.Context, request TranscribeRequest) (*TranscribeResponse, error) {
	if err := validatePurposeTier(request.Purpose, request.Tier); err != nil {
		return nil, err
	}
	if strings.TrimSpace(request.Audio.Ref) == "" {
		return nil, bizerr.NewCode(aicommon.CodeAssetRefRequired)
	}
	return nil, unavailable(CapabilityMethodTranscribe)
}

// Synthesize validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Synthesize(_ context.Context, request SynthesizeRequest) (*SynthesizeResponse, error) {
	if err := validatePurposeTier(request.Purpose, request.Tier); err != nil {
		return nil, err
	}
	if strings.TrimSpace(request.Text) == "" {
		return nil, bizerr.NewCode(aicommon.CodeAssetRefRequired)
	}
	return nil, unavailable(CapabilityMethodSynthesize)
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
