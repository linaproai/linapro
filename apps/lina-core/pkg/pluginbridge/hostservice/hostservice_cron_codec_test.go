// This file verifies cron host service request codec round trips.

package hostservice

import "testing"

// TestHostServiceCronRegisterRequestRoundTrip verifies dynamic-plugin cron
// declarations preserve all supported fields across host-service encoding.
func TestHostServiceCronRegisterRequestRoundTrip(t *testing.T) {
	original := &HostServiceCronRegisterRequest{
		Contract: &CronContract{
			Name:           "heartbeat",
			DisplayName:    "Heartbeat",
			Description:    "Runs one plugin heartbeat.",
			Pattern:        "# */10 * * * *",
			Timezone:       "Asia/Shanghai",
			Scope:          CronScopeAllNode,
			Concurrency:    CronConcurrencySingleton,
			MaxConcurrency: 1,
			TimeoutSeconds: 30,
			RequestType:    "CronHeartbeatReq",
			InternalPath:   "/cron-heartbeat",
		},
	}

	data := MarshalHostServiceCronRegisterRequest(original)
	decoded, err := UnmarshalHostServiceCronRegisterRequest(data)
	if err != nil {
		t.Fatalf("expected cron register request decode to succeed, got %v", err)
	}
	if decoded == nil || decoded.Contract == nil {
		t.Fatalf("expected decoded cron register contract, got %#v", decoded)
	}
	if decoded.Contract.Name != original.Contract.Name || decoded.Contract.Pattern != original.Contract.Pattern {
		t.Fatalf("expected decoded cron register contract to preserve key fields, got %#v", decoded.Contract)
	}
	if decoded.Contract.InternalPath != original.Contract.InternalPath || decoded.Contract.RequestType != original.Contract.RequestType {
		t.Fatalf("expected decoded cron register dispatch fields to match, got %#v", decoded.Contract)
	}
	if decoded.Contract.Scope != original.Contract.Scope || decoded.Contract.Concurrency != original.Contract.Concurrency {
		t.Fatalf("expected decoded cron register enums to match, got %#v", decoded.Contract)
	}
}
