// This file tests lock host service request and response codec round trips.

package pluginbridge

import "testing"

func TestHostServiceLockAcquireRoundTrip(t *testing.T) {
	original := &HostServiceLockAcquireResponse{
		Acquired: true,
		Ticket:   "ticket-1",
		ExpireAt: "2026-04-15T12:10:00Z",
	}

	requestData := MarshalHostServiceLockAcquireRequest(&HostServiceLockAcquireRequest{LeaseMillis: 5000})
	request, err := UnmarshalHostServiceLockAcquireRequest(requestData)
	if err != nil {
		t.Fatalf("request unmarshal failed: %v", err)
	}
	if request.LeaseMillis != 5000 {
		t.Fatalf("leaseMillis: got %d want 5000", request.LeaseMillis)
	}

	data := MarshalHostServiceLockAcquireResponse(original)
	decoded, err := UnmarshalHostServiceLockAcquireResponse(data)
	if err != nil {
		t.Fatalf("response unmarshal failed: %v", err)
	}
	if !decoded.Acquired || decoded.Ticket != original.Ticket || decoded.ExpireAt != original.ExpireAt {
		t.Fatalf("response: got %#v want %#v", decoded, original)
	}
}

func TestHostServiceLockRenewRoundTrip(t *testing.T) {
	requestData := MarshalHostServiceLockRenewRequest(&HostServiceLockRenewRequest{Ticket: "ticket-2"})
	request, err := UnmarshalHostServiceLockRenewRequest(requestData)
	if err != nil {
		t.Fatalf("request unmarshal failed: %v", err)
	}
	if request.Ticket != "ticket-2" {
		t.Fatalf("ticket: got %q want %q", request.Ticket, "ticket-2")
	}

	original := &HostServiceLockRenewResponse{ExpireAt: "2026-04-15T12:20:00Z"}
	data := MarshalHostServiceLockRenewResponse(original)
	decoded, err := UnmarshalHostServiceLockRenewResponse(data)
	if err != nil {
		t.Fatalf("response unmarshal failed: %v", err)
	}
	if decoded.ExpireAt != original.ExpireAt {
		t.Fatalf("expireAt: got %q want %q", decoded.ExpireAt, original.ExpireAt)
	}
}

func TestHostServiceLockReleaseRequestRoundTrip(t *testing.T) {
	original := &HostServiceLockReleaseRequest{Ticket: "ticket-3"}

	data := MarshalHostServiceLockReleaseRequest(original)
	decoded, err := UnmarshalHostServiceLockReleaseRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Ticket != original.Ticket {
		t.Fatalf("ticket: got %q want %q", decoded.Ticket, original.Ticket)
	}
}
