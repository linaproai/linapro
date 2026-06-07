// Package aicommon defines shared value objects used by typed AI capability
// subpackages. It intentionally contains no provider registry or service
// dispatch logic, so the root ai namespace can aggregate subpackages without
// import cycles.
package aicommon

import (
	"lina-core/pkg/plugin/capability/capmodel"
	"strings"
)

const (
	// CapabilityTypeText identifies text generation and related text-only methods.
	CapabilityTypeText CapabilityType = "text"
	// CapabilityTypeImage identifies image generation and editing methods.
	CapabilityTypeImage CapabilityType = "image"
	// CapabilityTypeEmbedding identifies vector embedding methods.
	CapabilityTypeEmbedding CapabilityType = "embedding"
	// CapabilityTypeAudio identifies audio transcription and synthesis methods.
	CapabilityTypeAudio CapabilityType = "audio"
	// CapabilityTypeVision identifies image, screenshot, and diagram analysis methods.
	CapabilityTypeVision CapabilityType = "vision"
	// CapabilityTypeDocument identifies document understanding and citation methods.
	CapabilityTypeDocument CapabilityType = "document"
	// CapabilityTypeSafety identifies safety moderation methods.
	CapabilityTypeSafety CapabilityType = "safety"
	// CapabilityTypeVideo identifies video generation, editing, extension, and operation methods.
	CapabilityTypeVideo CapabilityType = "video"
)

const (
	// CapabilityMethodTextGenerate identifies synchronous text generation.
	CapabilityMethodTextGenerate CapabilityMethod = "generate"
	// CapabilityMethodImageGenerate identifies image generation.
	CapabilityMethodImageGenerate CapabilityMethod = "generate"
	// CapabilityMethodImageEdit identifies image editing.
	CapabilityMethodImageEdit CapabilityMethod = "edit"
	// CapabilityMethodEmbeddingCreate identifies embedding creation.
	CapabilityMethodEmbeddingCreate CapabilityMethod = "create"
	// CapabilityMethodAudioTranscribe identifies speech-to-text transcription.
	CapabilityMethodAudioTranscribe CapabilityMethod = "transcribe"
	// CapabilityMethodAudioSynthesize identifies text-to-speech synthesis.
	CapabilityMethodAudioSynthesize CapabilityMethod = "synthesize"
	// CapabilityMethodVisionAnalyze identifies visual understanding.
	CapabilityMethodVisionAnalyze CapabilityMethod = "analyze"
	// CapabilityMethodDocumentAnalyze identifies document analysis.
	CapabilityMethodDocumentAnalyze CapabilityMethod = "analyze"
	// CapabilityMethodDocumentCite identifies citation-aware document analysis.
	CapabilityMethodDocumentCite CapabilityMethod = "cite"
	// CapabilityMethodSafetyModerate identifies safety moderation.
	CapabilityMethodSafetyModerate CapabilityMethod = "moderate"
	// CapabilityMethodVideoGenerate identifies video generation.
	CapabilityMethodVideoGenerate CapabilityMethod = "generate"
	// CapabilityMethodVideoEdit identifies video editing.
	CapabilityMethodVideoEdit CapabilityMethod = "edit"
	// CapabilityMethodVideoExtend identifies video extension.
	CapabilityMethodVideoExtend CapabilityMethod = "extend"
	// CapabilityMethodVideoOperationGet identifies provider operation status lookup.
	CapabilityMethodVideoOperationGet CapabilityMethod = "operation.get"
	// CapabilityMethodVideoOperationCancel identifies provider operation cancellation.
	CapabilityMethodVideoOperationCancel CapabilityMethod = "operation.cancel"
)

const (
	// TierBasic is the low-cost AI tier.
	TierBasic Tier = "basic"
	// TierStandard is the default AI tier.
	TierStandard Tier = "standard"
	// TierAdvanced is the high-capability AI tier.
	TierAdvanced Tier = "advanced"
)

const (
	// OperationStatusQueued reports that a provider operation has been accepted but not started.
	OperationStatusQueued OperationStatus = "queued"
	// OperationStatusRunning reports that a provider operation is in progress.
	OperationStatusRunning OperationStatus = "running"
	// OperationStatusSucceeded reports that a provider operation completed successfully.
	OperationStatusSucceeded OperationStatus = "succeeded"
	// OperationStatusFailed reports that a provider operation failed.
	OperationStatusFailed OperationStatus = "failed"
	// OperationStatusCanceled reports that a provider operation was canceled.
	OperationStatusCanceled OperationStatus = "canceled"
)

const reasonNoProvider = "no_provider"

// CapabilityType identifies one AI capability family.
type CapabilityType string

// CapabilityMethod identifies one method inside a capability family.
type CapabilityMethod string

// Tier identifies the governed platform service level requested by AI callers.
type Tier string

// OperationStatus identifies a provider operation lifecycle state.
type OperationStatus string

// AssetRef is an opaque governed reference to an input or output asset. It must
// not carry inline file bytes or provider-authenticated download URLs.
type AssetRef struct {
	// Ref is the opaque asset reference understood by a governed asset service.
	Ref string `json:"ref"`
	// Scope identifies the owner namespace, such as host, plugin, or temp.
	Scope string `json:"scope,omitempty"`
	// Purpose optionally records the governed business purpose for policy checks.
	Purpose string `json:"purpose,omitempty"`
	// MimeType is the projected MIME type supplied by the asset owner.
	MimeType string `json:"mimeType,omitempty"`
	// SizeBytes is the projected asset size in bytes.
	SizeBytes int64 `json:"sizeBytes,omitempty"`
}

// AssetResult describes a generated or derived asset without exposing content bytes.
type AssetResult struct {
	// AssetRef is the governed output asset reference.
	AssetRef AssetRef `json:"assetRef"`
	// MimeType is the output MIME type.
	MimeType string `json:"mimeType"`
	// SizeBytes is the output size in bytes when known.
	SizeBytes int64 `json:"sizeBytes,omitempty"`
	// Checksum is an optional content checksum owned by the asset service.
	Checksum string `json:"checksum,omitempty"`
	// Width is the output media width in pixels when applicable.
	Width int `json:"width,omitempty"`
	// Height is the output media height in pixels when applicable.
	Height int `json:"height,omitempty"`
	// DurationMs is the output media duration in milliseconds when applicable.
	DurationMs int64 `json:"durationMs,omitempty"`
	// CreatedAt is a Unix timestamp in milliseconds.
	CreatedAt int64 `json:"createdAt,omitempty"`
}

// ProviderProjection is the public provider/model identity snapshot returned
// by provider-backed AI calls.
type ProviderProjection struct {
	// ProviderName is the public provider display name.
	ProviderName string `json:"providerName,omitempty"`
	// ModelName is the public model display name or model identifier.
	ModelName string `json:"modelName,omitempty"`
	// Protocol is the provider protocol family used for the call.
	Protocol string `json:"protocol,omitempty"`
}

// ProviderOperationRef is an opaque provider protocol operation projection. It
// is not a business job identifier and must not leak provider credentials.
type ProviderOperationRef struct {
	// OperationRef is the opaque provider operation reference.
	OperationRef string `json:"operationRef"`
	// CapabilityType identifies the operation capability family.
	CapabilityType CapabilityType `json:"capabilityType"`
	// CapabilityMethod identifies the operation method.
	CapabilityMethod CapabilityMethod `json:"capabilityMethod"`
	// Provider contains public provider/model identity.
	Provider ProviderProjection `json:"provider"`
	// Status is the current provider operation status.
	Status OperationStatus `json:"status"`
	// NextPollAfterMs is the recommended delay before the next status query.
	NextPollAfterMs int64 `json:"nextPollAfterMs,omitempty"`
	// ExpiresAt is a Unix timestamp in milliseconds when the operation reference expires.
	ExpiresAt int64 `json:"expiresAt,omitempty"`
}

// Usage describes generic token, character, or media-unit usage summaries.
type Usage struct {
	// InputTokens is the input token count when available.
	InputTokens int `json:"inputTokens,omitempty"`
	// OutputTokens is the output token count when available.
	OutputTokens int `json:"outputTokens,omitempty"`
	// InputAssets is the number of referenced input assets.
	InputAssets int `json:"inputAssets,omitempty"`
	// OutputAssets is the number of generated output assets.
	OutputAssets int `json:"outputAssets,omitempty"`
}

// MethodStatus describes method-level availability without leaking provider internals.
type MethodStatus struct {
	// CapabilityType identifies the AI capability family.
	CapabilityType CapabilityType `json:"capabilityType"`
	// CapabilityMethod identifies the method inside the capability family.
	CapabilityMethod CapabilityMethod `json:"capabilityMethod"`
	// Available reports whether the method currently has an active provider.
	Available bool `json:"available"`
	// Reason is a stable diagnostic reason for unavailable status.
	Reason string `json:"reason,omitempty"`
	// CapabilityStatus is the underlying framework capability status projection.
	CapabilityStatus capmodel.CapabilityStatus `json:"capabilityStatus"`
}

// Valid reports whether the tier is one of the stable platform tiers.
func (t Tier) Valid() bool {
	switch t {
	case TierBasic, TierStandard, TierAdvanced:
		return true
	default:
		return false
	}
}

// Valid reports whether the operation status is supported by the common contract.
func (s OperationStatus) Valid() bool {
	switch s {
	case OperationStatusQueued, OperationStatusRunning, OperationStatusSucceeded, OperationStatusFailed, OperationStatusCanceled:
		return true
	default:
		return false
	}
}

// PurposeResourceRef builds the governed host-service resource reference for a purpose.
func PurposeResourceRef(purpose string) string {
	trimmed := strings.TrimSpace(purpose)
	if trimmed == "" {
		return ""
	}
	return "purpose:" + trimmed
}

// UnavailableStatus returns a common unavailable capability projection.
func UnavailableStatus(capabilityID string) capmodel.CapabilityStatus {
	return capmodel.CapabilityStatus{
		CapabilityID: capabilityID,
		Available:    false,
		Reason:       reasonNoProvider,
	}
}

// UnavailableMethodStatus returns a common unavailable method projection.
func UnavailableMethodStatus(capabilityID string, capabilityType CapabilityType, capabilityMethod CapabilityMethod) MethodStatus {
	status := UnavailableStatus(capabilityID)
	return MethodStatus{
		CapabilityType:   capabilityType,
		CapabilityMethod: capabilityMethod,
		Available:        false,
		Reason:           status.Reason,
		CapabilityStatus: status,
	}
}
