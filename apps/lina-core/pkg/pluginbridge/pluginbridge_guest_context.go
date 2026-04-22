// This file stores guest controller request context and response state so
// typed dynamic-plugin handlers can read bridge metadata and emit custom
// responses without depending on the raw envelope signature.

package pluginbridge

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// guestControllerContextKey is the private context key used to stash one guest
// request envelope and in-flight response state for typed guest controllers.
type guestControllerContextKey struct{}

// guestControllerContextState stores the raw request envelope together with the
// mutable response state accumulated by typed guest controllers.
type guestControllerContextState struct {
	request  *BridgeRequestEnvelopeV1
	response *guestControllerResponseState
}

// guestControllerResponseState stores headers, status, and optional raw body
// written by one typed guest controller.
type guestControllerResponseState struct {
	statusCode  int
	contentType string
	headers     map[string][]string
	body        []byte
	written     bool
}

// ResponseError exposes one prebuilt bridge response through the error
// channel so typed guest controllers can reuse ErrorClassifier results.
type ResponseError interface {
	error
	// Response returns the bridge response that should be returned to the host.
	Response() *BridgeResponseEnvelopeV1
}

// guestResponseError is the default ResponseError implementation.
type guestResponseError struct {
	response *BridgeResponseEnvelopeV1
}

// newGuestControllerContext creates one typed-controller context backed by the
// supplied request envelope and an empty mutable response state.
func newGuestControllerContext(request *BridgeRequestEnvelopeV1) context.Context {
	return context.WithValue(context.Background(), guestControllerContextKey{}, &guestControllerContextState{
		request: request,
		response: &guestControllerResponseState{
			headers: make(map[string][]string),
		},
	})
}

// guestControllerStateFromContext returns the typed guest controller state
// stored on the given context, if present.
func guestControllerStateFromContext(ctx context.Context) *guestControllerContextState {
	if ctx == nil {
		return nil
	}
	state, _ := ctx.Value(guestControllerContextKey{}).(*guestControllerContextState)
	return state
}

// RequestEnvelopeFromContext returns the raw bridge request envelope stored on
// one typed guest controller context.
func RequestEnvelopeFromContext(ctx context.Context) *BridgeRequestEnvelopeV1 {
	state := guestControllerStateFromContext(ctx)
	if state == nil {
		return nil
	}
	return state.request
}

// SetResponseHeader appends one response header value to the typed guest
// controller response state.
func SetResponseHeader(ctx context.Context, key string, values ...string) error {
	state := guestControllerStateFromContext(ctx)
	if state == nil || state.response == nil {
		return gerror.New("typed guest controller context is missing response state")
	}

	normalizedKey := strings.TrimSpace(key)
	if normalizedKey == "" {
		return gerror.New("typed guest controller response header key cannot be empty")
	}

	normalizedValues := make([]string, 0, len(values))
	for _, value := range values {
		normalizedValue := strings.TrimSpace(value)
		if normalizedValue == "" {
			continue
		}
		normalizedValues = append(normalizedValues, normalizedValue)
	}
	if len(normalizedValues) == 0 {
		return nil
	}

	state.response.headers[normalizedKey] = append(
		state.response.headers[normalizedKey],
		normalizedValues...,
	)
	return nil
}

// SetResponseStatusCode stores one custom status code on the typed guest
// controller response state.
func SetResponseStatusCode(ctx context.Context, statusCode int) error {
	state := guestControllerStateFromContext(ctx)
	if state == nil || state.response == nil {
		return gerror.New("typed guest controller context is missing response state")
	}
	if statusCode <= 0 {
		return gerror.Newf("typed guest controller status code must be positive: %d", statusCode)
	}
	state.response.statusCode = statusCode
	return nil
}

// WriteResponse writes one fully manual response body for the typed guest
// controller and marks the response as already materialized.
func WriteResponse(
	ctx context.Context,
	statusCode int,
	contentType string,
	body []byte,
) error {
	state := guestControllerStateFromContext(ctx)
	if state == nil || state.response == nil {
		return gerror.New("typed guest controller context is missing response state")
	}
	if statusCode <= 0 {
		return gerror.Newf("typed guest controller status code must be positive: %d", statusCode)
	}

	state.response.statusCode = statusCode
	state.response.contentType = strings.TrimSpace(contentType)
	state.response.body = append(state.response.body[:0], body...)
	state.response.written = true
	return nil
}

// WriteNoContent marks one typed guest controller response as explicit empty
// success without requiring a placeholder DTO.
func WriteNoContent(ctx context.Context, statusCode int) error {
	if statusCode <= 0 {
		statusCode = 204
	}
	return WriteResponse(ctx, statusCode, "", nil)
}

// buildGuestControllerResponse materializes the final bridge response for one
// typed guest controller return value and its accumulated response state.
func buildGuestControllerResponse(
	ctx context.Context,
	payload interface{},
) (*BridgeResponseEnvelopeV1, error) {
	state := guestControllerStateFromContext(ctx)
	if state == nil || state.response == nil {
		if payload == nil {
			return nil, gerror.New("typed guest controller returned no payload without response state")
		}
		return WriteJSON(200, payload)
	}

	if state.response.written {
		response := NewSuccessResponse(
			defaultGuestResponseStatus(state.response.statusCode, 200),
			state.response.contentType,
			state.response.body,
		)
		response.Headers = cloneGuestResponseHeaders(state.response.headers)
		return response, nil
	}

	if payload != nil {
		response, err := WriteJSON(defaultGuestResponseStatus(state.response.statusCode, 200), payload)
		if err != nil {
			return nil, err
		}
		response.Headers = cloneGuestResponseHeaders(state.response.headers)
		return response, nil
	}

	if state.response.statusCode > 0 || len(state.response.headers) > 0 {
		response := NewSuccessResponse(
			defaultGuestResponseStatus(state.response.statusCode, 200),
			state.response.contentType,
			nil,
		)
		response.Headers = cloneGuestResponseHeaders(state.response.headers)
		return response, nil
	}

	return nil, gerror.New("typed guest controller returned nil payload without writing a response")
}

// defaultGuestResponseStatus returns fallback when the stored status code is
// zero or negative.
func defaultGuestResponseStatus(statusCode int, fallback int) int {
	if statusCode > 0 {
		return statusCode
	}
	return fallback
}

// cloneGuestResponseHeaders deep-copies one response header map so the final
// bridge response cannot be mutated through shared state.
func cloneGuestResponseHeaders(headers map[string][]string) map[string][]string {
	if len(headers) == 0 {
		return nil
	}

	cloned := make(map[string][]string, len(headers))
	for key, values := range headers {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

// Error returns one readable summary of the bridge response error for logs or
// debugging surfaces.
func (e *guestResponseError) Error() string {
	if e == nil || e.response == nil {
		return "bridge response error"
	}
	if e.response.Failure != nil && strings.TrimSpace(e.response.Failure.Message) != "" {
		return e.response.Failure.Message
	}
	if len(e.response.Body) > 0 {
		return string(e.response.Body)
	}
	return fmt.Sprintf("bridge response error (status=%d)", e.response.StatusCode)
}

// Response returns the prebuilt bridge response carried by the error wrapper.
func (e *guestResponseError) Response() *BridgeResponseEnvelopeV1 {
	if e == nil {
		return nil
	}
	return e.response
}

// NewResponseError wraps one prebuilt bridge response so typed guest
// controllers can return it through the error channel.
func NewResponseError(response *BridgeResponseEnvelopeV1) error {
	if response == nil {
		return gerror.New("bridge response cannot be nil")
	}
	return &guestResponseError{response: response}
}

// ResponseFromError extracts one prebuilt bridge response from an error, if
// the error implements ResponseError.
func ResponseFromError(err error) *BridgeResponseEnvelopeV1 {
	if err == nil {
		return nil
	}

	var target ResponseError
	if !errors.As(err, &target) {
		return nil
	}
	return target.Response()
}

// Static interface compliance guards for the default response error wrapper.
var _ ResponseError = (*guestResponseError)(nil)
