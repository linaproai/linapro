// This file tests Jobs host service request codec round trips.

package protocol

import "testing"

// TestHostServiceJobsRegisterRequestRoundTrip verifies Jobs registration
// requests preserve declaration metadata used by dynamic-plugin discovery.
func TestHostServiceJobsRegisterRequestRoundTrip(t *testing.T) {
	original := &HostServiceJobsRegisterRequest{
		Contract: &JobContract{
			Name:           "heartbeat",
			DisplayName:    "Dynamic Plugin Heartbeat",
			Description:    "Runs a dynamic plugin heartbeat.",
			Pattern:        "# */10 * * * *",
			Timezone:       DefaultJobContractTimezone,
			Scope:          JobScopeAllNode,
			Concurrency:    JobConcurrencySingleton,
			MaxConcurrency: 1,
			TimeoutSeconds: 30,
			RequestType:    "JobHeartbeatReq",
			InternalPath:   "/job-heartbeat",
		},
	}

	data := MarshalHostServiceJobsRegisterRequest(original)
	decoded, err := UnmarshalHostServiceJobsRegisterRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Contract == nil {
		t.Fatal("expected decoded contract")
	}
	if decoded.Contract.Name != original.Contract.Name ||
		decoded.Contract.DisplayName != original.Contract.DisplayName ||
		decoded.Contract.Description != original.Contract.Description ||
		decoded.Contract.Pattern != original.Contract.Pattern ||
		decoded.Contract.Timezone != original.Contract.Timezone ||
		decoded.Contract.Scope != original.Contract.Scope ||
		decoded.Contract.Concurrency != original.Contract.Concurrency ||
		decoded.Contract.MaxConcurrency != original.Contract.MaxConcurrency ||
		decoded.Contract.TimeoutSeconds != original.Contract.TimeoutSeconds ||
		decoded.Contract.RequestType != original.Contract.RequestType ||
		decoded.Contract.InternalPath != original.Contract.InternalPath {
		t.Fatalf("decoded contract mismatch: got %#v want %#v", decoded.Contract, original.Contract)
	}
}

// TestHostServiceJobsRegisterEmptyRequestRoundTrip verifies empty registration
// payloads decode without synthesizing a partial declaration.
func TestHostServiceJobsRegisterEmptyRequestRoundTrip(t *testing.T) {
	decoded, err := UnmarshalHostServiceJobsRegisterRequest(nil)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded == nil {
		t.Fatal("expected decoded request")
	}
	if decoded.Contract != nil {
		t.Fatalf("expected empty request to have nil contract, got %#v", decoded.Contract)
	}
}
