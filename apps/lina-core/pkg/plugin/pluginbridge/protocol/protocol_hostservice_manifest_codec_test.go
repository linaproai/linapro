// This file tests manifest host service request and response codec round trips.

package protocol

import "testing"

// TestHostServiceManifestGetRequestRoundTrip verifies manifest paths are preserved.
func TestHostServiceManifestGetRequestRoundTrip(t *testing.T) {
	original := &HostServiceManifestGetRequest{Path: "metadata.yaml"}

	data := MarshalHostServiceManifestGetRequest(original)
	decoded, err := UnmarshalHostServiceManifestGetRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Path != original.Path {
		t.Fatalf("path: got %q want %q", decoded.Path, original.Path)
	}
}

// TestHostServiceManifestGetResponseRoundTrip verifies manifest bodies preserve found flags.
func TestHostServiceManifestGetResponseRoundTrip(t *testing.T) {
	original := &HostServiceManifestGetResponse{
		Found: true,
		Body:  []byte("name: demo\n"),
	}

	data := MarshalHostServiceManifestGetResponse(original)
	decoded, err := UnmarshalHostServiceManifestGetResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !decoded.Found {
		t.Fatal("found: expected true")
	}
	if string(decoded.Body) != string(original.Body) {
		t.Fatalf("body: got %q want %q", string(decoded.Body), string(original.Body))
	}
}
