// This file exposes reusable guest-side request and response helpers so
// dynamic plugin controllers can decode envelope inputs, emit JSON responses,
// and classify business errors without duplicating bridge scaffolding.

package pluginbridge

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// errGuestBindJSONEmptyBody indicates the envelope request body is missing and
// the guest controller should translate the failure into a 400 response.
var errGuestBindJSONEmptyBody = gerror.New("请求体不能为空")

// errGuestBindJSONInvalidJSON indicates the envelope body is not valid JSON and
// the guest controller should translate the failure into a 400 response.
var errGuestBindJSONInvalidJSON = gerror.New("请求体 JSON 无法解析")

// IsGuestBindJSONError reports whether the error originates from BindJSON body
// decoding and therefore represents a client-supplied payload problem.
func IsGuestBindJSONError(err error) bool {
	if err == nil {
		return false
	}
	return gerror.Is(err, errGuestBindJSONEmptyBody) || gerror.Is(err, errGuestBindJSONInvalidJSON)
}

// BindJSON decodes the envelope request body into T. It returns a typed sentinel
// error for empty bodies and malformed JSON so plugin controllers can translate
// both cases into a 400 response via ClassifyBindJSONError.
func BindJSON[T any](request *BridgeRequestEnvelopeV1) (*T, error) {
	out := new(T)
	if err := bindJSONBody(request, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ClassifyBindJSONError converts a BindJSON sentinel error into the canonical
// 400 bridge response. Callers that also need to recognize other business
// errors should compose this with ErrorClassifier cases.
func ClassifyBindJSONError(err error) *BridgeResponseEnvelopeV1 {
	if err == nil {
		return nil
	}
	if IsGuestBindJSONError(err) {
		return NewBadRequestResponse(err.Error())
	}
	return nil
}

// WriteJSON marshals the payload into a JSON bridge response using the given
// HTTP status code. Marshal failures are returned as errors so the guest
// runtime can surface them via the standard internal-error fallback path.
func WriteJSON(statusCode int, payload any) (*BridgeResponseEnvelopeV1, error) {
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal bridge JSON response failed")
	}
	return NewJSONResponse(statusCode, content), nil
}

// PathParam reads one trimmed path parameter from the matched route snapshot.
// It returns an empty string when the envelope or snapshot is absent.
func PathParam(request *BridgeRequestEnvelopeV1, key string) string {
	if request == nil || request.Route == nil || len(request.Route.PathParams) == 0 {
		return ""
	}
	return strings.TrimSpace(request.Route.PathParams[key])
}

// QueryValue reads the first trimmed query value for the given key from the
// matched route snapshot. It returns an empty string when the envelope or
// snapshot is absent, or when the key has no associated value.
func QueryValue(request *BridgeRequestEnvelopeV1, key string) string {
	if request == nil || request.Route == nil || len(request.Route.QueryValues) == 0 {
		return ""
	}

	values := request.Route.QueryValues[key]
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

// QueryInt reads the first query value for the given key and parses it as an
// int. Missing keys and parse failures both return zero so controllers can
// treat the value as an optional pagination or filter input.
func QueryInt(request *BridgeRequestEnvelopeV1, key string) int {
	value := QueryValue(request, key)
	if value == "" {
		return 0
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

// QueryFlag reports whether any value associated with the given query key is a
// truthy flag value (1, true, yes, on). Comparison is case-insensitive and
// whitespace-tolerant.
func QueryFlag(request *BridgeRequestEnvelopeV1, key string) bool {
	if request == nil || request.Route == nil || len(request.Route.QueryValues) == 0 {
		return false
	}

	for _, value := range request.Route.QueryValues[key] {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "1", "true", "yes", "on":
			return true
		}
	}
	return false
}

// ErrorResponseBuilder synthesizes a bridge response for a classified error.
// Classifier composition wraps each matching predicate with one builder.
type ErrorResponseBuilder func(message string) *BridgeResponseEnvelopeV1

// ErrorMatcher reports whether one classifier case applies to the given error.
type ErrorMatcher func(err error) bool

// ErrorCase binds one matcher to one response builder, letting ErrorClassifier
// map specific business sentinel checks onto normalized bridge responses.
type ErrorCase struct {
	// Match reports whether the case applies to the given error.
	Match ErrorMatcher
	// Build synthesizes the bridge response when Match returns true.
	Build ErrorResponseBuilder
}

// NewErrorCase constructs one classifier case from a matcher predicate and a
// response builder.
func NewErrorCase(match ErrorMatcher, build ErrorResponseBuilder) ErrorCase {
	return ErrorCase{Match: match, Build: build}
}

// ErrorClassifier maps plugin business errors onto bridge responses. Cases are
// evaluated in registration order and the first matching case wins. Errors
// that no case recognizes fall back to NewInternalErrorResponse.
type ErrorClassifier interface {
	// Classify translates one error into a bridge response envelope. The nil
	// input is treated as an anonymous internal error so guest controllers can
	// always emit a response.
	Classify(err error) *BridgeResponseEnvelopeV1
}

// errorClassifier is the default ErrorClassifier implementation evaluated in
// registration order.
type errorClassifier struct {
	cases []ErrorCase
}

// NewErrorClassifier returns an ErrorClassifier that evaluates each case in
// order. BindJSON sentinel errors are always recognized first so plugin
// controllers do not need to register that case explicitly.
func NewErrorClassifier(cases ...ErrorCase) ErrorClassifier {
	filtered := make([]ErrorCase, 0, len(cases))
	for _, item := range cases {
		if item.Match == nil || item.Build == nil {
			continue
		}
		filtered = append(filtered, item)
	}
	return &errorClassifier{cases: filtered}
}

// Classify evaluates each registered case and returns the first matching
// bridge response. BindJSON errors are handled before plugin cases so malformed
// client inputs always normalize to 400. Unmatched errors fall back to a 500
// response carrying the error message.
func (c *errorClassifier) Classify(err error) *BridgeResponseEnvelopeV1 {
	if err == nil {
		return NewInternalErrorResponse("Bridge execution failed")
	}
	if response := ClassifyBindJSONError(err); response != nil {
		return response
	}
	for _, item := range c.cases {
		if item.Match(err) {
			return item.Build(err.Error())
		}
	}
	return NewInternalErrorResponse(err.Error())
}

// Static interface compliance guard for the default classifier.
var _ ErrorClassifier = (*errorClassifier)(nil)
