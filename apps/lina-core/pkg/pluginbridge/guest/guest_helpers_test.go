// This file verifies the guest-side request binding, response writing,
// envelope accessors, and error classifier helpers published to dynamic
// plugin controllers.

package guest

import (
	"errors"
	"math"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
)

// guestHelpersTestPayload exercises BindJSON generic decoding across typed
// scalar and nested fields.
type guestHelpersTestPayload struct {
	Title  string `json:"title"`
	Amount int    `json:"amount"`
}

// TestBindJSONDecodesBody verifies the generic helper decodes valid envelope
// bodies into the requested type.
func TestBindJSONDecodesBody(t *testing.T) {
	request := &BridgeRequestEnvelopeV1{
		Request: &HTTPRequestSnapshotV1{
			Body: []byte(`{"title":"demo","amount":3}`),
		},
	}

	payload, err := BindJSON[guestHelpersTestPayload](request)
	if err != nil {
		t.Fatalf("expected BindJSON to succeed, got error: %v", err)
	}
	if payload == nil {
		t.Fatalf("expected BindJSON to return non-nil payload")
	}
	if payload.Title != "demo" || payload.Amount != 3 {
		t.Fatalf("expected decoded payload {demo 3}, got %#v", *payload)
	}
}

// TestBindJSONRejectsEmptyBody verifies the empty-body sentinel error is
// returned for nil envelopes, nil request snapshots, and zero-length bodies.
func TestBindJSONRejectsEmptyBody(t *testing.T) {
	cases := []struct {
		name    string
		request *BridgeRequestEnvelopeV1
	}{
		{name: "nil envelope", request: nil},
		{name: "nil request snapshot", request: &BridgeRequestEnvelopeV1{}},
		{name: "empty body", request: &BridgeRequestEnvelopeV1{Request: &HTTPRequestSnapshotV1{}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := BindJSON[guestHelpersTestPayload](c.request); !IsGuestBindJSONError(err) {
				t.Fatalf("expected empty-body sentinel error, got %v", err)
			}
		})
	}
}

// TestBindJSONRejectsInvalidJSON verifies malformed bodies return the invalid
// JSON sentinel wrapped with the underlying decoder message.
func TestBindJSONRejectsInvalidJSON(t *testing.T) {
	request := &BridgeRequestEnvelopeV1{
		Request: &HTTPRequestSnapshotV1{Body: []byte(`{"title":`)},
	}

	payload, err := BindJSON[guestHelpersTestPayload](request)
	if payload != nil {
		t.Fatalf("expected BindJSON to return nil payload on parse error, got %#v", payload)
	}
	if !IsGuestBindJSONError(err) {
		t.Fatalf("expected invalid-json sentinel error, got %v", err)
	}
}

// TestClassifyBindJSONErrorReturnsBadRequest verifies the helper maps the
// BindJSON sentinel to a normalized 400 response and leaves unrelated errors
// unhandled.
func TestClassifyBindJSONErrorReturnsBadRequest(t *testing.T) {
	_, bindErr := BindJSON[guestHelpersTestPayload](nil)
	response := ClassifyBindJSONError(bindErr)
	if response == nil || response.StatusCode != 400 {
		t.Fatalf("expected 400 response for BindJSON error, got %#v", response)
	}

	if unrelated := ClassifyBindJSONError(errors.New("unrelated")); unrelated != nil {
		t.Fatalf("expected nil response for non-bindjson error, got %#v", unrelated)
	}
	if ClassifyBindJSONError(nil) != nil {
		t.Fatalf("expected nil response for nil error")
	}
}

// TestWriteJSONMarshalsPayload verifies the helper produces a 200 JSON
// response with serialized body bytes.
func TestWriteJSONMarshalsPayload(t *testing.T) {
	response, err := WriteJSON(200, guestHelpersTestPayload{Title: "demo", Amount: 5})
	if err != nil {
		t.Fatalf("expected WriteJSON to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != 200 {
		t.Fatalf("expected 200 response, got %#v", response)
	}
	if string(response.Body) != `{"title":"demo","amount":5}` {
		t.Fatalf("expected marshaled payload body, got %q", string(response.Body))
	}
	if response.ContentType != "application/json" {
		t.Fatalf("expected application/json content type, got %q", response.ContentType)
	}
}

// unmarshalableValue triggers a json.Marshal failure so WriteJSON's error path
// can be covered.
type unmarshalableValue struct{}

// MarshalJSON returns a non-finite float to force json.Marshal to fail.
func (unmarshalableValue) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal boom")
}

// TestWriteJSONReturnsMarshalErrors verifies marshal failures are reported to
// the caller rather than swallowed into a partial response.
func TestWriteJSONReturnsMarshalErrors(t *testing.T) {
	if _, err := WriteJSON(200, math.Inf(1)); err == nil {
		t.Fatalf("expected marshal failure for +Inf payload")
	}
	if _, err := WriteJSON(200, unmarshalableValue{}); err == nil {
		t.Fatalf("expected marshal failure for custom MarshalJSON error")
	}
}

// TestPathParamReadsTrimmedValue verifies path parameter access handles nil
// envelopes, missing snapshots, absent keys, and whitespace.
func TestPathParamReadsTrimmedValue(t *testing.T) {
	if PathParam(nil, "id") != "" {
		t.Fatalf("expected empty value for nil envelope")
	}
	if PathParam(&BridgeRequestEnvelopeV1{}, "id") != "" {
		t.Fatalf("expected empty value for missing route snapshot")
	}

	request := &BridgeRequestEnvelopeV1{
		Route: &RouteMatchSnapshotV1{
			PathParams: map[string]string{"id": "  abc\n"},
		},
	}
	if got := PathParam(request, "id"); got != "abc" {
		t.Fatalf("expected trimmed value abc, got %q", got)
	}
	if got := PathParam(request, "missing"); got != "" {
		t.Fatalf("expected empty value for missing key, got %q", got)
	}
}

// TestQueryValueReadsFirstEntry verifies query access returns the first value
// associated with the key and trims whitespace.
func TestQueryValueReadsFirstEntry(t *testing.T) {
	request := &BridgeRequestEnvelopeV1{
		Route: &RouteMatchSnapshotV1{
			QueryValues: map[string][]string{
				"keyword":  {" demo "},
				"empty":    {},
				"multiple": {"first", "second"},
			},
		},
	}

	if got := QueryValue(request, "keyword"); got != "demo" {
		t.Fatalf("expected trimmed keyword value demo, got %q", got)
	}
	if got := QueryValue(request, "empty"); got != "" {
		t.Fatalf("expected empty string for zero-length values slice, got %q", got)
	}
	if got := QueryValue(request, "multiple"); got != "first" {
		t.Fatalf("expected first query value, got %q", got)
	}
	if got := QueryValue(nil, "keyword"); got != "" {
		t.Fatalf("expected empty value for nil envelope")
	}
}

// TestQueryIntParsesValidIntegers verifies parse failures and missing keys
// both return zero, matching the previous plugin-local helper behavior.
func TestQueryIntParsesValidIntegers(t *testing.T) {
	request := &BridgeRequestEnvelopeV1{
		Route: &RouteMatchSnapshotV1{
			QueryValues: map[string][]string{
				"pageNum":  {"7"},
				"pageSize": {"abc"},
			},
		},
	}

	if got := QueryInt(request, "pageNum"); got != 7 {
		t.Fatalf("expected parsed pageNum 7, got %d", got)
	}
	if got := QueryInt(request, "pageSize"); got != 0 {
		t.Fatalf("expected parse failure to return 0, got %d", got)
	}
	if got := QueryInt(request, "missing"); got != 0 {
		t.Fatalf("expected missing key to return 0, got %d", got)
	}
}

// TestQueryFlagMatchesTruthyValues verifies truthy literals are recognized
// case-insensitively while unrelated values return false.
func TestQueryFlagMatchesTruthyValues(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{"one", "1", true},
		{"true lower", "true", true},
		{"true upper", "TRUE", true},
		{"yes", "yes", true},
		{"on", "on", true},
		{"zero", "0", false},
		{"false", "false", false},
		{"empty", "", false},
		{"garbage", "maybe", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			request := &BridgeRequestEnvelopeV1{
				Route: &RouteMatchSnapshotV1{
					QueryValues: map[string][]string{"flag": {c.value}},
				},
			}
			if got := QueryFlag(request, "flag"); got != c.want {
				t.Fatalf("expected QueryFlag(%q)=%v, got %v", c.value, c.want, got)
			}
		})
	}

	if QueryFlag(nil, "flag") {
		t.Fatalf("expected nil envelope to return false")
	}
}

// errDemoInvalidInput mocks a plugin-side sentinel used by classifier tests.
var errDemoInvalidInput = gerror.New("demo invalid input")

// errDemoNotFound mocks a plugin-side sentinel used by classifier tests.
var errDemoNotFound = gerror.New("demo not found")

// TestErrorClassifierPrefersBindJSONErrorsFirst verifies the classifier maps
// BindJSON sentinels to 400 before consulting plugin-supplied cases.
func TestErrorClassifierPrefersBindJSONErrorsFirst(t *testing.T) {
	classifier := NewErrorClassifier(
		NewErrorCase(func(err error) bool { return gerror.Is(err, errDemoInvalidInput) }, NewBadRequestResponse),
	)

	_, bindErr := BindJSON[guestHelpersTestPayload](nil)
	response := classifier.Classify(bindErr)
	if response == nil || response.StatusCode != 400 {
		t.Fatalf("expected BindJSON error to yield 400 response, got %#v", response)
	}
}

// TestErrorClassifierDispatchesByRegistrationOrder verifies plugin cases are
// evaluated in order and the first match wins.
func TestErrorClassifierDispatchesByRegistrationOrder(t *testing.T) {
	classifier := NewErrorClassifier(
		NewErrorCase(func(err error) bool { return gerror.Is(err, errDemoInvalidInput) }, NewBadRequestResponse),
		NewErrorCase(func(err error) bool { return gerror.Is(err, errDemoNotFound) }, NewNotFoundResponse),
	)

	if response := classifier.Classify(gerror.Wrap(errDemoInvalidInput, "bad")); response == nil || response.StatusCode != 400 {
		t.Fatalf("expected invalid-input case to yield 400 response, got %#v", response)
	}
	if response := classifier.Classify(gerror.Wrap(errDemoNotFound, "missing")); response == nil || response.StatusCode != 404 {
		t.Fatalf("expected not-found case to yield 404 response, got %#v", response)
	}
}

// TestErrorClassifierFallsBackToInternalError verifies unmatched errors and
// nil inputs both produce a 500 response so plugins always emit a response.
func TestErrorClassifierFallsBackToInternalError(t *testing.T) {
	classifier := NewErrorClassifier()

	if response := classifier.Classify(errors.New("boom")); response == nil || response.StatusCode != 500 {
		t.Fatalf("expected unmatched error to yield 500 response, got %#v", response)
	}
	if response := classifier.Classify(nil); response == nil || response.StatusCode != 500 {
		t.Fatalf("expected nil error to yield 500 response, got %#v", response)
	}
}

// TestNewErrorClassifierSkipsNilCases verifies cases with nil matcher or
// builder are discarded so invalid registrations do not panic at runtime.
func TestNewErrorClassifierSkipsNilCases(t *testing.T) {
	classifier := NewErrorClassifier(
		NewErrorCase(nil, NewBadRequestResponse),
		NewErrorCase(func(error) bool { return true }, nil),
	)
	if response := classifier.Classify(errors.New("boom")); response == nil || response.StatusCode != 500 {
		t.Fatalf("expected invalid cases to be skipped and yield 500 fallback, got %#v", response)
	}
}
