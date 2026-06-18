// Package aitext owns the stable text AI capability contract exposed through
// capability. The host owns provider discovery, validation, and fallback
// behavior; official plugins own provider/model/tier storage and external
// protocol adapters.
package aitext

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/capmodel"
	internalregistry "lina-core/pkg/plugin/capability/internal/capabilityregistry"
)

const (
	// CapabilityAITextV1 identifies the versioned text AI framework capability.
	CapabilityAITextV1 = "framework.ai.text.v1"
	// ProviderPluginID is the official source-plugin identifier that provides text AI capability.
	ProviderPluginID = "linapro-ai-core"
)

// CapabilityType identifies one AI capability family.
type CapabilityType string

const (
	// CapabilityTypeText is the first supported AI capability family.
	CapabilityTypeText CapabilityType = CapabilityType(aicommon.CapabilityTypeText)
)

// CapabilityMethod identifies one AI method inside a capability family.
type CapabilityMethod string

const (
	// CapabilityMethodGenerate identifies synchronous text generation.
	CapabilityMethodGenerate CapabilityMethod = CapabilityMethod(aicommon.CapabilityMethodTextGenerate)
)

// Tier identifies the platform text AI service level requested by callers.
type Tier string

const (
	// TierBasic is the low-cost text AI tier for simple generation tasks.
	TierBasic Tier = "basic"
	// TierStandard is the default text AI tier for regular generation tasks.
	TierStandard Tier = "standard"
	// TierAdvanced is the high-capability text AI tier for complex generation tasks.
	TierAdvanced Tier = "advanced"
)

// ThinkingEffort identifies the abstract reasoning effort requested by callers.
type ThinkingEffort string

const (
	// ThinkingEffortLow requests low reasoning effort.
	ThinkingEffortLow ThinkingEffort = "low"
	// ThinkingEffortMedium requests medium reasoning effort.
	ThinkingEffortMedium ThinkingEffort = "medium"
	// ThinkingEffortHigh requests high reasoning effort.
	ThinkingEffortHigh ThinkingEffort = "high"
	// ThinkingEffortXHigh requests extra-high reasoning effort.
	ThinkingEffortXHigh ThinkingEffort = "xhigh"
	// ThinkingEffortMax requests the maximum model-supported reasoning effort.
	ThinkingEffortMax ThinkingEffort = "max"
)

// MessageRole identifies the role of one text generation message.
type MessageRole string

const (
	// MessageRoleSystem carries system instructions.
	MessageRoleSystem MessageRole = "system"
	// MessageRoleUser carries user input.
	MessageRoleUser MessageRole = "user"
	// MessageRoleAssistant carries prior assistant output.
	MessageRoleAssistant MessageRole = "assistant"
)

// Message carries one plain-text message in a generation request.
type Message struct {
	// Role identifies the message author role.
	Role MessageRole `json:"role"`
	// Content is the plain-text message body. It must not contain hidden thinking content.
	Content string `json:"content"`
}

// GenerateRequest carries one synchronous text generation request.
type GenerateRequest struct {
	// Purpose identifies the governed calling scenario, such as content.summary.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier Tier `json:"tier"`
	// Messages carries ordered plain-text generation context.
	Messages []Message `json:"messages"`
	// MaxOutputTokens optionally caps generated output tokens.
	MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
	// Temperature optionally controls sampling.
	Temperature *float64 `json:"temperature,omitempty"`
	// ThinkingEffort optionally requests abstract model reasoning effort.
	ThinkingEffort *ThinkingEffort `json:"thinkingEffort,omitempty"`
	// Metadata carries short audit keys and must not include prompt or response bodies.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ProviderRequest carries a provider-internal text generation request after
// the host service layer has attached governed caller identity.
type ProviderRequest struct {
	// GenerateRequest carries the ordinary caller-visible request fields.
	GenerateRequest
	// SourcePluginID identifies the dynamic or source plugin that initiated the call.
	SourcePluginID string `json:"sourcePluginId,omitempty"`
}

// Usage describes token usage returned by a text provider.
type Usage struct {
	// InputTokens is the prompt/input token count reported by the provider.
	InputTokens int `json:"inputTokens"`
	// OutputTokens is the completion/output token count reported by the provider.
	OutputTokens int `json:"outputTokens"`
}

// GenerateResponse carries the successful synchronous text generation result.
type GenerateResponse struct {
	// Text is the generated plain text.
	Text string `json:"text"`
	// Tier is the actual platform tier used by the provider.
	Tier Tier `json:"tier"`
	// ProviderName is the public provider display name.
	ProviderName string `json:"providerName"`
	// ModelName is the public model display name or model identifier.
	ModelName string `json:"modelName"`
	// Protocol is the provider protocol family used for the call.
	Protocol string `json:"protocol"`
	// Usage contains reported token counts.
	Usage Usage `json:"usage"`
	// LatencyMs is the provider call latency in milliseconds.
	LatencyMs int `json:"latencyMs"`
	// GeneratedAt is a Unix timestamp in milliseconds.
	GeneratedAt int64 `json:"generatedAt"`
	// ThinkingEffort is the actual effort applied by the provider when available.
	ThinkingEffort *ThinkingEffort `json:"thinkingEffort,omitempty"`
}

// ProviderEnv carries explicit host construction inputs for a text AI provider.
type ProviderEnv struct {
	// PluginID is the provider plugin being constructed.
	PluginID string
	// BizCtx exposes the current request business context for provider-side
	// audit projection without leaking host-internal context models.
	BizCtx bizctxcap.Service
	// Cache exposes the plugin-scoped shared cache backend for non-authoritative
	// revision markers and other runtime acceleration metadata.
	Cache cachecap.Service
}

// ProviderRuntime defines the narrow plugin state and environment capability
// required by aitext to use declared providers.
//
// ProviderRuntime 定义 aitext 在延迟创建文本 AI 提供方时所需的最小宿主运行时入口，适用于判断官方插件是否处于可服务状态，
// 并为 provider 工厂构造受治理的宿主环境。
type ProviderRuntime interface {
	// IsProviderEnabled reports whether pluginID may serve framework provider calls.
	//
	// IsProviderEnabled 判断指定插件是否允许承接框架文本 AI 能力调用，通常用于能力服务在调用 provider 前确认插件已启用且处于可服务状态。
	IsProviderEnabled(ctx context.Context, pluginID string) bool
	// AITextProviderEnv returns typed, plugin-scoped construction inputs for one provider plugin.
	//
	// AITextProviderEnv 返回指定文本 AI 插件的类型化构造环境，适用于 provider 工厂获取受治理宿主输入，同时避免宿主消费方依赖插件内部实现。
	AITextProviderEnv(pluginID string) ProviderEnv
}

// Provider defines the text AI capability implemented by provider plugins.
//
// Provider 定义文本 AI 能力插件必须实现的提供方契约，适用于 linapro-ai-core 等插件向宿主提供按 purpose 和档位治理的同步文本生成能力。
type Provider interface {
	// GenerateText executes one synchronous text generation request.
	//
	// GenerateText 根据已校验的 purpose、档位、消息、生成参数和可选 thinkingEffort 执行文本生成；实现必须脱敏日志并返回结构化业务错误。
	GenerateText(ctx context.Context, request ProviderRequest) (*GenerateResponse, error)
}

// Service defines the optional text AI capability consumed by host core
// services and plugins without depending on a concrete provider implementation.
//
// Service 定义宿主核心服务、源码插件和动态插件可消费的文本 AI 能力，适用于按稳定档位执行同步文本生成，并在官方插件缺失时获得安全降级错误。
type Service interface {
	// Available reports whether an active text AI provider is available.
	//
	// Available 判断当前是否存在可用文本 AI 提供方，适用于调用方决定展示、降级或提示缺少智能中心配置。
	Available(ctx context.Context) bool
	// Status returns the current text AI capability activation state.
	//
	// Status 返回文本 AI 能力激活状态，适用于诊断、治理检查和插件能力状态展示。
	Status(ctx context.Context) capmodel.CapabilityStatus
	// MethodStatus returns method-level text AI availability without exposing provider internals.
	//
	// MethodStatus 返回文本 AI 子能力单个方法的可用状态，适用于插件按方法降级；结果不得包含 provider 私有配置。
	MethodStatus(ctx context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus
	// GenerateText executes one synchronous text generation request.
	//
	// GenerateText 按 purpose 和档位执行同步文本生成；请求无效、provider 不可用或插件配置缺失时返回结构化业务错误。
	GenerateText(ctx context.Context, request GenerateRequest) (*GenerateResponse, error)
}

// ProviderFactory creates one text AI provider from an explicit typed
// construction environment during lazy capability use.
type ProviderFactory func(ctx context.Context, env ProviderEnv) (Provider, error)

// Manager owns text AI provider declarations and lazy provider instances.
type Manager struct {
	registry *internalregistry.Manager[ProviderEnv]
}

// NewManager creates an empty text AI provider manager.
func NewManager() *Manager {
	return &Manager{registry: internalregistry.NewManager[ProviderEnv]()}
}

// RegisterFactory records one plugin-provided text AI capability factory.
func (m *Manager) RegisterFactory(pluginID string, factory ProviderFactory) error {
	return m.registry.RegisterFactory(
		CapabilityAITextV1,
		pluginID,
		func(ctx context.Context, env ProviderEnv) (any, error) {
			return factory(ctx, env)
		},
	)
}

// serviceImpl delegates text AI calls to the active provider and returns
// structured fallback errors when no provider is usable.
type serviceImpl struct {
	manager        *Manager
	runtime        ProviderRuntime
	sourcePluginID string
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// New creates an optional text AI capability service from explicit runtime-owned dependencies.
func New(manager *Manager, runtime ProviderRuntime) Service {
	if manager == nil {
		manager = NewManager()
	}
	if runtime == nil {
		runtime = noopProviderRuntime{}
	}
	return &serviceImpl{manager: manager, runtime: runtime}
}

// ForPlugin returns a text AI service that injects pluginID into provider
// requests when the supplied service supports host-managed provider dispatch.
func ForPlugin(service Service, pluginID string) Service {
	if service == nil {
		return &serviceImpl{
			manager:        NewManager(),
			runtime:        noopProviderRuntime{},
			sourcePluginID: strings.TrimSpace(pluginID),
		}
	}
	impl, ok := service.(*serviceImpl)
	if !ok {
		return service
	}
	return &serviceImpl{
		manager:        impl.manager,
		runtime:        impl.runtime,
		sourcePluginID: strings.TrimSpace(pluginID),
	}
}

// CapabilityType returns the fixed capability family for text generation.
func (r GenerateRequest) CapabilityType() CapabilityType {
	return CapabilityTypeText
}

// CapabilityMethod returns the fixed capability method for text generation.
func (r GenerateRequest) CapabilityMethod() CapabilityMethod {
	return CapabilityMethodGenerate
}

// Valid reports whether the tier is one of the stable platform text tiers.
func (t Tier) Valid() bool {
	switch t {
	case TierBasic, TierStandard, TierAdvanced:
		return true
	default:
		return false
	}
}

// Valid reports whether the effort is one of the stable platform effort values.
func (e ThinkingEffort) Valid() bool {
	switch e {
	case ThinkingEffortLow, ThinkingEffortMedium, ThinkingEffortHigh, ThinkingEffortXHigh, ThinkingEffortMax:
		return true
	default:
		return false
	}
}

// Valid reports whether the message role is supported by the v1 text contract.
func (r MessageRole) Valid() bool {
	switch r {
	case MessageRoleSystem, MessageRoleUser, MessageRoleAssistant:
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

// noopProviderRuntime reports all plugins as disabled when aitext is
// constructed without an explicit provider runtime.
type noopProviderRuntime struct{}
