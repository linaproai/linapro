// This file verifies the AI namespace keeps text capability fallback access
// available when no provider is configured.

package aicap

import (
	"context"
	"reflect"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/aicap/aiaudio"
	"lina-core/pkg/plugin/capability/aicap/aicommon"
	"lina-core/pkg/plugin/capability/aicap/aiimage"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/aicap/aivideo"
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
	if service == nil || service.Text() == nil || service.Image() == nil || service.Embedding() == nil ||
		service.Audio() == nil || service.Vision() == nil || service.Document() == nil ||
		service.Safety() == nil || service.Video() == nil {
		t.Fatal("expected plugin-scoped AI fallback service")
	}
}

func TestMultimodalFallbackServicesReturnUnavailable(t *testing.T) {
	t.Parallel()

	service := New(nil)
	_, err := service.Image().Generate(context.Background(), imageGenerateRequest())
	if !bizerr.Is(err, aicommon.CodeProviderUnavailable) {
		t.Fatalf("expected image provider unavailable, got %v", err)
	}
	_, err = service.Audio().Transcribe(context.Background(), audioTranscribeRequest())
	if !bizerr.Is(err, aicommon.CodeProviderUnavailable) {
		t.Fatalf("expected audio provider unavailable, got %v", err)
	}
	_, err = service.Video().OperationGet(context.Background(), videoOperationGetRequest())
	if !bizerr.Is(err, aicommon.CodeProviderUnavailable) {
		t.Fatalf("expected video operation provider unavailable, got %v", err)
	}
}

func TestAINamespaceDoesNotExposeWeakGateway(t *testing.T) {
	t.Parallel()

	serviceType := reflect.TypeOf((*Service)(nil)).Elem()
	for _, method := range []string{"Invoke", "Generate"} {
		if _, ok := serviceType.MethodByName(method); ok {
			t.Fatalf("AI namespace must not expose weak gateway method %s", method)
		}
	}
}

func imageGenerateRequest() aiimage.GenerateRequest {
	return aiimage.GenerateRequest{
		Purpose: "asset.preview",
		Tier:    aicommon.TierStandard,
		Prompt:  "draw a chart",
		Count:   1,
	}
}

func audioTranscribeRequest() aiaudio.TranscribeRequest {
	return aiaudio.TranscribeRequest{
		Purpose: "meeting.transcript",
		Tier:    aicommon.TierStandard,
		Audio: aicommon.AssetRef{
			Ref:       "asset/audio-1",
			MimeType:  "audio/mpeg",
			SizeBytes: 1024,
		},
	}
}

func videoOperationGetRequest() aivideo.OperationGetRequest {
	return aivideo.OperationGetRequest{
		Purpose:      "video.preview",
		OperationRef: "operation-1",
	}
}
