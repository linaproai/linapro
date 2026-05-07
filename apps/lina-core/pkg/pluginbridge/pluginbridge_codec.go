// This file implements public bridge codec entry points and response builders.

package pluginbridge

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// EncodeRequestEnvelope encodes one request envelope into protobuf wire bytes.
func EncodeRequestEnvelope(in *BridgeRequestEnvelopeV1) ([]byte, error) {
	if in == nil {
		return nil, gerror.New("bridge request envelope cannot be nil")
	}
	return marshalRequestEnvelope(in), nil
}

// DecodeRequestEnvelope decodes one request envelope from protobuf wire bytes.
func DecodeRequestEnvelope(content []byte) (*BridgeRequestEnvelopeV1, error) {
	out := &BridgeRequestEnvelopeV1{}
	if err := unmarshalRequestEnvelope(content, out); err != nil {
		return nil, err
	}
	return out, nil
}

// EncodeResponseEnvelope encodes one response envelope into protobuf wire bytes.
func EncodeResponseEnvelope(in *BridgeResponseEnvelopeV1) ([]byte, error) {
	if in == nil {
		return nil, gerror.New("bridge response envelope cannot be nil")
	}
	return marshalResponseEnvelope(in), nil
}

// DecodeResponseEnvelope decodes one response envelope from protobuf wire bytes.
func DecodeResponseEnvelope(content []byte) (*BridgeResponseEnvelopeV1, error) {
	out := &BridgeResponseEnvelopeV1{}
	if err := unmarshalResponseEnvelope(content, out); err != nil {
		return nil, err
	}
	return out, nil
}

// NewSuccessResponse builds one normalized bridge success response.
func NewSuccessResponse(statusCode int, contentType string, body []byte) *BridgeResponseEnvelopeV1 {
	return &BridgeResponseEnvelopeV1{
		StatusCode:  int32(statusCode),
		ContentType: strings.TrimSpace(contentType),
		Body:        append([]byte(nil), body...),
	}
}

// NewJSONResponse builds one JSON response using the provided raw bytes.
func NewJSONResponse(statusCode int, body []byte) *BridgeResponseEnvelopeV1 {
	return NewSuccessResponse(statusCode, "application/json", body)
}

// NewFailureResponse builds one normalized failure response with a plain-text body.
func NewFailureResponse(statusCode int, code string, message string) *BridgeResponseEnvelopeV1 {
	content := strings.TrimSpace(message)
	response := &BridgeResponseEnvelopeV1{
		StatusCode:  int32(statusCode),
		ContentType: "text/plain; charset=utf-8",
		Body:        []byte(content),
		Failure: &BridgeFailureV1{
			Code:    strings.TrimSpace(code),
			Message: content,
		},
	}
	return response
}

// NewUnauthorizedResponse builds a normalized 401 response.
func NewUnauthorizedResponse(message string) *BridgeResponseEnvelopeV1 {
	return NewFailureResponse(401, bridgeFailureCodeUnauthorized, messageOrDefault(message, "Unauthorized"))
}

// NewForbiddenResponse builds a normalized 403 response.
func NewForbiddenResponse(message string) *BridgeResponseEnvelopeV1 {
	return NewFailureResponse(403, bridgeFailureCodeForbidden, messageOrDefault(message, "Forbidden"))
}

// NewBadRequestResponse builds a normalized 400 response.
func NewBadRequestResponse(message string) *BridgeResponseEnvelopeV1 {
	return NewFailureResponse(400, bridgeFailureCodeBadRequest, messageOrDefault(message, "Bad Request"))
}

// NewNotFoundResponse builds a normalized 404 response.
func NewNotFoundResponse(message string) *BridgeResponseEnvelopeV1 {
	return NewFailureResponse(404, bridgeFailureCodeNotFound, messageOrDefault(message, "Not Found"))
}

// NewInternalErrorResponse builds a normalized 500 response.
func NewInternalErrorResponse(message string) *BridgeResponseEnvelopeV1 {
	return NewFailureResponse(500, bridgeFailureCodeInternal, messageOrDefault(message, "Internal Server Error"))
}

// messageOrDefault returns the trimmed message when present or the supplied
// fallback otherwise.
func messageOrDefault(value string, fallback string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return fallback
	}
	return normalized
}
