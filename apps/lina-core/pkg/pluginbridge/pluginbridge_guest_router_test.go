// This file tests reflected guest controller dispatch and guest route fallback behavior.

package pluginbridge

import (
	"errors"
	"testing"
)

// guestRouteTestController provides reflected handlers for guest route
// dispatcher tests.
type guestRouteTestController struct{}

// BackendSummary returns a deterministic JSON payload for successful dispatch
// assertions.
func (c *guestRouteTestController) BackendSummary(
	request *BridgeRequestEnvelopeV1,
) (*BridgeResponseEnvelopeV1, error) {
	return NewJSONResponse(200, []byte(`{"requestId":"`+request.RequestID+`"}`)), nil
}

// Failure returns a controller error so dispatcher error propagation can be
// asserted.
func (c *guestRouteTestController) Failure(
	_ *BridgeRequestEnvelopeV1,
) (*BridgeResponseEnvelopeV1, error) {
	return nil, errors.New("boom")
}

// IgnoredMethod does not match the dispatcher signature and should be ignored
// during controller reflection.
func (c *guestRouteTestController) IgnoredMethod() {}

// TestGuestControllerRouteDispatcherDispatchesByRequestType verifies reflected
// request-type dispatch finds the matching controller method.
func TestGuestControllerRouteDispatcherDispatchesByRequestType(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestRouteTestController{})
	if err != nil {
		t.Fatalf("expected dispatcher creation to succeed, got error: %v", err)
	}

	response, err := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		RequestID: "req-1",
		Route: &RouteMatchSnapshotV1{
			RequestType: "BackendSummaryReq",
		},
	})
	if err != nil {
		t.Fatalf("expected request dispatch to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != 200 {
		t.Fatalf("expected dispatcher to return 200 response, got %#v", response)
	}
	if string(response.Body) != `{"requestId":"req-1"}` {
		t.Fatalf("expected dispatcher to preserve controller response body, got %q", string(response.Body))
	}
}

// TestGuestControllerRouteDispatcherReturnsControllerError verifies controller
// errors are returned without a synthesized bridge response.
func TestGuestControllerRouteDispatcherReturnsControllerError(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestRouteTestController{})
	if err != nil {
		t.Fatalf("expected dispatcher creation to succeed, got error: %v", err)
	}

	response, dispatchErr := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		Route: &RouteMatchSnapshotV1{
			RequestType: "FailureReq",
		},
	})
	if dispatchErr == nil || dispatchErr.Error() != "boom" {
		t.Fatalf("expected controller error boom, got %v", dispatchErr)
	}
	if response != nil {
		t.Fatalf("expected dispatcher to return nil response on error, got %#v", response)
	}
}

// TestGuestControllerRouteDispatcherRejectsMissingRequestType verifies the
// dispatcher responds with a bad request when requestType is absent.
func TestGuestControllerRouteDispatcherRejectsMissingRequestType(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestRouteTestController{})
	if err != nil {
		t.Fatalf("expected dispatcher creation to succeed, got error: %v", err)
	}

	response, err := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		Route: &RouteMatchSnapshotV1{},
	})
	if err != nil {
		t.Fatalf("expected missing request type to return bridge response, got error: %v", err)
	}
	if response == nil || response.StatusCode != 400 {
		t.Fatalf("expected 400 response for missing request type, got %#v", response)
	}
}

// TestGuestControllerRouteDispatcherFallsBackToInternalPath verifies internal
// path dispatch works when requestType is not supplied.
func TestGuestControllerRouteDispatcherFallsBackToInternalPath(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestRouteTestController{})
	if err != nil {
		t.Fatalf("expected dispatcher creation to succeed, got error: %v", err)
	}

	response, err := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		RequestID: "req-2",
		Route: &RouteMatchSnapshotV1{
			InternalPath: "/backend-summary",
		},
	})
	if err != nil {
		t.Fatalf("expected internal path fallback to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != 200 {
		t.Fatalf("expected 200 response for internal path fallback, got %#v", response)
	}
}

// TestGuestControllerRouteDispatcherReturnsNotFound verifies unknown reflected
// handlers return a not-found bridge response.
func TestGuestControllerRouteDispatcherReturnsNotFound(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestRouteTestController{})
	if err != nil {
		t.Fatalf("expected dispatcher creation to succeed, got error: %v", err)
	}

	response, err := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		Route: &RouteMatchSnapshotV1{
			RequestType: "UnknownReq",
		},
	})
	if err != nil {
		t.Fatalf("expected unknown request type to return bridge response, got error: %v", err)
	}
	if response == nil || response.StatusCode != 404 {
		t.Fatalf("expected 404 response for unknown request type, got %#v", response)
	}
}
