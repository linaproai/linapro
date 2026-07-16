// This file verifies execution-identity helpers used by scheduled-job handlers
// for at-least-once idempotency.

package jobmeta

import (
	"context"
	"testing"
)

// TestExecutionLogIDRoundTrip verifies log ids are stored and read back.
func TestExecutionLogIDRoundTrip(t *testing.T) {
	ctx := WithExecutionLogID(context.Background(), 42)
	logID, ok := ExecutionLogID(ctx)
	if !ok {
		t.Fatal("expected ExecutionLogID to be present")
	}
	if logID != 42 {
		t.Fatalf("expected ExecutionLogID=42, got %d", logID)
	}
}

// TestExecutionLogIDRejectsMissingAndInvalid verifies empty and non-positive ids.
func TestExecutionLogIDRejectsMissingAndInvalid(t *testing.T) {
	if _, ok := ExecutionLogID(context.Background()); ok {
		t.Fatal("expected empty context to report missing execution log id")
	}
	if _, ok := ExecutionLogID(WithExecutionLogID(context.Background(), 0)); ok {
		t.Fatal("expected non-positive log id to be ignored")
	}
	if _, ok := ExecutionLogID(WithExecutionLogID(context.Background(), -1)); ok {
		t.Fatal("expected negative log id to be ignored")
	}
}
