// This file tests host call request and response codec round trips.

package hostcall

import (
	"testing"

	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

// TestHostCallResponseEnvelopeRoundTrip verifies host call response envelopes
// survive a marshal/unmarshal round trip.
func TestHostCallResponseEnvelopeRoundTrip(t *testing.T) {
	original := &HostCallResponseEnvelope{
		Status:  HostCallStatusCapabilityDenied,
		Payload: []byte("missing host:runtime capability"),
	}
	data := MarshalHostCallResponse(original)
	decoded, err := UnmarshalHostCallResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Status != original.Status {
		t.Errorf("status: got %d, want %d", decoded.Status, original.Status)
	}
	if string(decoded.Payload) != string(original.Payload) {
		t.Errorf("payload: got %q, want %q", decoded.Payload, original.Payload)
	}
}

// TestHostCallSuccessResponseRoundTrip verifies the empty success helper
// preserves the expected success status through codec round trips.
func TestHostCallSuccessResponseRoundTrip(t *testing.T) {
	original := NewHostCallEmptySuccessResponse()
	data := MarshalHostCallResponse(original)
	decoded, err := UnmarshalHostCallResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Status != HostCallStatusSuccess {
		t.Errorf("status: got %d, want %d", decoded.Status, HostCallStatusSuccess)
	}
}

// TestHostCallErrorPayloadIsStructured verifies failed host calls expose
// stable metadata instead of a bare string payload.
func TestHostCallErrorPayloadIsStructured(t *testing.T) {
	response := NewHostCallErrorResponse(HostCallStatusCapabilityDenied, "missing host:runtime capability")
	payload, err := UnmarshalHostCallErrorPayload(response.Payload)
	if err != nil {
		t.Fatalf("decode error payload failed: %v", err)
	}
	if payload.ErrorCode != hostCallErrorCodeCapabilityDenied {
		t.Fatalf("expected capability error code, got %#v", payload)
	}
	if payload.MessageKey != "error.host_call.capability_denied" || payload.Fallback != "missing host:runtime capability" {
		t.Fatalf("unexpected structured error payload: %#v", payload)
	}
}

// TestHostCallErrorPayloadPreservesBizerrMetadata verifies host-call errors can
// carry business error localization metadata across the bridge.
func TestHostCallErrorPayloadPreservesBizerrMetadata(t *testing.T) {
	code := bizerr.MustDefine("PLUGIN_TARGET_USER_INVISIBLE", "Target user is not visible", gcode.CodeNotAuthorized)
	response := NewHostCallErrorResponseFromError(
		HostCallStatusCapabilityDenied,
		bizerr.NewCode(code, bizerr.P("userId", 42)),
	)
	payload, err := UnmarshalHostCallErrorPayload(response.Payload)
	if err != nil {
		t.Fatalf("decode error payload failed: %v", err)
	}
	if payload.ErrorCode != "PLUGIN_TARGET_USER_INVISIBLE" || payload.MessageKey != "error.plugin.target.user.invisible" {
		t.Fatalf("unexpected bizerr metadata: %#v", payload)
	}
	if payload.MessageParams["userId"] != float64(42) {
		t.Fatalf("expected userId parameter, got %#v", payload.MessageParams)
	}
}

// TestHostCallLogRequestRoundTrip verifies structured log requests preserve
// levels, messages, and fields through the codec.
func TestHostCallLogRequestRoundTrip(t *testing.T) {
	original := &HostCallLogRequest{
		Level:   LogLevelWarning,
		Message: "test warning message",
		Fields:  map[string]string{"key1": "val1", "key2": "val2"},
	}
	data := MarshalHostCallLogRequest(original)
	decoded, err := UnmarshalHostCallLogRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Level != original.Level {
		t.Errorf("level: got %d, want %d", decoded.Level, original.Level)
	}
	if decoded.Message != original.Message {
		t.Errorf("message: got %q, want %q", decoded.Message, original.Message)
	}
	if len(decoded.Fields) != 2 || decoded.Fields["key1"] != "val1" {
		t.Errorf("fields: got %v, want %v", decoded.Fields, original.Fields)
	}
}

// TestHostCallStateGetRequestRoundTrip verifies runtime state get requests
// round-trip through the codec.
func TestHostCallStateGetRequestRoundTrip(t *testing.T) {
	original := &HostCallStateGetRequest{Key: "counter"}
	data := MarshalHostCallStateGetRequest(original)
	decoded, err := UnmarshalHostCallStateGetRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Key != original.Key {
		t.Errorf("key: got %q, want %q", decoded.Key, original.Key)
	}
}

// TestHostCallStateGetResponseRoundTrip verifies runtime state get responses
// preserve both value and found flag through the codec.
func TestHostCallStateGetResponseRoundTrip(t *testing.T) {
	original := &HostCallStateGetResponse{Value: "42", Found: true}
	data := MarshalHostCallStateGetResponse(original)
	decoded, err := UnmarshalHostCallStateGetResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Value != original.Value {
		t.Errorf("value: got %q, want %q", decoded.Value, original.Value)
	}
	if decoded.Found != original.Found {
		t.Errorf("found: got %v, want %v", decoded.Found, original.Found)
	}
}

// TestHostCallStateSetRequestRoundTrip verifies runtime state set requests
// round-trip through the codec.
func TestHostCallStateSetRequestRoundTrip(t *testing.T) {
	original := &HostCallStateSetRequest{Key: "counter", Value: "43"}
	data := MarshalHostCallStateSetRequest(original)
	decoded, err := UnmarshalHostCallStateSetRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Key != original.Key {
		t.Errorf("key: got %q, want %q", decoded.Key, original.Key)
	}
	if decoded.Value != original.Value {
		t.Errorf("value: got %q, want %q", decoded.Value, original.Value)
	}
}

// TestHostCallStateDeleteRequestRoundTrip verifies runtime state delete
// requests round-trip through the codec.
func TestHostCallStateDeleteRequestRoundTrip(t *testing.T) {
	original := &HostCallStateDeleteRequest{Key: "counter"}
	data := MarshalHostCallStateDeleteRequest(original)
	decoded, err := UnmarshalHostCallStateDeleteRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Key != original.Key {
		t.Errorf("key: got %q, want %q", decoded.Key, original.Key)
	}
}
