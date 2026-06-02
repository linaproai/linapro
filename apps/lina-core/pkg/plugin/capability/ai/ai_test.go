// This file verifies the AI namespace keeps text capability fallback access
// available when no provider is configured.

package ai

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/ai/aitext"
)

func TestTextReturnsFallbackService(t *testing.T) {
	t.Parallel()

	service := New(nil)
	text := service.Text()
	if text == nil {
		t.Fatal("expected non-nil text AI fallback service")
	}
	_, err := text.GenerateText(context.Background(), aitext.GenerateRequest{
		Purpose: "test.summary",
		Tier:    aitext.TierStandard,
		Messages: []aitext.Message{
			{Role: aitext.MessageRoleUser, Content: "hello"},
		},
	})
	if !bizerr.Is(err, aitext.CodeTextProviderUnavailable) {
		t.Fatalf("expected provider unavailable fallback error, got %v", err)
	}
}

func TestForPluginReturnsFallbackService(t *testing.T) {
	t.Parallel()

	service := ForPlugin(nil, "source-plugin")
	if service == nil || service.Text() == nil {
		t.Fatal("expected plugin-scoped AI fallback service")
	}
}
