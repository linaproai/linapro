// This file verifies text AI capability fallback, validation, and singleton
// provider behavior without requiring the official AI plugin.

package aitext

import (
	"context"
	"fmt"
	"testing"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
)

func TestGenerateTextReturnsUnavailableWithoutProvider(t *testing.T) {
	t.Parallel()

	service := New(nil, nil)
	_, err := service.GenerateText(context.Background(), validGenerateRequest())
	if !bizerr.Is(err, CodeTextProviderUnavailable) {
		t.Fatalf("expected provider unavailable error, got %v", err)
	}
	if service.Available(context.Background()) {
		t.Fatal("expected text AI capability to be unavailable without provider")
	}
}

func TestGenerateTextValidatesTierAndThinkingEffort(t *testing.T) {
	t.Parallel()

	service := New(nil, nil)
	request := validGenerateRequest()
	request.Tier = Tier("custom")
	_, err := service.GenerateText(context.Background(), request)
	if !bizerr.Is(err, CodeTextTierInvalid) {
		t.Fatalf("expected invalid tier error, got %v", err)
	}

	request = validGenerateRequest()
	invalidEffort := ThinkingEffort("extreme")
	request.ThinkingEffort = &invalidEffort
	_, err = service.GenerateText(context.Background(), request)
	if !bizerr.Is(err, CodeTextThinkingEffortInvalid) {
		t.Fatalf("expected invalid thinking effort error, got %v", err)
	}
}

func TestGenerateRequestCapabilityIdentity(t *testing.T) {
	t.Parallel()

	request := validGenerateRequest()
	if request.CapabilityType() != CapabilityTypeText {
		t.Fatalf("expected capability type %q, got %q", CapabilityTypeText, request.CapabilityType())
	}
	if request.CapabilityMethod() != CapabilityMethodGenerate {
		t.Fatalf("expected capability method %q, got %q", CapabilityMethodGenerate, request.CapabilityMethod())
	}
}

func TestGenerateTextDelegatesToActiveProvider(t *testing.T) {
	t.Parallel()

	pluginID := fmt.Sprintf("plugin-test-ai-provider-%d", time.Now().UnixNano())
	manager := NewManager()
	if err := manager.RegisterFactory(pluginID, func(context.Context, ProviderEnv) (Provider, error) {
		return fakeProvider{}, nil
	}); err != nil {
		t.Fatalf("register provider failed: %v", err)
	}

	service := New(manager, testRuntime{pluginID: pluginID})
	response, err := service.GenerateText(context.Background(), validGenerateRequest())
	if err != nil {
		t.Fatalf("expected active provider success, got %v", err)
	}
	if response == nil || response.Text != "generated text" || response.Tier != TierStandard {
		t.Fatalf("unexpected response: %#v", response)
	}
	status := service.Status(context.Background())
	if !status.Available || status.ActiveProvider != pluginID {
		t.Fatalf("expected active provider status, got %#v", status)
	}
}

func TestForPluginInjectsSourcePluginID(t *testing.T) {
	t.Parallel()

	pluginID := fmt.Sprintf("plugin-test-ai-provider-source-%d", time.Now().UnixNano())
	var seenSourcePluginID string
	manager := NewManager()
	if err := manager.RegisterFactory(pluginID, func(context.Context, ProviderEnv) (Provider, error) {
		return fakeProviderFunc(func(_ context.Context, request ProviderRequest) (*GenerateResponse, error) {
			seenSourcePluginID = request.SourcePluginID
			return fakeProvider{}.GenerateText(context.Background(), request)
		}), nil
	}); err != nil {
		t.Fatalf("register provider failed: %v", err)
	}

	service := ForPlugin(New(manager, testRuntime{pluginID: pluginID}), " source-plugin ")
	if _, err := service.GenerateText(context.Background(), validGenerateRequest()); err != nil {
		t.Fatalf("expected active provider success, got %v", err)
	}
	if seenSourcePluginID != "source-plugin" {
		t.Fatalf("expected scoped source plugin id, got %q", seenSourcePluginID)
	}
}

func TestGenerateTextRejectsProviderConflict(t *testing.T) {
	t.Parallel()

	firstPluginID := fmt.Sprintf("plugin-test-ai-provider-a-%d", time.Now().UnixNano())
	secondPluginID := fmt.Sprintf("plugin-test-ai-provider-b-%d", time.Now().UnixNano())
	manager := NewManager()
	for _, pluginID := range []string{firstPluginID, secondPluginID} {
		pluginID := pluginID
		if err := manager.RegisterFactory(pluginID, func(context.Context, ProviderEnv) (Provider, error) {
			return fakeProvider{}, nil
		}); err != nil {
			t.Fatalf("register provider %s failed: %v", pluginID, err)
		}
	}

	service := New(manager, testRuntime{pluginID: firstPluginID, secondPluginID: secondPluginID})
	_, err := service.GenerateText(context.Background(), validGenerateRequest())
	if !bizerr.Is(err, capmodel.CodeCapabilityProviderConflict) {
		t.Fatalf("expected provider conflict error, got %v", err)
	}
	status := service.Status(context.Background())
	if status.Available || status.Reason == "" {
		t.Fatalf("expected unavailable conflict status, got %#v", status)
	}
}

func validGenerateRequest() GenerateRequest {
	return GenerateRequest{
		Purpose: "test.summary",
		Tier:    TierStandard,
		Messages: []Message{
			{Role: MessageRoleUser, Content: "hello"},
		},
		MaxOutputTokens: 128,
	}
}

type fakeProvider struct{}

func (fakeProvider) GenerateText(context.Context, ProviderRequest) (*GenerateResponse, error) {
	return &GenerateResponse{
		Text:         "generated text",
		Tier:         TierStandard,
		ProviderName: "Fake",
		ModelName:    "fake-model",
		Protocol:     "test",
		Usage:        Usage{InputTokens: 1, OutputTokens: 2},
		GeneratedAt:  time.Now().UnixMilli(),
	}, nil
}

type fakeProviderFunc func(context.Context, ProviderRequest) (*GenerateResponse, error)

func (f fakeProviderFunc) GenerateText(ctx context.Context, request ProviderRequest) (*GenerateResponse, error) {
	return f(ctx, request)
}

type testRuntime struct {
	pluginID       string
	secondPluginID string
}

func (r testRuntime) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return pluginID == r.pluginID || pluginID == r.secondPluginID
}

func (r testRuntime) AITextProviderEnv(pluginID string) ProviderEnv {
	return ProviderEnv{PluginID: pluginID}
}
