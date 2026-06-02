// hostservice_ai_codec.go exposes AI host-service payload codecs through the
// public protocol facade.

package protocol

import "lina-core/pkg/plugin/pluginbridge/internal/hostservice"

var (
	MarshalHostServiceAITextGenerateRequest   = hostservice.MarshalHostServiceAITextGenerateRequest
	UnmarshalHostServiceAITextGenerateRequest = hostservice.UnmarshalHostServiceAITextGenerateRequest
)
