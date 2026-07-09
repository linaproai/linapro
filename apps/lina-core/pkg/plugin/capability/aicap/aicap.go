// Package aicap defines the host AI capability namespace exposed through the
// plugin capability directory. The package only aggregates typed AI sub
// capabilities; each sub capability keeps its own DTOs, status, fallback, and
// provider boundary when the corresponding host provider SPI exists.
package aicap

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/aicap/aiaudio"
	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/aicap/aidocument"
	"lina-core/pkg/plugin/capability/aicap/aiembedding"
	"lina-core/pkg/plugin/capability/aicap/aiimage"
	"lina-core/pkg/plugin/capability/aicap/aisafety"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/aicap/aivideo"
	"lina-core/pkg/plugin/capability/aicap/aivision"
	"lina-core/pkg/plugin/capability/capmodel"
)

// Service aggregates typed AI sub capabilities under one stable namespace.
type Service interface {
	// MethodStatuses returns method-level availability for AI sub capabilities.
	MethodStatuses(ctx context.Context, input MethodStatusesInput) (*MethodStatusesResult, error)
	// Text returns the text AI capability service.
	Text() aitext.Service
	// Image returns the image AI capability service.
	Image() aiimage.Service
	// Embedding returns the embedding AI capability service.
	Embedding() aiembedding.Service
	// Audio returns the audio AI capability service.
	Audio() aiaudio.Service
	// Vision returns the vision AI capability service.
	Vision() aivision.Service
	// Document returns the document AI capability service.
	Document() aidocument.Service
	// Safety returns the safety AI capability service.
	Safety() aisafety.Service
	// Video returns the video AI capability service.
	Video() aivideo.Service
}

type (
	// AssetRef is an opaque governed reference to an input or output asset.
	AssetRef = aicommon.AssetRef
	// AssetResult describes a generated or derived asset without content bytes.
	AssetResult = aicommon.AssetResult
	// ProviderOperationRef is an opaque provider protocol operation reference.
	ProviderOperationRef = aicommon.ProviderOperationRef
	// ProviderInfo is the public provider/model identity snapshot.
	ProviderInfo = aicommon.ProviderInfo
	// CapabilityType identifies one AI capability family.
	CapabilityType = aicommon.CapabilityType
	// CapabilityMethod identifies one method inside a capability family.
	CapabilityMethod = aicommon.CapabilityMethod
	// Tier identifies the governed platform service level requested by AI callers.
	Tier = aicommon.Tier
	// MethodStatus describes method-level availability without leaking provider internals.
	MethodStatus = aicommon.MethodStatus
)

const (
	// CapabilityTypeText identifies text generation and related text-only methods.
	CapabilityTypeText = aicommon.CapabilityTypeText
	// CapabilityTypeImage identifies image generation and editing methods.
	CapabilityTypeImage = aicommon.CapabilityTypeImage
	// CapabilityTypeEmbedding identifies vector embedding methods.
	CapabilityTypeEmbedding = aicommon.CapabilityTypeEmbedding
	// CapabilityTypeAudio identifies audio transcription and synthesis methods.
	CapabilityTypeAudio = aicommon.CapabilityTypeAudio
	// CapabilityTypeVision identifies image, screenshot, and diagram analysis methods.
	CapabilityTypeVision = aicommon.CapabilityTypeVision
	// CapabilityTypeDocument identifies document understanding and citation methods.
	CapabilityTypeDocument = aicommon.CapabilityTypeDocument
	// CapabilityTypeSafety identifies safety moderation methods.
	CapabilityTypeSafety = aicommon.CapabilityTypeSafety
	// CapabilityTypeVideo identifies video generation, editing, extension, and operation methods.
	CapabilityTypeVideo = aicommon.CapabilityTypeVideo

	// TierBasic is the low-cost AI tier.
	TierBasic = aicommon.TierBasic
	// TierStandard is the default AI tier.
	TierStandard = aicommon.TierStandard
	// TierAdvanced is the high-capability AI tier.
	TierAdvanced = aicommon.TierAdvanced
)

const (
	// MaxMethodStatusBatchSize limits cross-sub-capability AI status reads.
	MaxMethodStatusBatchSize = 100
)

// MethodStatusQuery identifies one AI sub-capability method status to read.
type MethodStatusQuery struct {
	// CapabilityType identifies the AI capability family.
	CapabilityType CapabilityType `json:"capabilityType"`
	// CapabilityMethod identifies the method inside the capability family.
	CapabilityMethod CapabilityMethod `json:"capabilityMethod"`
}

// MethodStatusesInput carries one bounded AI method status batch request.
type MethodStatusesInput struct {
	// Methods contains AI method status queries.
	Methods []MethodStatusQuery `json:"methods"`
}

// MethodStatusesResult carries AI method statuses in request order.
type MethodStatusesResult struct {
	// Items contains one status per requested method.
	Items []MethodStatus `json:"items"`
}

// serviceImpl stores typed AI sub capability services.
type serviceImpl struct {
	text      aitext.Service
	image     aiimage.Service
	embedding aiembedding.Service
	audio     aiaudio.Service
	vision    aivision.Service
	document  aidocument.Service
	safety    aisafety.Service
	video     aivideo.Service
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// MethodStatuses returns method-level availability for AI sub capabilities.
func (s *serviceImpl) MethodStatuses(ctx context.Context, input MethodStatusesInput) (*MethodStatusesResult, error) {
	if len(input.Methods) > MaxMethodStatusBatchSize {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", MaxMethodStatusBatchSize))
	}
	result := &MethodStatusesResult{Items: make([]MethodStatus, 0, len(input.Methods))}
	for _, query := range input.Methods {
		result.Items = append(result.Items, s.methodStatus(ctx, query))
	}
	return result, nil
}

// methodStatus returns one sub-capability status without exposing provider internals.
func (s *serviceImpl) methodStatus(ctx context.Context, query MethodStatusQuery) MethodStatus {
	method := aicommon.CapabilityMethod(query.CapabilityMethod)
	switch aicommon.CapabilityType(query.CapabilityType) {
	case aicommon.CapabilityTypeText:
		return s.Text().MethodStatus(ctx, method)
	case aicommon.CapabilityTypeImage:
		return s.Image().MethodStatus(ctx, method)
	case aicommon.CapabilityTypeEmbedding:
		return s.Embedding().MethodStatus(ctx, method)
	case aicommon.CapabilityTypeAudio:
		return s.Audio().MethodStatus(ctx, method)
	case aicommon.CapabilityTypeVision:
		return s.Vision().MethodStatus(ctx, method)
	case aicommon.CapabilityTypeDocument:
		return s.Document().MethodStatus(ctx, method)
	case aicommon.CapabilityTypeSafety:
		return s.Safety().MethodStatus(ctx, method)
	case aicommon.CapabilityTypeVideo:
		return s.Video().MethodStatus(ctx, method)
	default:
		return aicommon.UnavailableMethodStatus("", aicommon.CapabilityType(query.CapabilityType), method)
	}
}

// New creates an AI namespace service from the text AI capability service.
func New(text aitext.Service) Service {
	if text == nil {
		text = aitext.New(nil, nil, nil)
	}
	service := &serviceImpl{text: text}
	service.ensureFallbacks()
	return service
}

// newWithSubCapabilities creates an AI namespace with runtime-owned sub capability services.
func newWithSubCapabilities(
	text aitext.Service,
	image aiimage.Service,
	embedding aiembedding.Service,
	audio aiaudio.Service,
	vision aivision.Service,
	document aidocument.Service,
	safety aisafety.Service,
	video aivideo.Service,
) Service {
	if text == nil {
		text = aitext.New(nil, nil, nil)
	}
	service := &serviceImpl{
		text:      text,
		image:     image,
		embedding: embedding,
		audio:     audio,
		vision:    vision,
		document:  document,
		safety:    safety,
		video:     video,
	}
	service.ensureFallbacks()
	return service
}

// Text returns the text AI capability service.
func (s *serviceImpl) Text() aitext.Service {
	if s == nil || s.text == nil {
		return aitext.New(nil, nil, nil)
	}
	return s.text
}

// Image returns the image AI capability service.
func (s *serviceImpl) Image() aiimage.Service {
	if s == nil || s.image == nil {
		return aiimage.New()
	}
	return s.image
}

// Embedding returns the embedding AI capability service.
func (s *serviceImpl) Embedding() aiembedding.Service {
	if s == nil || s.embedding == nil {
		return aiembedding.New()
	}
	return s.embedding
}

// Audio returns the audio AI capability service.
func (s *serviceImpl) Audio() aiaudio.Service {
	if s == nil || s.audio == nil {
		return aiaudio.New()
	}
	return s.audio
}

// Vision returns the vision AI capability service.
func (s *serviceImpl) Vision() aivision.Service {
	if s == nil || s.vision == nil {
		return aivision.New()
	}
	return s.vision
}

// Document returns the document AI capability service.
func (s *serviceImpl) Document() aidocument.Service {
	if s == nil || s.document == nil {
		return aidocument.New()
	}
	return s.document
}

// Safety returns the safety AI capability service.
func (s *serviceImpl) Safety() aisafety.Service {
	if s == nil || s.safety == nil {
		return aisafety.New()
	}
	return s.safety
}

// Video returns the video AI capability service.
func (s *serviceImpl) Video() aivideo.Service {
	if s == nil || s.video == nil {
		return aivideo.New()
	}
	return s.video
}

// ForPlugin returns a plugin-scoped AI namespace service while preserving the
// runtime-owned AI sub capability implementations. The scoped namespace binds
// pluginID to downstream AI provider requests through each sub capability, so
// source plugins and dynamic plugins can consume host AI services without
// manually supplying or spoofing the caller identity.
//
// This host-injected source identity is important for AI invocation audit,
// usage attribution, troubleshooting, and future plugin-level governance such
// as quota, rate limit, tier access, or purpose policy decisions. When service
// is nil, the returned namespace still exposes fallback sub capabilities so
// callers receive structured unavailable errors instead of nil services.
func ForPlugin(service Service, pluginID string) Service {
	if service == nil {
		return New(aitext.ForPlugin(nil, pluginID))
	}
	return newWithSubCapabilities(
		aitext.ForPlugin(service.Text(), pluginID),
		aiimage.ForPlugin(service.Image(), pluginID),
		aiembedding.ForPlugin(service.Embedding(), pluginID),
		aiaudio.ForPlugin(service.Audio(), pluginID),
		aivision.ForPlugin(service.Vision(), pluginID),
		aidocument.ForPlugin(service.Document(), pluginID),
		aisafety.ForPlugin(service.Safety(), pluginID),
		aivideo.ForPlugin(service.Video(), pluginID),
	)
}

// ensureFallbacks fills every typed sub capability with a safe fallback service.
func (s *serviceImpl) ensureFallbacks() {
	if s == nil {
		return
	}
	if s.image == nil {
		s.image = aiimage.New()
	}
	if s.embedding == nil {
		s.embedding = aiembedding.New()
	}
	if s.audio == nil {
		s.audio = aiaudio.New()
	}
	if s.vision == nil {
		s.vision = aivision.New()
	}
	if s.document == nil {
		s.document = aidocument.New()
	}
	if s.safety == nil {
		s.safety = aisafety.New()
	}
	if s.video == nil {
		s.video = aivideo.New()
	}
}
