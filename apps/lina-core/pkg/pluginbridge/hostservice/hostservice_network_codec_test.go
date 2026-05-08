// This file tests network host service codec round trips.

package hostservice

import "testing"

// TestHostServiceNetworkRequestRoundTrip verifies network requests preserve
// method, headers, and body payloads through the codec.
func TestHostServiceNetworkRequestRoundTrip(t *testing.T) {
	original := &HostServiceNetworkRequest{
		Method: "POST",
		Headers: map[string]string{
			"content-type": "application/json",
		},
		Body: []byte(`{"name":"ticket"}`),
	}

	data := MarshalHostServiceNetworkRequest(original)
	decoded, err := UnmarshalHostServiceNetworkRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Method != original.Method {
		t.Fatalf("request: got %#v want %#v", decoded, original)
	}
	if decoded.Headers["content-type"] != "application/json" {
		t.Fatalf("headers: got %#v", decoded.Headers)
	}
	if string(decoded.Body) != string(original.Body) {
		t.Fatalf("body: got %q want %q", decoded.Body, original.Body)
	}
}

// TestHostServiceNetworkResponseRoundTrip verifies network responses preserve
// status, headers, body, and content type through the codec.
func TestHostServiceNetworkResponseRoundTrip(t *testing.T) {
	original := &HostServiceNetworkResponse{
		StatusCode:  201,
		Headers:     map[string]string{"Content-Type": "application/json"},
		Body:        []byte(`{"ok":true}`),
		ContentType: "application/json",
	}

	data := MarshalHostServiceNetworkResponse(original)
	decoded, err := UnmarshalHostServiceNetworkResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.StatusCode != original.StatusCode {
		t.Fatalf("statusCode: got %d want %d", decoded.StatusCode, original.StatusCode)
	}
	if decoded.Headers["Content-Type"] != "application/json" {
		t.Fatalf("headers: got %#v", decoded.Headers)
	}
	if string(decoded.Body) != string(original.Body) {
		t.Fatalf("body: got %q want %q", decoded.Body, original.Body)
	}
	if decoded.ContentType != original.ContentType {
		t.Fatalf("contentType: got %q want %q", decoded.ContentType, original.ContentType)
	}
}
