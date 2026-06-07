// Package aiembedding defines the typed embedding AI capability contract
// exposed under AI().Embedding().
package aiembedding

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	// CapabilityAIEmbeddingV1 identifies the versioned embedding AI framework capability.
	CapabilityAIEmbeddingV1 = "framework.ai.embedding.v1"
	// CapabilityType identifies the embedding capability family.
	CapabilityType = aicommon.CapabilityTypeEmbedding
	// CapabilityMethodCreate identifies embedding creation.
	CapabilityMethodCreate = aicommon.CapabilityMethodEmbeddingCreate
)

// Input carries one text or asset embedding input.
type Input struct {
	// Text is the text input when embedding text.
	Text string `json:"text,omitempty"`
	// AssetRef references an input asset when embedding non-text content.
	AssetRef *aicommon.AssetRef `json:"assetRef,omitempty"`
	// MimeType is the projected MIME type for asset inputs.
	MimeType string `json:"mimeType,omitempty"`
}

// CreateRequest carries one governed embedding request.
type CreateRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aicommon.Tier `json:"tier"`
	// Inputs contains bounded embedding inputs.
	Inputs []Input `json:"inputs"`
	// Dimensions optionally requests an embedding dimensionality.
	Dimensions int `json:"dimensions,omitempty"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Embedding contains one vector result.
type Embedding struct {
	// Index is the input index.
	Index int `json:"index"`
	// Values contains vector values returned by the provider.
	Values []float32 `json:"values"`
}

// CreateResponse carries embedding results and provider identity.
type CreateResponse struct {
	// Embeddings contains ordered vector results.
	Embeddings []Embedding `json:"embeddings"`
	// Provider contains public provider/model identity.
	Provider aicommon.ProviderProjection `json:"provider"`
	// Usage contains minimal usage details.
	Usage aicommon.Usage `json:"usage,omitempty"`
	// LatencyMs is the provider call latency in milliseconds.
	LatencyMs int `json:"latencyMs,omitempty"`
	// CreatedAt is a Unix timestamp in milliseconds.
	CreatedAt int64 `json:"createdAt,omitempty"`
}

// ProviderRequest carries provider-internal embedding requests with source identity.
type ProviderRequest struct {
	CreateRequest
	// SourcePluginID identifies the dynamic or source plugin that initiated the call.
	SourcePluginID string `json:"sourcePluginId,omitempty"`
}

// Provider defines embedding AI capability implemented by provider plugins.
type Provider interface {
	// CreateEmbedding executes one governed embedding request.
	CreateEmbedding(ctx context.Context, request ProviderRequest) (*CreateResponse, error)
}

// Service defines the embedding AI capability consumed by host services and plugins.
type Service interface {
	// Available reports whether an active embedding provider is available.
	Available(ctx context.Context) bool
	// Status returns the current embedding AI capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// MethodStatus returns the current activation state for one embedding method.
	MethodStatus(ctx context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus
	// Create executes one governed embedding request.
	Create(ctx context.Context, request CreateRequest) (*CreateResponse, error)
}

type serviceImpl struct {
	sourcePluginID string
}

var _ Service = (*serviceImpl)(nil)

// New creates a fallback embedding AI service.
func New() Service { return &serviceImpl{} }

// ForPlugin returns a plugin-scoped embedding service.
func ForPlugin(service Service, pluginID string) Service {
	if _, ok := service.(*serviceImpl); service == nil || ok {
		return &serviceImpl{sourcePluginID: strings.TrimSpace(pluginID)}
	}
	return service
}

// Available reports whether an active embedding provider is available.
func (*serviceImpl) Available(context.Context) bool { return false }

// Status returns the fallback embedding capability status.
func (*serviceImpl) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(CapabilityAIEmbeddingV1)
}

// MethodStatus returns a fallback embedding method status.
func (*serviceImpl) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(CapabilityAIEmbeddingV1, CapabilityType, method)
}

// Create validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Create(_ context.Context, request CreateRequest) (*CreateResponse, error) {
	if strings.TrimSpace(request.Purpose) == "" {
		return nil, bizerr.NewCode(aicommon.CodePurposeRequired)
	}
	if !request.Tier.Valid() {
		return nil, bizerr.NewCode(aicommon.CodeTierInvalid, bizerr.P("tier", string(request.Tier)))
	}
	if len(request.Inputs) == 0 {
		return nil, bizerr.NewCode(aicommon.CodeAssetRefRequired)
	}
	return nil, bizerr.NewCode(
		aicommon.CodeProviderUnavailable,
		bizerr.P("capabilityType", string(CapabilityType)),
		bizerr.P("capabilityMethod", string(CapabilityMethodCreate)),
	)
}
