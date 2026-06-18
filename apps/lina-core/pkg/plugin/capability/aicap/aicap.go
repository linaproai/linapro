// Package aicap defines the host AI capability namespace exposed through the
// plugin capability directory. The package only aggregates typed AI sub
// capabilities; each sub capability keeps its own DTOs, status, fallback, and
// provider contract.
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

type (
	// AssetRef is an opaque governed reference to an input or output asset.
	AssetRef = aicommon.AssetRef
	// AssetResult describes a generated or derived asset without content bytes.
	AssetResult = aicommon.AssetResult
	// ProviderOperationRef is an opaque provider protocol operation projection.
	ProviderOperationRef = aicommon.ProviderOperationRef
	// ProviderProjection is the public provider/model identity snapshot.
	ProviderProjection = aicommon.ProviderProjection
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

// Service aggregates typed AI sub capabilities under one stable namespace.
//
// Service 聚合宿主发布的 AI 子能力，适用于源码插件、动态插件和宿主模块通过统一入口访问文本等能力，同时避免在根能力目录继续追加 AI 子能力方法。
type Service interface {
	// MethodStatuses returns method-level availability for AI sub capabilities.
	//
	// MethodStatuses 批量返回 AI 子能力方法状态，用于插件降级判断；结果只包含能力、方法、可用性和状态原因，不暴露 provider 配置。
	MethodStatuses(ctx context.Context, input MethodStatusesInput) (*MethodStatusesResult, error)
	// Text returns the text AI capability service.
	//
	// Text 返回文本 AI 子能力服务；未配置 provider 时也必须返回可降级服务，由子能力自身返回结构化不可用错误。
	Text() aitext.Service
	// Image returns the image AI capability service.
	//
	// Image 返回图片 AI 子能力服务；未配置 provider 时仍返回可降级服务。
	Image() aiimage.Service
	// Embedding returns the embedding AI capability service.
	//
	// Embedding 返回向量 AI 子能力服务；未配置 provider 时仍返回可降级服务。
	Embedding() aiembedding.Service
	// Audio returns the audio AI capability service.
	//
	// Audio 返回音频 AI 子能力服务；未配置 provider 时仍返回可降级服务。
	Audio() aiaudio.Service
	// Vision returns the vision AI capability service.
	//
	// Vision 返回视觉 AI 子能力服务；未配置 provider 时仍返回可降级服务。
	Vision() aivision.Service
	// Document returns the document AI capability service.
	//
	// Document 返回文档 AI 子能力服务；未配置 provider 时仍返回可降级服务。
	Document() aidocument.Service
	// Safety returns the safety AI capability service.
	//
	// Safety 返回安全审核 AI 子能力服务；未配置 provider 时仍返回可降级服务。
	Safety() aisafety.Service
	// Video returns the video AI capability service.
	//
	// Video 返回视频 AI 子能力服务；未配置 provider 时仍返回可降级服务。
	Video() aivideo.Service
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

// Option customizes optional non-text AI sub capability services.
type Option func(*serviceImpl)

// WithImage sets the image AI sub capability service.
func WithImage(service aiimage.Service) Option {
	return func(s *serviceImpl) { s.image = service }
}

// WithEmbedding sets the embedding AI sub capability service.
func WithEmbedding(service aiembedding.Service) Option {
	return func(s *serviceImpl) { s.embedding = service }
}

// WithAudio sets the audio AI sub capability service.
func WithAudio(service aiaudio.Service) Option {
	return func(s *serviceImpl) { s.audio = service }
}

// WithVision sets the vision AI sub capability service.
func WithVision(service aivision.Service) Option {
	return func(s *serviceImpl) { s.vision = service }
}

// WithDocument sets the document AI sub capability service.
func WithDocument(service aidocument.Service) Option {
	return func(s *serviceImpl) { s.document = service }
}

// WithSafety sets the safety AI sub capability service.
func WithSafety(service aisafety.Service) Option {
	return func(s *serviceImpl) { s.safety = service }
}

// WithVideo sets the video AI sub capability service.
func WithVideo(service aivideo.Service) Option {
	return func(s *serviceImpl) { s.video = service }
}

// New creates an AI namespace service from explicit sub capability services.
func New(text aitext.Service, options ...Option) Service {
	if text == nil {
		text = aitext.New(nil, nil)
	}
	service := &serviceImpl{text: text}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	service.ensureFallbacks()
	return service
}

// Text returns the text AI capability service.
func (s *serviceImpl) Text() aitext.Service {
	if s == nil || s.text == nil {
		return aitext.New(nil, nil)
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
//
// ForPlugin 返回绑定插件身份的 AI 命名空间服务，同时保留宿主运行期持有的 AI 子能力实现。
// 该方法会把 pluginID 通过各个子能力注入到后续 provider 请求中，使源码插件和动态插件可以消费宿主
// AI 能力，而不需要也不能由调用方手动填写或伪造来源身份。
//
// 宿主可信注入的来源身份用于 AI 调用审计、用量归因、问题定位，以及后续插件级配额、限流、
// 档位访问和 purpose 策略等治理能力。service 为空时仍返回带 fallback 子能力的命名空间，
// 确保调用方获得结构化不可用错误，而不是 nil service。
func ForPlugin(service Service, pluginID string) Service {
	if service == nil {
		return New(aitext.ForPlugin(nil, pluginID))
	}
	return New(
		aitext.ForPlugin(service.Text(), pluginID),
		WithImage(aiimage.ForPlugin(service.Image(), pluginID)),
		WithEmbedding(aiembedding.ForPlugin(service.Embedding(), pluginID)),
		WithAudio(aiaudio.ForPlugin(service.Audio(), pluginID)),
		WithVision(aivision.ForPlugin(service.Vision(), pluginID)),
		WithDocument(aidocument.ForPlugin(service.Document(), pluginID)),
		WithSafety(aisafety.ForPlugin(service.Safety(), pluginID)),
		WithVideo(aivideo.ForPlugin(service.Video(), pluginID)),
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
