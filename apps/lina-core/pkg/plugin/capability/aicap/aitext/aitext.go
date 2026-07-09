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

// Service defines the optional text AI capability consumed by host core
// services and plugins without depending on a concrete provider implementation.
type Service interface {
	// Available reports whether an active text AI provider is available.
	Available(ctx context.Context) bool
	// Status returns the current text AI capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// MethodStatus returns method-level text AI availability without exposing provider internals.
	MethodStatus(ctx context.Context, method aicommon.CapabilityMethod) aicommon.MethodStatus
	// GenerateText executes one synchronous text generation request.
	GenerateText(ctx context.Context, request GenerateRequest) (*GenerateResponse, error)
}

// Provider defines the text AI capability implemented by provider plugins.
type Provider interface {
	// GenerateText executes one synchronous text generation request.
	GenerateText(ctx context.Context, request ProviderRequest) (*GenerateResponse, error)
}

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
	// audit metadata without leaking host-internal context models.
	BizCtx bizctxcap.Service
	// Cache exposes the plugin-scoped shared cache backend for non-authoritative
	// revision markers and other runtime acceleration metadata.
	Cache cachecap.Service
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
	enablement     internalregistry.EnablementReader
	envFactory     internalregistry.ProviderEnvFactory[ProviderEnv]
	sourcePluginID string
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// New creates an optional text AI capability service from explicit runtime-owned dependencies.
func New(
	manager *Manager,
	enablement internalregistry.EnablementReader,
	envFactory internalregistry.ProviderEnvFactory[ProviderEnv],
) Service {
	if manager == nil {
		manager = NewManager()
	}
	if enablement == nil {
		enablement = noopEnablementReader{}
	}
	if envFactory == nil {
		envFactory = defaultProviderEnv
	}
	return &serviceImpl{manager: manager, enablement: enablement, envFactory: envFactory}
}

// ForPlugin returns a text AI service that injects pluginID into provider
// requests when the supplied service supports host-managed provider dispatch.
func ForPlugin(service Service, pluginID string) Service {
	if service == nil {
		return &serviceImpl{
			manager:        NewManager(),
			enablement:     noopEnablementReader{},
			envFactory:     defaultProviderEnv,
			sourcePluginID: strings.TrimSpace(pluginID),
		}
	}
	impl, ok := service.(*serviceImpl)
	if !ok {
		return service
	}
	return &serviceImpl{
		manager:        impl.manager,
		enablement:     impl.enablement,
		envFactory:     impl.envFactory,
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

// noopEnablementReader reports all provider plugins as disabled.
type noopEnablementReader struct{}
