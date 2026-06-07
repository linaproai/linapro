// Package aidocument defines the typed document AI capability contract exposed
// under AI().Document().
package aidocument

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	// CapabilityAIDocumentV1 identifies the versioned document AI framework capability.
	CapabilityAIDocumentV1 = "framework.ai.document.v1"
	// CapabilityType identifies the document capability family.
	CapabilityType = aicommon.CapabilityTypeDocument
	// CapabilityMethodAnalyze identifies document analysis.
	CapabilityMethodAnalyze = aicommon.CapabilityMethodDocumentAnalyze
	// CapabilityMethodCite identifies citation-aware document analysis.
	CapabilityMethodCite = aicommon.CapabilityMethodDocumentCite
)

// AnalyzeRequest carries one governed document analysis request.
type AnalyzeRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aicommon.Tier `json:"tier"`
	// Documents references input documents.
	Documents []aicommon.AssetRef `json:"documents"`
	// Prompt is the analysis instruction.
	Prompt string `json:"prompt"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CiteRequest carries one governed citation-aware document request.
type CiteRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aicommon.Tier `json:"tier"`
	// Documents references input documents.
	Documents []aicommon.AssetRef `json:"documents"`
	// Question is the citation-aware question.
	Question string `json:"question"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Citation describes one cited document span without leaking document content.
type Citation struct {
	// DocumentRef references the cited document.
	DocumentRef string `json:"documentRef"`
	// Page optionally identifies the page number.
	Page int `json:"page,omitempty"`
	// Locator optionally stores a provider-neutral span locator.
	Locator string `json:"locator,omitempty"`
}

// Response carries document analysis text, citations, and provider identity.
type Response struct {
	// Text is the document analysis result.
	Text string `json:"text"`
	// Citations contains citation projections when available.
	Citations []Citation `json:"citations,omitempty"`
	// Provider contains public provider/model identity.
	Provider aicommon.ProviderProjection `json:"provider"`
	// Usage contains minimal usage details.
	Usage aicommon.Usage `json:"usage,omitempty"`
	// LatencyMs is the provider call latency in milliseconds.
	LatencyMs int `json:"latencyMs,omitempty"`
	// CreatedAt is a Unix timestamp in milliseconds.
	CreatedAt int64 `json:"createdAt,omitempty"`
}

// ProviderRequest carries provider-internal document requests with source identity.
type ProviderRequest struct {
	// SourcePluginID identifies the dynamic or source plugin that initiated the call.
	SourcePluginID string `json:"sourcePluginId,omitempty"`
	// Analyze contains the document analysis request when Method is analyze.
	Analyze *AnalyzeRequest `json:"analyze,omitempty"`
	// Cite contains the citation-aware request when Method is cite.
	Cite *CiteRequest `json:"cite,omitempty"`
}

// Provider defines document AI capability implemented by provider plugins.
type Provider interface {
	// AnalyzeDocument executes one governed document analysis request.
	AnalyzeDocument(ctx context.Context, request ProviderRequest) (*Response, error)
	// CiteDocument executes one governed citation-aware document request.
	CiteDocument(ctx context.Context, request ProviderRequest) (*Response, error)
}

// Service defines the document AI capability consumed by host services and plugins.
type Service interface {
	// Available reports whether an active document provider is available.
	Available(ctx context.Context) bool
	// Status returns the current document AI capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// MethodStatus returns the current activation state for one document method.
	MethodStatus(ctx context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus
	// Analyze executes one governed document analysis request.
	Analyze(ctx context.Context, request AnalyzeRequest) (*Response, error)
	// Cite executes one governed citation-aware document request.
	Cite(ctx context.Context, request CiteRequest) (*Response, error)
}

type serviceImpl struct {
	sourcePluginID string
}

var _ Service = (*serviceImpl)(nil)

// New creates a fallback document AI service.
func New() Service { return &serviceImpl{} }

// ForPlugin returns a plugin-scoped document service.
func ForPlugin(service Service, pluginID string) Service {
	if _, ok := service.(*serviceImpl); service == nil || ok {
		return &serviceImpl{sourcePluginID: strings.TrimSpace(pluginID)}
	}
	return service
}

// Available reports whether an active document provider is available.
func (*serviceImpl) Available(context.Context) bool { return false }

// Status returns the fallback document capability status.
func (*serviceImpl) Status(context.Context) capmodel.CapabilityStatus {
	return aicommon.UnavailableStatus(CapabilityAIDocumentV1)
}

// MethodStatus returns a fallback document method status.
func (*serviceImpl) MethodStatus(_ context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus {
	return aicommon.UnavailableMethodStatus(CapabilityAIDocumentV1, CapabilityType, method)
}

// Analyze validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Analyze(_ context.Context, request AnalyzeRequest) (*Response, error) {
	if err := validateDocumentBoundary(request.Purpose, request.Tier, request.Documents); err != nil {
		return nil, err
	}
	return nil, unavailable(CapabilityMethodAnalyze)
}

// Cite validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Cite(_ context.Context, request CiteRequest) (*Response, error) {
	if err := validateDocumentBoundary(request.Purpose, request.Tier, request.Documents); err != nil {
		return nil, err
	}
	return nil, unavailable(CapabilityMethodCite)
}

func validateDocumentBoundary(purpose string, tier aicommon.Tier, documents []aicommon.AssetRef) error {
	if strings.TrimSpace(purpose) == "" {
		return bizerr.NewCode(aicommon.CodePurposeRequired)
	}
	if !tier.Valid() {
		return bizerr.NewCode(aicommon.CodeTierInvalid, bizerr.P("tier", string(tier)))
	}
	if len(documents) == 0 || strings.TrimSpace(documents[0].Ref) == "" {
		return bizerr.NewCode(aicommon.CodeAssetRefRequired)
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
