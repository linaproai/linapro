// This file implements transport codecs for governed AI host-service calls.
// The bridge owns only payload serialization; text capability semantics remain
// in capability/ai/aitext.

package hostservice

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/ai/aitext"
)

// HostServiceAITextGenerateRequest carries one text generation host-service request.
type HostServiceAITextGenerateRequest struct {
	// Purpose is the governed calling scenario and must match resourceRef.
	Purpose string `json:"purpose"`
	// Tier is the requested text AI tier; empty lets the resource default apply.
	Tier aitext.Tier `json:"tier,omitempty"`
	// Messages carries ordered plain-text generation context.
	Messages []aitext.Message `json:"messages"`
	// MaxOutputTokens optionally caps generated output tokens.
	MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
	// Temperature optionally controls sampling.
	Temperature *float64 `json:"temperature,omitempty"`
	// ThinkingEffort optionally requests abstract model reasoning effort.
	ThinkingEffort *aitext.ThinkingEffort `json:"thinkingEffort,omitempty"`
	// Metadata carries short audit keys and must not include prompt or response bodies.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// MarshalHostServiceAITextGenerateRequest encodes one AI text request as JSON.
func MarshalHostServiceAITextGenerateRequest(req *HostServiceAITextGenerateRequest) []byte {
	if req == nil {
		return nil
	}
	content, err := json.Marshal(req)
	if err != nil {
		return nil
	}
	return content
}

// UnmarshalHostServiceAITextGenerateRequest decodes one AI text request.
func UnmarshalHostServiceAITextGenerateRequest(data []byte) (*HostServiceAITextGenerateRequest, error) {
	if len(data) == 0 {
		return nil, gerror.New("ai text generate request is empty")
	}
	out := &HostServiceAITextGenerateRequest{}
	if err := json.Unmarshal(data, out); err != nil {
		return nil, gerror.Wrap(err, "decode ai text generate request failed")
	}
	return out, nil
}
