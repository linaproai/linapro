// protocol_codec.go exposes bridge request and response codecs through the public protocol facade.
// Only delegate aliases live here; encoding behavior is owned by the internal codec package.

package protocol

import "lina-core/pkg/plugin/pluginbridge/internal/codec"

var (
	EncodeRequestEnvelope    = codec.EncodeRequestEnvelope
	DecodeRequestEnvelope    = codec.DecodeRequestEnvelope
	EncodeResponseEnvelope   = codec.EncodeResponseEnvelope
	DecodeResponseEnvelope   = codec.DecodeResponseEnvelope
	EncodeBodyBase64         = codec.EncodeBodyBase64
	NewSuccessResponse       = codec.NewSuccessResponse
	NewJSONResponse          = codec.NewJSONResponse
	NewFailureResponse       = codec.NewFailureResponse
	NewUnauthorizedResponse  = codec.NewUnauthorizedResponse
	NewForbiddenResponse     = codec.NewForbiddenResponse
	NewBadRequestResponse    = codec.NewBadRequestResponse
	NewNotFoundResponse      = codec.NewNotFoundResponse
	NewInternalErrorResponse = codec.NewInternalErrorResponse
)
