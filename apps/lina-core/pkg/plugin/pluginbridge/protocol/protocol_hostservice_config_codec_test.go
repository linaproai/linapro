// This file tests shared configuration payload codec round trips.

package protocol

import "testing"

// TestHostServiceConfigKeyRequestRoundTrip verifies config key requests preserve keys.
func TestHostServiceConfigKeyRequestRoundTrip(t *testing.T) {
	original := &HostServiceConfigKeyRequest{Key: "monitor.interval"}

	data := MarshalHostServiceConfigKeyRequest(original)
	decoded, err := UnmarshalHostServiceConfigKeyRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Key != original.Key {
		t.Fatalf("key: got %q want %q", decoded.Key, original.Key)
	}
}

// TestHostServiceConfigValueResponseRoundTrip verifies config values preserve found flags.
func TestHostServiceConfigValueResponseRoundTrip(t *testing.T) {
	original := &HostServiceConfigValueResponse{
		Value: `{"interval":"1m"}`,
		Found: true,
	}

	data := MarshalHostServiceConfigValueResponse(original)
	decoded, err := UnmarshalHostServiceConfigValueResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Value != original.Value {
		t.Fatalf("value: got %q want %q", decoded.Value, original.Value)
	}
	if !decoded.Found {
		t.Fatal("found: expected true")
	}
}
