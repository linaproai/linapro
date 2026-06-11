// This file implements optional text AI capability delegation. It validates
// request boundaries before forwarding calls to the active provider, returning
// structured business errors when the official provider is absent or disabled.

package aitext

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	maxMetadataEntries  = 16
	maxMetadataKeyLen   = 64
	maxMetadataValueLen = 256
)

// IsProviderEnabled always returns false.
func (noopProviderRuntime) IsProviderEnabled(_ context.Context, _ string) bool {
	return false
}

// AITextProviderEnv returns an empty typed provider environment.
func (noopProviderRuntime) AITextProviderEnv(pluginID string) ProviderEnv {
	return ProviderEnv{PluginID: pluginID}
}

// Available reports whether an active text AI provider is available.
func (s *serviceImpl) Available(ctx context.Context) bool {
	if s == nil {
		return false
	}
	return s.manager.registry.StatusWithProvider(ctx, CapabilityAITextV1, s.runtime, s.providerEnv).Available
}

// Status returns the current text AI capability activation state.
func (s *serviceImpl) Status(ctx context.Context) capmodel.CapabilityStatus {
	if s == nil {
		return convertCapabilityStatus(NewManager().registry.Status(ctx, CapabilityAITextV1, nil))
	}
	return convertCapabilityStatus(s.manager.registry.StatusWithProvider(ctx, CapabilityAITextV1, s.runtime, s.providerEnv))
}

// GenerateText executes one synchronous text generation request.
func (s *serviceImpl) GenerateText(ctx context.Context, request GenerateRequest) (*GenerateResponse, error) {
	if err := validateGenerateRequest(request); err != nil {
		return nil, err
	}
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeTextProviderUnavailable)
	}
	if provider == nil {
		return nil, bizerr.NewCode(CodeTextProviderUnavailable)
	}
	return provider.GenerateText(ctx, ProviderRequest{
		GenerateRequest: request,
		SourcePluginID:  s.sourcePluginID,
	})
}

// currentProvider returns the currently registered text AI capability provider.
func (s *serviceImpl) currentProvider(ctx context.Context) (Provider, error) {
	if s == nil {
		return nil, nil
	}
	provider, err := s.manager.registry.ActiveProviderWithError(ctx, CapabilityAITextV1, s.runtime, s.providerEnv)
	if err != nil || provider == nil {
		return nil, err
	}
	typedProvider, ok := provider.(Provider)
	if !ok {
		return nil, nil
	}
	return typedProvider, nil
}

// providerEnv builds lazy construction inputs for one text AI provider.
func (s *serviceImpl) providerEnv(_ context.Context, pluginID string) ProviderEnv {
	env := ProviderEnv{PluginID: pluginID}
	if s != nil && s.runtime != nil {
		env = s.runtime.AITextProviderEnv(pluginID)
	}
	if env.PluginID == "" {
		env.PluginID = pluginID
	}
	return env
}

// validateGenerateRequest checks provider-independent request boundaries.
func validateGenerateRequest(request GenerateRequest) error {
	if strings.TrimSpace(request.Purpose) == "" {
		return bizerr.NewCode(CodeTextPurposeRequired)
	}
	if !request.Tier.Valid() {
		return bizerr.NewCode(CodeTextTierInvalid, bizerr.P("tier", string(request.Tier)))
	}
	if len(request.Messages) == 0 {
		return bizerr.NewCode(CodeTextMessagesRequired)
	}
	for _, message := range request.Messages {
		if !message.Role.Valid() {
			return bizerr.NewCode(CodeTextMessageRoleInvalid, bizerr.P("role", string(message.Role)))
		}
	}
	if request.ThinkingEffort != nil && !request.ThinkingEffort.Valid() {
		return bizerr.NewCode(
			CodeTextThinkingEffortInvalid,
			bizerr.P("effort", string(*request.ThinkingEffort)),
		)
	}
	if request.MaxOutputTokens < 0 {
		return bizerr.NewCode(CodeTextMaxOutputTokensInvalid)
	}
	return validateMetadata(request.Metadata)
}

// validateMetadata prevents pluginbridge callers from smuggling large prompts
// or responses through audit metadata.
func validateMetadata(metadata map[string]string) error {
	if len(metadata) > maxMetadataEntries {
		return bizerr.NewCode(CodeTextMetadataTooLarge)
	}
	for key, value := range metadata {
		if strings.TrimSpace(key) == "" || len(key) > maxMetadataKeyLen || len(value) > maxMetadataValueLen {
			return bizerr.NewCode(CodeTextMetadataTooLarge)
		}
	}
	return nil
}
