// This file tests representative AI host service request codec round trips.

package protocol

import (
	"reflect"
	"testing"

	"lina-core/pkg/plugin/capability/aicap/aitext"
)

// TestHostServiceAITextGenerateRequestRoundTrip verifies AI text requests keep
// governance, prompt, sampling, and audit metadata through JSON transport.
func TestHostServiceAITextGenerateRequestRoundTrip(t *testing.T) {
	temperature := 0.2
	thinkingEffort := aitext.ThinkingEffortHigh
	original := &HostServiceAITextGenerateRequest{
		Purpose: "content.summary",
		Tier:    aitext.TierAdvanced,
		Messages: []aitext.Message{
			{Role: aitext.MessageRoleSystem, Content: "Summarize the input."},
			{Role: aitext.MessageRoleUser, Content: "Quarterly usage report"},
		},
		MaxOutputTokens: 512,
		Temperature:     &temperature,
		ThinkingEffort:  &thinkingEffort,
		Metadata: map[string]string{
			"trace": "trace-1",
		},
	}

	data := MarshalHostServiceAITextGenerateRequest(original)
	decoded, err := UnmarshalHostServiceAITextGenerateRequest(data)
	if err != nil {
		t.Fatalf("decode AI text request failed: %v", err)
	}
	if !reflect.DeepEqual(decoded, original) {
		t.Fatalf("unexpected AI text request: %#v", decoded)
	}
}
